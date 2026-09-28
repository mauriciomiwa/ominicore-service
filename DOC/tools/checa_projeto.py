#!/usr/bin/env python3
"""Audita um repositório inteiro contra o padrão SDD da casa e diz o que falta
para todas as skills rodarem.

Uso:
  python3 checa_projeto.py <raiz-do-repo> [--modo local|ci|precommit|stop] [--fonte DIR]

  --modo local      (default) relatório completo para um humano ou agente
         ci         o que o CI deve barrar: exige o CLI do OpenSpec; não olha
                    hooks locais nem o registro da sentinel
         precommit  rápido, sem rede: o que barra o commit
         stop       usado pelo hook Stop do Claude Code: só o que impede dar
                    o trabalho por concluído (inclui a auditoria da sentinel)
  --fonte           pasta do repositório de skills (para medir drift das
                    cópias em .agents/skills/). Default: descoberta automática
                    quando o script roda de dentro da skill.

O que confere, por skill:
  estrutura       DOC/INDEX.md, AGENTS.md, CLAUDE.md, .agents/skills/ com as
                  skills da casa (e nomad-worker quando há job Nomad),
                  openspec/config.yaml com o contexto da casa, comandos /opsx,
                  garantias instaladas (pre-commit, CI, hooks do Claude)
  novo-projeto-sdd  valida_rastreabilidade.py por projeto, openspec validate
                  --strict e --archived, todo T-n vigente citado em algum teste
  dependências    .github/dependabot.yml cobrindo os manifestos
  ambientes       DOC/Infra/AMBIENTES.md quando há artefato de deploy, pares
                  de config hml/prod sem chave órfã, campos A DEFINIR
  mapa            System/mapa/ em dia com o código (mapeia_projeto.py --check):
                  rota, endpoint, chamada, fila ou variável nova fora do mapa
  nomad-worker    jobs Nomad sem imagem :latest/sem tag e sem segredo literal
  sentinel        (só no git local, nunca no CI) cada arquivo de código do commit
                  (precommit) ou do trabalho em aberto (stop/local) foi auditado no
                  conteúdo atual — registro em .sdd/auditoria.jsonl, gravado por
                  `sdd_hooks.py auditoria`

Saída em texto simples. Exit 0 = sem FURO; 1 = há FURO; 2 = uso incorreto.
"""
import argparse
import fnmatch
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

AQUI = Path(__file__).resolve().parent
SKILLS_CASA = ("conhecimento", "sentinel", "ambientes", "sabatina", "passagem")
IGNORAR_DIRS = {".git", "node_modules", ".venv", "venv", "__pycache__", "dist", "build",
                "target", ".next", ".idea", ".vscode", "vendor", ".sdd", ".gradle"}
RE_TESTE_ARQ = re.compile(
    r"(^|/)(tests?|__tests__|spec|specs|e2e|it)/|(^|/)test_[^/]+$|_test\.[a-z]+$|"
    r"\.(test|spec)\.[a-z]+$|Tests?\.(java|kt|cs|scala)$|_spec\.rb$", re.I)
# T-n no nome do teste: test_T12_…, it("T-12 …"), e também o estilo Go/Java
# (TestT12Algo, TestT16T19…): basta não vir colado a outra maiúscula (HTTP2, UTF8).
RE_T_CODIGO = re.compile(r"(?<![A-Z0-9])T[-_]?(\d+(?:\.\d+)?)(?![0-9])")  # T-12 ou T-28.15; não o T de 2026-09-19T00:00
def chave_tn(s):
    return tuple(int(x) for x in s.split("."))


RE_T_SPEC = re.compile(r"^####\s+Scenario:\s*.*?\bT-(\d+(?:\.\d+)?)\b", re.M)
EXT_CODIGO = {".py", ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".java", ".kt", ".go", ".rb",
              ".cs", ".php", ".rs", ".swift", ".scala", ".dart", ".ex", ".exs", ".feature", ".robot"}


class Relatorio:
    def __init__(self):
        self.itens = []  # (nivel, area, msg)

    def add(self, nivel, area, msg):
        self.itens.append((nivel, area, msg))

    def furo(self, area, msg):
        self.add("FURO", area, msg)

    def alerta(self, area, msg):
        self.add("ALERTA", area, msg)

    def ok(self, area, msg):
        self.add("OK", area, msg)

    def pulado(self, area, msg):
        self.add("PULADO", area, msg)

    @property
    def furos(self):
        return [i for i in self.itens if i[0] == "FURO"]


# ------------------------------------------------------------------ util
def run(cmd, cwd=None, timeout=300):
    env = {**os.environ, "OPENSPEC_TELEMETRY": "0", "OPENSPEC_NO_UPDATE_CHECK": "1",
           "PYTHONDONTWRITEBYTECODE": "1", "CI": os.environ.get("CI", "1")}
    try:
        p = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True, timeout=timeout, env=env)
        return p.returncode, (p.stdout + p.stderr).strip()
    except (OSError, subprocess.TimeoutExpired) as e:
        return 99, str(e)


