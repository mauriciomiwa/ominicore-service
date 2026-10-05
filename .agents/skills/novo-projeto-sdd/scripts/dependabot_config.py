#!/usr/bin/env python3
"""Gera ou confere o .github/dependabot.yml a partir dos manifestos do repo.

Uso:
  python3 dependabot_config.py <raiz-do-repo>            # gera se nao existir
  python3 dependabot_config.py <raiz-do-repo> --check    # so confere cobertura
  python3 dependabot_config.py <raiz-do-repo> --stdout   # imprime, nao grava

Padrao da casa (DOC/Infra/seguranca-dependencias.md):
  - version updates semanais (segunda 06:00, America/Sao_Paulo);
  - minor+patch agrupados num PR por ecossistema; major em PR proprio;
  - no maximo 5 PRs abertos por ecossistema;
  - security updates ficam a cargo do repo (Settings > Code security) e nao
    sao limitados por este arquivo.

Comportamento:
  - NUNCA sobrescreve um dependabot.yml existente: sem --check, reporta e sai.
  - --check compara ecossistemas/diretorios detectados com os declarados no
    arquivo. Diretorio com manifesto e sem cobertura = furo.
  - Diretorio com bun.lock (e seus workspaces sem lockfile proprio) e so do
    ecossistema bun; o mesmo vale para uv.lock sobre o pip.
  - Workspace npm sem lockfile proprio fica coberto pelo diretorio do lockfile.
  - Parser do YAML existente e raso (regex em package-ecosystem, directory e
    directories). Arquivo muito fora do padrao -> rc 2, confira no olho.

Saida: rc 0 = ok/gerado · 1 = furo de cobertura · 2 = nao deu para avaliar.
"""
import argparse
import json
import re
import sys
from pathlib import Path

# arquivo (nome exato ou glob) -> ecossistema do Dependabot
REGRAS = [
    ("package.json", "npm"),            # cobre npm, yarn e pnpm
    ("bun.lock", "bun"), ("bun.lockb", "bun"),
    ("pom.xml", "maven"),
    ("build.gradle", "gradle"), ("build.gradle.kts", "gradle"),
    ("settings.gradle", "gradle"), ("settings.gradle.kts", "gradle"),
    ("go.mod", "gomod"),
    ("uv.lock", "uv"),
    ("pyproject.toml", "pip"), ("requirements*.txt", "pip"),
    ("Pipfile", "pip"), ("setup.py", "pip"),
    ("*.csproj", "nuget"), ("*.fsproj", "nuget"), ("packages.config", "nuget"),
    ("Directory.Packages.props", "nuget"),
    ("global.json", "dotnet-sdk"),
    ("composer.json", "composer"),
    ("pubspec.yaml", "pub"),
    ("Cargo.toml", "cargo"),
    ("Gemfile", "bundler"),
    ("mix.exs", "mix"),
    ("Dockerfile", "docker"), ("*.Dockerfile", "docker"), ("Dockerfile.*", "docker"),
    ("docker-compose*.yml", "docker-compose"), ("docker-compose*.yaml", "docker-compose"),
    ("compose*.yml", "docker-compose"), ("compose*.yaml", "docker-compose"),
    ("Chart.yaml", "helm"),
    ("*.tf", "terraform"),
    ("devcontainer.json", "devcontainers"),
]
IGNORAR = {".git", "node_modules", "vendor", "target", "build", "dist", ".venv",
           "venv", "__pycache__", ".gradle", ".idea", "bin", "obj", ".next",
           ".terraform", "Garbage", ".agents", ".claude", ".worktrees"}
PROFUNDIDADE = 4
# lockfile proprio de um package.json: com ele o diretorio tem gerente proprio
LOCKS_NPM = ("package-lock.json", "npm-shrinkwrap.json", "yarn.lock", "pnpm-lock.yaml")


