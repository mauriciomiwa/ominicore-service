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
  - Parser do YAML existente e raso (regex em package-ecosystem, directory e
    directories). Arquivo muito fora do padrao -> rc 2, confira no olho.

Saida: rc 0 = ok/gerado · 1 = furo de cobertura · 2 = nao deu para avaliar.
"""
import argparse
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
           ".terraform", "Garbage", ".agents"}
PROFUNDIDADE = 4


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
    # uv.lock presente: o pyproject daquele diretorio e do uv, nao do pip
    for d in achados.get("uv", set()):
        achados.get("pip", set()).discard(d)
    if "pip" in achados and not achados["pip"]:
        del achados["pip"]
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


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("raiz")
    ap.add_argument("--check", action="store_true")
    ap.add_argument("--stdout", action="store_true")
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
        if not alvo.exists():
            if achados:
                print(f"FURO: {alvo.relative_to(raiz)} nao existe e ha manifestos: "
                      + ", ".join(sorted(achados)))
                print("      gere com: python3 dependabot_config.py <raiz>")
                return 1
            print("OK: sem manifestos e sem dependabot.yml (greenfield).")
            return 0
        decl = ler_declarado(alvo.read_text(encoding="utf-8"))
        if decl is None:
            print(f"NAO AVALIADO: {alvo.name} fora do formato esperado — confira no olho.")
            return 2
        furos = []
        for eco, dirs in sorted(achados.items()):
            for d in sorted(dirs):
                if not coberto(d, decl.get(eco, set())):
                    furos.append(f"{eco} em {d}")
        sobras = sorted(set(decl) - set(achados))
        for f in furos:
            print(f"FURO: manifesto sem cobertura no dependabot.yml — {f}")
        for s in sobras:
            print(f"AVISO: '{s}' declarado mas nenhum manifesto detectado "
                  f"(profundidade {PROFUNDIDADE}) — sobra ou manifesto fora do alcance.")
        if not furos:
            print(f"OK: {len(achados)} ecossistema(s) detectado(s), todos cobertos.")
        return 1 if furos else 0

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