def arquivos(raiz: Path):
    for dirpath, dirnames, filenames in os.walk(raiz):
        dirnames[:] = [d for d in dirnames if d not in IGNORAR_DIRS]
        for f in filenames:
            yield Path(dirpath) / f


def achar_script(nome: str):
    for cand in (AQUI / nome, AQUI.parents[1] / "ambientes" / "scripts" / nome,
                 AQUI.parents[1] / "novo-projeto-sdd" / "scripts" / nome):
        if cand.is_file():
            return cand
    return None


def docs_do_repo(raiz: Path):
    """Pastas DOC do repositório: DOC/ (app único) ou DOC/<projeto>/ (monorepo)."""
    base = raiz / "DOC"
    if (base / "INDEX.md").is_file():
        return [base]
    return sorted(p.parent for p in base.glob("*/INDEX.md")) if base.is_dir() else []


EXCLUI_AUDITORIA = ("DOC/", "openspec/", ".sdd/", ".agents/", ".claude/", ".githooks/")


# o que a sentinel audita: código-fonte e código de infra. Config gerada (yml/json),
# texto e documentação ficam de fora — cobrá-los só gera atrito sem ganho.
EXT_AUDITAVEL = EXT_CODIGO - {".feature", ".robot"} | {
    ".vue", ".svelte", ".astro", ".sql", ".sh", ".bash", ".ps1", ".hcl", ".tf", ".nomad",
    ".c", ".h", ".cpp", ".hpp", ".m", ".lua", ".groovy", ".gradle", ".kts", ".erl", ".clj", ".fs"}
NOMES_AUDITAVEIS = {"Dockerfile", "Makefile", "Jenkinsfile", "Containerfile"}


def e_codigo_auditavel(p: str) -> bool:
    if not p or p.startswith(EXCLUI_AUDITORIA):
        return False
    nome = p.rsplit("/", 1)[-1]
    return Path(nome).suffix.lower() in EXT_AUDITAVEL or nome in NOMES_AUDITAVEIS \
        or nome.startswith("Dockerfile.")


def padroes_sentinel(raiz: Path):
    """Padrões do .sentinelignore da raiz: código gerado ou de terceiros que não
    passa pela sentinel (ex.: modelos gerados pelo framework). Uma linha por
    padrão; "pasta/" ignora a pasta inteira; # comenta."""
    f = raiz / ".sentinelignore"
    if not f.is_file():
        return []
    return [l.strip() for l in f.read_text(encoding="utf-8", errors="replace").splitlines()
            if l.strip() and not l.strip().startswith("#")]


def padrao_amplo_sentinel(pad: str) -> bool:
    """`*`, `**/*`, `*.py`… desligam a sentinel para o repositório inteiro numa linha:
    padrão sem pasta nem nome específico é furo, não configuração."""
    pad = pad.strip().lstrip("/")
    return bool(re.fullmatch(r"[*/?]*(\.\w+)?", pad))


def ignorado_sentinel(p: str, padroes) -> bool:
    nome = p.rsplit("/", 1)[-1]
    for pad in padroes:
        pad = pad.lstrip("/")
        if pad.endswith("/"):
            if p.startswith(pad):
                return True
        elif fnmatch.fnmatch(p, pad) or ("/" not in pad and fnmatch.fnmatch(nome, pad)):
            return True
    return False


def arquivos_alterados(raiz: Path, escopo: str = "trabalho"):
    """{caminho: blob-sha} dos arquivos de código alterados.

    escopo "trabalho": tudo o que não foi commitado (índice + árvore + não rastreados),
    com o conteúdo da árvore. escopo "staged": só o que vai no commit, com o conteúdo
    do índice. Arquivo removido = "REMOVIDO". None quando não é repositório git."""
    rc, _ = run(["git", "rev-parse", "--is-inside-work-tree"], cwd=raiz, timeout=20)
    if rc != 0:
        return None
    out = {}
    ignorar = padroes_sentinel(raiz)
    if escopo == "staged":
        rc, nomes = run(["git", "diff", "--cached", "--name-only", "-z", "--diff-filter=ACMRTD"],
                        cwd=raiz, timeout=60)
        for p in [n for n in nomes.split("\0") if n]:
            if not e_codigo_auditavel(p) or ignorado_sentinel(p, ignorar):
                continue
            rc, ls = run(["git", "ls-files", "-s", "--", p], cwd=raiz, timeout=20)
            out[p] = ls.split()[1] if rc == 0 and ls.strip() else "REMOVIDO"
        return out
    # sem run(): ele faz strip() e comeria o espaço inicial do "XY" da 1ª linha
    try:
        status = subprocess.run(["git", "status", "--porcelain", "--untracked-files=all"], cwd=raiz,
                                capture_output=True, text=True, timeout=60).stdout
    except (OSError, subprocess.TimeoutExpired):
        return None
    for ln in status.splitlines():
        p = ln[3:].split(" -> ")[-1].strip().strip('"')
        if not e_codigo_auditavel(p) or ignorado_sentinel(p, ignorar):
            continue
        f = raiz / p
        if f.is_file():
            rc, sha = run(["git", "hash-object", "--", p], cwd=raiz, timeout=20)
            out[p] = sha.strip() if rc == 0 else "?"
        else:
            out[p] = "REMOVIDO"
    return out