def detectar(raiz: Path) -> dict:
    """Retorna {ecossistema: set(diretorios relativos '/x/y')}."""
    achados: dict = {}

    def anda(d: Path, nivel: int):
        if nivel > PROFUNDIDADE:
            return
        try:
            itens = sorted(d.iterdir())
        except OSError:
            return
        for p in itens:
            if p.is_dir():
                if p.name not in IGNORAR:
                    anda(p, nivel + 1)
                continue
            for padrao, eco in REGRAS:
                if p.match(padrao):
                    rel = "/" + p.parent.relative_to(raiz).as_posix()
                    rel = "/" if rel == "/." else rel
                    # devcontainer.json vive em .devcontainer/ -> Dependabot quer "/"
                    if eco == "devcontainers":
                        rel = "/"
                    achados.setdefault(eco, set()).add(rel)

    anda(raiz, 0)

    def tem_lock(d: str) -> bool:
        return any((raiz / d.lstrip("/") / f).exists() for f in LOCKS_NPM)

    # uv.lock presente: o pyproject daquele diretorio e do uv, nao do pip
    for d in achados.get("uv", set()):
        achados.get("pip", set()).discard(d)
    # bun.lock presente: o package.json daquele diretorio e dos workspaces abaixo
    # dele (sem lockfile proprio) e do bun. O npm do Dependabot falha nesses
    # diretorios ("cannot update bun.lock") — nao emitir.
    for b in achados.get("bun", set()):
        base = "" if b == "/" else b
        for d in list(achados.get("npm", set())):
            if d == b or (d.startswith(base + "/") and not tem_lock(d)):
                achados["npm"].discard(d)
    # workspaces npm: package.json sem lockfile proprio abaixo de um diretorio
    # npm com lockfile e coberto por ele. Declarar o workspace a parte faz o
    # Dependabot mexer so no package.json dele, sem o lock da raiz -> npm ci falha.
    for r in [d for d in achados.get("npm", set()) if tem_lock(d)]:
        base = "" if r == "/" else r
        for d in list(achados.get("npm", set())):
            if d != r and d.startswith(base + "/") and not tem_lock(d):
                achados["npm"].discard(d)
    for eco in ("pip", "npm"):
        if eco in achados and not achados[eco]:
            del achados[eco]
    if (raiz / ".github" / "workflows").is_dir():
        achados.setdefault("github-actions", set()).add("/")
    return achados


def gerar(achados: dict) -> str:
    linhas = [
        "# Gerado por dependabot_config.py — padrao da casa.",
        "# Politica, SLA e regra de merge: DOC/Infra/seguranca-dependencias.md",
        "# Ecossistema/diretorio novo? Re-rode o script com --check.",
        "version: 2",
        "updates:",
    ]
    for eco in sorted(achados):
        dirs = sorted(achados[eco])
        linhas.append(f'  - package-ecosystem: "{eco}"')
        if len(dirs) == 1:
            linhas.append(f'    directory: "{dirs[0]}"')
        else:
            linhas.append("    directories:")
            linhas += [f'      - "{d}"' for d in dirs]
        linhas += [
            "    schedule:",
            '      interval: "weekly"',
            '      day: "monday"',
            '      time: "06:00"',
            '      timezone: "America/Sao_Paulo"',
            "    open-pull-requests-limit: 5",
            "    commit-message:",
            '      prefix: "chore(deps)"',
            "    groups:",
            "      minor-e-patch:",
            '        update-types: ["minor", "patch"]',
        ]
    return "\n".join(linhas) + "\n"


def ler_declarado(texto: str):
    """{ecossistema: set(dirs)} do YAML existente, por regex. None se nao parseou."""
    decl: dict = {}
    blocos = re.split(r"^\s*-\s*package-ecosystem:\s*", texto, flags=re.M)[1:]
    if not blocos:
        return None
    for b in blocos:
        m = re.match(r"""["']?([\w-]+)["']?""", b)
        if not m:
            return None
        eco = m.group(1)
        dirs = set()
        m1 = re.search(r"""^\s*directory:\s*["']?([^"'\n#]+?)["']?\s*$""", b, re.M)
        if m1:
            dirs.add(m1.group(1).strip())
        m2 = re.search(r"^\s*directories:\s*\n((?:\s*-\s*.+\n?)+)", b, re.M)
        if m2:
            for item in re.findall(r"""-\s*["']?([^"'\n#]+?)["']?\s*$""", m2.group(1), re.M):
                dirs.add(item.strip())
        decl.setdefault(eco, set()).update(dirs)
    return decl


