#!/usr/bin/env python3
"""Hooks do Claude Code que fazem as skills da casa rodarem de fato.

Instalado pelo scaffold em <DOC>/tools/ e ligado em .claude/settings.json:

  session-start   SessionStart     injeta o estado.md (onde o trabalho parou, ≤ 40
                                   linhas), os nomes das skills e o ciclo — o resto a
                                   IA lê sob demanda pelo INDEX.md
  prompt          UserPromptSubmit lê o pedido e injeta qual skill ele aciona
                                   (ambientes, nomad-worker, sentinel, conhecimento,
                                   OpenSpec) — a IA não depende de "lembrar"
  pre-bash        PreToolUse/Bash  comando com efeito (deploy, migration, apply,
                                   escrita em banco/Vault) em prod ou sem alvo
                                   identificado vira pedido de permissão ao
                                   usuário; `git commit --no-verify` é negado
  stop            Stop             roda checa_projeto.py --modo stop; com FURO
                                   (rastreabilidade, T-n sem teste, openspec
                                   validate, chave sem par, diff não auditado pela
                                   sentinel) a IA não encerra e recebe o motivo; sem
                                   furo, lembra uma vez por sessão de atualizar o
                                   estado.md (skill passagem) quando ele ficou velho
  auditoria       (manual)         a skill sentinel registra o veredito sobre o
                                   conteúdo atual de cada arquivo alterado; o
                                   pre-commit barra commit de arquivo não auditado
                                   (a sentinel roda só no git local, não no CI)

Entrada dos hooks: JSON no stdin (formato do Claude Code). Saída: JSON no stdout
ou exit 2 com o motivo no stderr, conforme o evento. Nunca imprime segredos.
"""
import argparse
import datetime
import hashlib
import json
import os
import re
import subprocess
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
sys.dont_write_bytecode = True
sys.path.insert(0, str(AQUI))

MAX_BLOQUEIOS_STOP = 3
MAX_FUROS_MSG = 5  # a mensagem do Stop entra no contexto: só o necessário para agir
VEREDITOS = ("aprovado", "aprovado com ressalvas", "corrigir antes de concluir")

# ------------------------------------------------------------- utilidades


def ler_stdin():
    try:
        dados = sys.stdin.read()
        return json.loads(dados) if dados.strip() else {}
    except (json.JSONDecodeError, OSError):
        return {}


def raiz_projeto(entrada=None):
    for cand in (os.environ.get("CLAUDE_PROJECT_DIR"), (entrada or {}).get("cwd")):
        if cand and Path(cand).is_dir():
            p = Path(cand).resolve()
            break
    else:
        p = Path.cwd()
    try:
        top = subprocess.run(["git", "rev-parse", "--show-toplevel"], cwd=p,
                             capture_output=True, text=True, timeout=10).stdout.strip()
        if top:
            return Path(top)
    except (OSError, subprocess.TimeoutExpired):
        pass
    return p


def emitir(evento, **campos):
    print(json.dumps({"hookSpecificOutput": {"hookEventName": evento, **campos}}, ensure_ascii=False))


# ---------------------------------------------------------- session-start
MAX_LINHAS_ESTADO = 40


def ler_estado(doc: Path, raiz: Path):
    """Linhas do estado.md (cortado em MAX_LINHAS_ESTADO) ou None se não existe."""
    f = doc / "estado.md"
    if not f.is_file():
        return None
    linhas = f.read_text(encoding="utf-8", errors="replace").rstrip("\n").splitlines()
    if len(linhas) > MAX_LINHAS_ESTADO:
        linhas = linhas[:MAX_LINHAS_ESTADO] + [f"(… cortado em {MAX_LINHAS_ESTADO} linhas — enxugue o "
                                               f"{f.relative_to(raiz).as_posix()} na próxima passagem)"]
    return linhas


def cartao_projeto(raiz: Path) -> str:
    """Fatos do projeto em uma linha, lidos do disco (a IA não precisa descobrir)."""
    osp = raiz / "openspec"
    if not osp.is_dir():
        return ""
    changes = sorted(p.name for p in (osp / "changes").glob("*") if p.is_dir() and p.name != "archive") \
        if (osp / "changes").is_dir() else []
    caps = sorted(p.parent.name for p in (osp / "specs").glob("*/spec.md")) if (osp / "specs").is_dir() else []
    n_req = sum(len(re.findall(r"^### Requirement:", f.read_text(encoding="utf-8", errors="replace"), re.M))
                for f in (osp / "specs").glob("*/spec.md")) if caps else 0
    partes = [f"specs vigentes: {len(caps)} capacidade(s), {n_req} REQ"]
    partes.append("changes ativas: " + (", ".join(changes) if changes else "nenhuma"))
    return "Projeto: " + " · ".join(partes) + "."


