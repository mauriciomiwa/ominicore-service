#!/usr/bin/env python3
"""Gera a vista HTML do mapa do projeto a partir de <DOC>/System/mapa/*.md (a fonte da verdade).

Uso:
  python3 mapa_html.py <DOC>/System/mapa [saida.html]      # default: <mapa>/mapa.html

Lê visao-geral.md, inventario.json/.md e as páginas de área. A ordem e o grupo
das áreas vêm da tabela "Áreas" da visao-geral.md (links para <area>.md); área
sem grupo declarado cai em "Áreas". Só biblioteca padrão: usa o pacote
`markdown` se estiver instalado e, senão, um renderizador embutido que cobre o
que as páginas do mapa usam (títulos, listas, tabelas, código, citações). O HTML sai pronto para publicar como artifact ou abrir
no navegador; não contém valores de .env (o inventário só guarda nomes).
"""
import html
import json
import re
import sys
import unicodedata
from pathlib import Path

RE_HREF_SEGURO = re.compile(r"^(#|/|\./|\.\./|[\w-]+\.md|https?://|mailto:)")


def href_seguro(url: str) -> str:
    """Só âncora, caminho relativo, http(s) e mailto; o resto (javascript:, data:, aspas) vira '#'."""
    return url if RE_HREF_SEGURO.match(url) and '"' not in url else "#"


def sanitiza_hrefs(h: str) -> str:
    return re.sub(r'href="([^"]*)"',
                  lambda m: f'href="{html.escape(href_seguro(html.unescape(m.group(1))), quote=True)}"', h)

try:  # usa o pacote markdown se existir; senão, o renderizador embutido abaixo
    import markdown as _markdown
except ImportError:
    _markdown = None


def _inline(t: str) -> str:
    """Markdown inline: `código`, **negrito**, *itálico*, [texto](link)."""
    codigos = []

    def guarda(m):
        codigos.append(f"<code>{html.escape(m.group(1))}</code>")
        return f"\x00{len(codigos) - 1}\x00"

    t = re.sub(r"`([^`]+)`", guarda, t)
    t = html.escape(t, quote=False)
    t = re.sub(r"\*\*(.+?)\*\*", r"<strong>\1</strong>", t)
    t = re.sub(r"(?<![\w*])\*(?!\s)(.+?)(?<!\s)\*(?![\w*])", r"<em>\1</em>", t)
    t = re.sub(r"\[([^\]]+)\]\(([^)\s]+)\)",
               lambda m: f'<a href="{html.escape(href_seguro(html.unescape(m.group(2))), quote=True)}">{m.group(1)}</a>', t)
    return re.sub(r"\x00(\d+)\x00", lambda m: codigos[int(m.group(1))], t)


def _celulas(linha: str):
    linha = linha.strip()
    if linha.startswith("|"):
        linha = linha[1:]
    if linha.endswith("|") and not linha.endswith("\\|"):
        linha = linha[:-1]
    cel, atual, em_code = [], "", False
    for ch in linha:
        if ch == "`":
            em_code = not em_code
        if ch == "|" and not em_code:
            cel.append(atual.strip())
            atual = ""
        else:
            atual += ch
    cel.append(atual.strip())
    return cel


def _md_basico(texto: str) -> str:
    """Subconjunto de Markdown usado nas páginas do mapa: títulos, parágrafos,
    listas (com aninhamento simples), tabelas, blocos de código, citações."""
    linhas = texto.splitlines()
    out, i, par = [], 0, []

    def fecha_par():
        if par:
            out.append("<p>" + _inline(" ".join(x.strip() for x in par)) + "</p>")
            par.clear()

    while i < len(linhas):
        ln = linhas[i]
        cerca = re.match(r"^\s*(```|~~~)\s*(\w*)", ln)
        if cerca:
            fecha_par()
            marca, lang = cerca.group(1), cerca.group(2)
            corpo = []
            i += 1
            while i < len(linhas) and not linhas[i].strip().startswith(marca):
                corpo.append(linhas[i])
                i += 1
            classe = f' class="language-{lang}"' if lang else ""
            out.append(f"<pre><code{classe}>{html.escape(chr(10).join(corpo))}</code></pre>")
            i += 1
            continue
        h = re.match(r"^(#{1,6})\s+(.*)$", ln)
        if h:
            fecha_par()
            n = len(h.group(1))
            out.append(f"<h{n}>{_inline(h.group(2).strip())}</h{n}>")
            i += 1
            continue
        if ln.strip().startswith("|") and i + 1 < len(linhas) and re.match(r"^\s*\|?[\s:|-]+\|?\s*$", linhas[i + 1]) \
                and "-" in linhas[i + 1]:
            fecha_par()
            cab = _celulas(ln)
            i += 2
            corpo = []
            while i < len(linhas) and linhas[i].strip().startswith("|"):
                corpo.append(_celulas(linhas[i]))
                i += 1
            t = ["<table><thead><tr>" + "".join(f"<th>{_inline(c)}</th>" for c in cab) + "</tr></thead><tbody>"]
            for r in corpo:
                t.append("<tr>" + "".join(f"<td>{_inline(c)}</td>" for c in r) + "</tr>")
            out.append("".join(t) + "</tbody></table>")
            continue
        lst = re.match(r"^(\s*)([-*+]|\d+[.)])\s+(.*)$", ln)
        if lst:
            fecha_par()
            ordenada = lst.group(2)[0].isdigit()
            tag = "ol" if ordenada else "ul"
            itens = []
            while i < len(linhas):
                m = re.match(r"^(\s*)([-*+]|\d+[.)])\s+(.*)$", linhas[i])
                if m:
                    itens.append([len(m.group(1)), m.group(3)])
                    i += 1
                elif linhas[i].strip() and linhas[i].startswith((" ", "\t")) and itens:
                    itens[-1][1] += " " + linhas[i].strip()
                    i += 1
                else:
                    break
            base = itens[0][0]
            html_l, aberto_sub = [f"<{tag}>"], False
            for nivel, txt in itens:
                if nivel > base and not aberto_sub:
                    html_l[-1] = html_l[-1].replace("</li>", "") + "<ul>"
                    aberto_sub = True
                elif nivel <= base and aberto_sub:
                    html_l.append("</ul></li>")
                    aberto_sub = False
                html_l.append(f"<li>{_inline(txt)}</li>")
            if aberto_sub:
                html_l.append("</ul></li>")
            out.append("".join(html_l) + f"</{tag}>")
            continue
        if ln.strip().startswith(">"):
            fecha_par()
            cit = []
            while i < len(linhas) and linhas[i].strip().startswith(">"):
                cit.append(linhas[i].strip()[1:].strip())
                i += 1
            out.append("<blockquote><p>" + _inline(" ".join(cit)) + "</p></blockquote>")
            continue
        if not ln.strip():
            fecha_par()
        elif re.match(r"^\s*(-{3,}|\*{3,})\s*$", ln):
            fecha_par()
            out.append("<hr>")
        else:
            par.append(ln)
        i += 1
    fecha_par()
    return "\n".join(out)