def hash_diff(raiz: Path):
    """Resumo curto do trabalho de código não commitado ("" = nada; None = sem git)."""
    arqs = arquivos_alterados(raiz)
    if arqs is None:
        return None
    if not arqs:
        return ""
    return hashlib.sha1(json.dumps(arqs, sort_keys=True).encode()).hexdigest()[:16]


def ler_auditorias(raiz: Path):
    """Registros da sentinel, do mais novo para o mais antigo."""
    log = raiz / ".sdd" / "auditoria.jsonl"
    regs = []
    if log.is_file():
        for ln in log.read_text(encoding="utf-8").splitlines():
            try:
                regs.append(json.loads(ln))
            except json.JSONDecodeError:
                continue
    return list(reversed(regs))


def situacao_auditoria(raiz: Path, arqs: dict):
    """(nao_auditados, mandou_corrigir) para o conjunto de arquivos informado."""
    regs = ler_auditorias(raiz)
    nao, corrigir = [], []
    for p, sha in sorted(arqs.items()):
        reg = next((r for r in regs if r.get("arquivos", {}).get(p) == sha), None)
        if reg is None:
            nao.append(p)
        elif str(reg.get("veredito", "")).startswith("corrigir"):
            corrigir.append(p)
    return nao, corrigir


# ------------------------------------------------------------- estrutura
def termos_glossario(doc: Path):
    """Quantos termos o Negocio/glossario.md define (None = arquivo ausente).
    Conta "**Termo**:" fora de comentário HTML e linhas de tabela preenchidas
    (formato antigo | Termo | Significado |)."""
    g = doc / "Negocio" / "glossario.md"
    if not g.is_file():
        return None
    txt = re.sub(r"<!--.*?-->", "", g.read_text(encoding="utf-8", errors="replace"), flags=re.S)
    novos = re.findall(r"^\*\*[^*\n]+\*\*:", txt, re.M)  # **Termo**: com ou sem definição na mesma linha
    tabela = [l for l in re.findall(r"^\|([^|\n]+)\|", txt, re.M)
              if l.strip() and not set(l.strip()) <= set("-: ") and l.strip().lower() not in ("termo",)
              and not l.strip().startswith("<")]
    return len(novos) + len(tabela)


# O que a IA lê no começo de toda sessão precisa ficar curto (ver skill passagem).
LIMITE_AGENTS = 4096        # bytes
LIMITE_ESTADO = 40          # linhas
LIMITE_LOG = 15 * 1024      # bytes


def tamanhos_contexto(raiz: Path, doc: Path):
    """Problemas de tamanho do que é lido no início: lista de (chave, mensagem)."""
    out = []
    ag = raiz / "AGENTS.md"
    if ag.is_file() and ag.stat().st_size > LIMITE_AGENTS:
        out.append(("agents", f"AGENTS.md com {ag.stat().st_size // 1024} KB (> {LIMITE_AGENTS // 1024} KB) — ele entra "
                              "em toda sessão; mova detalhe para a DOC e deixe só regras e ponteiros."))
    rel = doc.relative_to(raiz).as_posix()
    est = doc / "estado.md"
    if not est.is_file():
        out.append(("estado", f"{rel}/estado.md não existe — o chat novo não sabe onde o trabalho parou "
                              "(crie com a skill passagem)."))
    else:
        n = len(est.read_text(encoding="utf-8", errors="replace").splitlines())
        if n > LIMITE_ESTADO:
            out.append(("estado", f"{rel}/estado.md com {n} linhas (> {LIMITE_ESTADO}) — é retrato, não "
                                  "histórico; reescreva com a skill passagem."))
    lg = doc / "log.md"
    if lg.is_file() and lg.stat().st_size > LIMITE_LOG:
        out.append(("log", f"{rel}/log.md com {lg.stat().st_size // 1024} KB (> {LIMITE_LOG // 1024} KB) — arquive "
                           f"as entradas mais antigas: python3 {rel}/tools/arquiva_log.py {rel}"))
    return out


