#!/usr/bin/env python3
"""Monta a estrutura DOC/ de um projeto novo a partir de assets/templates/.

Uso:
  python3 scaffold.py <destino da DOC/> --projeto NOME [--porte nucleo|padrao|completo] [--raiz CAMINHO]

  <destino>   onde a DOC/ vai viver: .../DOC ou .../DOC/<projeto>
  --projeto   nome do projeto (substitui {{PROJETO}} nos templates)
  --porte     nucleo | padrao | completo (default: padrao) — ver a seção
              "Proporcionalidade do scaffold" de references/kickoff.md da skill
  --raiz      raiz do repositório, onde vão AGENTS.md e CLAUDE.md
              (default: deduzida — o pai do diretório DOC)

Comportamento:
  - NUNCA sobrescreve arquivo existente (pula e reporta) — seguro re-rodar.
  - Substitui {{PROJETO}}, {{DATA}}, {{PORTE}} e {{DOC_REL}}.
  - Escreve a seção "Quando precisar, crie" do INDEX.md com os artefatos que
    o porte deixou de fora — o lugar reservado é a linha no índice.
  - Copia valida_rastreabilidade.py e dependabot_config.py para <destino>/tools/.
  - OpenSpec (qualquer porte): garante <raiz>/openspec/{specs,changes}/ e o
    openspec/config.yaml com o contexto da casa (pt-BR, DOC/, convencao
    REQ-n / Origem: RN-n / Scenario: T-n). Se o config ja existe (ex.: criado
    pelo `openspec init`), MESCLA: acrescenta o que falta, nunca remove; o
    marcador "sdd-rastreabilidade" evita duplicar ao re-rodar. O que nao der
    para mesclar com seguranca vira AVISO com o trecho a colar.
  - Gera <raiz>/.github/dependabot.yml a partir dos manifestos (qualquer porte;
    greenfield = aviso) e, do porte padrao em diante, instala o workflow de
    auto-merge do Dependabot.
"""
import argparse
import datetime
import json
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

PORTES = {"nucleo": 1, "padrao": 2, "completo": 3}

# (caminho relativo no template, porte mínimo, descrição p/ "Quando precisar, crie")
MANIFEST = [
    ("INDEX.md",                          1, None),
    ("CLAUDE.md",                         1, None),
    ("log.md",                            1, None),
    ("estado.md",                         1, None),
    ("to-do.md",                          1, None),
    ("brainstorm.md",                     1, None),
    ("Negocio/README.md",                 1, None),
    ("Negocio/especificacao-negocio.md",  1, None),
    ("System/README.md",                  1, None),
    ("System/backlog-mvp.md",             1, None),
    ("tools/README.md",                   1, None),
    ("Padroes/README.md",                 2, None),
    ("Padroes/stack-tecnologico.md",      2, "stack + versões + decisões em aberto"),
    ("Padroes/goldrules.md",              2, "princípios de engenharia que a sentinel audita"),
    ("Padroes/fontes-externas.md",        2, "fontes oficiais a consultar antes da web"),
    ("System/plano-de-testes.md",         2, "status dos testes T-n dos cenarios OpenSpec"),
    ("System/arquitetura.md",             2, "componentes, fluxos-chave, decisões"),
    ("Infra/infraestrutura.md",           1, None),
    ("Infra/README.md",                   2, "infra, deploy, rede (nunca segredos)"),
    ("Infra/seguranca-dependencias.md",   2, "politica Dependabot: SLA, regra de merge, ignore"),
    ("Skill/README.md",                   2, "skills do projeto para a IA"),
    ("Padroes/i18n-string-files.md",      3, "convenção i18n: nenhum texto hardcoded"),
    ("System/i18n/strings.pt-BR.json",    3, "arquivo de strings chave-valor"),
    ("Negocio/glossario.md",              1, None),
    ("System/decisoes/README.md",         2, "decisões (ADR): difíceis de reverter, surpreendentes e com alternativa real"),
    ("DRAW/README.md",                    3, "diagramas (fonte editável + export)"),
    ("Suporte/README.md",                 3, "materiais de referência (nunca segredos)"),
    ("Garbage/README.md",                 3, "inventário de descartáveis"),
]
# (template, destino na raiz, porte minimo)
RAIZ_FILES = [
    ("_raiz/AGENTS.md", "AGENTS.md", 1),
    ("_raiz/CLAUDE.md", "CLAUDE.md", 1),
    ("_raiz/openspec/specs/.gitkeep", "openspec/specs/.gitkeep", 1),
    ("_raiz/openspec/changes/.gitkeep", "openspec/changes/.gitkeep", 1),
    ("_raiz/.github/workflows/dependabot-automerge.yml",
     ".github/workflows/dependabot-automerge.yml", 2),
    # garantias: pre-commit + CI em qualquer porte (hooks do Claude: mescla abaixo)
    ("_raiz/.githooks/pre-commit", ".githooks/pre-commit", 1),
    ("_raiz/.github/workflows/sdd-checks.yml", ".github/workflows/sdd-checks.yml", 1),
]