def cmd_session_start(_args):
    raiz = raiz_projeto(ler_stdin())
    from checa_projeto import docs_do_repo
    docs = docs_do_repo(raiz)
    nomes = sorted(s.parent.name for s in (raiz / ".agents" / "skills").glob("*/SKILL.md")
                   if not s.parent.name.startswith("openspec-"))
    linhas = ["Projeto sob SDD com OpenSpec. Siga o AGENTS.md (gatilho de cada skill na tabela dele)."]
    if nomes:
        linhas.append("Skills do projeto em .agents/skills/ (leia o SKILL.md quando o pedido cair no gatilho): "
                      + ", ".join(nomes) + ".")
    cartao = cartao_projeto(raiz)
    if cartao:
        linhas.append(cartao)
    for d in docs:
        rel = d.relative_to(raiz).as_posix()
        estado = ler_estado(d, raiz)
        if estado is None:
            linhas.append(f"Sem {rel}/estado.md: ao encerrar a sessão, crie-o com a skill passagem.")
        else:
            linhas += ["", f"=== {rel}/estado.md (onde o trabalho parou; confirme com o usuário se o próximo "
                           "passo ainda vale) ===", *estado, "=== fim do estado ==="]
    linhas += [
        "",
        "Leitura sob demanda: para o resto, use o INDEX.md da DOC e leia só o que a tarefa pedir — nunca o "
        "log.md inteiro (só a entrada do topo ou uma busca), no mapa só a página da área, nas specs só a "
        "capacidade envolvida.",
        "Ciclo de mudança de comportamento: sabatina → /opsx:propose (RN de origem) → /opsx:apply (sentinel "
        "audita cada trecho) → openspec validate + valida_rastreabilidade → deploy pela skill ambientes → "
        "/opsx:archive → passagem ao encerrar.",
        "Garantias: o hook Stop recusa encerrar com FURO; comandos com efeito em prod pedem permissão; o "
        "pre-commit e o CI repetem as checagens.",
    ]
    emitir("SessionStart", additionalContext="\n".join(linhas))
    return 0


# ------------------------------------------------------------------ prompt
ROTAS = [
    ("ambientes", r"\b(deploy|implant|sob(e|ir)|subir|publicar|promov|release|rollback|migra(tion|ção|cao|r)|hml|homolog|"
                  r"prod(u[cç][aã]o)?|staging|ambiente|vari[aá]ve(l|is) de ambiente|segredo|secret|\.env)\b"),
    ("nomad-worker", r"\b(nomad|jobspec|\.hcl|consul|vault|worker|cron|fila|queue)\b"),
    ("sentinel", r"\b(implement|codific|c[oó]digo|refator|corrig|fix|bug|review|revis|pr\b|pull request|"
                 r"teste|test|função|funcao|classe|endpoint|feature)\w*"),
    ("openspec", r"\b(nova funcionalidade|feature|requisito|mudan[cç]a de comportamento|regra de neg[oó]cio|"
                 r"\brn-\d+|change|proposta|escopo)\w*"),
    ("conhecimento", r"\b(onde fica|como funciona|por que|porque|o que (é|faz)|arquitetura|documenta|explica)\w*"),
    ("sabatina", r"\b(sabatin|grill|me questione|me pergunte|vamos alinhar|alinhar antes|pensar junto|"
                 r"nova funcionalidade|mudan[cç]a grande|redesenh|arquitetur|proposta|planejar|planej)\w*"),
]
TEXTO_ROTA = {  # uma frase cada: o SKILL.md tem o resto
    "ambientes": "ambientes: bloco de alvo antes de executar; prod só com confirmação explícita.",
    "nomad-worker": "nomad-worker: copie o padrão de infra/*.hcl; segredo só via Vault; tag imutável.",
    "sentinel": "sentinel: audite o trecho e registre o veredito (sdd_hooks.py auditoria).",
    "openspec": "OpenSpec: mudança de comportamento vira change (/opsx:propose) com RN, REQ-n e T-n.",
    "conhecimento": "conhecimento: responda pela DOC/ (estado.md, INDEX.md) e openspec/, citando o arquivo.",
    "sabatina": "sabatina: rodadas com recomendação antes de propor; só aja depois do \"fechado\".",
}