def checa_estrutura(raiz, docs, rel, modo, fonte):
    area = "estrutura"
    if not docs:
        rel.furo(area, "sem DOC/INDEX.md nem DOC/<projeto>/INDEX.md — rode o kickoff "
                       "(novo-projeto-sdd) ou o scaffold.py para criar a base.")
    else:
        rel.ok(area, "DOC: " + ", ".join(d.relative_to(raiz).as_posix() for d in docs))

    if modo == "local":  # glossário é julgamento, não mecânica: só alerta, e só na checagem local
        for d in docs:
            for _, msg in tamanhos_contexto(raiz, d):
                rel.alerta(area, msg)
            n = termos_glossario(d)
            nome = d.relative_to(raiz).as_posix()
            if n is None:
                rel.alerta(area, f"{nome}/Negocio/glossario.md não existe — a linguagem do domínio fica na cabeça "
                                 "de cada um; crie-o e alimente-o durante a skill sabatina.")
            elif n == 0:
                rel.alerta(area, f"{nome}/Negocio/glossario.md sem nenhum termo — defina os termos do domínio "
                                 "(skill sabatina ou peça ao Claude: \"monte o glossário a partir das RN e do código\").")

    agents = raiz / "AGENTS.md"
    if not agents.is_file():
        rel.furo(area, "sem AGENTS.md na raiz — outras IAs (Codex, Cursor, Devin) não sabem que a DOC/ existe.")
        texto_agents = ""
    else:
        texto_agents = agents.read_text(encoding="utf-8", errors="replace")
    claude = raiz / "CLAUDE.md"
    if not claude.is_file():
        rel.furo(area, "sem CLAUDE.md na raiz apontando para o AGENTS.md.")
    elif "AGENTS.md" not in claude.read_text(encoding="utf-8", errors="replace"):
        rel.alerta(area, "CLAUDE.md não menciona o AGENTS.md — duas fontes de instrução divergem.")

    skills_dir = raiz / ".agents" / "skills"
    esperadas = list(SKILLS_CASA)
    if jobs_nomad(raiz):
        esperadas.append("nomad-worker")
    for s in esperadas:
        if not (skills_dir / s / "SKILL.md").is_file():
            rel.furo(area, f"skill '{s}' ausente em .agents/skills/ — rode o sync-skills.py do repositório "
                           f"de skills (ou o scaffold) para instalá-la.")
        elif texto_agents and s not in texto_agents:
            rel.alerta(area, f"AGENTS.md não cita a skill '{s}' — agentes sem hook podem não saber quando usá-la; "
                             "acrescente a tabela 'Skills do projeto' do template _raiz/AGENTS.md.")

    # drift das cópias
    if fonte and (fonte / "conhecimento" / "SKILL.md").is_file():
        drift = []
        for s in esperadas:
            src, dst = fonte / s, skills_dir / s
            if not dst.is_dir() or not src.is_dir():
                continue
            for f in src.rglob("*"):
                if f.is_file() and f.name != ".DS_Store" and "__pycache__" not in f.parts:
                    alvo = dst / f.relative_to(src)
                    if not alvo.is_file() or alvo.read_bytes() != f.read_bytes():
                        drift.append(f"{s}/{f.relative_to(src)}")
        if drift:
            rel.alerta(area, f"{len(drift)} arquivo(s) de skill desatualizado(s) em .agents/skills/ "
                             f"(ex.: {drift[0]}) — rode sync-skills.py.")
        else:
            rel.ok(area, "cópias em .agents/skills/ em dia com a fonte.")
    else:
        rel.pulado(area, "drift das skills não medido (sem --fonte).")

    # openspec
    cfg = raiz / "openspec" / "config.yaml"
    if not cfg.is_file():
        rel.furo(area, "sem openspec/config.yaml — rode `openspec init --tools claude,agents` e o scaffold.")
    elif "sdd-rastreabilidade" not in cfg.read_text(encoding="utf-8", errors="replace"):
        rel.furo(area, "openspec/config.yaml sem o contexto da casa (marcador sdd-rastreabilidade) — "
                       "re-rode o scaffold para mesclar.")
    else:
        rel.ok(area, "openspec/config.yaml com o contexto da casa.")
    opsx = [p for base in (".claude/skills", ".agents/skills", ".cursor/skills", ".claude/commands")
            if (raiz / base).is_dir() for p in (raiz / base).glob("*opsx*")] + \
           [p for base in (".claude/skills", ".agents/skills") if (raiz / base).is_dir()
            for p in (raiz / base).glob("openspec-*")]
    if not opsx and modo != "ci":
        rel.alerta(area, "comandos /opsx não instalados — rode `openspec init --tools claude,agents`.")

    # garantias
    if modo in ("local", "stop"):
        settings = raiz / ".claude" / "settings.json"
        try:
            hooks_ok = "sdd_hooks.py" in settings.read_text(encoding="utf-8")
        except OSError:
            hooks_ok = False
        if not hooks_ok:
            rel.furo(area, "hooks do Claude Code não instalados (.claude/settings.json sem sdd_hooks.py) — "
                           "sem eles nada impede a IA de encerrar com furo ou rodar comando em prod. "
                           "Re-rode o scaffold.")
        rc, hp = run(["git", "config", "--get", "core.hooksPath"], cwd=raiz, timeout=20)
        if not (raiz / ".githooks" / "pre-commit").is_file():
            rel.furo(area, "sem .githooks/pre-commit — re-rode o scaffold.")
        elif hp.strip() != ".githooks":
            rel.furo(area, "pre-commit existe mas o git não o usa — rode `git config core.hooksPath .githooks`.")
    wf = raiz / ".github" / "workflows" / "sdd-checks.yml"
    if not wf.is_file():
        rel.furo(area, "sem .github/workflows/sdd-checks.yml — o CI não barra furo no merge. Re-rode o scaffold "
                       "e marque o check 'sdd-checks' como obrigatório na proteção da branch.")