def mesclar_settings(existente: dict, nosso: dict):
    """Acrescenta os hooks da casa ao .claude/settings.json sem remover nada."""
    hooks = existente.setdefault("hooks", {})
    mudou = False
    for evento, entradas in nosso["hooks"].items():
        atuais = hooks.setdefault(evento, [])
        comandos = [h.get("command", "") for e in atuais if isinstance(e, dict)
                    for h in e.get("hooks", []) if isinstance(h, dict)]
        for entrada in entradas:
            sub = entrada["hooks"][0]["command"].rsplit(" ", 1)[-1]
            if any(re.search(rf'sdd_hooks\.py"?\s+{re.escape(sub)}\b', c) for c in comandos):
                continue
            atuais.append(entrada)
            mudou = True
    return existente, mudou
MARCADOR = "<!-- QUANDO-PRECISAR-CRIE -->"
MARCADOR_OPENSPEC = "sdd-rastreabilidade"


def blocos_yaml(texto: str):
    """Divide um YAML simples em {chave_topo: bloco} preservando o texto."""
    blocos, atual = {}, None
    for ln in texto.splitlines(keepends=True):
        m = re.match(r"^([A-Za-z_][\w-]*):", ln)
        if m:
            atual = m.group(1)
            blocos[atual] = ln
        elif atual:
            blocos[atual] += ln
    return blocos