if len(sys.argv) < 2:
    print(__doc__)
    sys.exit(2)
MAPA = Path(sys.argv[1])
SAIDA = Path(sys.argv[2]) if len(sys.argv) > 2 else MAPA / "mapa.html"

N_CORES_GRUPO = 6  # --g-1 … --g-6 no CSS; grupos além disso reciclam a paleta


def slug(texto: str) -> str:
    s = unicodedata.normalize("NFKD", texto).encode("ascii", "ignore").decode().lower()
    return re.sub(r"[^a-z0-9]+", "-", s).strip("-") or "areas"


def grupos_da_visao(visao: str):
    """Grupo de cada área = coluna "Grupo" da tabela de áreas da visao-geral.md
    (linha com link para <area>.md). Sem a coluna, tudo cai em "Áreas".
    Devolve (grupo_de, titulos, ordem)."""
    grupo_de, titulos, ordem = {}, {}, []
    cabecalho = []
    linhas = visao.splitlines()
    for i, ln in enumerate(linhas):
        if not ln.strip().startswith("|"):
            continue
        celulas = [c.strip() for c in ln.strip().strip("|").split("|")]
        if i + 1 < len(linhas) and re.match(r"^\s*\|?\s*:?-{3,}", linhas[i + 1]):
            cabecalho = celulas
            continue
        m = re.search(r"\]\(([\w-]+)\.md\)", ln)
        if not m or not cabecalho:
            continue
        idx = next((k for k, c in enumerate(cabecalho) if re.search(r"grupo", c, re.I)), None)
        if idx is None or idx >= len(celulas) or not celulas[idx]:
            continue
        titulo = re.sub(r"[*`_]", "", celulas[idx]).strip()
        g = slug(titulo)
        grupo_de[m.group(1)] = g
        if g not in titulos:
            titulos[g] = titulo
            ordem.append(g)
    titulos.setdefault("areas", "Áreas")
    if "areas" not in ordem:
        ordem.append("areas")
    return grupo_de, titulos, ordem


def render(md_texto: str) -> str:
    if _markdown is not None:
        # o pacote deixa HTML cru passar; escapa o que parece tag (sem mexer em "> citação" nem "a < b")
        h = _markdown.markdown(re.sub(r"<(?=[A-Za-z/!?])", "&lt;", md_texto),
                               extensions=["tables", "fenced_code", "sane_lists"])
    else:
        h = _md_basico(md_texto)
    h = sanitiza_hrefs(h)
    h = re.sub(r'<pre><code class="language-mermaid">(.*?)</code></pre>',
               lambda m: f'<div class="diagram"><pre class="mermaid">{m.group(1)}</pre></div>', h, flags=re.S)
    h = re.sub(r"<table>", '<div class="table-wrap"><table>', h)
    h = re.sub(r"</table>", "</table></div>", h)
    h = re.sub(r'href="inventario\.md[^"]*"', 'href="#fatos"', h)
    h = re.sub(r'href="([\w-]+)\.md(#[^"]*)?"', lambda m: f'href="#area-{m.group(1)}"', h)
    h = re.sub(r'href="visao-geral"', 'href="#visao-geral"', h)
    h = h.replace('href="#area-visao-geral"', 'href="#visao-geral"')
    return h