def rotear(texto: str):
    t = texto.lower()
    return [nome for nome, rx in ROTAS if re.search(rx, t)]


def cmd_prompt(_args):
    entrada = ler_stdin()
    rotas = rotear(entrada.get("prompt", ""))
    if not rotas:
        return 0
    emitir("UserPromptSubmit", additionalContext="Skills deste pedido (.agents/skills/): " + " | ".join(TEXTO_ROTA[r] for r in rotas))
    return 0


# ---------------------------------------------------------------- pre-bash
RE_EFEITO = [re.compile(p, re.I) for p in (
    r"\bnomad\s+(job\s+)?(run|stop|dispatch|revert|scale|plan\s+-apply)\b",
    r"\bterraform\s+(apply|destroy|import|state\s+rm)\b",
    r"\bterragrunt\s+(apply|destroy)\b",
    r"\bkubectl\s+(apply|delete|rollout|scale|patch|replace|create|set|edit)\b",
    r"\bhelm\s+(install|upgrade|uninstall|rollback)\b",
    r"\b(flyway|liquibase)\b.*\b(migrate|update|clean|repair|rollback)\b",
    r"\bprisma\s+(migrate\s+(deploy|reset)|db\s+push)\b",
    r"\balembic\s+(upgrade|downgrade|stamp)\b",
    r"\b(manage\.py|rails|rake|artisan|knex|sequelize(-cli)?|typeorm|goose|migrate)\b.*\b(migrat|up|down|reset|rollback)",
    r"\bvault\s+(kv\s+)?(put|write|delete|patch|destroy)\b",
    r"\bconsul\s+kv\s+(put|delete)\b",
    r"\bdocker\s+(push|stack\s+deploy|service\s+update)\b",
    r"\b(psql|mysql|mariadb|mongosh?|redis-cli|sqlcmd)\b",
    r"\bgh\s+(workflow\s+run|release\s+create)\b",
    r"\b(npm|pnpm|yarn|bun|make|just|task)\s+(run\s+)?(deploy|release|migrate|db:migrate|publish)\b",
    r"\b(aws|gcloud|az)\s+\S+.*\b(deploy|delete|put|create|update|apply)\b",
    r"\bansible-playbook\b",
    r"\bscp\b|\brsync\b.*@",
    r"\bcurl\b.*(?:(?:-X|--request)\s*(?:POST|PUT|PATCH|DELETE)\b|\s(?:-d|--data\S*)\b)",
    r"\b(?:npx\s+)?supabase\s+(?:db\s+push|functions\s+deploy|secrets\s+set|migration\s+up\b.*--linked)\b",
    r"\b(?:npx\s+)?wrangler\s+(?:deploy|publish|secret\s+(?:put|bulk|delete)|d1\s+(?:execute|migrations\s+apply)"
    r"|kv[: ]key\s+put|versions\s+(?:upload|deploy))\b",
    r"\bssh\s+\S+@",
)]  # lista aberta: heurística de última linha — a barreira primária é a skill ambientes
RE_PROD = re.compile(r"(?<![a-z0-9])(prod|production|prd|produ[cç][aã]o|live)(?![a-z0-9])", re.I)
RE_HML = re.compile(r"(?<![a-z0-9/])(hml|homolog\w*|staging|stg|dev|local|localhost|127\.0\.0\.1|sandbox)(?![a-z0-9])", re.I)
# commit: --no-verify ou -n, inclusive colado a outras flags curtas (-anm, -qn);
# push: só --no-verify (-n em push é dry-run); -c core.hooksPath=… desliga o hook do mesmo jeito
RE_NO_VERIFY = re.compile(r"\bgit\b.*\bcommit\b.*(?:\s--no-verify\b|\s-[A-Za-z]*n[A-Za-z]*\b)"
                          r"|\bgit\s+push\b.*\s--no-verify\b"
                          r"|\bgit\b.*\s-c\s*core\.hooksPath\b", re.I)
RE_FORCE = re.compile(r"\bgit\s+push\b.*(?:\s--force(?:-with-lease)?\b|\s-f\b|\s\+\S+)")