# ------------------------------------------------------ rastreabilidade
def specs_t(raiz: Path):
    """T-n vigentes (openspec/specs) e em changes ativas."""
    vig, ativos = set(), set()
    base = raiz / "openspec"
    for f in (base / "specs").rglob("spec.md") if (base / "specs").is_dir() else []:
        vig |= set(RE_T_SPEC.findall(f.read_text(encoding="utf-8", errors="replace")))
    ch = base / "changes"
    if ch.is_dir():
        for d in ch.iterdir():
            if d.is_dir() and d.name != "archive":
                for f in d.rglob("spec.md"):
                    ativos |= set(RE_T_SPEC.findall(f.read_text(encoding="utf-8", errors="replace")))
    return vig, ativos - vig


# Roteiro de teste manual (tests/manual/*.md ou *.manual.md): cada T-n manual
# vira um passo com resultado esperado, versionado junto do código.
RE_ROTEIRO_MANUAL = re.compile(r"(^|/)(tests?|testes)/manual/[^/]+\.md$|\.manual\.md$", re.I)


def testes_citados(raiz: Path):
    """T-n citados em testes: arquivos de teste de código, verificações do
    pipeline (workflows de CI e Makefile), onde vivem testes como "SDK gerado
    bate com o contrato", e roteiros de teste manual (tests/manual/*.md)."""
    citados = {}
    for f in arquivos(raiz):
        rel = f.relative_to(raiz).as_posix()
        pipeline = rel.startswith(".github/workflows/") or f.name in ("Makefile", "Taskfile.yml", "justfile") \
            or rel.endswith((".gitlab-ci.yml", "Jenkinsfile")) \
            or RE_ROTEIRO_MANUAL.search(rel) is not None  # teste manual: roteiro versionado
        if not pipeline and (f.suffix.lower() not in EXT_CODIGO | {".sql"} or not RE_TESTE_ARQ.search(rel)):  # .sql: pgTAP
            continue
        try:
            txt = f.read_text(encoding="utf-8", errors="replace")
        except OSError:
            continue
        for n in RE_T_CODIGO.findall(txt):
            citados.setdefault(n, rel)
    return citados