def sem_titulo(md_texto: str, padrao: str = ""):
    """(título, corpo). O título é o primeiro "# " — pode vir depois de linhas de
    citação ("> Revisado por: …", gravada pelo Aprovar) ou em branco, que ficam no corpo."""
    linhas = md_texto.splitlines()
    for i, ln in enumerate(linhas):
        if ln.startswith("# "):
            return ln[2:].strip(), "\n".join(linhas[:i] + linhas[i + 1:])
        if ln.strip() and not ln.lstrip().startswith(">"):
            break
    return padrao, md_texto


def conta_rn(md_texto: str) -> int:
    m = re.search(r"^##\s+Regras de negócio.*?$(.*?)(?=^##\s|\Z)", md_texto, re.M | re.S)
    if not m:
        return 0
    linhas = [l for l in m.group(1).splitlines() if l.strip().startswith("|")]
    tabela = max(0, len(linhas) - 2)
    bullets = len([l for l in m.group(1).splitlines() if re.match(r"^\s*[-*]\s+\S", l)])
    return tabela or bullets


def resumo(md_texto: str) -> str:
    m = re.search(r"^##\s+Para que serve\s*$(.*?)(?=^##\s|\Z)", md_texto, re.M | re.S)
    txt = m.group(1) if m else ""
    txt = re.sub(r"\[([^\]]+)\]\([^)]+\)", r"\1", txt)
    par = next((p.strip() for p in txt.split("\n\n") if p.strip() and not p.strip().startswith(("-", "|", "`"))), "")
    par = re.sub(r"[*`]", "", re.sub(r"\s+", " ", par))
    return (par[:180] + "…") if len(par) > 180 else par