def classificar_bash(cmd: str):
    """Retorna (decisão, motivo) ou (None, None) para deixar passar."""
    if RE_NO_VERIFY.search(cmd):
        return "deny", ("Pular o pre-commit (--no-verify / -n / -c core.hooksPath) é proibido neste projeto: ele "
                        "roda as checagens SDD. Corrija o furo que o hook apontou em vez de contorná-lo.")
    if RE_FORCE.search(cmd):
        return "ask", "git push --force reescreve histórico remoto. Confirme com o usuário antes."
    if not any(rx.search(cmd) for rx in RE_EFEITO):
        return None, None
    base = ("Comando com efeito fora da máquina local. Pela skill ambientes: mostre ao usuário o bloco de "
            "alvo (Ambiente / Provado por / Ação / Reversível) e só execute com a confirmação dele.")
    if RE_PROD.search(cmd):
        return "ask", "ALVO: PRODUÇÃO. " + base
    if RE_HML.search(cmd):
        return None, None
    return "ask", ("ALVO NÃO IDENTIFICADO no comando — ambiente não se presume; sem evidência, trate como "
                   "produção. " + base)


def cmd_pre_bash(_args):
    entrada = ler_stdin()
    cmd = ((entrada.get("tool_input") or {}).get("command") or "")
    decisao, motivo = classificar_bash(cmd)
    if decisao:
        emitir("PreToolUse", permissionDecision=decisao, permissionDecisionReason=motivo)
    return 0


# -------------------------------------------------------------------- stop
def _stop_inativo(motivo: str):
    """Fail-open avisado: não tranca o usuário, mas diz que a garantia não rodou."""
    print(json.dumps({"systemMessage": f"SDD: checagem Stop não rodou ({motivo}) — garantia inativa nesta "
                                       "resposta; rode checa_projeto.py à mão."}, ensure_ascii=False))
    return 0


def cmd_stop(_args):
    entrada = ler_stdin()
    raiz = raiz_projeto(entrada)
    try:
        status = subprocess.run(["git", "status", "--porcelain"], cwd=raiz, capture_output=True,
                                text=True, timeout=30)
    except (OSError, subprocess.TimeoutExpired) as e:
        return _stop_inativo(f"git status: {type(e).__name__}")
    if status.returncode != 0 or not status.stdout.strip():
        return 0  # sem repo git ou nada mudou nesta sessão: conversa não é barrada

    checa = AQUI / "checa_projeto.py"
    if not checa.is_file():
        return _stop_inativo("checa_projeto.py não encontrado ao lado do hook")
    try:
        p = subprocess.run([sys.executable, str(checa), str(raiz), "--modo", "stop", "--json"],
                           capture_output=True, text=True, timeout=240)
        if p.returncode == 2 or not p.stdout.strip():  # crash do checa: rc 2 e stdout vazio
            return _stop_inativo("checa_projeto.py falhou: " + (p.stderr.strip().splitlines() or ["sem saída"])[-1][:160])
        itens = json.loads(p.stdout)
    except (OSError, subprocess.TimeoutExpired, json.JSONDecodeError) as e:
        return _stop_inativo(f"checa_projeto.py: {type(e).__name__}")
    furos = [i for i in itens if i.get("nivel") == "FURO"]
    estado_arq = raiz / ".sdd" / "stop.json"
    if not furos:
        estado_arq.unlink(missing_ok=True)  # corrigiu: a trava anti-laço recomeça do zero
        return lembrar_passagem(raiz, entrada, status.stdout)

    # trava anti-laço: no máximo N bloqueios seguidos, na mesma sessão, para o mesmo conjunto de furos
    sessao = str(entrada.get("session_id") or "-")
    chave = hashlib.sha1(json.dumps(furos, sort_keys=True).encode()).hexdigest()[:16]
    try:
        estado = json.loads(estado_arq.read_text(encoding="utf-8"))
        if not isinstance(estado, dict) or estado.get("sessao") != sessao:
            estado = {}
    except (OSError, json.JSONDecodeError):
        estado = {}
    n = (estado.get("contagens") or {}).get(chave, 0) + 1
    if n > MAX_BLOQUEIOS_STOP:
        print(json.dumps({"systemMessage": f"SDD: {len(furos)} furo(s) continuam abertos após "
                                           f"{MAX_BLOQUEIOS_STOP} tentativas — encerrando; o pre-commit e o CI "
                                           "ainda vão barrar. Rode checa_projeto.py para ver."}, ensure_ascii=False))
        return 0
    estado_arq.parent.mkdir(parents=True, exist_ok=True)
    estado_arq.write_text(json.dumps({"sessao": sessao, "contagens": {chave: n}}), encoding="utf-8")

    msg = ["Não encerre ainda: as checagens SDD encontraram furo(s) no trabalho desta sessão.", ""]
    for f in furos[:MAX_FUROS_MSG]:
        texto = f["msg"] if len(f["msg"]) <= 220 else f["msg"][:217] + "…"
        msg.append(f"- [{f['area']}] {texto}")
    if len(furos) > MAX_FUROS_MSG:
        msg.append(f"- … e mais {len(furos) - MAX_FUROS_MSG}: python3 <DOC>/tools/checa_projeto.py . --modo stop")
    msg += ["", "Corrija (ou, se não for possível agora, registre no to-do.md da DOC/ e explique ao usuário "
                "por quê) e tente concluir de novo."]
    print("\n".join(msg), file=sys.stderr)
    return 2


