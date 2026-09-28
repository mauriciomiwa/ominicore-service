#!/usr/bin/env python3
"""Registro das execuções dos scripts SDD em <DOC>/log/ — para tratar erros e
dependências faltando depois, e para acompanhar o projeto ao longo do tempo.

Estrutura (criada sob demanda):

  <DOC>/log/
    README.md          o que é cada arquivo
    historico.jsonl    uma linha por execução: data, script, rc, faltas, alertas, erro
    problemas.md       o que está aberto AGORA: a última execução de cada script,
                       com FALTA/ALERTA, erros e o comando de cada correção
    execucoes/         detalhe de cada execução (JSON); fica fora do git
    .gitignore         ignora execucoes/

Uso nos scripts:

    import registro_log
    registro_log.registrar(raiz, "verifica-projeto", rc, resumo={...},
                           problemas=[{"nivel": "FALTA", "etapa": ..., "msg": ..., "como": ..., "comando": ...}])

    if __name__ == "__main__":
        sys.exit(registro_log.executar("mapeia_projeto", main, raiz_de_argv))

`executar` roda o main e, se ele explodir com exceção, grava o traceback no
log antes de sair com código 2 — o erro não se perde no terminal.
Nunca grava valores de variáveis de ambiente nem conteúdo de arquivos do projeto.
"""
import datetime
import json
import os
import sys
import traceback
from pathlib import Path

MAX_EXECUCOES = 200
LIMITE_SAIDA = 20_000

README = """# log/ — execuções dos scripts SDD

Gerado pelos scripts das skills (verifica-projeto, checa_projeto, mapeia_projeto,
mapa_html, sync-skills) e pelo painel do openmd.

| Arquivo | O que é |
|---|---|
| `problemas.md` | o que está aberto agora: última execução de cada script, com FALTA/ALERTA, erros e como corrigir |
| `historico.jsonl` | uma linha por execução (data, script, resultado, faltas, alertas) — dá a evolução do projeto |
| `execucoes/` | detalhe de cada execução em JSON (fora do git) |

Para tratar: abra `problemas.md`, rode o comando de cada item (ou peça ao Claude:
"trate os problemas de DOC/.../log/problemas.md") e rode a verificação de novo.
"""


def doc_do_projeto(raiz: Path):
    """Pasta da DOC: DOC/ (app único) ou a única DOC/<projeto>/ com INDEX.md."""
    raiz = Path(raiz)
    if (raiz / "DOC" / "INDEX.md").is_file():
        return raiz / "DOC"
    base = raiz / "DOC"
    if base.is_dir():
        docs = sorted(p.parent for p in base.glob("*/INDEX.md"))
        if len(docs) == 1:
            return docs[0]
        if docs:
            return docs[0]
    return None


def _agora():
    return datetime.datetime.now().replace(microsecond=0)


def registrar(raiz, script, rc, resumo=None, problemas=None, saida=None, erro=None, doc=None):
    """Grava a execução. Devolve o caminho do detalhe ou None se não há DOC."""
    try:
        raiz = Path(raiz).resolve()
        doc = Path(doc) if doc else doc_do_projeto(raiz)
        if doc is None:
            return None
        log = doc / "log"
        (log / "execucoes").mkdir(parents=True, exist_ok=True)
        if not (log / "README.md").exists():
            (log / "README.md").write_text(README, encoding="utf-8")
        if not (log / ".gitignore").exists():
            (log / ".gitignore").write_text("execucoes/\n", encoding="utf-8")

        quando = _agora()
        problemas = list(problemas or [])
        faltas = sum(1 for p in problemas if p.get("nivel") == "FALTA")
        alertas = sum(1 for p in problemas if p.get("nivel") == "ALERTA")
        nome = f"{quando:%Y-%m-%d_%H%M%S}_{script}.json"
        detalhe = {
            "data": quando.isoformat(), "script": script, "rc": rc, "resumo": resumo or {},
            "problemas": problemas, "erro": erro,
            "saida": (saida or "")[-LIMITE_SAIDA:],
        }
        (log / "execucoes" / nome).write_text(json.dumps(detalhe, ensure_ascii=False, indent=1), encoding="utf-8")
        linha = {"data": quando.isoformat(), "script": script, "rc": rc, "faltas": faltas, "alertas": alertas,
                 "erro": bool(erro), "detalhe": f"execucoes/{nome}"}
        linha.update({k: v for k, v in (resumo or {}).items() if isinstance(v, (int, float, str, bool))})
        with (log / "historico.jsonl").open("a", encoding="utf-8") as h:
            h.write(json.dumps(linha, ensure_ascii=False) + "\n")
        _podar(log / "execucoes")
        _problemas_md(log)
        return log / "execucoes" / nome
    except Exception as e:  # o log nunca derruba o script, mas deixa rastro
        print(f"registro_log: não gravou o log da execução ({type(e).__name__}: {e})", file=sys.stderr)
        return None


