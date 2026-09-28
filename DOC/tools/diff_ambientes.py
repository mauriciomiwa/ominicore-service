#!/usr/bin/env python3
"""
diff_ambientes.py — compara as CHAVES de configuracao declaradas entre dois
ambientes (tipicamente hml e prod) e reporta o que falta de cada lado.

Principio: compara NOMES, nunca imprime VALORES. Um valor de producao nao pode
vazar para o terminal, o log de CI nem a resposta de um agente. A unica coisa
que o script diz sobre valores e se dois sao identicos entre ambientes — o que,
para um segredo, e um achado e nao uma conveniencia.

Formatos aceitos (detectados pela extensao e pelo conteudo):
  .env / .env.example / .env.hml      KEY=valor
  .yaml / .yml                        chave: valor  (nivel raiz e aninhado)
  .hcl / .tf                          blocos variable "nome" {...}
  .json                               objeto plano ou aninhado

Uso:
    python3 diff_ambientes.py <arquivo-hml> <arquivo-prod>
    python3 diff_ambientes.py .env.hml .env.prod --rotulos hml prod
    python3 diff_ambientes.py a.env b.env --valores-iguais
    python3 diff_ambientes.py a.env b.env --ignorar 'DEBUG_*' --ignorar LOCAL_PORT

Saida (codigo de retorno):
    0  conjuntos de chaves identicos
    1  ha divergencia (chave so de um lado)
    2  erro de uso (arquivo ausente, formato nao reconhecido)
"""

from __future__ import annotations

import argparse
import fnmatch
import json
import re
import sys
from pathlib import Path

# --------------------------------------------------------------------------
# Parsers — cada um devolve dict {chave: valor_bruto}
# --------------------------------------------------------------------------


def parse_env(texto: str) -> dict[str, str]:
    """KEY=valor, uma por linha. Ignora comentario, linha vazia e 'export '."""
    saida: dict[str, str] = {}
    for linha in texto.splitlines():
        linha = linha.strip()
        if not linha or linha.startswith("#"):
            continue
        if linha.startswith("export "):
            linha = linha[len("export "):].strip()
        if "=" not in linha:
            continue
        chave, valor = linha.split("=", 1)
        chave = chave.strip()
        if not chave:
            continue
        saida[chave] = _desaspar(valor.strip())
    return saida


def parse_yaml(texto: str) -> dict[str, str]:
    """
    YAML por indentacao, sem dependencia externa. Cobre mapeamentos aninhados,
    que e o caso de values.yaml. Listas viram uma chave so (o conteudo da lista
    nao e comparado — o que importa aqui e a chave existir).

    Cuidado que custou um bug: `sentry_dsn:` sem valor pode ser bloco pai OU
    chave declarada vazia. Distinguir exige olhar a proxima linha util — a
    chave vazia e justamente o caso que este script existe para pegar, entao
    nao da para tratar toda linha sem valor como pai.
    """
    # Passo 1: linhas uteis, com indentacao.
    linhas: list[tuple[int, str, str]] = []  # (indent, nome, valor)
    for bruta in texto.splitlines():
        conteudo = bruta.strip()
        if not conteudo or conteudo.startswith("#") or conteudo.startswith("- "):
            continue
        if ":" not in conteudo:
            continue
        nome, _, valor = conteudo.partition(":")
        nome = nome.strip()
        if not nome:
            continue
        valor = valor.strip()
        if valor.startswith("#"):
            valor = ""
        linhas.append((len(bruta) - len(bruta.lstrip()), nome, valor))

    # Passo 2: e pai quem nao tem valor E tem filho mais indentado logo abaixo.
    saida: dict[str, str] = {}
    pilha: list[tuple[int, str]] = []
    for i, (indent, nome, valor) in enumerate(linhas):
        while pilha and pilha[-1][0] >= indent:
            pilha.pop()
        caminho = ".".join([n for _, n in pilha] + [nome])

        tem_filho = i + 1 < len(linhas) and linhas[i + 1][0] > indent
        if valor == "" and tem_filho:
            pilha.append((indent, nome))
        else:
            saida[caminho] = _desaspar(valor.split(" #")[0].strip())
    return saida