def estados_velhos(raiz: Path, porcelain: str):
    """estado.md ausente ou mais antigo que o arquivo mais recente mexido na sessão."""
    from checa_projeto import docs_do_repo
    mexidos = []
    for ln in porcelain.splitlines():
        p = ln[3:].split(" -> ")[-1].strip().strip('"')
        if not p or p.startswith(".sdd/") or p.endswith("estado.md") or "/log/" in p:
            continue
        f = raiz / p
        if f.is_file():
            mexidos.append(f.stat().st_mtime)
    if not mexidos:
        return []
    velhos = []
    for d in docs_do_repo(raiz):
        e = d / "estado.md"
        if not e.is_file() or e.stat().st_mtime < max(mexidos):
            velhos.append(e.relative_to(raiz).as_posix())
    return velhos


def lembrar_passagem(raiz: Path, entrada: dict, porcelain: str):
    """Cobrança leve: uma vez por sessão, pede a passagem."""
    velhos = estados_velhos(raiz, porcelain)
    if not velhos:
        return 0
    sessao = str(entrada.get("session_id") or "-")
    marca = raiz / ".sdd" / "passagem.json"
    try:
        feito = json.loads(marca.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        feito = {}
    if feito.get("sessao") == sessao:  # já lembrou nesta sessão: não insiste
        return 0
    marca.parent.mkdir(parents=True, exist_ok=True)
    marca.write_text(json.dumps({"sessao": sessao}), encoding="utf-8")
    print("Antes de encerrar: " + ", ".join(velhos) + " ficou para trás do trabalho desta sessão. Atualize-o "
          "com a skill passagem (retrato de onde parou, ≤ 40 linhas) — ou, se a sessão não mudou o rumo, "
          "diga isso ao usuário em uma linha. Este lembrete não se repete nesta sessão.", file=sys.stderr)
    return 2


# --------------------------------------------------------------- auditoria
def cmd_auditoria(args):
    from checa_projeto import arquivos_alterados  # mesmo diretório
    raiz = Path(args.raiz).resolve() if args.raiz else raiz_projeto()
    veredito = args.veredito.strip().lower()
    if veredito not in VEREDITOS:
        print(f"ERRO: veredito deve ser um de: {', '.join(VEREDITOS)}")
        return 2
    arqs = arquivos_alterados(raiz)
    if arqs is None:
        print("ERRO: não é repositório git — nada a registrar.")
        return 2
    if not arqs:
        print("Sem mudança de código não commitada — nada a registrar.")
        return 0
    log = raiz / ".sdd" / "auditoria.jsonl"
    log.parent.mkdir(parents=True, exist_ok=True)
    reg = {"data": datetime.datetime.now().isoformat(timespec="seconds"), "veredito": veredito,
           "arquivos": arqs}
    with log.open("a", encoding="utf-8") as f:
        f.write(json.dumps(reg, ensure_ascii=False) + "\n")
    print(f"Registrado: {len(arqs)} arquivo(s) — {veredito}.")
    if veredito.startswith("corrigir"):
        print("O commit e o encerramento ficam barrados até os arquivos mudarem e serem auditados de novo.")
    return 0


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)
    for nome in ("session-start", "prompt", "pre-bash", "stop"):
        sub.add_parser(nome)
    a = sub.add_parser("auditoria")
    a.add_argument("--veredito", required=True)
    a.add_argument("--raiz", default=None)
    args = ap.parse_args()
    return {"session-start": cmd_session_start, "prompt": cmd_prompt, "pre-bash": cmd_pre_bash,
            "stop": cmd_stop, "auditoria": cmd_auditoria}[args.cmd](args)


if __name__ == "__main__":
    sys.exit(main())