def _podar(pasta: Path):
    arquivos = sorted(pasta.glob("*.json"))
    for f in arquivos[:-MAX_EXECUCOES]:
        try:
            f.unlink()  # só detalhe de execução gerado por este módulo
        except OSError:
            pass


def ultimas_por_script(log: Path):
    """Última execução de cada script (dict script -> detalhe)."""
    ultimas = {}
    hist = log / "historico.jsonl"
    if not hist.is_file():
        return ultimas
    for ln in hist.read_text(encoding="utf-8").splitlines():
        try:
            d = json.loads(ln)
        except json.JSONDecodeError:
            continue
        ultimas[d.get("script")] = d
    out = {}
    for script, d in ultimas.items():
        f = log / d.get("detalhe", "")
        try:
            out[script] = json.loads(f.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError):
            out[script] = {"data": d.get("data"), "script": script, "rc": d.get("rc"), "problemas": [], "erro": None}
    return out


def _problemas_md(log: Path):
    ultimas = ultimas_por_script(log)
    L = ["# Problemas abertos", "",
         f"> Gerado em {_agora():%Y-%m-%d %H:%M} a partir da última execução de cada script. "
         "Não edite: rode os scripts de novo.", ""]
    abertos = 0
    for script in sorted(ultimas):
        d = ultimas[script]
        probs = [p for p in d.get("problemas", []) if p.get("nivel") in ("FALTA", "ALERTA")]
        if not probs and not d.get("erro"):
            continue
        L += [f"## {script} — {str(d.get('data', ''))[:16].replace('T', ' ')}", ""]
        if d.get("erro"):
            abertos += 1
            ultima = [l for l in str(d["erro"]).strip().splitlines() if l.strip()][-1:] or ["erro"]
            L += [f"- [ ] **ERRO** — o script falhou: `{ultima[0][:300]}`",
                  f"  - traceback completo em `log/execucoes/` (execução de {str(d.get('data', ''))[:19]})", ""]
        for p in sorted(probs, key=lambda p: p.get("nivel") != "FALTA"):
            abertos += 1
            etapa = f" [{p['etapa']}]" if p.get("etapa") else ""
            L.append(f"- [ ] **{p['nivel']}**{etapa} {p.get('msg', '')}")
            if p.get("como"):
                L.append(f"  - como corrigir: {p['como']}")
            if p.get("comando"):
                L.append(f"  - `{p['comando']}`")
        L.append("")
    if not abertos:
        L += ["Nenhum problema aberto na última execução de cada script.", ""]
    (log / "problemas.md").write_text("\n".join(L), encoding="utf-8")


def executar(script, main, raiz=None):
    """Roda main(); em exceção não tratada, grava o traceback em <DOC>/log/ e sai com 2."""
    try:
        return main()
    except SystemExit:
        raise
    except KeyboardInterrupt:
        return 130
    except Exception:
        tb = traceback.format_exc()
        print(tb, file=sys.stderr)
        r = raiz() if callable(raiz) else raiz
        if r:
            destino = registrar(r, script, 2, erro=tb)
            if destino:
                print(f"Erro registrado em {destino.parent.parent / 'problemas.md'}", file=sys.stderr)
        return 2


def raiz_do_argv(indice=1):
    """Primeiro argumento posicional como raiz (para `executar`)."""
    pos = [a for a in sys.argv[1:] if not a.startswith("-")]
    if len(pos) >= indice:
        p = Path(os.path.expanduser(pos[indice - 1]))
        # aceita tanto a raiz quanto um caminho dentro dela (ex.: DOC/System/mapa)
        for c in [p, *p.resolve().parents]:
            if (c / "DOC").is_dir() or (c / ".git").exists():
                return c
        return p
    return None