TEMPLATE_HTML = r"""<meta charset="utf-8">
<title>Mapa do {{PROJETO}}</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Public+Sans:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;600&display=swap">
<style>
  :root {
    --bg: #f4f5f3; --panel: #fbfbfa; --panel-2: #eceee9; --line: #dcdfd8;
    --ink: #13161a; --ink-2: #464b52; --ink-3: #737982;
    --g-1: #2d6fd6; --g-2: #13936a; --g-3: #c27208; --g-4: #6b5bd2; --g-5: #b4285a; --g-6: #0f8a9c;
    --warn: #c2410c;
    --sans: "Public Sans", -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    --mono: "JetBrains Mono", ui-monospace, "SF Mono", Menlo, Consolas, monospace;
  }
  @media (prefers-color-scheme: dark) {
    :root:not([data-theme="light"]) {
      color-scheme: dark;
      --bg: #111315; --panel: #181b1e; --panel-2: #22262a; --line: #31363b;
      --ink: #eef0f1; --ink-2: #c1c6cc; --ink-3: #8b9199;
      --g-1: #6aa1f2; --g-2: #3cc497; --g-3: #f0a541; --g-4: #a497f2; --g-5: #e9729a; --g-6: #4fc3d3;
      --warn: #f07a4a;
    }
  }
  :root[data-theme="dark"] {
    color-scheme: dark;
    --bg: #111315; --panel: #181b1e; --panel-2: #22262a; --line: #31363b;
    --ink: #eef0f1; --ink-2: #c1c6cc; --ink-3: #8b9199;
    --g-1: #6aa1f2; --g-2: #3cc497; --g-3: #f0a541; --g-4: #a497f2; --g-5: #e9729a; --g-6: #4fc3d3;
    --warn: #f07a4a;
  }
  * { box-sizing: border-box; }
  body { margin: 0; background: var(--bg); color: var(--ink); font-family: var(--sans); font-size: 15px; line-height: 1.6;
         padding-inline: 16px; padding-block: 32px 80px; }
  h1, h2, h3, h4 { margin: 0; text-wrap: balance; }
  code { font-family: var(--mono); font-size: 0.82em; background: var(--panel-2); padding: 1px 5px; border-radius: 5px; overflow-wrap: anywhere; }
  a { color: inherit; }
  {{GRUPOS_CSS}}

  .shell { max-width: 1180px; margin: 0 auto; display: grid; grid-template-columns: 250px minmax(0, 1fr); gap: 36px; align-items: start; }
  header.top { max-width: 1180px; margin: 0 auto 28px; display: grid; gap: 10px; }
  .eyebrow { font-size: 0.72rem; font-weight: 700; letter-spacing: 0.12em; text-transform: uppercase; color: var(--ink-3); }
  h1 { font-size: clamp(1.8rem, 4.5vw, 2.6rem); font-weight: 800; letter-spacing: -0.025em; line-height: 1.1; }
  .sub { color: var(--ink-2); max-width: 70ch; font-size: 1.02rem; }
  .stats { display: flex; flex-wrap: wrap; gap: 8px 18px; font-size: 0.9rem; color: var(--ink-2); }
  .stats b { color: var(--ink); font-variant-numeric: tabular-nums; font-size: 1.05rem; }
  .aviso { font-size: 0.84rem; color: var(--ink-2); border-left: 3px solid var(--warn); padding-left: 10px; max-width: 80ch; }

  nav.side { position: sticky; top: calc(env(safe-area-inset-top, 0px) + 16px); max-height: calc(100vh - 32px); overflow-y: auto;
             display: grid; gap: 14px; padding-bottom: 12px; }
  .filtro { width: 100%; font: inherit; font-size: 0.9rem; padding: 8px 10px; border-radius: 8px; border: 1px solid var(--line);
            background: var(--panel); color: var(--ink); }
  .filtro:focus-visible { outline: 2px solid var(--g-1); outline-offset: 1px; }
  nav.side ul { list-style: none; margin: 0; padding: 0; display: grid; gap: 2px; }
  .nav-top a, nav.side li li a { display: flex; align-items: center; gap: 8px; text-decoration: none; padding: 5px 8px; border-radius: 7px; font-size: 0.88rem; color: var(--ink-2); }
  .nav-top a:hover, nav.side li li a:hover, nav.side a:focus-visible { background: var(--panel-2); color: var(--ink); outline: none; }
  nav.side a.ativo { background: var(--panel-2); color: var(--ink); font-weight: 600; }
  .nav-gt { display: block; font-size: 0.7rem; font-weight: 700; letter-spacing: 0.1em; text-transform: uppercase; color: var(--ink-3); padding: 8px 8px 4px; }
  .nav-t { flex: 1; min-width: 0; }
  .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--g); flex: none; }
  .badge { font-size: 0.7rem; font-variant-numeric: tabular-nums; color: var(--ink-3); border: 1px solid var(--line); border-radius: 999px; padding: 0 6px; }
  .vazio { font-size: 0.84rem; color: var(--ink-3); padding: 4px 8px; }

  main { display: grid; gap: 48px; min-width: 0; }
  section { scroll-margin-top: 16px; display: grid; gap: 16px; min-width: 0; }
  .cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(230px, 1fr)); gap: 10px; }
  .card { display: grid; gap: 5px; align-content: start; text-decoration: none; background: var(--panel); border: 1px solid var(--line);
          border-top: 3px solid var(--g); border-radius: 10px; padding: 12px 14px; }
  .card:hover, .card:focus-visible { border-color: var(--g); outline: none; }
  .card-g { font-size: 0.68rem; font-weight: 700; letter-spacing: 0.1em; text-transform: uppercase; color: var(--g); }
  .card b { font-size: 0.95rem; line-height: 1.3; }
  .card-r { font-size: 0.82rem; color: var(--ink-2); }
  .card-m { font-size: 0.74rem; color: var(--ink-3); font-variant-numeric: tabular-nums; }

  .area { border-top: 3px solid var(--g); padding-top: 18px; }
  .area-h { display: grid; gap: 4px; }
  .area-h .eyebrow { color: var(--g); }
  #situacao h3 { font-size: 1rem; font-weight: 700; display: flex; gap: 10px; align-items: baseline; flex-wrap: wrap; }
  .prog { font-size: 0.8rem; font-weight: 500; color: var(--ink-3); font-variant-numeric: tabular-nums; }
  .st { font-size: 0.72rem; font-weight: 700; letter-spacing: 0.04em; border-radius: 999px; padding: 1px 8px; border: 1px solid var(--line); white-space: nowrap; }
  .st-concluida .st { color: var(--g-2); border-color: var(--g-2); }
  .st-em-curso .st { color: var(--g-1); border-color: var(--g-1); }
  .st-a-fazer .st { color: var(--ink-3); }
  .estado { background: var(--panel); border: 1px solid var(--line); border-radius: 10px; padding: 12px 16px; }
  .area-h h2, #situacao > h2, #infraestrutura > h2, #visao-geral > h2, #fatos > h2 { font-size: 1.5rem; font-weight: 800; letter-spacing: -0.015em; }

  .md { display: grid; gap: 12px; min-width: 0; color: var(--ink); }
  .md > * { margin: 0; }
  .md h2 { font-size: 1.1rem; font-weight: 700; margin-top: 12px; }
  .md h3 { font-size: 1rem; font-weight: 700; margin-top: 6px; }
  .md p, .md li { color: var(--ink-2); max-width: 78ch; }
  .md ul, .md ol { padding-left: 20px; display: grid; gap: 4px; }
  .md strong { color: var(--ink); }
  .md blockquote { margin: 0; border-left: 3px solid var(--line); padding-left: 12px; color: var(--ink-3); font-size: 0.9rem; }
  .table-wrap { overflow-x: auto; border: 1px solid var(--line); border-radius: 10px; background: var(--panel); }
  table { border-collapse: collapse; width: 100%; font-size: 0.84rem; }
  th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid var(--line); vertical-align: top; }
  th { font-size: 0.72rem; text-transform: uppercase; letter-spacing: 0.07em; color: var(--ink-3); background: var(--panel-2); white-space: nowrap; }
  tr:last-child td { border-bottom: 0; }
  td { color: var(--ink-2); }
  .diagram { overflow-x: auto; background: var(--panel); border: 1px solid var(--line); border-radius: 10px; padding: 12px; }
  .diagram pre.mermaid { margin: 0; min-width: min-content; background: transparent; }
  details { background: var(--panel); border: 1px solid var(--line); border-radius: 10px; padding: 10px 14px; }
  details[open] { display: grid; gap: 10px; }
  summary { cursor: pointer; font-weight: 600; }
  .topo { font-size: 0.8rem; margin: 0; }
  .topo a { color: var(--ink-3); }
  footer.page { color: var(--ink-3); font-size: 0.8rem; border-top: 1px solid var(--line); padding-top: 14px; }

  @media (max-width: 900px) {
    .shell { grid-template-columns: 1fr; gap: 20px; }
    nav.side { position: static; max-height: none; border: 1px solid var(--line); border-radius: 10px; padding: 12px; background: var(--panel); }
    nav.side .nav-g ul { grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); }
  }
</style>

<header class="top" id="top">
  <span class="eyebrow">{{PROJETO}} · mapa do código</span>
  <h1>Mapa do {{PROJETO}}</h1>
  <p class="sub">O código por área de negócio: o que cada parte faz, por onde se entra, o que chama em
  outros sistemas, as regras que o código aplica e os pontos de atenção.</p>
  <div class="stats">{{STATS}}</div>
  <p class="aviso">Gerado de <code>DOC/System/mapa/</code> (inventário de {{DATA}}). Ainda não revisado por uma pessoa:
  regras, bugs e riscos de segurança estão marcados como "a confirmar". A fonte é a DOC; esta página é só a vista.</p>
</header>

<div class="shell">
  <nav class="side" aria-label="Áreas do mapa">
    <label class="eyebrow" for="filtro">Filtrar áreas</label>
    <input class="filtro" id="filtro" type="search" placeholder="ex.: rota, tela, integração" autocomplete="off">
    <ul class="nav-top">
      <li><a href="#situacao">Situação do projeto</a></li>
      {{NAV_INFRA}}
      <li><a href="#visao-geral">Visão geral</a></li>
      <li><a href="#fatos">Fatos do inventário</a></li>
    </ul>
    <ul>{{NAV}}</ul>
    <p class="vazio" id="nada" hidden>Nenhuma área menciona esse termo.</p>
  </nav>

  <main>
    {{SITUACAO}}

    {{INFRA}}

    <section id="visao-geral">
      <h2>Visão geral</h2>
      <div class="cards">{{CARDS}}</div>
      <div class="md">{{VISAO}}</div>
    </section>

    {{SECOES}}

    <section id="fatos">
      <h2>Fatos do inventário</h2>
      <div class="md">
        <p>Extraídos do código por <code>DOC/tools/mapeia_projeto.py</code>, sem interpretação.
        O <code>checa_projeto.py</code> avisa quando o código muda e esta lista fica para trás.</p>
        <h3>Módulos e quem importa quem</h3>
        <div class="diagram"><pre class="mermaid">{{MODULOS}}</pre></div>
        {{DETALHE_ROTAS}}
        {{DETALHE_ENDPOINTS}}
        <details><summary>Chamadas a outros sistemas ({{N_HTTP}})</summary>
          <div class="table-wrap"><table><thead><tr><th>Método</th><th>URL</th><th>Onde</th></tr></thead><tbody>{{HTTP}}</tbody></table></div>
        </details>
      </div>
    </section>

    <footer class="page">Para atualizar: rode <code>python3 DOC/tools/mapeia_projeto.py .</code>, revise as páginas
    das áreas afetadas em <code>DOC/System/mapa/</code> e gere esta vista de novo.</footer>
  </main>
</div>

<script type="module">
  import mermaid from "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs";
  mermaid.initialize({ startOnLoad: true, theme: matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "default" });
</script>
<script>
  (function () {
    var input = document.getElementById("filtro");
    var areas = Array.prototype.slice.call(document.querySelectorAll("section.area"));
    var links = Array.prototype.slice.call(document.querySelectorAll("nav.side a[data-area]"));
    var cards = Array.prototype.slice.call(document.querySelectorAll(".card"));
    var nada = document.getElementById("nada");
    try { var salvo = localStorage.getItem("mapa-filtro"); if (salvo) input.value = salvo; } catch (e) {}
    function aplicar() {
      var t = input.value.trim().toLowerCase();
      var visiveis = 0;
      areas.forEach(function (s) {
        var ok = !t || s.getAttribute("data-busca").indexOf(t) !== -1;
        s.hidden = !ok;
        if (ok) visiveis++;
        var id = s.id.replace("area-", "");
        links.forEach(function (a) { if (a.getAttribute("data-area") === id) a.parentNode.hidden = !ok; });
        cards.forEach(function (c) { if (c.getAttribute("href") === "#" + s.id) c.hidden = !ok; });
      });
      nada.hidden = visiveis > 0;
      try { localStorage.setItem("mapa-filtro", input.value); } catch (e) {}
    }
    input.addEventListener("input", aplicar);
    aplicar();

    if ("IntersectionObserver" in window) {
      var obs = new IntersectionObserver(function (entradas) {
        entradas.forEach(function (en) {
          if (!en.isIntersecting) return;
          links.forEach(function (a) { a.classList.toggle("ativo", "area-" + a.getAttribute("data-area") === en.target.id); });
        });
      }, { rootMargin: "0px 0px -70% 0px" });
      areas.forEach(function (s) { obs.observe(s); });
    }
  })();
</script>
"""