def _corpo_bloco(texto: str, inicio: int) -> tuple[str, int]:
    """
    Devolve o conteudo entre a chave aberta em `inicio` e a que a fecha, por
    contagem — nao por regex. Regex com `.*?\\n\\}` so acerta bloco na coluna 0
    e engole blocos aninhados; num job Nomad indentado isso perde variaveis
    silenciosamente, que e o pior defeito possivel neste script.
    """
    profundidade, i, dentro_str, n = 0, inicio, False, len(texto)
    while i < n:
        c = texto[i]
        if dentro_str:
            if c == "\\":
                i += 2
                continue
            if c == '"':
                dentro_str = False
        elif c == '"':
            dentro_str = True
        elif c == "#":
            while i < n and texto[i] != "\n":
                i += 1
        elif c == "{":
            profundidade += 1
        elif c == "}":
            profundidade -= 1
            if profundidade == 0:
                return texto[inicio + 1:i], i + 1
        i += 1
    return texto[inicio + 1:], n


def parse_hcl(texto: str) -> dict[str, str]:
    """
    Blocos variable "nome" { default = "x" } do HCL2/Terraform, em qualquer
    indentacao e inclusive numa linha so. Pega tambem atribuicoes simples em
    blocos env { CHAVE = "valor" }.
    """
    saida: dict[str, str] = {}

    for m in re.finditer(r'variable\s+"([^"]+)"\s*(\{)', texto):
        corpo, _ = _corpo_bloco(texto, m.start(2))
        d = re.search(r"default\s*=\s*([^\n}]+)", corpo)
        saida[m.group(1)] = _desaspar(d.group(1).strip()) if d else ""

    for m in re.finditer(r"env\s*(\{)", texto):
        corpo, _ = _corpo_bloco(texto, m.start(1))
        for linha in corpo.splitlines():
            linha = linha.strip()
            if not linha or linha.startswith("#") or "=" not in linha:
                continue
            chave, valor = linha.split("=", 1)
            chave = chave.strip()
            if re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", chave):
                saida.setdefault(chave, _desaspar(valor.strip()))
    return saida


def parse_json(texto: str) -> dict[str, str]:
    dados = json.loads(texto)
    if not isinstance(dados, dict):
        raise ValueError("JSON precisa ter um objeto no topo")

    saida: dict[str, str] = {}

    def caminhar(obj: dict, prefixo: str = "") -> None:
        for chave, valor in obj.items():
            completo = f"{prefixo}{chave}"
            if isinstance(valor, dict):
                caminhar(valor, f"{completo}.")
            else:
                saida[completo] = "" if valor is None else str(valor)

    caminhar(dados)
    return saida


def _desaspar(valor: str) -> str:
    if len(valor) >= 2 and valor[0] == valor[-1] and valor[0] in "\"'":
        return valor[1:-1]
    return valor


# --------------------------------------------------------------------------
# Carga
# --------------------------------------------------------------------------


def carregar(caminho: Path) -> dict[str, str]:
    if not caminho.is_file():
        erro(f"arquivo nao encontrado: {caminho}")

    texto = caminho.read_text(encoding="utf-8", errors="replace")
    nome = caminho.name.lower()
    sufixos = "".join(caminho.suffixes).lower()

    if nome.endswith(".json") or sufixos.endswith(".json"):
        try:
            return parse_json(texto)
        except (json.JSONDecodeError, ValueError) as exc:
            erro(f"{caminho}: JSON invalido — {exc}")
    if re.search(r"\.(ya?ml)(\.|$)", nome):
        return parse_yaml(texto)
    if re.search(r"\.(hcl|tf)(\.|$)", nome):
        return parse_hcl(texto)
    if "env" in nome or sufixos == "":
        return parse_env(texto)

    # Ultimo recurso: cheirar o conteudo.
    if re.search(r"^\s*variable\s+\"", texto, re.M):
        return parse_hcl(texto)
    if re.search(r"^\s*[\w.-]+\s*:\s", texto, re.M) and "=" not in texto:
        return parse_yaml(texto)
    return parse_env(texto)