def checa_rastreabilidade(raiz, docs, rel, modo):
    area = "novo-projeto-sdd"
    valida = achar_script("valida_rastreabilidade.py")
    if not valida:
        rel.pulado(area, "valida_rastreabilidade.py não encontrado ao lado deste script.")
    for d in docs:
        rc, out = run([sys.executable, str(valida), str(d)], cwd=raiz) if valida else (0, "")
        nome = d.relative_to(raiz).as_posix()
        if not valida:
            break
        if rc == 1:
            furos = [l[2:] for l in out.splitlines() if l.startswith("- ")][:3]
            rel.furo(area, f"rastreabilidade de {nome}: " + " | ".join(furos)
                     + f" (detalhe: python3 {valida.name} {nome})")
        elif rc == 2:
            rel.alerta(area, f"rastreabilidade de {nome}: não deu para parsear — confira à mão.")
        else:
            rel.ok(area, f"rastreabilidade de {nome} sem furo.")

    if not (raiz / "openspec").is_dir():
        return
    cli = shutil.which("openspec")
    if not cli:
        (rel.furo if modo == "ci" else rel.pulado)(
            area, "CLI do OpenSpec ausente — `openspec validate` não rodou "
                  "(npm install -g @fission-ai/openspec@latest).")
    elif modo != "precommit":
        rc, out = run([cli, "validate", "--all", "--strict", "--no-interactive"], cwd=raiz)
        if rc != 0:
            rel.furo(area, "openspec validate --all --strict reprovou: "
                     + " | ".join(l.strip() for l in out.splitlines() if "✗" in l or "ERROR" in l or "WARNING" in l)[:400])
        else:
            rel.ok(area, "openspec validate --all --strict aprovado.")
        rc, out = run([cli, "validate", "--archived", "--no-interactive"], cwd=raiz)
        if rc != 0 and "archive" in out.lower():
            rel.furo(area, "change arquivada com tarefa em aberto (openspec validate --archived).")

    vig, ativos = specs_t(raiz)
    if vig or ativos:
        citados = testes_citados(raiz)
        sem = sorted(vig - set(citados), key=chave_tn)
        if sem:
            rel.furo(area, f"{len(sem)} cenário(s) vigente(s) sem teste que cite o ID: "
                           + ", ".join(f"T-{n}" for n in sem[:12])
                           + (" …" if len(sem) > 12 else "")
                           + " — o nome do teste deve conter T-n (ex.: test_T12_…, it('T-12 …')).")
        else:
            rel.ok(area, f"todos os {len(vig)} T-n vigentes citados em testes.")
        em_curso = sorted(ativos - set(citados), key=chave_tn)
        if em_curso and modo == "local":
            rel.alerta(area, "T-n de changes ativas ainda sem teste (em curso): "
                             + ", ".join(f"T-{n}" for n in em_curso[:12]))
        orfaos = sorted(set(citados) - vig - ativos, key=chave_tn)
        if orfaos and modo == "local":
            rel.alerta(area, "testes citam T-n que não existe nas specs: "
                             + ", ".join(f"T-{n} ({citados[n]})" for n in orfaos[:6]))


# ---------------------------------------------------------- dependências
def checa_dependencias(raiz, rel, modo):
    area = "dependências"
    dep = achar_script("dependabot_config.py")
    if not dep:
        rel.pulado(area, "dependabot_config.py não encontrado.")
        return
    rc, out = run([sys.executable, str(dep), str(raiz), "--check"], cwd=raiz)
    if rc == 1:
        rel.furo(area, "Dependabot sem cobertura: " + " | ".join(
            l.strip("- ") for l in out.splitlines() if "furo" in l.lower() or l.startswith("-"))[:300]
            + " (python3 dependabot_config.py <raiz> para gerar).")
    elif rc == 2:
        if "nao existe" in out.lower() or "não existe" in out.lower() or "nenhum" in out.lower():
            rel.alerta(area, out.splitlines()[-1][:200] if out else "dependabot não avaliado.")
        else:
            rel.alerta(area, "dependabot.yml fora do padrão — confira à mão.")
    else:
        rel.ok(area, "Dependabot cobre os manifestos.")


# ------------------------------------------------------------ ambientes
def jobs_nomad(raiz: Path):
    out = []
    for f in arquivos(raiz):
        if f.suffix in (".hcl", ".nomad") and "openspec" not in f.parts:
            try:
                if re.search(r'^\s*job\s+"', f.read_text(encoding="utf-8", errors="replace"), re.M):
                    out.append(f)
            except OSError:
                pass
    return out


def tem_deploy(raiz: Path):
    marcas = ("Dockerfile", "docker-compose.yml", "compose.yml", "compose.yaml")
    for f in arquivos(raiz):
        if f.name in marcas or f.suffix in (".hcl", ".tf", ".nomad"):
            return True
        if "k8s" in f.parts or "helm" in f.parts:
            return True
    return False


RE_AMB = re.compile(r"(?<![a-z])(hml|homolog|homologacao|staging|stg)(?![a-z])", re.I)


def pares_ambiente(raiz: Path):
    pares = []
    for f in arquivos(raiz):
        if f.suffix not in ("", ".env", ".yaml", ".yml", ".json", ".hcl", ".tf", ".properties") and \
                not f.name.startswith(".env"):
            continue
        if not RE_AMB.search(f.name):
            continue
        for alvo in ("prod", "production", "prd"):
            cand = f.with_name(RE_AMB.sub(alvo, f.name))
            if cand.is_file() and cand != f:
                pares.append((f, cand))
                break
    return pares


def divergencias_declaradas(amb: Path):
    txt = amb.read_text(encoding="utf-8", errors="replace")
    m = re.search(r"^#+\s*Diverg[eê]ncias deliberadas.*?$(.*?)(?=^#+\s|\Z)", txt, re.M | re.S | re.I)
    return set(re.findall(r"`([A-Za-z0-9_.*-]+)`", m.group(1))) if m else set()