# ------------------------------------------------------------ conteúdo
visao = (MAPA / "visao-geral.md").read_text(encoding="utf-8")
_, visao_corpo = sem_titulo(visao)
visao_corpo = re.sub(r"^>.*\n(>.*\n)*", "", visao_corpo.lstrip("\n"))
inv = json.loads((MAPA / "inventario.json").read_text(encoding="utf-8"))
inv_md = (MAPA / "inventario.md").read_text(encoding="utf-8")
modulos_mermaid = re.search(r"```mermaid\n(.*?)```", inv_md, re.S).group(1)

GRUPO_DE, GRUPOS, ORDEM_GRUPOS = grupos_da_visao(visao)
grupos_css = "".join(f".g-{g} {{ --g: var(--g-{i % N_CORES_GRUPO + 1}); }}\n  " for i, g in enumerate(ORDEM_GRUPOS))
links = re.findall(r"\]\(([\w-]+)\.md\)", visao)
ordem = [n for n in dict.fromkeys(links) if n not in ("visao-geral", "inventario") and (MAPA / f"{n}.md").is_file()]
AREAS = [(n, GRUPO_DE.get(n, "areas")) for n in ordem] or \
        [(p.stem, "areas") for p in sorted(MAPA.glob("*.md")) if p.stem not in ("visao-geral", "inventario")]
