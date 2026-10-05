#!/usr/bin/env python3
"""Inventário determinístico do código: a base factual do mapa do projeto.

Uso:
  python3 mapeia_projeto.py <raiz-do-repo> [--saida <DOC>/System/mapa]   # gera
  python3 mapeia_projeto.py <raiz-do-repo> --check [--saida ...]         # só compara

Extrai o que o código afirma, sem interpretar:
  stack        manifestos, linguagem, frameworks principais
  módulos      pastas de código (src/<x> ou <x>/ no topo) e quem importa quem
  entradas     rotas de front (vue-router, react-router, Next/Nuxt por arquivo),
               endpoints de back (Express/Fastify/Koa, Nest, FastAPI/Flask,
               Spring, Go net/http/gin/echo/chi), filas consumidas, crons
  saídas       chamadas HTTP (axios/fetch/http()/api...), filas publicadas
  config       variáveis de ambiente lidas no código e declaradas em .env*
  dados        migrations, schema do ORM, arquivos .sql
  infra        Dockerfile, compose, jobs Nomad, workflows de CI

Grava <saida>/inventario.json (fonte para o --check) e <saida>/inventario.md
(resumo legível + diagrama Mermaid dos módulos). Os dois são GERADOS: não
edite. A interpretação (o que cada área faz, fluxos, RN candidatas) fica nas
outras páginas de <saida>/, escritas pela IA a partir deste inventário.

--check compara o código atual com o inventario.json salvo e lista o que
surgiu ou sumiu (rota, endpoint, chamada HTTP, env, módulo, fila). Exit 0 =
mapa em dia; 1 = drift; 2 = sem inventário salvo ou uso incorreto.
Nunca imprime valores de variáveis de ambiente — só nomes.
"""
import argparse
import datetime
import json
import os
import re
import sys
from pathlib import Path

IGNORAR_DIRS = {".git", "node_modules", ".venv", "venv", "__pycache__", "dist", "build", "target",
                ".next", ".nuxt", ".output", ".idea", ".vscode", "vendor", ".sdd", ".gradle", "coverage",
                ".wrangler", "DOC", "openspec", ".agents", ".claude", ".githooks", "public", "static"}
