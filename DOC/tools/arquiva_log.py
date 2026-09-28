#!/usr/bin/env python3
"""Arquiva as entradas antigas do log.md da DOC, sem apagar nada.

O log.md fica com as entradas dos últimos N dias (padrão 30) e, se mesmo assim
passar do limite (padrão 15 KB), só com as mais recentes que couberem (no
mínimo as 3 últimas). O resto vai para <DOC>/log/arquivo-AAAA.md (uma por ano,
mais recente no topo). Assim o que a IA lê no dia a dia continua curto e a
história inteira segue versionada.

Uso:
  python3 arquiva_log.py <pasta da DOC> [--dias 30] [--limite-kb 15] [--check]

Formato esperado do log.md: cabeçalho livre e entradas "## AAAA-MM-DD — título"
(ou "## AAAA-MM-DD a AAAA-MM-DD — …"; vale a primeira data). Entrada sem data
fica no log.md. --check só mostra o que moveria (exit 1 se houver o que mover).
"""
import argparse
import datetime
import re
import sys
from pathlib import Path

RE_ENTRADA = re.compile(r"^## +(\d{4})-(\d{2})-(\d{2})", re.M)


def separar(texto: str):
    """(cabeçalho, [entradas]) — cada entrada começa em "## "."""
    partes = re.split(r"(?m)^(?=## )", texto)
    if partes and not partes[0].startswith("## "):
        cab, entradas = partes[0], partes[1:]
    else:
        cab, entradas = "", partes
    return cab, entradas


def data_de(entrada: str):
    m = RE_ENTRADA.match(entrada)
    if not m:
        return None
    try:
        return datetime.date(int(m.group(1)), int(m.group(2)), int(m.group(3)))
    except ValueError:
        return None


MIN_ENTRADAS = 3


def gravar(arq: Path, texto: str):
    """Escrita atômica: um .tmp ao lado e replace, para nunca deixar o arquivo pela metade."""
    tmp = arq.with_name(arq.name + ".tmp")
    tmp.write_text(texto, encoding="utf-8")
    tmp.replace(arq)


def arquivar(doc: Path, dias: int = 30, hoje: datetime.date = None, check: bool = False,
             limite: int = 15 * 1024):
    """Devolve {ano: n_entradas} movidas (ou que seriam movidas, com check)."""
    log = doc / "log.md"
    if not log.is_file():
        return {}
    hoje = hoje or datetime.date.today()
    corte = hoje - datetime.timedelta(days=dias)
    cab, entradas = separar(log.read_text(encoding="utf-8"))
    ficam, vao = [], {}
    for e in entradas:
        d = data_de(e)
        if d is not None and d < corte:
            vao.setdefault(d.year, []).append(e)
        else:
            ficam.append(e)
    # ainda grande: fica só o mais recente que couber no limite (entrada sem data nunca sai)
    tamanho = lambda xs: len((cab + "".join(xs)).encode("utf-8"))
    if limite and tamanho(ficam) > limite:
        datadas = sorted((i for i, e in enumerate(ficam) if data_de(e)), key=lambda i: data_de(ficam[i]),
                         reverse=True)
        manter, usado = set(), tamanho([e for e in ficam if not data_de(e)])
        for n, i in enumerate(datadas):
            t = len(ficam[i].encode("utf-8"))
            if n < MIN_ENTRADAS or usado + t <= limite:
                manter.add(i)
                usado += t
            else:
                break
        novas = []
        for i, e in enumerate(ficam):
            if data_de(e) and i not in manter:
                vao.setdefault(data_de(e).year, []).append(e)
            else:
                novas.append(e)
        ficam = novas
    if check or not vao:
        return {a: len(v) for a, v in vao.items()}
    pasta = doc / "log"
    pasta.mkdir(exist_ok=True)
    for ano, lista in vao.items():
        arq = pasta / f"arquivo-{ano}.md"
        cab_arq = f"# Log arquivado — {ano}\n\nEntradas antigas do `log.md`, mais recente no topo.\n\n"
        atual = arq.read_text(encoding="utf-8") if arq.is_file() else cab_arq
        cab_a, antigas = separar(atual)
        juntas = sorted(lista + antigas, key=lambda e: data_de(e) or datetime.date.min, reverse=True)
        gravar(arq, (cab_a or cab_arq) + "".join(x if x.endswith("\n") else x + "\n" for x in juntas))
    nota = "" if "log/arquivo-" in cab else \
        "\nEntradas antigas: `log/arquivo-AAAA.md` (`tools/arquiva_log.py`).\n\n"
    corpo = cab.rstrip("\n") + "\n" + nota if cab else nota
    gravar(log, (corpo if corpo.endswith("\n\n") else corpo + "\n") + "".join(ficam))
    return {a: len(v) for a, v in vao.items()}


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("doc")
    ap.add_argument("--dias", type=int, default=30)
    ap.add_argument("--limite-kb", type=int, default=15)
    ap.add_argument("--check", action="store_true")
    args = ap.parse_args()
    if args.dias < 0 or args.limite_kb < 0:
        ap.error("--dias e --limite-kb precisam ser >= 0")
    doc = Path(args.doc).expanduser().resolve()
    movidas = arquivar(doc, args.dias, check=args.check, limite=args.limite_kb * 1024)
    if not movidas:
        print("Nada a arquivar: o log.md só tem entradas recentes.")
        return 0
    total = sum(movidas.values())
    anos = ", ".join(f"{a}: {n}" for a, n in sorted(movidas.items()))
    if args.check:
        print(f"Arquivaria {total} entrada(s) ({anos}) em {doc.name}/log/arquivo-AAAA.md.")
        return 1
    print(f"Arquivadas {total} entrada(s) ({anos}) em {doc.name}/log/arquivo-AAAA.md. Nada foi apagado.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