areas = []
for nome, grupo in AREAS:
    texto = (MAPA / f"{nome}.md").read_text(encoding="utf-8")
    titulo, corpo = sem_titulo(texto, nome.replace("-", " ").capitalize())
    corpo = re.sub(r"^\s*>.*\n(>.*\n)*", "", corpo.lstrip("\n"))
    areas.append({"id": nome, "grupo": grupo, "titulo": titulo, "html": render(corpo),
                  "rn": conta_rn(texto), "resumo": resumo(texto),
                  "busca": re.sub(r"\s+", " ", texto.lower())})

n_rn = sum(a["rn"] for a in areas)
# Números do topo: só o que o projeto tem (serviço de back-end não tem tela; front não expõe endpoint).
ENDPOINTS = inv.get("endpoints", [])
_contagens = [("telas (rotas do front)", len(inv["rotas"])), ("endpoints expostos", len(ENDPOINTS)),
              ("chamadas a outros sistemas", len(inv["chamadas_http"])),
              ("filas", len(inv.get("filas_consumidas", [])) + len(inv.get("filas_publicadas", []))),
              ("crons", len(inv.get("crons", [])))]
stats = [("áreas", len(areas))] + [c for c in _contagens if c[1]] + \
        [("regras candidatas", n_rn), ("módulos", len(inv["modulos"]))]

# ------------------------------------------------------------ situação do projeto
# Onde o projeto está, lido dos arquivos que o fluxo SDD já mantém: estado.md (onde
# parou), backlog + openspec/changes (o que foi concluído, está em curso, falta) e
# to-do.md (aonde é preciso voltar). Nada disso é escrito à mão aqui.
DOC_DIR = MAPA.resolve().parent.parent
def _raiz_do_repo(doc_dir: Path) -> Path:
    """Sobe da DOC até o repositório (para no primeiro .git; nunca lê fora dele)."""
    for c in [doc_dir, *doc_dir.parents]:
        if (c / ".git").exists():
            return c
        if (c / "openspec").is_dir():
            return c
    return doc_dir.parent


RAIZ = _raiz_do_repo(DOC_DIR)
CHANGES = RAIZ / "openspec" / "changes"


def progresso_tasks(pasta: Path):
    f = pasta / "tasks.md"
    if not f.is_file():
        return 0, 0
    t = f.read_text(encoding="utf-8", errors="replace")
    return len(re.findall(r"^\s*- \[[xX]\]", t, re.M)), len(re.findall(r"^\s*- \[[ xX]\]", t, re.M))


def fatias_do_backlog():
    f = DOC_DIR / "System" / "backlog-mvp.md"
    if not f.is_file():
        return []
    txt = f.read_text(encoding="utf-8", errors="replace")
    arquivadas = [p.name for p in (CHANGES / "archive").glob("*")] if (CHANGES / "archive").is_dir() else []
    specs = RAIZ / "openspec" / "specs"
    vigentes = set(re.findall(r"^###\s+Requirement:\s*(REQ-[\d.]+)",
                              "\n".join(f.read_text(encoding="utf-8", errors="replace") for f in specs.glob("*/spec.md")),
                              re.M)) if specs.is_dir() else set()
    out = []
    for m in re.finditer(r"^###\s+(Fatia\s+[\w.]+)\s*[—-]\s*(.+?)\s*$(.*?)(?=^###\s|\Z)", txt, re.M | re.S):
        nome, titulo, corpo = m.group(1), m.group(2), m.group(3)
        ch = re.search(r"^Change:\s*`([^`]+)`", corpo, re.M)
        change = ch.group(1) if ch else ""
        if change and any(a.endswith(change) for a in arquivadas):
            status, prog = "concluida", ""
        elif change and (CHANGES / change).is_dir():
            feitas, total = progresso_tasks(CHANGES / change)
            status, prog = "em-curso", (f"{feitas}/{total} tarefas" if total else "")
        elif change:
            status, prog = "a-fazer", ""
        else:
            reqs = set(re.findall(r"\bREQ-[\d.]+\b", corpo))
            if reqs and reqs <= vigentes:  # entregue antes do OpenSpec: REQ já vigentes
                status, prog = "concluida", "REQ vigentes"
            else:
                status, prog = "sem-change", "sem change no OpenSpec"
        out.append({"nome": nome, "titulo": titulo, "change": change, "status": status, "prog": prog})
    return out