def checa_ambientes(raiz, docs, rel, modo):
    area = "ambientes"
    amb = None
    for d in docs:
        c = d / "Infra" / "AMBIENTES.md"
        if c.is_file():
            amb = c
            break
    if amb is None:
        if tem_deploy(raiz):
            (rel.furo if modo == "ci" else rel.alerta)(
                area, "há artefato de deploy mas nenhum DOC/.../Infra/AMBIENTES.md — a skill ambientes "
                      "não tem contrato para provar o alvo (template em ambientes/references/).")
    else:
        n = amb.read_text(encoding="utf-8", errors="replace").count("A DEFINIR")
        if n:
            rel.alerta(area, f"{amb.relative_to(raiz)} tem {n} campo(s) 'A DEFINIR'.")
        else:
            rel.ok(area, f"{amb.relative_to(raiz)} preenchido.")

    diff = achar_script("diff_ambientes.py")
    pares = pares_ambiente(raiz)
    if pares and not diff:
        rel.pulado(area, "pares hml/prod encontrados, mas diff_ambientes.py não está ao lado deste script.")
        return
    ignorar = divergencias_declaradas(amb) if amb else set()
    for a, b in pares:
        cmd = [sys.executable, str(diff), str(a), str(b), "--rotulos", "hml", "prod"]
        for i in sorted(ignorar):
            cmd += ["--ignorar", i]
        rc, out = run(cmd, cwd=raiz)
        par = f"{a.relative_to(raiz)} × {b.relative_to(raiz)}"
        if rc == 1:
            rel.furo(area, f"chave sem par entre ambientes em {par} — declare em 'Divergências deliberadas' "
                           f"do AMBIENTES.md ou crie a chave do outro lado.")
        elif rc == 2:
            rel.alerta(area, f"não deu para comparar {par}.")
        else:
            rel.ok(area, f"{par}: mesmas chaves.")


# ---------------------------------------------------------------- mapa
def checa_mapa(raiz, docs, rel, modo):
    area = "mapa"
    mapeia = achar_script("mapeia_projeto.py")
    if not mapeia or not docs:
        return
    for d in docs:
        pasta = d / "System" / "mapa"
        nome = d.relative_to(raiz).as_posix()
        if not (pasta / "inventario.json").is_file():
            n_codigo = sum(1 for f in arquivos(raiz) if f.suffix.lower() in EXT_CODIGO)
            if n_codigo >= 20 and modo == "local":
                rel.alerta(area, f"{nome} sem mapa do código ({n_codigo} arquivos de código) — gere com "
                                 f"python3 {nome}/tools/mapeia_projeto.py . e escreva System/mapa/visao-geral.md "
                                 "(procedimento em references/mapa-projeto.md da novo-projeto-sdd).")
            continue
        raiz_proj = raiz if len(docs) == 1 else raiz
        rc, out = run([sys.executable, str(mapeia), str(raiz_proj), "--saida", str(pasta), "--check"], cwd=raiz)
        if rc == 1:
            itens = [l.strip()[2:] for l in out.splitlines() if l.strip().startswith("- ")][:5]
            rel.alerta(area, f"mapa de {nome} desatualizado: " + "; ".join(itens)
                       + " — rode mapeia_projeto.py e revise as páginas das áreas afetadas.")
        elif rc == 0:
            rel.ok(area, f"mapa de {nome} em dia com o código.")
            if not (pasta / "visao-geral.md").is_file() and modo == "local":
                rel.alerta(area, f"{nome}/System/mapa tem o inventário mas não a visao-geral.md (interpretação).")


# --------------------------------------------------------------- nomad
def checa_nomad(raiz, rel):
    area = "nomad-worker"
    for f in jobs_nomad(raiz):
        txt = f.read_text(encoding="utf-8", errors="replace")
        nome = f.relative_to(raiz)
        for img in re.findall(r'image\s*=\s*"([^"]+)"', txt):
            if "${" in img or "{{" in img:
                continue
            tag = img.rsplit("/", 1)[-1]
            if ":" not in tag and "@sha256" not in img:
                rel.furo(area, f"{nome}: imagem '{img}' sem tag — use tag imutável (SHA).")
            elif tag.endswith(":latest"):
                rel.furo(area, f"{nome}: imagem '{img}' com :latest — use tag imutável (SHA).")
        for bloco in re.findall(r"\benv\s*\{(.*?)\}", txt, re.S):
            for k, v in re.findall(r'([A-Za-z0-9_]+)\s*=\s*"([^"]*)"', bloco):
                if re.search(r"PASS|SECRET|TOKEN|PRIVATE|API_?KEY|CREDENTIAL", k, re.I) and v \
                        and "${" not in v and "{{" not in v:
                    rel.furo(area, f"{nome}: segredo literal em env ({k}) — segredo só via template do Vault.")