def coberto(d: str, declarados: set) -> bool:
    for x in declarados:
        if x == d:
            return True
        if "*" in x:  # glob de directories
            rx = "^" + re.escape(x).replace(r"\*\*", ".*").replace(r"\*", "[^/]*") + "$"
            if re.match(rx, d):
                return True
    return False


def checar(raiz: Path, alvo: Path, achados: dict) -> dict:
    """Cobertura do dependabot.yml: {rc, furos: [{ecossistema, diretorio}], sobras, sem_arquivo, nao_avaliado}."""
    res = {"rc": 0, "furos": [], "sobras": [], "sem_arquivo": False, "nao_avaliado": False,
           "ecossistemas": sorted(achados), "arquivo": alvo.relative_to(raiz).as_posix()}
    if not alvo.exists():
        res["sem_arquivo"] = bool(achados)
        res["rc"] = 1 if achados else 0
        return res
    decl = ler_declarado(alvo.read_text(encoding="utf-8"))
    if decl is None:
        res.update(nao_avaliado=True, rc=2)
        return res
    res["furos"] = [{"ecossistema": eco, "diretorio": d} for eco, dirs in sorted(achados.items())
                    for d in sorted(dirs) if not coberto(d, decl.get(eco, set()))]
    res["sobras"] = sorted(set(decl) - set(achados))
    res["rc"] = 1 if res["furos"] else 0
    return res


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("raiz")
    ap.add_argument("--check", action="store_true")
    ap.add_argument("--stdout", action="store_true")
    ap.add_argument("--json", action="store_true", help="com --check: resultado em JSON (para outros scripts)")
    a = ap.parse_args()

    raiz = Path(a.raiz).resolve()
    if not raiz.is_dir():
        print(f"ERRO: {raiz} nao e diretorio")
        return 2
    alvo = raiz / ".github" / "dependabot.yml"
    if not alvo.exists() and (raiz / ".github" / "dependabot.yaml").exists():
        alvo = raiz / ".github" / "dependabot.yaml"
    achados = detectar(raiz)

    if a.stdout:
        sys.stdout.write(gerar(achados) if achados else "# nada detectado\n")
        return 0

    if a.check:
        res = checar(raiz, alvo, achados)
        if a.json:
            print(json.dumps(res, ensure_ascii=False))
            return res["rc"]
        if res["sem_arquivo"]:
            print(f"FURO: {res['arquivo']} nao existe e ha manifestos: " + ", ".join(res["ecossistemas"]))
            print("      gere com: python3 dependabot_config.py <raiz>")
        elif not alvo.exists():
            print("OK: sem manifestos e sem dependabot.yml (greenfield).")
        elif res["nao_avaliado"]:
            print(f"NAO AVALIADO: {alvo.name} fora do formato esperado — confira no olho.")
        else:
            for f in res["furos"]:
                print(f"FURO: manifesto sem cobertura no dependabot.yml — {f['ecossistema']} em {f['diretorio']}")
            for s_ in res["sobras"]:
                print(f"AVISO: '{s_}' declarado mas nenhum manifesto detectado "
                      f"(profundidade {PROFUNDIDADE}) — sobra ou manifesto fora do alcance.")
            if not res["furos"]:
                print(f"OK: {len(achados)} ecossistema(s) detectado(s), todos cobertos.")
        return res["rc"]

    if alvo.exists():
        print(f"JA EXISTE: {alvo} — nada foi sobrescrito. Use --check para conferir cobertura.")
        return 0
    if not achados:
        print("NADA DETECTADO: greenfield. Re-rode quando o primeiro manifesto existir.")
        return 0
    alvo.parent.mkdir(parents=True, exist_ok=True)
    alvo.write_text(gerar(achados), encoding="utf-8")
    print(f"GERADO: {alvo}")
    for eco in sorted(achados):
        print(f"  + {eco}: {', '.join(sorted(achados[eco]))}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