def pendencias_todo(limite=15):
    f = DOC_DIR / "to-do.md"
    if not f.is_file():
        return [], 0
    abertas = [re.sub(r"^\s*- \[ \]\s*", "", l).strip()
               for l in f.read_text(encoding="utf-8", errors="replace").splitlines() if re.match(r"^\s*- \[ \]", l)]
    return abertas[:limite], len(abertas)


ROTULO = {"concluida": "Concluída", "em-curso": "Em curso", "a-fazer": "A fazer", "sem-change": "—"}
FATIAS = fatias_do_backlog()
TODO, N_TODO = pendencias_todo()
ESTADO_F = DOC_DIR / "estado.md"
estado_html = ""
if ESTADO_F.is_file():
    _, est = sem_titulo(ESTADO_F.read_text(encoding="utf-8", errors="replace"))
    est = re.sub(r"<!--.*?-->", "", est, flags=re.S)
    est = re.sub(r"^Retrato de onde o trabalho parou.*?(?:\n\s*\n|\Z)", "", est, flags=re.M | re.S)  # texto do modelo
    estado_html = render(est.strip())
n_ok = sum(f["status"] == "concluida" for f in FATIAS)
fatias_rows = "".join(
    f'<tr class="st-{f["status"]}"><td>{html.escape(f["nome"])}</td><td>{html.escape(f["titulo"])}</td>'
    f'<td>{("<code>" + html.escape(f["change"]) + "</code>") if f["change"] else "—"}</td>'
    f'<td><span class="st">{ROTULO[f["status"]]}</span> <span class="prog">{html.escape(f["prog"])}</span></td></tr>'
    for f in FATIAS)
todo_html = "".join(f"<li>{render(t)[3:-4] if render(t).startswith('<p>') else html.escape(t)}</li>" for t in TODO)
situacao = ['<section id="situacao"><h2>Situação do projeto</h2>',
            '<p class="sub">Lida de <code>estado.md</code>, <code>System/backlog-mvp.md</code>, '
            '<code>openspec/changes/</code> e <code>to-do.md</code> na hora em que esta vista foi gerada.</p>']
if estado_html:
    situacao.append(f'<h3>Onde parou</h3><div class="md estado">{estado_html}</div>')
else:
    situacao.append('<p class="vazio">Sem <code>estado.md</code>: ao encerrar uma sessão, faça a passagem de bastão (skill passagem).</p>')
if FATIAS:
    situacao.append(f'<h3>Fatias do backlog <span class="prog">{n_ok} de {len(FATIAS)} concluídas</span></h3>'
                    '<div class="table-wrap"><table><thead><tr><th>Fatia</th><th>O quê</th><th>Change</th><th>Situação</th></tr></thead>'
                    f'<tbody>{fatias_rows}</tbody></table></div>')
if TODO:
    mais = f" (mostrando {len(TODO)} de {N_TODO})" if N_TODO > len(TODO) else ""
    situacao.append(f'<h3>Aonde voltar <span class="prog">{N_TODO} pendência(s) no to-do.md{mais}</span></h3>'
                    f'<div class="md"><ul>{todo_html}</ul></div>')
situacao.append('</section>')
SITUACAO = "".join(situacao)

# Infraestrutura (stack, onde roda, bancos de hml e prod): Infra/infraestrutura.md
INFRA_F = next((DOC_DIR / n / "infraestrutura.md" for n in ("Infra", "infra")
                if (DOC_DIR / n / "infraestrutura.md").is_file()), None)
if INFRA_F:
    _, inf = sem_titulo(INFRA_F.read_text(encoding="utf-8", errors="replace"))
    inf = re.sub(r"^\s*>.*\n(>.*\n)*", "", inf.lstrip("\n"))  # nota do modelo
    n_def = inf.split("## Detectado no repositório")[0].count("A DEFINIR")  # mesma regra do checa_projeto
    aviso_inf = (f'<p class="aviso">{n_def} campo(s) ainda "A DEFINIR".</p>' if n_def else "")
    INFRA = (f'<section id="infraestrutura"><h2>Infraestrutura</h2><p class="sub">De '
             f'<code>{html.escape(INFRA_F.relative_to(DOC_DIR).as_posix())}</code>: stack, onde roda e onde estão os '
             f'dados de cada ambiente.</p>{aviso_inf}<div class="md">{render(inf)}</div></section>')
else:
    INFRA = ('<section id="infraestrutura"><h2>Infraestrutura</h2><p class="vazio">Sem <code>Infra/infraestrutura.md</code>: '
             'peça ao Claude "documente a infraestrutura (stack, onde roda, bancos de hml e prod) com a skill ambientes".</p></section>')
NAV_INFRA = '<li><a href="#infraestrutura">Infraestrutura</a></li>'
if FATIAS:
    stats.insert(0, ("fatias concluídas", f"{n_ok}/{len(FATIAS)}"))


# ------------------------------------------------------------ HTML
def badge(n):
    return f'<span class="badge" title="regras candidatas">{n}</span>' if n else ""