# ------------------------------------------------------------- sentinel
def checa_sentinel(raiz, rel, modo):
    """Sentinel só no git local: o que vai ser commitado (precommit) ou o trabalho
    em aberto (stop/local) precisa ter sido auditado, arquivo a arquivo."""
    area = "sentinel"
    amplos = [p for p in padroes_sentinel(raiz) if padrao_amplo_sentinel(p)]
    if amplos:
        rel.furo(area, "padrão amplo demais no .sentinelignore (desliga a sentinel para o repositório): "
                       + ", ".join(amplos) + ". Liste só pasta ou nome de código gerado/terceiros.")
    escopo = "staged" if modo == "precommit" else "trabalho"
    arqs = arquivos_alterados(raiz, escopo)
    if arqs is None:
        rel.pulado(area, "não é repositório git — auditoria não verificável.")
        return
    if not arqs:
        rel.ok(area, "nenhum arquivo de código " + ("no commit." if escopo == "staged" else "alterado."))
        return
    nao, corrigir = situacao_auditoria(raiz, arqs)
    if corrigir:
        rel.furo(area, "a sentinel mandou corrigir antes de concluir e estes arquivos não mudaram desde então: "
                       + ", ".join(corrigir[:8]) + (" …" if len(corrigir) > 8 else ""))
    if nao:
        (rel.furo if modo in ("stop", "precommit") else rel.alerta)(
            area, f"{len(nao)} arquivo(s) de código sem auditoria da sentinel no conteúdo atual: "
                  + ", ".join(nao[:8]) + (" …" if len(nao) > 8 else "")
                  + ". Rode a skill sentinel sobre o diff e registre: python3 <DOC>/tools/sdd_hooks.py "
                    "auditoria --veredito \"aprovado|aprovado com ressalvas|corrigir antes de concluir\".")
    if not nao and not corrigir:
        rel.ok(area, f"{len(arqs)} arquivo(s) de código auditado(s) pela sentinel.")


# ----------------------------------------------------------------- main
def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("raiz")
    ap.add_argument("--modo", choices=("local", "ci", "precommit", "stop"), default="local")
    ap.add_argument("--fonte", default=None)
    ap.add_argument("--json", action="store_true", help="saída JSON (para hooks e CI)")
    ap.add_argument("--sem-log", action="store_true", help="não grava em <DOC>/log/")
    args = ap.parse_args()

    raiz = Path(args.raiz).resolve()
    if not raiz.is_dir():
        print(f"ERRO: diretório não encontrado: {raiz}")
        return 2
    fonte = Path(args.fonte).resolve() if args.fonte else (
        AQUI.parents[1] if (AQUI.parents[1] / "conhecimento" / "SKILL.md").is_file() else None)

    rel = Relatorio()
    docs = docs_do_repo(raiz)
    if args.modo != "stop":
        checa_estrutura(raiz, docs, rel, args.modo, fonte)
    checa_rastreabilidade(raiz, docs, rel, args.modo)
    if args.modo != "stop":
        checa_dependencias(raiz, rel, args.modo)
    checa_ambientes(raiz, docs, rel, args.modo)
    if args.modo in ("local", "ci"):
        checa_mapa(raiz, docs, rel, args.modo)
    checa_nomad(raiz, rel)
    if args.modo in ("local", "stop", "precommit"):
        checa_sentinel(raiz, rel, args.modo)

    if args.modo in ("local", "ci") and not args.sem_log:
        try:
            import registro_log
            registro_log.registrar(raiz, f"checa_projeto-{args.modo}", 1 if rel.furos else 0,
                                   resumo={"furos": len(rel.furos)},
                                   problemas=[{"nivel": "FALTA" if n == "FURO" else n, "etapa": a, "msg": m}
                                              for n, a, m in rel.itens if n in ("FURO", "ALERTA")])
        except ImportError:
            pass
    if args.json:
        print(json.dumps([{"nivel": n, "area": a, "msg": m} for n, a, m in rel.itens], ensure_ascii=False))
        return 1 if rel.furos else 0

    print(f"=== Checagem SDD ({args.modo}) — {raiz} ===")
    ordem = ["estrutura", "novo-projeto-sdd", "dependências", "ambientes", "mapa", "nomad-worker", "sentinel"]
    for area in ordem:
        itens = [i for i in rel.itens if i[1] == area]
        if not itens:
            continue
        print(f"\n[{area}]")
        for nivel, _, msg in sorted(itens, key=lambda i: ["FURO", "ALERTA", "PULADO", "OK"].index(i[0])):
            print(f"  {nivel:<6} {msg}")
    n_f = len(rel.furos)
    n_a = sum(1 for i in rel.itens if i[0] == "ALERTA")
    print(f"\nResultado: {n_f} furo(s), {n_a} alerta(s)." + (" Corrija os FUROS — eles barram commit, CI e o "
                                                           "encerramento da IA." if n_f else ""))
    return 1 if n_f else 0


if __name__ == "__main__":
    try:
        import registro_log
        sys.exit(registro_log.executar("checa_projeto", main, registro_log.raiz_do_argv))
    except ImportError:
        sys.exit(main())