def erro(msg: str) -> None:
    print(f"ERRO: {msg}", file=sys.stderr)
    sys.exit(2)


# --------------------------------------------------------------------------
# Relatorio
# --------------------------------------------------------------------------


def filtrar(chaves: set[str], padroes: list[str]) -> set[str]:
    if not padroes:
        return chaves
    return {c for c in chaves if not any(fnmatch.fnmatch(c, p) for p in padroes)}


def main() -> int:
    ap = argparse.ArgumentParser(
        description="Compara chaves de configuracao entre dois ambientes. "
                    "Nunca imprime valores.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    ap.add_argument("arquivo_a", type=Path)
    ap.add_argument("arquivo_b", type=Path)
    ap.add_argument(
        "--rotulos", nargs=2, metavar=("A", "B"), default=None,
        help="nomes dos ambientes no relatorio (padrao: os nomes dos arquivos)",
    )
    ap.add_argument(
        "--ignorar", action="append", default=[], metavar="PADRAO",
        help="chave a ignorar; aceita glob (DEBUG_*). Pode repetir.",
    )
    ap.add_argument(
        "--valores-iguais", action="store_true",
        help="tambem aponta chaves cujo valor e identico nos dois ambientes "
             "(para segredo, e um achado). Os valores nao sao impressos.",
    )
    args = ap.parse_args()

    a, b = carregar(args.arquivo_a), carregar(args.arquivo_b)
    rot_a, rot_b = args.rotulos or (args.arquivo_a.name, args.arquivo_b.name)

    chaves_a = filtrar(set(a), args.ignorar)
    chaves_b = filtrar(set(b), args.ignorar)

    so_a = sorted(chaves_a - chaves_b)
    so_b = sorted(chaves_b - chaves_a)
    comuns = sorted(chaves_a & chaves_b)

    print(f"Comparando chaves declaradas")
    print(f"  {rot_a}: {args.arquivo_a}  ({len(chaves_a)} chaves)")
    print(f"  {rot_b}: {args.arquivo_b}  ({len(chaves_b)} chaves)")
    if args.ignorar:
        print(f"  ignorando: {', '.join(args.ignorar)}")
    print()

    if so_a:
        print(f"BLOQUEANTE  so em {rot_a} — falta em {rot_b} ({len(so_a)}):")
        for c in so_a:
            print(f"  - {c}")
        print()

    if so_b:
        print(f"BLOQUEANTE  so em {rot_b} — falta em {rot_a} ({len(so_b)}):")
        for c in so_b:
            print(f"  - {c}")
        print()

    if args.valores_iguais:
        iguais = [c for c in comuns if a[c] and a[c] == b[c]]
        if iguais:
            print(f"ALERTA  valor identico nos dois ambientes ({len(iguais)}):")
            for c in iguais:
                print(f"  - {c}")
            print("  Se alguma dessas for credencial, o mesmo segredo esta em")
            print("  hml e prod: comprometer um ambiente compromete os dois.")
            print()

    print(f"Presentes nos dois: {len(comuns)}")
    print()

    if so_a or so_b:
        print("VEREDITO: nao promove — toda chave precisa existir nos dois")
        print("ambientes, declarada, mesmo que o valor mude. Crie a que falta")
        print("com valor e dono definidos, ou registre a ausencia como")
        print("divergencia deliberada no contrato (DOC/Infra/AMBIENTES.md).")
        return 1

    print("VEREDITO: conjuntos de chaves identicos.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