nav = []
for g in ORDEM_GRUPOS:
    itens = "".join(
        f'<li><a href="#area-{a["id"]}" data-area="{a["id"]}"><span class="dot g-{g}"></span>'
        f'<span class="nav-t">{html.escape(a["titulo"].split(" (")[0])}</span>{badge(a["rn"])}</a></li>'
        for a in areas if a["grupo"] == g)
    if itens:  # grupo sem área não ganha cabeçalho vazio no menu
        nav.append(f'<li class="nav-g"><span class="nav-gt">{html.escape(GRUPOS[g])}</span><ul>{itens}</ul></li>')

cards = "".join(
    f'<a class="card g-{a["grupo"]}" href="#area-{a["id"]}"><span class="card-g">{GRUPOS[a["grupo"]]}</span>'
    f'<b>{html.escape(a["titulo"])}</b><span class="card-r">{html.escape(a["resumo"])}</span>'
    f'<span class="card-m">{a["rn"]} regras candidatas</span></a>'
    for g in ORDEM_GRUPOS for a in areas if a["grupo"] == g)

secoes = "".join(
    f'<section class="area g-{a["grupo"]}" id="area-{a["id"]}" data-busca="{html.escape(a["busca"], quote=True)}">'
    f'<header class="area-h"><span class="eyebrow">{GRUPOS[a["grupo"]]}</span><h2>{html.escape(a["titulo"])}</h2></header>'
    f'<div class="md">{a["html"]}</div><p class="topo"><a href="#top">Voltar ao topo</a></p></section>'
    for g in ORDEM_GRUPOS for a in areas if a["grupo"] == g)

rotas_rows = "".join(
    f'<tr><td><code>{html.escape(r["path"])}</code></td><td>{html.escape(r["titulo"] or "—")}</td>'
    f'<td><code>{html.escape(r["componente"] or "—")}</code></td></tr>' for r in inv["rotas"])
http_rows = "".join(
    f'<tr><td>{c["metodo"]}</td><td><code>{html.escape(c["url"])}</code></td><td><code>{html.escape(c["arquivo"])}:{c["linha"]}</code></td></tr>'
    for c in sorted(inv["chamadas_http"], key=lambda c: (c["url"], c["metodo"])))

endpoints_rows = "".join(
    f'<tr><td>{html.escape(e["metodo"])}</td><td><code>{html.escape(e["path"])}</code></td>'
    f'<td><code>{html.escape(e["arquivo"])}{(":" + str(e["linha"])) if e.get("linha") else ""}</code></td></tr>'
    for e in sorted(ENDPOINTS, key=lambda e: (e["path"], e["metodo"])))
DETALHE_ROTAS = (f'<details><summary>Telas — rotas do front ({len(inv["rotas"])})</summary><div class="table-wrap"><table>'
                 f'<thead><tr><th>Rota</th><th>Título</th><th>Componente</th></tr></thead><tbody>{rotas_rows}</tbody></table></div></details>'
                 if inv["rotas"] else "")
DETALHE_ENDPOINTS = (f'<details><summary>Endpoints expostos ({len(ENDPOINTS)})</summary><div class="table-wrap"><table>'
                     f'<thead><tr><th>Método</th><th>Caminho</th><th>Onde</th></tr></thead><tbody>{endpoints_rows}</tbody></table></div></details>'
                     if ENDPOINTS else "")
stats_html = "".join(f'<span><b>{v}</b> {k}</span>' for k, v in stats)
TEMPLATE = TEMPLATE_HTML
out = (TEMPLATE.replace("{{NAV}}", "".join(nav)).replace("{{CARDS}}", cards)
       .replace("{{VISAO}}", render(visao_corpo)).replace("{{SECOES}}", secoes).replace("{{SITUACAO}}", SITUACAO)
       .replace("{{INFRA}}", INFRA).replace("{{NAV_INFRA}}", NAV_INFRA)
       .replace("{{STATS}}", stats_html).replace("{{DATA}}", inv["gerado_em"][:10])
       .replace("{{MODULOS}}", html.escape(modulos_mermaid))
       .replace("{{DETALHE_ROTAS}}", DETALHE_ROTAS).replace("{{DETALHE_ENDPOINTS}}", DETALHE_ENDPOINTS)
       .replace("{{HTTP}}", http_rows)
       .replace("{{N_HTTP}}", str(len(inv["chamadas_http"]))).replace("{{N_ROTAS}}", str(len(inv["rotas"])))
       .replace("{{PROJETO}}", html.escape(inv.get("raiz", "projeto"))).replace("{{GRUPOS_CSS}}", grupos_css))
SAIDA.write_text(out, encoding="utf-8")
try:
    sys.path.insert(0, str(Path(__file__).resolve().parent))
    import registro_log
    _raiz = next((c for c in [MAPA.resolve(), *MAPA.resolve().parents] if (c / "DOC").is_dir()), None)
    if _raiz:
        registro_log.registrar(_raiz, "mapa_html", 0, resumo={"areas": len(areas), "regras_candidatas": n_rn})
except ImportError:
    print("AVISO: registro_log.py ausente — execução não registrada no log da DOC.", file=sys.stderr)
print(f"{SAIDA} — {len(out)//1024} KB, {n_rn} regras candidatas")