def mesclar_config(existente: str, nosso: str):
    """Mescla o config.yaml da casa num config existente. Retorna (texto, avisos).

    Nunca remove nada. O marcador so controla o bloco context; rules/operations
    sao conferidas a cada execucao (aviso repete enquanto faltar)."""
    avisos = []
    nossos = blocos_yaml(nosso)
    deles = blocos_yaml(existente)
    texto = existente if existente.endswith("\n") else existente + "\n"

    if MARCADOR_OPENSPEC not in existente:
        ctx_nosso = nossos["context"].split("\n", 1)[1].rstrip("\n") + "\n"
        if "context" not in deles:
            texto += "\n" + nossos["context"]
        elif re.match(r"^context:\s*\|[-+]?\s*$", deles["context"].splitlines()[0]):
            linhas = texto.splitlines(keepends=True)
            ini = next(i for i, l in enumerate(linhas) if re.match(r"^context:", l))
            fim = ini + 1
            while fim < len(linhas) and (not linhas[fim].strip() or linhas[fim].startswith((" ", "\t"))):
                fim += 1
            while fim > ini + 1 and not linhas[fim - 1].strip():
                fim -= 1
            # indentacao do bloco existente (o nosso usa 2 espacos)
            corpo = [l for l in linhas[ini + 1:fim] if l.strip()]
            recuo = re.match(r"^[ \t]*", corpo[0]).group(0) if corpo else "  "
            if "Language:" in deles["context"]:
                ctx_nosso = ctx_nosso[ctx_nosso.index("  # " + MARCADOR_OPENSPEC):]
            reind = [(recuo + l[2:]) if l.strip() else l for l in ctx_nosso.splitlines(keepends=True)]
            linhas[fim:fim] = ["\n"] + reind
            texto = "".join(linhas)
        else:
            avisos.append("openspec/config.yaml tem 'context' fora do formato literal (context: |) — "
                          "cole manualmente o bloco context do template da skill.")

    for chave, assinatura in (("rules", "valida_rastreabilidade"), ("operations", "sentinel")):
        if chave not in deles:
            texto += "\n" + nossos[chave]
        elif assinatura not in deles[chave]:
            avisos.append(f"openspec/config.yaml ja tem '{chave}:' sem as entradas da casa — acrescente-as "
                          f"a mao (template: assets/templates/_raiz/openspec/config.yaml).")

    # nao grava YAML quebrado: confere com PyYAML quando disponivel
    try:
        import yaml  # type: ignore
        try:
            yaml.safe_load(texto)
        except yaml.YAMLError as e:
            avisos.append(f"mescla do openspec/config.yaml geraria YAML invalido ({e.__class__.__name__}) — "
                          "nada gravado; mescle a mao a partir do template.")
            return existente, avisos
    except ImportError:
        if texto != existente:
            avisos.append("PyYAML ausente — confira o openspec/config.yaml mesclado "
                          "(abra o arquivo ou rode `openspec validate`).")
    return texto, avisos


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("destino")
    ap.add_argument("--projeto", required=True)
    ap.add_argument("--porte", choices=sorted(PORTES), default="padrao")
    ap.add_argument("--raiz", default=None)
    args = ap.parse_args()

    aqui = Path(__file__).resolve().parent
    templates = aqui.parent / "assets" / "templates"
    if not templates.is_dir():
        sys.exit(f"ERRO: templates nao encontrados em {templates}")

    destino = Path(args.destino).resolve()
    nivel = PORTES[args.porte]

    # raiz do repositorio: o pai do diretorio chamado DOC (ou o pai do destino)
    if args.raiz:
        raiz = Path(args.raiz).resolve()
    else:
        p = destino
        while p.name.upper() != "DOC" and p.parent != p:
            p = p.parent
        raiz = p.parent if p.name.upper() == "DOC" else destino.parent

    if not destino.is_relative_to(raiz):
        sys.exit(f"ERRO: destino {destino} fora da raiz {raiz} — os hooks apontariam para caminho inexistente. "
                 "Confira --raiz.")
    doc_rel = destino.relative_to(raiz).as_posix()
    openspec_rel = Path(os.path.relpath(raiz / "openspec", destino)).as_posix()
    subst = {
        "{{PROJETO}}": args.projeto,
        "{{DATA}}": datetime.date.today().isoformat(),
        "{{PORTE}}": args.porte,
        "{{DOC_REL}}": doc_rel,
        "{{OPENSPEC_REL}}": openspec_rel,
    }

    sys.path.insert(0, str(aqui))
    import checa_projeto as cp
    so_dkm = () if cp.e_dkm(raiz) else cp.SKILLS_SO_DKM  # skill da casa DKM fica fora do AGENTS.md alheio

    def render(caminho: Path) -> str:
        texto = caminho.read_text(encoding="utf-8")
        for k, v in subst.items():
            texto = texto.replace(k, v)
        if caminho.name == "AGENTS.md" and so_dkm:
            texto = cp.sem_linhas_de_skill(texto, so_dkm)
        return texto

    criados, pulados = [], []

    def escrever(origem: Path, alvo: Path):
        if alvo.exists():
            pulados.append(alvo)
            return
        alvo.parent.mkdir(parents=True, exist_ok=True)
        alvo.write_text(render(origem), encoding="utf-8")
        criados.append(alvo)

    fora = []
    for rel, minimo, descr in MANIFEST:
        if minimo <= nivel:
            escrever(templates / rel, destino / rel)
        elif descr:
            fora.append((rel, descr))

    for rel_tpl, nome, minimo in RAIZ_FILES:
        if minimo <= nivel:
            escrever(templates / rel_tpl, raiz / nome)

    # openspec/config.yaml: cria ou mescla (nunca remove nada do existente)
    avisos_openspec = []
    monorepo = destino.parent.name.upper() == "DOC"
    cfg_tpl = templates / "_raiz" / "openspec" / "config.yaml"

    def render_cfg() -> str:
        texto = cfg_tpl.read_text(encoding="utf-8")
        if monorepo:
            # config e do REPOSITORIO: nao amarrar ao primeiro projeto
            nota = ("  Monorepo: cada projeto tem sua DOC em DOC/<projeto>/ e suas capacidades no\n"
                    "  OpenSpec prefixadas pelo nome dele (openspec/specs/<projeto>-<capacidade>/).\n"
                    "  RN-n valem por projeto; REQ-n e T-n sao unicos no repositorio inteiro.\n"
                    "  Nas linhas abaixo, <projeto> = prefixo da capacidade em questao.")
            texto = texto.replace("{{PROJETO}}", "<projeto>").replace("{{DOC_REL}}", "DOC/<projeto>")
            texto = texto.replace("  {{MONOREPO_NOTA}}", nota)
        else:
            texto = texto.replace("  {{MONOREPO_NOTA}}\n", "")
        for k, v in subst.items():
            texto = texto.replace(k, v)
        return texto

    cfg = raiz / "openspec" / "config.yaml"
    if not cfg.exists() and (raiz / "openspec" / "config.yml").exists():
        cfg = raiz / "openspec" / "config.yml"
    if not cfg.exists():
        cfg.parent.mkdir(parents=True, exist_ok=True)
        cfg.write_text(render_cfg(), encoding="utf-8")
        criados.append(cfg)
    else:
        atual = cfg.read_text(encoding="utf-8")
        novo, avisos_openspec = mesclar_config(atual, render_cfg())
        if novo != atual:
            cfg.write_text(novo, encoding="utf-8")
            criados.append(f"{cfg} (mesclado)")
        else:
            pulados.append(cfg)

    # scripts de manutencao -> tools/
    for nome_script, src in cp.scripts_tools(aqui.parents[1]).items():
        if not src.is_file():
            avisos_openspec.append(f"{nome_script} nao encontrado (fonte da skill ambientes ausente?) — "
                                   f"copie-o para {destino / 'tools'} para a checagem de ambientes rodar.")
            continue
        if src.is_file():
            alvo = destino / "tools" / src.name
            if alvo.exists() and alvo.read_text(encoding="utf-8") != src.read_text(encoding="utf-8"):
                avisos_openspec.append(f"{alvo} difere da versao da skill (desatualizado?) — nada "
                                       f"sobrescrito; para atualizar: cp {src} {alvo}")
            if not alvo.exists():
                alvo.parent.mkdir(parents=True, exist_ok=True)
                alvo.write_text(src.read_text(encoding="utf-8"), encoding="utf-8")
                criados.append(alvo)

    # .github/dependabot.yml a partir dos manifestos (nunca sobrescreve)
    dep_yml = raiz / ".github" / "dependabot.yml"
    if dep_yml.exists() or (raiz / ".github" / "dependabot.yaml").exists():
        pulados.append(dep_yml)
    else:
        sys.dont_write_bytecode = True  # nao sujar a pasta da skill
        sys.path.insert(0, str(aqui))
        import dependabot_config as dc
        achados = dc.detectar(raiz)
        if achados:
            dep_yml.parent.mkdir(parents=True, exist_ok=True)
            dep_yml.write_text(dc.gerar(achados), encoding="utf-8")
            criados.append(dep_yml)
        else:
            print("AVISO: nenhum manifesto detectado (greenfield) — .github/dependabot.yml "
                  f"nao gerado. Quando o primeiro existir: python3 {doc_rel}/tools/"
                  "dependabot_config.py <raiz>")

    # skills que viajam com o projeto -> <raiz>/.agents/skills/
    # Fontes ficam ao lado da pasta desta skill: no repo skills-miwa e no
    # diretorio de skills sincronizadas do perfil, o layout e o mesmo.
    skills_fonte = aqui.parents[1]  # pasta que contem as skills irmas
    sys.dont_write_bytecode = True
    skills_projeto = cp.skills_padrao(skills_fonte, raiz)  # a mesma regra do sync-skills.py
    sem_fonte = [n for n in cp.SKILLS_CASA if n not in skills_projeto]
    for nome in skills_projeto:
        src = skills_fonte / nome
        for f in sorted(src.rglob("*")):
            if f.is_file() and f.name != ".DS_Store" and "__pycache__" not in f.parts and f.suffix != ".pyc":
                # cópia fiel (sem render): os templates da novo-projeto-sdd têm {{PROJETO}} de propósito
                alvo = raiz / ".agents" / "skills" / nome / f.relative_to(src)
                if alvo.exists():
                    pulados.append(alvo)
                    continue
                alvo.parent.mkdir(parents=True, exist_ok=True)
                shutil.copyfile(f, alvo)
                criados.append(alvo)
    if sem_fonte:
        print(f"AVISO: fonte de {', '.join(sem_fonte)} nao encontrada ao lado da "
              f"skill — copie-as para {raiz / '.agents' / 'skills'} manualmente "
              f"(ou rode o sync-skills.py do repositorio skills-miwa).")

    # pre-commit: o nosso sai executavel; um que ja existia nao e mexido, mas tem de chamar a checagem
    pc = raiz / ".githooks" / "pre-commit"
    if pc in criados:
        pc.chmod(0o755)
    elif pc.is_file() and "checa_projeto" not in pc.read_text(encoding="utf-8", errors="replace"):
        avisos_openspec.append(f"{pc.relative_to(raiz)} ja existia e nao chama o checa_projeto.py — o commit nao "
                               f"passa pelas garantias. Acrescente a chamada do template "
                               f"(assets/templates/_raiz/.githooks/pre-commit); nada foi alterado nele.")

    # hooks do Claude Code: cria ou mescla .claude/settings.json
    nosso_settings = json.loads(render(templates / "_raiz" / ".claude" / "settings.json"))
    st = raiz / ".claude" / "settings.json"
    if not st.exists():
        st.parent.mkdir(parents=True, exist_ok=True)
        st.write_text(json.dumps(nosso_settings, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
        criados.append(st)
    else:
        try:
            atual = json.loads(st.read_text(encoding="utf-8"))
            novo, mudou = mesclar_settings(atual, nosso_settings)
            if mudou:
                st.write_text(json.dumps(novo, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
                criados.append(f"{st} (hooks mesclados)")
            else:
                pulados.append(st)
        except (json.JSONDecodeError, AttributeError):
            avisos_openspec.append(f"{st} nao e JSON valido — hooks nao instalados; mescle a mao a partir de "
                                   "assets/templates/_raiz/.claude/settings.json.")

    # git: usar .githooks/ e ignorar .sdd/ (estado local dos hooks)
    e_git = (raiz / ".git").exists() and subprocess.run(
        ["git", "rev-parse", "--git-dir"], cwd=raiz, capture_output=True, text=True).returncode == 0
    if e_git:
        atual_hp = subprocess.run(["git", "config", "--get", "core.hooksPath"], cwd=raiz,
                                  capture_output=True, text=True).stdout.strip()
        if not atual_hp:
            r = subprocess.run(["git", "config", "core.hooksPath", ".githooks"], cwd=raiz,
                               capture_output=True, text=True)
            if r.returncode == 0:
                criados.append("git config core.hooksPath .githooks")
            else:
                avisos_openspec.append("nao consegui rodar `git config core.hooksPath .githooks` — rode a mao.")
        elif atual_hp != ".githooks":
            avisos_openspec.append(f"core.hooksPath ja aponta para '{atual_hp}' — chame "
                                   ".githooks/pre-commit a partir do hook existente.")
    gi = raiz / ".gitignore"
    linhas_gi = gi.read_text(encoding="utf-8").splitlines() if gi.exists() else []
    faltam_gi = [l for l in (".sdd/", "__pycache__/") if l not in linhas_gi]
    if faltam_gi:
        with gi.open("a", encoding="utf-8") as h:
            h.write(("\n" if linhas_gi and linhas_gi[-1] else "")
                    + "# estado local dos hooks SDD e cache do Python das ferramentas\n" + "\n".join(faltam_gi) + "\n")
        criados.append(f"{gi} (+ {', '.join(faltam_gi)})")

    # "Quando precisar, crie" no INDEX.md
    index = destino / "INDEX.md"
    if index.exists() and MARCADOR in index.read_text(encoding="utf-8"):
        if fora:
            linhas = "\n".join(f"- `{rel}` — {descr}" for rel, descr in fora)
        else:
            linhas = "- (nada — porte completo: todos os artefatos criados)"
        index.write_text(
            index.read_text(encoding="utf-8").replace(MARCADOR, linhas),
            encoding="utf-8",
        )

    print(f"Scaffold '{args.projeto}' (porte {args.porte}) em {destino}")
    print(f"Raiz do repositorio: {raiz} (DOC_REL={doc_rel})")
    print(f"Criados: {len(criados)}")
    for c in criados:
        print(f"  + {c}")
    if pulados:
        print(f"Pulados (ja existiam — nada foi sobrescrito): {len(pulados)}")
        for p in pulados:
            print(f"  = {p}")
    if fora:
        print(f"Fora do porte (listados no INDEX.md): {len(fora)}")
    for a in avisos_openspec:
        print(f"AVISO: {a}")
    print(f"CONFERIR: python3 {doc_rel}/tools/checa_projeto.py {raiz} — e marque o check "
          "'sdd-checks' como obrigatorio na protecao da branch.")
    instalados = [d for base in (".claude/skills", ".agents/skills", ".cursor/skills")
                  if (raiz / base).is_dir() for d in (raiz / base).glob("openspec-*")]
    if not instalados:
        if shutil.which("openspec"):
            print("PROXIMO: comandos /opsx ainda nao instalados — na raiz rode "
                  "`openspec init --tools claude,agents` (acrescente cursor, codex... conforme o time) "
                  "e depois re-rode este scaffold para garantir a mescla do config.yaml.")
        else:
            print("PROXIMO: OpenSpec CLI nao encontrado (requer Node 20.19+). Instale com "
                  "`npm install -g @fission-ai/openspec@latest`, rode `openspec init --tools claude,agents` "
                  "na raiz e re-rode este scaffold. Sem o CLI, as specs funcionam como markdown, "
                  "mas sem /opsx e sem `openspec validate`/`archive`.")


if __name__ == "__main__":
    main()