EXT_JS = {".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".vue", ".svelte"}
EXT_CODIGO = EXT_JS | {".py", ".go", ".java", ".kt", ".rb", ".php", ".cs"}
LIMITE_ARQUIVO = 600_000  # bytes; arquivos maiores (bundles) são ignorados

MANIFESTOS = {
    "package.json": "js/ts", "pyproject.toml": "python", "requirements.txt": "python",
    "go.mod": "go", "pom.xml": "java", "build.gradle": "java/kotlin", "build.gradle.kts": "kotlin",
    "composer.json": "php", "Gemfile": "ruby", "Cargo.toml": "rust", "pubspec.yaml": "dart",
}
FRAMEWORKS = ["vue", "nuxt", "react", "next", "@angular/core", "svelte", "express", "fastify", "koa",
              "@nestjs/core", "fastapi", "flask", "django", "gin-gonic/gin", "labstack/echo", "go-chi/chi",
              "spring-boot", "laravel/framework", "rails", "vue-router", "react-router", "axios",
              "socket.io-client", "socket.io", "amqplib", "bullmq", "prisma", "typeorm", "sequelize",
              "sqlalchemy", "alembic", "gorm"]

M = "(get|post|put|patch|delete)"
RX_ENDPOINT = [
    ("express", re.compile(r"\b(?:app|router|server|fastify|routes?)\s*\.\s*" + M + r"\s*\(\s*['\"`]([^'\"`]+)", re.I)),
    ("fastapi/flask", re.compile(r"@\w+\.(get|post|put|patch|delete|route)\(\s*['\"]([^'\"]+)")),
    ("go", re.compile(r"\b\w+\.(GET|POST|PUT|PATCH|DELETE|Get|Post|Put|Patch|Delete|HandleFunc|Handle)\(\s*\"(/[^\"]*)\"")),
]
RX_NEST_CTRL = re.compile(r"@Controller\(\s*['\"]?([^'\")]*)")
RX_NEST = re.compile(r"@(Get|Post|Put|Patch|Delete)\(\s*['\"]?([^'\")]*)")
RX_SPRING_CLS = re.compile(r"@RequestMapping\(\s*(?:value\s*=\s*|path\s*=\s*)?\"([^\"]*)\"[^)]*\)\s*(?:public\s+)?(?:class|@)")
RX_SPRING = re.compile(r"@(Get|Post|Put|Patch|Delete)Mapping\(\s*(?:value\s*=\s*|path\s*=\s*)?\"?([^\")]*)")
# receptor explícito (instância HTTP): qualquer URL vale; receptor genérico (…Api…, …Http…, request):
# só quando o argumento parece URL (/…, http…, `${base}…`) — senão `therapist.get('x')` vira chamada
_HTTP_CHAMADA = r"(?:\([^()]*\))?\s*\.\s*" + M + r"\s*(?:<[^>()]*>)?\s*\(\s*([`'\"])(.*?)\2"
RX_HTTP = re.compile(r"\b(?:axios|http|api|client|instance|\$http)" + _HTTP_CHAMADA, re.S)
RX_HTTP_GENERICO = re.compile(r"\b(?:\w*[Hh]ttp\w*|\w*[Aa]pi\w*|request)" + _HTTP_CHAMADA, re.S)
# URL passada como constante: http(true).get(RESOURCE, …) / api.post(API_PEDIDOS)
# Go: net/http 1.22 "MÉTODO /caminho" e tabelas {metodo: http.MethodPost, caminho: "/v1/..."}
RX_GO_PADRAO = re.compile(r"\.(?:HandleFunc|Handle)\(\s*\"(GET|POST|PUT|PATCH|DELETE)\s+(/[^\"]*)\"")
RX_GO_TABELA = re.compile(r"http\.Method(Get|Post|Put|Patch|Delete)\b[^\n]*?(?:caminho|path|rota|route|pattern)\w*\s*:\s*\"(/[^\"]*)\"", re.I)
RX_GO_SAIDA = re.compile(r"http\.NewRequest(?:WithContext)?\([^,\n]*?,?\s*(?:http\.Method(Get|Post|Put|Patch|Delete)|\"(GET|POST|PUT|PATCH|DELETE)\")\s*,[^\n]*?\"(/[^\"]*)\"")
# cliente openapi-fetch / ky / ofetch no front: core.GET("/apps"), client.POST(...)
RX_HTTP_MAIUSC = re.compile(r"\b\w+\.(GET|POST|PUT|PATCH|DELETE)\(\s*([`'\"])(/[^`'\"]*)\2")
RX_HTTP_CONST = re.compile(r"\b(?:axios|http|api|client|request|instance|\$http|\w*[Hh]ttp\w*|\w*[Aa]pi\w*)"
                           r"(?:\([^()]*\))?\s*\.\s*" + M + r"\s*(?:<[^>()]*>)?\s*\(\s*([A-Z_][A-Z0-9_]*)\s*[,)]")
RX_FETCH = re.compile(r"\bfetch\(\s*([`'\"])(.*?)\1")
RX_CONST_STR = re.compile(r"\b(?:const|let|var|export\s+const)\s+([A-Z_][A-Z0-9_]*)\s*=\s*([`'\"])([^`'\"]*)\2")
RX_ENV = [re.compile(p) for p in (
    r"process\.env\.([A-Z][A-Z0-9_]+)", r"process\.env\[['\"]([A-Z][A-Z0-9_]+)",
    r"import\.meta\.env\.([A-Z][A-Z0-9_]+)", r"os\.environ(?:\.get)?\(?\[?\s*['\"]([A-Z][A-Z0-9_]+)",
    r"os\.getenv\(\s*['\"]([A-Z][A-Z0-9_]+)", r"os\.Getenv\(\s*\"([A-Z][A-Z0-9_]+)",
    r"getenv\(\s*['\"]([A-Z][A-Z0-9_]+)", r"System\.getenv\(\s*\"([A-Z][A-Z0-9_]+)")]
# ${VAR} só fora de JS/TS (lá é template literal): properties, yaml, shell, hcl
RX_ENV_PLACEHOLDER = re.compile(r"\$\{([A-Z][A-Z0-9_]+)(?::[^}]*)?\}")
ENV_EMBUTIDAS = {"DEV", "PROD", "MODE", "SSR", "BASE_URL"}
RX_LIB_FILA = re.compile(r"amqp|rabbit|bullmq|\bbull\b|kafka|nats|sqs|pubsub|@nestjs/microservices|redis", re.I)
RX_FILA_IN = [re.compile(p) for p in (
    r"\.consume\(\s*['\"`]([^'\"`]+)", r"@RabbitListener\(\s*queues\s*=\s*\{?\s*\"([^\"]+)",
    r"new\s+Worker\(\s*['\"`]([^'\"`]+)", r"@Process\(\s*['\"`]([^'\"`]+)", r"\.subscribe\(\s*['\"`]([^'\"`]+)")]
RX_FILA_OUT = [re.compile(p) for p in (
    r"\.sendToQueue\(\s*['\"`]([^'\"`]+)", r"\.publish\(\s*['\"`]([^'\"`]+)",
    r"new\s+Queue\(\s*['\"`]([^'\"`]+)", r"convertAndSend\(\s*\"([^\"]+)")]
RX_CRON = [re.compile(p) for p in (
    r"cron\.schedule\(\s*['\"`]([^'\"`]+)", r"@Cron\(\s*['\"`]([^'\"`]+)",
    r"@Scheduled\(\s*cron\s*=\s*\"([^\"]+)", r"CronJob\(\s*['\"`]([^'\"`]+)", r"\bcron\s*=\s*\"([^\"]+)\"")]
RX_SOCKET = re.compile(r"\.(?:on|emit)\(\s*['\"`]([a-zA-Z][\w:.\-/]+)['\"`]")
RX_IMPORT_JS = re.compile(r"""(?:import\s[^'"`;]*?from\s*|import\s*\(\s*|require\(\s*|import\s+)['"]([^'"]+)['"]""")
RX_IMPORT_PY = re.compile(r"^\s*(?:from\s+([\w.]+)\s+import|import\s+([\w.]+))", re.M)
RX_ROTA_PATH = re.compile(r"\bpath\s*:\s*['\"`]([^'\"`]*)['\"`]")
RX_ROTA_COMP = re.compile(r"\bcomponent\s*:\s*(?:\(\s*\)\s*=>\s*import\(\s*['\"]([^'\"]+)['\"]|([A-Za-z_$][\w$]*))")
RX_ROTA_NOME = re.compile(r"\bname\s*:\s*['\"`]([^'\"`]+)['\"`]")
RX_ROTA_TITULO = re.compile(r"\btitle\s*:\s*['\"`]([^'\"`]+)['\"`]")
RX_REACT_ROUTE = re.compile(r"<Route\b[^>]*\bpath=['\"{`]+([^'\"}`]+)")


# ------------------------------------------------------------------ util
RX_ARQ_TESTE = re.compile(r"(^|/)(tests?|__tests__|spec|testdata)/|\.(test|spec)\.[a-z]+$|(^|/)test_[^/]*$|_test\.[a-z]+$")


def eh_migration(rel: str, sufixo: str) -> bool:
    """Arquivo de migration: "migrat" no caminho (migrate, migration, migrations) e código ou .sql.
    Mesma regra do Sentinel (fixo Migrations) e do verifica-projeto.py."""
    return "migrat" in rel.lower() and sufixo in EXT_CODIGO | {".sql"}


def pastas_migrations(raiz: Path, limite: int = 20) -> list:
    """Pastas com migrations, relativas à raiz (ignora DOC, dependências e build)."""
    achadas = []
    for f in arquivos(raiz):
        r = f.relative_to(raiz).as_posix()
        if eh_migration(r, f.suffix):
            d = f.parent.relative_to(raiz).as_posix()
            if d not in achadas:
                achadas.append(d)
                if len(achadas) >= limite:
                    break
    return achadas


def arquivos(raiz: Path):
    for dirpath, dirnames, filenames in os.walk(raiz):
        dirnames[:] = sorted(d for d in dirnames if d not in IGNORAR_DIRS and not d.startswith(".")
                             or d in (".github",))
        for f in sorted(filenames):
            yield Path(dirpath) / f


def ler(f: Path):
    try:
        if f.stat().st_size > LIMITE_ARQUIVO:
            return None
        return f.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return None


def rel(raiz, f):
    return f.relative_to(raiz).as_posix()


def normaliza_url(u: str, consts: dict):
    # só substitui constante cujo valor é caminho/URL; outro valor (token, chave) fica como {NOME}
    def const(m):
        v = consts.get(m.group(1))
        return v if isinstance(v, str) and v.startswith(("/", "http")) else "{" + m.group(1) + "}"
    u = re.sub(r"\$\{\s*([A-Z_][A-Z0-9_]*)\s*\}", const, u)
    u = re.sub(r"\$\{[^}]*\}", "{param}", u)
    u = re.sub(r"\s+", "", u)
    return u


def linha(texto, pos):
    return texto.count("\n", 0, pos) + 1


# ---------------------------------------------------------------- aliases
def aliases(raiz: Path):
    """Aliases de import (@ → src, @views → src/views). tsconfig/jsconfig primeiro
    (respeitando baseUrl); o bundler (vite/webpack) prevalece por ser o que roda."""
    out = {"@": "src", "~": "src"}
    for nome in ("tsconfig.json", "jsconfig.json"):
        t = ler(raiz / nome) or ""
        b = re.search(r"['\"]baseUrl['\"]\s*:\s*['\"]([^'\"]*)['\"]", t)
        base_url = (b.group(1) if b else ".").strip("/")
        base_url = "" if base_url in (".", "./") else base_url.lstrip("./")
        for m in re.finditer(r"['\"]([@~][\w-]*)/\*['\"]\s*:\s*\[\s*['\"](?:\./)?([^'\"*]*?)/?\*", t):
            destino = "/".join(x for x in (base_url, m.group(2).strip("/")) if x)
            out[m.group(1)] = destino
    for nome in ("vite.config.mjs", "vite.config.js", "vite.config.ts", "vue.config.js", "webpack.config.js"):
        t = ler(raiz / nome) or ""
        for m in re.finditer(r"find\s*:\s*['\"]([^'\"]+)['\"]\s*,\s*replacement\s*:\s*path\.resolve\([^,]+,\s*['\"]([^'\"]+)['\"]", t):
            out[m.group(1)] = re.sub(r"^\./", "", m.group(2)).strip("/")
        for m in re.finditer(r"['\"](@[\w-]*|~)['\"]\s*:\s*path\.resolve\([^,]+,\s*['\"]([^'\"]+)['\"]", t):
            out[m.group(1)] = re.sub(r"^\./", "", m.group(2)).strip("/")
    return out


def resolve_arquivo(raiz: Path, alvo: str):
    """Caminho relativo real de um import sem extensão (arquivo ou index da pasta)."""
    p = raiz / alvo
    if p.is_file():
        return alvo
    for ext in (".ts", ".js", ".vue", ".tsx", ".jsx", ".mjs", ".py", ".svelte"):
        if (raiz / (alvo + ext)).is_file():
            return alvo + ext
    if p.is_dir():
        for ext in (".ts", ".js", ".tsx", ".jsx", ".vue", ".py"):
            for nome in ("index", "__init__"):
                if (p / (nome + ext)).is_file():
                    return f"{alvo}/{nome}{ext}"
        return alvo + "/"
    return None


# ------------------------------------------------------------------ stack
def stack(raiz: Path):
    out = []
    for f in arquivos(raiz):
        if f.name not in MANIFESTOS:
            continue
        t = ler(f) or ""
        achados = []
        deps = 0
        if f.name == "package.json":
            try:
                pj = json.loads(t)
                todas = {**pj.get("dependencies", {}), **pj.get("devDependencies", {})}
                deps = len(todas)
                for fw in FRAMEWORKS:
                    if fw in todas:
                        achados.append(f"{fw}@{todas[fw]}")
            except json.JSONDecodeError:
                pass
        else:
            achados = [fw for fw in FRAMEWORKS if re.search(rf"[\"'/ ]{re.escape(fw)}[\"'/ @=<>~\n]", t)]
        out.append({"manifesto": rel(raiz, f), "linguagem": MANIFESTOS[f.name],
                    "frameworks": achados, "dependencias": deps})
    return out


# ---------------------------------------------------------------- módulos
def modulo_de(caminho: str, base: str):
    partes = caminho.split("/")
    if base and caminho.startswith(base + "/"):
        partes = partes[len(base.split("/")):]
        return f"{base}/{partes[0]}" if len(partes) > 1 else base
    return partes[0] if len(partes) > 1 else "(raiz)"


def resolve_import(esp: str, origem: str, als: dict):
    if esp.startswith("."):
        base = os.path.normpath(os.path.join(os.path.dirname(origem), esp))
        return base.replace(os.sep, "/")
    for a in sorted(als, key=len, reverse=True):
        if esp == a or esp.startswith(a + "/"):
            return (als[a] + esp[len(a):]).lstrip("/")
    return None  # pacote externo


def contrato_openapi(texto: str, nome: str):
    """(MÉTODO, caminho) de um contrato OpenAPI/Swagger, sem depender de biblioteca YAML."""
    out = []
    if nome.lower().endswith(".json"):
        try:
            for caminho, ops in (json.loads(texto).get("paths") or {}).items():
                out += [(m.upper(), caminho) for m in ops if m.lower() in ("get", "post", "put", "patch", "delete")]
        except (json.JSONDecodeError, AttributeError):
            pass
        return out
    em_paths, caminho, recuo_path = False, None, None
    srv = re.search(r"^servers:\s*\n\s*-\s*url:\s*['\"]?([^'\"\s]+)", texto, re.M)
    prefixo = re.sub(r"^\w+://[^/]+", "", srv.group(1)).rstrip("/") if srv else ""
    prefixo = "" if "{" in prefixo else prefixo
    for ln in texto.splitlines():
        if re.match(r"^paths:\s*$", ln):
            em_paths = True
            continue
        if em_paths and re.match(r"^\S", ln):
            break
        if not em_paths:
            continue
        m = re.match(r"^(\s+)['\"]?(/[^'\":]*)['\"]?\s*:\s*$", ln)
        if m and (recuo_path is None or len(m.group(1)) == recuo_path):
            recuo_path, caminho = len(m.group(1)), m.group(2)
            continue
        m = re.match(r"^(\s+)(get|post|put|patch|delete)\s*:", ln)
        if m and caminho and recuo_path is not None and len(m.group(1)) > recuo_path:
            out.append((m.group(2).upper(), prefixo + caminho))
    return out


# ----------------------------------------------------------------- main
def inventariar(raiz: Path):
    raiz = raiz.resolve()
    als = aliases(raiz)
    base = "src" if (raiz / "src").is_dir() else ""
    inv = {"gerado_em": datetime.datetime.now().isoformat(timespec="seconds"), "raiz": raiz.name,
           "stack": stack(raiz), "modulos": {}, "rotas": [], "endpoints": [], "chamadas_http": [],
           "sockets": [], "filas_consumidas": [], "filas_publicadas": [], "crons": [], "env": {},
           "env_declaradas": {}, "dados": {"migrations": [], "schemas": [], "sql": 0},
           "infra": {"docker": [], "compose": [], "nomad": [], "ci": []}}
    imports = {}

    for f in arquivos(raiz):
        r = rel(raiz, f)
        nome = f.name
        # infra / dados / env declaradas (independem de extensão de código)
        if nome.startswith("Dockerfile") or nome == "Containerfile":
            inv["infra"]["docker"].append(r)
        elif re.match(r"(docker-)?compose.*\.ya?ml$", nome):
            inv["infra"]["compose"].append(r)
        elif r.startswith(".github/workflows/") or nome in (".gitlab-ci.yml", "Jenkinsfile"):
            inv["infra"]["ci"].append(r)
        elif f.suffix in (".hcl", ".nomad") and re.search(r'^\s*job\s+"', ler(f) or "", re.M):
            inv["infra"]["nomad"].append(r)
        if eh_migration(r, f.suffix):
            d = f.parent.relative_to(raiz).as_posix()
            if d not in inv["dados"]["migrations"]:
                inv["dados"]["migrations"].append(d)
        if nome in ("schema.prisma",) or r.endswith(("models.py", "entities.ts")):
            inv["dados"]["schemas"].append(r)
        if f.suffix == ".sql":
            inv["dados"]["sql"] += 1
        if nome.startswith(".env") and not nome.endswith((".bkp", ".swp")) and "bkp" not in nome:
            t = ler(f) or ""
            chaves = sorted(set(re.findall(r"^\s*(?:export\s+)?([A-Z][A-Z0-9_]*)\s*=", t, re.M)))
            if chaves:
                inv["env_declaradas"][r] = chaves  # só NOMES, nunca valores

        if f.suffix in (".properties", ".yml", ".yaml", ".hcl", ".tf", ".sh", ".toml") and not r.startswith(".github/"):
            for m in RX_ENV_PLACEHOLDER.finditer(ler(f) or ""):
                inv["env"].setdefault(m.group(1), set()).add(r)
        if re.match(r"(openapi|swagger)[\w.-]*\.(ya?ml|json)$", nome, re.I):
            inv.setdefault("contratos", []).append(r)
            for met, caminho in contrato_openapi(ler(f) or "", nome):
                inv["endpoints"].append({"metodo": met, "path": caminho, "arquivo": r, "linha": 0, "tipo": "openapi"})
        if f.suffix not in EXT_CODIGO:
            continue
        t = ler(f)
        if t is None:
            continue
        mod = modulo_de(r, base)
        info = inv["modulos"].setdefault(mod, {"arquivos": 0, "importa": set()})
        info["arquivos"] += 1

        # imports -> grafo de módulos
        alvos = []
        if f.suffix in EXT_JS:
            for m in RX_IMPORT_JS.finditer(t):
                res = resolve_import(m.group(1), r, als)
                if res:
                    alvos.append(res)
        elif f.suffix == ".py":
            for m in RX_IMPORT_PY.finditer(t):
                alvos.append((m.group(1) or m.group(2)).replace(".", "/"))
        imports[r] = alvos

        # arquivo de teste entra no grafo de módulos, mas não nos fatos: rotas,
        # endpoints, chamadas, filas, crons e variáveis de um teste são fixtures
        if RX_ARQ_TESTE.search(r):
            continue

        consts = {m.group(1): m.group(3) for m in RX_CONST_STR.finditer(t)}

        # rotas de front (arquivos de rotas)
        if f.suffix in EXT_JS and re.search(r"(^|/)(router|routes?)(/|\.|$)|\.routes?\.", r, re.I):
            paths = list(RX_ROTA_PATH.finditer(t))
            for i, m in enumerate(paths):
                fim = paths[i + 1].start() if i + 1 < len(paths) else len(t)
                trecho = t[m.end():fim]
                c = RX_ROTA_COMP.search(trecho)
                n = RX_ROTA_NOME.search(trecho[:300]) or RX_ROTA_NOME.search(t[max(0, m.start() - 200):m.start()])
                ti = RX_ROTA_TITULO.search(trecho[:400])
                comp = None
                if c:
                    comp = c.group(1) or c.group(2)
                    if c.group(1):
                        comp = resolve_import(c.group(1), r, als) or c.group(1)
                inv["rotas"].append({"path": m.group(1), "nome": n.group(1) if n else None,
                                     "titulo": ti.group(1) if ti else None, "componente": comp,
                                     "arquivo": r, "linha": linha(t, m.start())})
            for m in RX_REACT_ROUTE.finditer(t):
                inv["rotas"].append({"path": m.group(1), "nome": None, "titulo": None, "componente": None,
                                     "arquivo": r, "linha": linha(t, m.start())})
        # rotas por arquivo (Next/Nuxt)
        mp = re.match(r"^(?:src/)?(pages|app)/(.+)\.(vue|tsx|jsx|ts|js)$", r)
        # em app/ (Next app router) só page.* é rota; o resto (componente, service do Angular) não
        if mp and mp.group(1) == "app" and not re.search(r"(^|/)page$", mp.group(2)):
            mp = None
        if mp and not re.search(r"(^|/)_|/api/|layout\.|loading\.|error\.", mp.group(2)):
            rota = "/" + re.sub(r"(^|/)(index|page)$", "", mp.group(2))
            rota = re.sub(r"\[\.\.\.(\w+)\]", r"*\1", re.sub(r"\[(\w+)\]", r"{\1}", rota))
            inv["rotas"].append({"path": rota or "/", "nome": None, "titulo": None, "componente": r,
                                 "arquivo": r, "linha": 1})

        # endpoints de back
        for tipo, rx in RX_ENDPOINT:
            if tipo == "go" and (f.suffix != ".go" or f.name.endswith("_test.go")):
                continue  # em JS/TS, core.GET("/x") é chamada de saída (openapi-fetch), não rota
            if tipo == "express" and f.suffix == ".go":
                continue
            for m in rx.finditer(t):
                met = m.group(1).upper()
                met = "ANY" if met in ("ROUTE", "HANDLEFUNC", "HANDLE") else met
                inv["endpoints"].append({"metodo": met, "path": m.group(2), "arquivo": r,
                                         "linha": linha(t, m.start()), "tipo": tipo})
        if f.suffix == ".go" and not f.name.endswith("_test.go"):
            for m in RX_GO_SAIDA.finditer(t):
                inv["chamadas_http"].append({"metodo": (m.group(1) or m.group(2)).upper(), "url": m.group(3),
                                             "arquivo": r, "linha": linha(t, m.start())})
            for rx in (RX_GO_PADRAO, RX_GO_TABELA):
                for m in rx.finditer(t):
                    inv["endpoints"].append({"metodo": m.group(1).upper(), "path": m.group(2), "arquivo": r,
                                             "linha": linha(t, m.start()), "tipo": "go"})
        ctrl = RX_NEST_CTRL.search(t)
        if ctrl:
            pref = "/" + ctrl.group(1).strip("/")
            for m in RX_NEST.finditer(t):
                p = (pref.rstrip("/") + "/" + m.group(2).strip("/")).rstrip("/") or "/"
                inv["endpoints"].append({"metodo": m.group(1).upper(), "path": p, "arquivo": r,
                                         "linha": linha(t, m.start()), "tipo": "nest"})
        if "Mapping(" in t:
            cls = RX_SPRING_CLS.search(t)
            pref = "/" + cls.group(1).strip("/") if cls else ""
            for m in RX_SPRING.finditer(t):
                p = (pref.rstrip("/") + "/" + m.group(2).strip("/")).rstrip("/") or "/"
                inv["endpoints"].append({"metodo": m.group(1).upper(), "path": p, "arquivo": r,
                                         "linha": linha(t, m.start()), "tipo": "spring"})

        # saídas
        if not re.search(r"(^|/)(tests?|__tests__|spec)/|\.(test|spec)\.", r):
            vistos = set()
            for rx, exige_url in ((RX_HTTP, False), (RX_HTTP_GENERICO, True)):
                for m in rx.finditer(t):
                    url = normaliza_url(m.group(3), consts)
                    if not url or url.startswith(("#", ".")) or len(url) >= 200 or m.start() in vistos:
                        continue
                    if exige_url and not url.startswith(("/", "http", "{")):
                        continue
                    vistos.add(m.start())
                    inv["chamadas_http"].append({"metodo": m.group(1).upper(), "url": url, "arquivo": r,
                                                 "linha": linha(t, m.start())})
            if f.suffix in EXT_JS:
                for m in RX_HTTP_MAIUSC.finditer(t):
                    inv["chamadas_http"].append({"metodo": m.group(1), "url": normaliza_url(m.group(3), consts),
                                                 "arquivo": r, "linha": linha(t, m.start())})
            for m in RX_HTTP_CONST.finditer(t):
                valor = consts.get(m.group(2))
                if valor is not None:
                    url = normaliza_url(valor, consts)
                    inv["chamadas_http"].append({"metodo": m.group(1).upper(), "url": url or "/", "arquivo": r,
                                                 "linha": linha(t, m.start())})
            for m in RX_FETCH.finditer(t):
                url = normaliza_url(m.group(2), consts)
                if url and len(url) < 200:
                    inv["chamadas_http"].append({"metodo": "FETCH", "url": url, "arquivo": r,
                                                 "linha": linha(t, m.start())})
            if "socket" in t.lower():
                for m in RX_SOCKET.finditer(t):
                    ev = m.group(1)
                    if ev not in ("click", "change", "input", "error", "load", "submit", "close", "open"):
                        inv["sockets"].append({"evento": ev, "arquivo": r})
        if RX_LIB_FILA.search(t):  # evita confundir event bus de UI com fila
            for rx in RX_FILA_IN:
                for m in rx.finditer(t):
                    inv["filas_consumidas"].append({"fila": m.group(1), "arquivo": r})
            for rx in RX_FILA_OUT:
                for m in rx.finditer(t):
                    inv["filas_publicadas"].append({"fila": m.group(1), "arquivo": r})
        for rx in RX_CRON:
            for m in rx.finditer(t):
                if re.match(r"^[\d*/,\- ?LW#@a-z]+$", m.group(1)):
                    inv["crons"].append({"expressao": m.group(1), "arquivo": r})
        for rx in RX_ENV + ([] if f.suffix in EXT_JS else [RX_ENV_PLACEHOLDER]):
            for m in rx.finditer(t):
                if m.group(1) not in ENV_EMBUTIDAS:
                    inv["env"].setdefault(m.group(1), set()).add(r)

    # grafo entre módulos
    for origem, alvos in imports.items():
        mo = modulo_de(origem, base)
        for a in alvos:
            real = resolve_arquivo(raiz, a)
            if not real:
                continue
            md = modulo_de(real.rstrip("/") + ("/x" if real.endswith("/") else ""), base)
            if md != mo and md in inv["modulos"]:
                inv["modulos"][mo]["importa"].add(md)
    for m in inv["modulos"].values():
        m["importa"] = sorted(m["importa"])
    inv["env"] = {k: sorted(v) for k, v in sorted(inv["env"].items())}
    # deduplica listas
    for chave, campos in (("chamadas_http", ("metodo", "url", "arquivo")), ("sockets", ("evento", "arquivo")),
                          ("endpoints", ("metodo", "path", "arquivo"))):
        vistos, uniq = set(), []
        for it in inv[chave]:
            k = tuple(it[c] for c in campos)
            if k not in vistos:
                vistos.add(k)
                uniq.append(it)
        inv[chave] = uniq
    return inv


def chaves(inv):
    """Conjunto comparável para o --check."""
    out = set()
    out |= {f"modulo {m}" for m in inv.get("modulos", {})}
    out |= {f"rota {r['path']} ({r['arquivo']})" for r in inv.get("rotas", [])}
    out |= {f"endpoint {e['metodo']} {e['path']}" for e in inv.get("endpoints", [])}
    out |= {f"http {c['metodo']} {c['url']}" for c in inv.get("chamadas_http", [])}
    out |= {f"env {e}" for e in inv.get("env", {})}
    out |= {f"fila-in {q['fila']}" for q in inv.get("filas_consumidas", [])}
    out |= {f"fila-out {q['fila']}" for q in inv.get("filas_publicadas", [])}
    out |= {f"cron {c['expressao']} ({c['arquivo']})" for c in inv.get("crons", [])}
    return out


# ------------------------------------------------------------- markdown
def id_mermaid(nome):
    return re.sub(r"\W", "_", nome) or "raiz"


def markdown(inv):
    L = [f"# Inventário do código — {inv['raiz']}", "",
         f"> GERADO por `tools/mapeia_projeto.py` em {inv['gerado_em']}. Não edite: rode o script de novo.",
         "> A interpretação (o que cada área faz, fluxos, RN candidatas) está nas outras páginas desta pasta.", ""]
    L += ["## Stack", "", "| Manifesto | Linguagem | Frameworks | Deps |", "|---|---|---|---|"]
    for s in inv["stack"]:
        L.append(f"| `{s['manifesto']}` | {s['linguagem']} | {', '.join(s['frameworks']) or '—'} | {s['dependencias'] or '—'} |")
    L += ["", "## Módulos", "", "```mermaid", "graph LR"]
    arestas = 0
    for nome, m in sorted(inv["modulos"].items()):
        L.append(f'  {id_mermaid(nome)}["{nome} ({m["arquivos"]})"]')
    for nome, m in sorted(inv["modulos"].items()):
        for alvo in m["importa"]:
            if arestas < 80:
                L.append(f"  {id_mermaid(nome)} --> {id_mermaid(alvo)}")
                arestas += 1
    L += ["```", ""]
    L += ["| Módulo | Arquivos | Importa |", "|---|---|---|"]
    for nome, m in sorted(inv["modulos"].items(), key=lambda x: -x[1]["arquivos"]):
        L.append(f"| `{nome}` | {m['arquivos']} | {', '.join(m['importa']) or '—'} |")

    if inv["rotas"]:
        L += ["", f"## Rotas de front ({len(inv['rotas'])})", "", "| Rota | Nome | Título | Componente | Onde |", "|---|---|---|---|---|"]
        for r in inv["rotas"]:
            L.append(f"| `{r['path']}` | {r['nome'] or '—'} | {r['titulo'] or '—'} | {('`' + r['componente'] + '`') if r['componente'] else '—'} | `{r['arquivo']}:{r['linha']}` |")
    if inv["endpoints"]:
        L += ["", f"## Endpoints ({len(inv['endpoints'])})", "", "| Método | Caminho | Onde |", "|---|---|---|"]
        for e in sorted(inv["endpoints"], key=lambda e: (e["path"], e["metodo"])):
            onde = f"`{e['arquivo']}`" + (" (contrato)" if e.get("tipo") == "openapi" else f":{e['linha']}")
            L.append(f"| {e['metodo']} | `{e['path']}` | {onde} |")
    if inv["chamadas_http"]:
        L += ["", f"## Chamadas HTTP de saída ({len(inv['chamadas_http'])})", "", "| Método | URL | Onde |", "|---|---|---|"]
        for c in sorted(inv["chamadas_http"], key=lambda c: (c["url"], c["metodo"])):
            L.append(f"| {c['metodo']} | `{c['url']}` | `{c['arquivo']}:{c['linha']}` |")
    for chave, titulo, campo in (("filas_consumidas", "Filas consumidas", "fila"),
                                 ("filas_publicadas", "Filas publicadas", "fila"),
                                 ("crons", "Crons", "expressao"), ("sockets", "Eventos de socket", "evento")):
        if inv[chave]:
            L += ["", f"## {titulo} ({len(inv[chave])})", "", "| Nome | Onde |", "|---|---|"]
            for it in inv[chave]:
                L.append(f"| `{it[campo]}` | `{it['arquivo']}` |")
    if inv["env"] or inv["env_declaradas"]:
        declaradas = {k for ks in inv["env_declaradas"].values() for k in ks}
        L += ["", f"## Variáveis de ambiente lidas no código ({len(inv['env'])})", "",
              "| Variável | Declarada em .env* | Onde é lida |", "|---|---|---|"]
        for k, onde in inv["env"].items():
            L.append(f"| `{k}` | {'sim' if k in declaradas else '**não**'} | {', '.join('`' + o + '`' for o in onde[:3])}{' …' if len(onde) > 3 else ''} |")
        if inv["env_declaradas"]:
            L += ["", "Arquivos `.env*` (só nomes de chave): " + ", ".join(f"`{a}` ({len(k)})" for a, k in inv["env_declaradas"].items())]
    d = inv["dados"]
    if d["migrations"] or d["schemas"] or d["sql"]:
        L += ["", "## Dados", "", f"- Migrations: {', '.join('`' + m + '`' for m in d['migrations']) or '—'}",
              f"- Schemas/modelos: {', '.join('`' + s + '`' for s in d['schemas']) or '—'}", f"- Arquivos .sql: {d['sql']}"]
    i = inv["infra"]
    L += ["", "## Infra", ""] + [f"- {k}: {', '.join('`' + x + '`' for x in v) or '—'}" for k, v in i.items()]
    return "\n".join(L) + "\n"


def _log(raiz, script, rc, resumo, problemas):
    try:
        import registro_log
        registro_log.registrar(raiz, script, rc, resumo=resumo, problemas=problemas)
    except ImportError:
        pass


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("raiz")
    ap.add_argument("--saida", default=None, help="pasta do mapa (default: <DOC>/System/mapa)")
    ap.add_argument("--check", action="store_true")
    ap.add_argument("--json", action="store_true", help="com --check: resultado em JSON (para outros scripts)")
    args = ap.parse_args()

    raiz = Path(args.raiz).resolve()
    if not raiz.is_dir():
        print(f"ERRO: diretório não encontrado: {raiz}")
        return 2
    if args.saida:
        saida = Path(args.saida)
        saida = saida if saida.is_absolute() else raiz / saida
    else:
        docs = [raiz / "DOC"] if (raiz / "DOC" / "INDEX.md").is_file() else \
            sorted(p.parent for p in (raiz / "DOC").glob("*/INDEX.md")) if (raiz / "DOC").is_dir() else []
        if len(docs) > 1:
            print("ERRO: monorepo com várias DOC/ — informe --saida DOC/<projeto>/System/mapa e aponte a raiz do projeto.")
            return 2
        saida = (docs[0] if docs else raiz / "DOC") / "System" / "mapa"

    inv = inventariar(raiz)
    arq_json = saida / "inventario.json"

    if args.check:
        if not arq_json.is_file():
            if args.json:
                print(json.dumps({"rc": 2, "sem_inventario": True, "surgiu": [], "sumiu": []}))
                return 2
            print(f"Sem inventário salvo em {arq_json.relative_to(raiz) if arq_json.is_relative_to(raiz) else arq_json} — "
                  "gere o mapa primeiro (sem --check).")
            return 2
        salvo = json.loads(arq_json.read_text(encoding="utf-8"))
        novos = sorted(chaves(inv) - chaves(salvo))
        sumiram = sorted(chaves(salvo) - chaves(inv))
        if args.json:
            print(json.dumps({"rc": 1 if (novos or sumiram) else 0, "gerado_em": salvo.get("gerado_em"),
                              "surgiu": novos, "sumiu": sumiram}, ensure_ascii=False))
        else:
            print(f"=== Mapa × código ({raiz.name}) — inventário salvo em {salvo.get('gerado_em')} ===")
            for titulo, lista in (("SURGIU no código e não está no mapa", novos), ("SUMIU do código e ainda está no mapa", sumiram)):
                if lista:
                    print(f"\n[{titulo}] {len(lista)}")
                    for x in lista[:40]:
                        print(f"  - {x}")
                    if len(lista) > 40:
                        print(f"  … +{len(lista) - 40}")
        _log(raiz, "mapeia_projeto-check", 1 if (novos or sumiram) else 0, {"surgiu": len(novos), "sumiu": len(sumiram)},
             [{"nivel": "ALERTA", "etapa": "mapa", "msg": f"surgiu no código: {x}", "comando": "python3 <DOC>/tools/mapeia_projeto.py ."}
              for x in novos[:30]] + [{"nivel": "ALERTA", "etapa": "mapa", "msg": f"sumiu do código: {x}"} for x in sumiram[:30]])
        if args.json:
            return 1 if (novos or sumiram) else 0
        if novos or sumiram:
            print("\nResultado: mapa desatualizado — rode mapeia_projeto.py sem --check e revise as páginas afetadas.")
            return 1
        print("\nResultado: mapa em dia com o código.")
        return 0

    saida.mkdir(parents=True, exist_ok=True)
    arq_json.write_text(json.dumps(inv, ensure_ascii=False, indent=1, default=sorted) + "\n", encoding="utf-8")
    (saida / "inventario.md").write_text(markdown(inv), encoding="utf-8")
    _log(raiz, "mapeia_projeto", 0, {"modulos": len(inv["modulos"]), "rotas": len(inv["rotas"]),
                                      "endpoints": len(inv["endpoints"]), "chamadas_http": len(inv["chamadas_http"]),
                                      "env": len(inv["env"])}, [])
    print(f"Inventário de {raiz.name} → {saida}")
    print(f"  módulos: {len(inv['modulos'])} · rotas: {len(inv['rotas'])} · endpoints: {len(inv['endpoints'])} · "
          f"chamadas HTTP: {len(inv['chamadas_http'])} · env: {len(inv['env'])} · filas: "
          f"{len(inv['filas_consumidas']) + len(inv['filas_publicadas'])} · crons: {len(inv['crons'])}")
    print("PRÓXIMO: a IA escreve as páginas interpretadas (visao-geral.md e uma por área) a partir do inventário — "
          "procedimento em references/mapa-projeto.md da skill novo-projeto-sdd.")
    return 0


if __name__ == "__main__":
    try:
        import registro_log
        sys.exit(registro_log.executar("mapeia_projeto", main, registro_log.raiz_do_argv))
    except ImportError:
        sys.exit(main())
