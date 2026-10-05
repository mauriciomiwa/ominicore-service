#!/usr/bin/env python3
"""Valida a cadeia RN -> REQ -> teste de um projeto SDD com OpenSpec.

Uso:
  python3 valida_rastreabilidade.py <DOC/ ou DOC/<projeto>/> [--openspec DIR] [--prefixo P]

  <DOC>        pasta da documentacao (onde fica Negocio/ e System/)
  --openspec   pasta openspec/ (default: procurada subindo a partir da DOC/)
  --prefixo    so considera capacidades cujo nome comeca com P (monorepo:
               convencao <projeto>-<capacidade>). Default: automatico — se a
               DOC/ e DOC/<projeto>/ e existe capacidade "<projeto>-*", filtra
               por ela. --prefixo "" desliga o filtro.

Le:
  Negocio/especificacao-negocio.md   RNs definidos (RN-n)
  openspec/specs/**/spec.md          REQ vigentes
  openspec/changes/<c>/specs/**/spec.md  deltas das changes ATIVAS (archive/
                                     ja foi mesclado nas specs e e ignorado)
  System/backlog-mvp.md              fatias (Change: `nome`, RN, REQ)
  System/plano-de-testes.md          opcional; testes T-n

Convencao (redacao em openspec/config.yaml do projeto):
  ### Requirement: REQ-12 <nome>
  <frase EARS com SHALL/MUST>
  Origem: RN-3
  #### Scenario: T-20 <nome>

Sem openspec/ mas com System/requisitos-ears.md: roda o modo LEGADO (tabela
EARS) e avisa para migrar.

Degrada com aviso, nunca inventa resultado. Saida em texto simples (sem emoji).
Exit 0 = sem furos; 1 = furos; 2 = nao conseguiu parsear o essencial.
Nao substitui `openspec validate` (formato); este script confere a CADEIA.
"""
import argparse
import contextlib
import io
import json
import re
import sys
from pathlib import Path

RE_RN = re.compile(r"\bRN-\d+\b")
# IDs: REQ-n/T-n (globais) ou REQ-<rn>.<k>/T-<rn>.<k> (prefixo = RN de origem; conferido abaixo)
RE_REQ = re.compile(r"\bREQ-\d+(?:\.\d+)?\b")
RE_TESTE = re.compile(r"\bT-\d+(?:\.\d+)?\b")
RE_SECAO_DELTA = re.compile(r"^##\s+(ADDED|MODIFIED|REMOVED|RENAMED)\s+Requirements\b", re.I)
RE_H2 = re.compile(r"^##\s")
RE_REQ_H = re.compile(r"^###\s+Requirement:\s*(.*?)\s*$")
RE_SCN_H = re.compile(r"^####\s+Scenario:\s*(.*?)\s*$")
RE_NORMATIVO = re.compile(r"\b(SHALL|MUST)\b")
RE_CHANGE_BACKLOG = re.compile(r"^\s*Change:\s*`([^`]+)`", re.M)
RE_DATA_ARQUIVO = re.compile(r"^\d{4}-\d{2}-\d{2}-")
RE_ORIGEM = re.compile(r"^\s*\**Origem\**\s*:", re.I)
RE_RENAME = re.compile(r"^\s*-\s*\**(FROM|TO)\**\s*:\s*`?\s*###\s+Requirement:\s*(.*?)\s*`?\s*$", re.I)


def nome_req(titulo):
    return RE_REQ.sub("", titulo).strip(" -—:").lower()


def num(ident):
    return tuple(int(x) for x in ident.split("-")[1].split("."))


def achar(doc: Path, rel: str):
    p = doc / rel
    if p.is_file():
        return p
    achados = sorted(doc.glob(f"**/{Path(rel).name}"))
    return achados[0] if achados else None


def achar_openspec(doc: Path):
    p = doc.resolve()
    while True:
        cand = p / "openspec"
        if cand.is_dir():
            return cand
        if (p / ".git").exists() or p.parent == p:
            return None
        p = p.parent


# ---------------------------------------------------------------- parse spec
def parse_spec(arquivo: Path, origem: str, cap: str, delta: bool, renames: list):
    """Extrai os requisitos de um spec.md (vigente ou delta).

    renames recebe (id, nome_de, nome_para) das secoes RENAMED."""
    reqs, atual, secao, em_codigo, cerca = [], None, "BASE", False, ""
    de = None
    for n, ln in enumerate(arquivo.read_text(encoding="utf-8").splitlines(), 1):
        marca = ln.lstrip()[:3]
        if marca in ("```", "~~~") and (not em_codigo or marca == cerca):
            em_codigo, cerca = not em_codigo, marca
            continue
        if em_codigo:
            continue
        m = RE_SECAO_DELTA.match(ln)
        if m:
            secao, atual = m.group(1).upper(), None
            continue
        if RE_H2.match(ln):
            # num delta, H2 desconhecido (## Purpose, ## Notes) nao e requisito vigente
            secao, atual = ("IGNORAR" if delta else "BASE"), None
            continue
        if secao == "RENAMED":
            m = RE_RENAME.match(ln)
            if m:
                if m.group(1).upper() == "FROM":
                    de = m.group(2)
                elif de is not None:
                    ids = RE_REQ.findall(de) or RE_REQ.findall(m.group(2))
                    if ids:
                        renames.append((ids[0], nome_req(de), nome_req(m.group(2))))
                    de = None
            continue
        if secao == "IGNORAR":
            continue
        m = RE_REQ_H.match(ln)
        if m:
            titulo = m.group(1)
            ids = RE_REQ.findall(titulo)
            atual = {
                "titulo": titulo,
                "nome": nome_req(titulo),
                "cap": cap,
                "rns_prosa": set(),
                "req": ids[0] if ids else None,
                "secao": secao,
                "rns": set(),
                "testes": [],
                "cenarios": 0,
                "cenarios_sem_id": 0,
                "normativo": False,
                "origem": origem,
                "onde": f"{arquivo}:{n}",
                "no_cenario": False,
            }
            reqs.append(atual)
            continue
        if atual is None:
            continue
        m = RE_SCN_H.match(ln)
        if m:
            atual["cenarios"] += 1
            atual["no_cenario"] = True
            ids_cen = RE_TESTE.findall(m.group(1))
            if not ids_cen:
                atual["cenarios_sem_id"] += 1
            atual["testes"] += ids_cen
            continue
        if RE_ORIGEM.match(ln):
            atual["rns"] |= set(RE_RN.findall(ln))
        elif not atual["no_cenario"]:
            atual["rns_prosa"] |= set(RE_RN.findall(ln))
            if RE_NORMATIVO.search(ln):
                atual["normativo"] = True
    return reqs


def capacidades(raiz_specs: Path):
    """(capacidade, arquivo) para cada spec.md sob raiz_specs."""
    out = []
    if raiz_specs.is_dir():
        for f in sorted(raiz_specs.rglob("spec.md")):
            out.append((f.parent.relative_to(raiz_specs).as_posix(), f))
    return out


# -------------------------------------------------------------- modo openspec
def validar_openspec(doc, openspec, prefixo_arg, rn_definidos, erros, alertas, avisos):
    # prefixo (monorepo)
    todas_caps = {c for c, _ in capacidades(openspec / "specs")}
    changes_dir = openspec / "changes"
    ativas, arquivadas = [], set()
    if changes_dir.is_dir():
        for d in sorted(p for p in changes_dir.iterdir() if p.is_dir()):
            if d.name == "archive":
                arquivadas = {RE_DATA_ARQUIVO.sub("", a.name) for a in d.iterdir() if a.is_dir()}
                continue
            ativas.append(d)
            todas_caps |= {c for c, _ in capacidades(d / "specs")}

    if prefixo_arg is not None:
        prefixo = prefixo_arg
    elif doc.resolve().parent.name.upper() == "DOC" and any(
            c.startswith(f"{doc.resolve().name}-") for c in todas_caps):
        prefixo = f"{doc.resolve().name}-"
    else:
        prefixo = ""
    if prefixo:
        print(f"Filtro de capacidade: '{prefixo}*'")

    def entra(cap):
        return cap.startswith(prefixo)

    # IDs sao unicos no REPOSITORIO: consistencia de ID olha tudo; RN/cobertura
    # olham so as capacidades do projeto (filtro)
    renames = []
    vig_all, del_all = [], []
    for cap, f in capacidades(openspec / "specs"):
        vig_all += parse_spec(f, f"spec {cap}", cap, False, renames)
    changes_com_delta = set()
    for d in ativas:
        for cap, f in capacidades(d / "specs"):
            achados = parse_spec(f, f"change {d.name} ({cap})", cap, True, renames)
            del_all += achados
            if achados or RE_REQ.search(f.read_text(encoding="utf-8")):
                changes_com_delta.add(d.name)
    vigentes = [r for r in vig_all if entra(r["cap"])]
    deltas = [r for r in del_all if entra(r["cap"])]
    todos = vigentes + deltas
    todos_all = vig_all + del_all
    if not todos:
        avisos.append(f"nenhum requisito em {openspec}/specs nem nos deltas das changes ativas"
                      + (f" (filtro '{prefixo}*')" if prefixo else "")
                      + " — projeto sem REQ ainda, ou formato fora do padrao.")

    # --- checagens por ocorrencia ---
    for r in todos:
        if r["secao"] == "REMOVED":
            continue
        rotulo = r["req"] or f"'{r['titulo']}'"
        if not r["req"]:
            erros.append(f"Requisito sem ID REQ-n no titulo: '{r['titulo']}' ({r['onde']}).")
        if not r["rns"]:
            dica = (f" (cita {', '.join(sorted(r['rns_prosa'], key=num))} no texto, mas origem so vale na linha 'Origem:')"
                    if r["rns_prosa"] else "")
            erros.append(f"{rotulo} sem RN de origem — falta a linha 'Origem: RN-n'{dica} ({r['origem']}).")
        for rn in sorted(r["rns"]):
            if rn_definidos and rn not in rn_definidos:
                erros.append(f"{rotulo} referencia {rn}, que nao existe na especificacao de negocio ({r['origem']}).")
        if not r["testes"]:
            erros.append(f"{rotulo} sem cenario de teste — nenhum '#### Scenario: T-n' ({r['origem']}).")
        elif r.get("cenarios_sem_id"):
            erros.append(f"{rotulo}: {r['cenarios_sem_id']} cenario(s) sem ID T-n no titulo — nao vira teste "
                         f"rastreavel ({r['origem']}).")
        if r["req"] and "." in r["req"] and r["rns"]:
            rn_prefixo = "RN-" + r["req"].split("-")[1].split(".")[0]
            if rn_prefixo not in r["rns"]:
                alertas.append(f"{rotulo} tem prefixo de {rn_prefixo} no ID, mas a Origem e "
                               f"{', '.join(sorted(r['rns']))} — ID e origem discordam ({r['origem']}).")
        if not r["normativo"]:
            alertas.append(f"{rotulo} sem SHALL/MUST na frase normativa — o openspec validate vai reclamar ({r['origem']}).")

    # --- consistencia de IDs ---
    ids_vigentes = {r["req"]: r for r in vig_all if r["req"]}
    renomeado = {}
    for rid, de_n, para_n in renames:
        renomeado[(rid, de_n)] = para_n
    por_id = {}
    for r in todos_all:
        if r["req"]:
            por_id.setdefault(r["req"], []).append(r)
    for rid, occ in por_id.items():
        nomes = {renomeado.get((rid, o["nome"]), o["nome"]) for o in occ if o["secao"] != "REMOVED"}
        if len(nomes) > 1:
            onde = "; ".join(f"{o['origem']}: '{o['titulo']}'" for o in occ)
            erros.append(f"{rid} usado para requisitos diferentes — IDs sao unicos e nunca reaproveitados ({onde}).")
        for o in occ:
            if o["secao"] == "ADDED" and rid in ids_vigentes:
                erros.append(f"{rid} aparece como ADDED em {o['origem']}, mas ja e vigente — use MODIFIED ou um ID novo.")
            if o["secao"] in ("MODIFIED", "REMOVED") and rid not in ids_vigentes:
                alertas.append(f"{rid} esta como {o['secao']} em {o['origem']}, mas nao existe nas specs vigentes.")

    dono_teste = {}
    for r in todos_all:
        if r["secao"] == "REMOVED" or not r["req"]:
            continue
        for t in r["testes"]:
            dono_teste.setdefault(t, set()).add(r["req"])
    for t, donos in sorted(dono_teste.items(), key=lambda x: num(x[0])):
        if len(donos) > 1:
            erros.append(f"{t} usado em mais de um requisito ({', '.join(sorted(donos, key=num))}) — cada cenario tem ID proprio.")

    # --- conjunto efetivo (vigente + planejado - removido) ---
    removidos = {r["req"] for r in deltas if r["secao"] == "REMOVED" and r["req"]}
    efetivos = {(r["cap"], r["req"]): r for r in todos
                if r["req"] and r["secao"] != "REMOVED" and r["req"] not in removidos}
    rn_cobertos = set()
    for r in todos:
        if r["req"] and r["secao"] != "REMOVED" and r["req"] not in removidos:
            rn_cobertos |= r["rns"]
    for rn in sorted(rn_definidos - rn_cobertos, key=num):
        alertas.append(f"{rn} nao tem nenhum REQ (nem vigente nem em change ativa) — e MVP sem requisito, ou e 'depois'? Confirme.")

    for d in ativas:
        meta = d / ".openspec.yaml"
        pula = meta.is_file() and re.search(r"(?m)^\s*skip_specs\s*:\s*true\b", meta.read_text(encoding="utf-8"))
        if d.name not in changes_com_delta and not pula:
            alertas.append(f"change '{d.name}' sem delta de spec — se nao muda comportamento, declare skip_specs: true no .openspec.yaml.")

    return {
        "efetivos": efetivos,
        "todos_ids": set(por_id),
        "ids_openspec": {r["req"] for r in todos_all if r["req"]},
        "ativas": {d.name for d in ativas},
        "arquivadas": arquivadas,
        "n_vigentes": len(vigentes),
        "n_deltas": len(deltas),
    }


# ---------------------------------------------------------------- modo legado
def linhas_tabela(texto):
    out = []
    for ln in texto.splitlines():
        s = ln.strip()
        if s.startswith("|") and s.endswith("|") and not re.match(r"^\|[\s:|-]+\|$", s):
            out.append([c.strip() for c in s.strip("|").split("|")])
    return out


def validar_legado(f_ears, rn_definidos, erros, alertas, avisos):
    alertas.append("modo LEGADO (System/requisitos-ears.md): migre para o OpenSpec — "
                   "procedimento em references/estrutura-doc.md da skill novo-projeto-sdd.")
    linhas = linhas_tabela(f_ears.read_text(encoding="utf-8"))
    ncols = max((len(l) for l in linhas), default=0)
    placar = [0] * ncols
    for l in linhas[1:] or linhas:
        for i, c in enumerate(l):
            if re.match(r"^\**REQ-\d+\**$", c):
                placar[i] += 1
    if not ncols or max(placar) == 0:
        avisos.append(f"{f_ears.name}: nao achei coluna com ids REQ-n — cheque a tabela manualmente.")
        return {"efetivos": {}, "todos_ids": set(), "ativas": set(), "arquivadas": set(),
                "n_vigentes": 0, "n_deltas": 0}
    col = placar.index(max(placar))
    efetivos = {}
    for l in linhas:
        if col >= len(l) or not RE_REQ.search(l[col]):
            continue
        rid = RE_REQ.search(l[col]).group(0)
        resto = " | ".join(c for i, c in enumerate(l) if i != col)
        efetivos[rid] = {"rns": set(RE_RN.findall(resto)), "testes": RE_TESTE.findall(resto)}
    rn_cobertos = set()
    for rid in sorted(efetivos, key=num):
        info = efetivos[rid]
        if not info["rns"]:
            erros.append(f"{rid} sem RN de origem.")
        for rn in info["rns"]:
            rn_cobertos.add(rn)
            if rn_definidos and rn not in rn_definidos:
                erros.append(f"{rid} referencia {rn}, que nao existe na especificacao de negocio.")
        if not info["testes"]:
            erros.append(f"{rid} sem teste (nenhum T-n na linha).")
    for rn in sorted(rn_definidos - rn_cobertos, key=num):
        alertas.append(f"{rn} nao tem nenhum REQ — e MVP sem requisito tecnico, ou e 'depois'? Confirme.")
    return {"efetivos": efetivos, "todos_ids": set(efetivos), "ativas": set(), "arquivadas": set(),
            "n_vigentes": len(efetivos), "n_deltas": 0}


# ----------------------------------------------------------------------- main
def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("doc")
    ap.add_argument("--openspec", default=None)
    ap.add_argument("--prefixo", default=None)
    ap.add_argument("--json", action="store_true", help="resultado em JSON (para outros scripts); mesmo exit code")
    args = ap.parse_args()
    if not args.json:
        return validar(args, {})
    saida = {}
    with contextlib.redirect_stdout(io.StringIO()):
        rc = validar(args, saida)
    print(json.dumps({"rc": rc, **saida}, ensure_ascii=False))
    return rc


def validar(args, saida: dict) -> int:
    """Valida e devolve o exit code; em `saida` deixa furos, alertas, avisos, contagens e o resultado."""

    doc = Path(args.doc)
    if not doc.is_dir():
        print(f"ERRO: diretorio nao encontrado: {doc}")
        return 2

    erros, alertas, avisos = [], [], []

    # RNs definidos
    rn_definidos = set()
    f_espec = achar(doc, "Negocio/especificacao-negocio.md")
    if f_espec:
        rn_definidos = set(RE_RN.findall(f_espec.read_text(encoding="utf-8")))
        if not rn_definidos:
            avisos.append(f"{f_espec.name}: nenhum RN-n encontrado — formato fora do padrao? Cheque manualmente.")
    else:
        avisos.append("especificacao-negocio.md nao encontrado — RNs nao verificaveis.")

    # requisitos: OpenSpec ou legado
    openspec = Path(args.openspec) if args.openspec else achar_openspec(doc)
    f_ears = achar(doc, "System/requisitos-ears.md")
    if openspec and openspec.is_dir():
        print(f"OpenSpec: {openspec}")
        res = validar_openspec(doc, openspec, args.prefixo, rn_definidos, erros, alertas, avisos)
        modo = "openspec"
        if f_ears:
            so_ears = set(RE_REQ.findall(f_ears.read_text(encoding="utf-8"))) - res["ids_openspec"]
            for rid in sorted(so_ears, key=num):
                erros.append(f"migracao incompleta: {rid} so existe em {f_ears.name} — leve-o para "
                             "openspec/specs/ (vigente) ou para uma change (planejado).")
    elif f_ears and RE_REQ.search(f_ears.read_text(encoding="utf-8")):
        res = validar_legado(f_ears, rn_definidos, erros, alertas, avisos)
        modo = "legado"
    else:
        avisos.append("nem openspec/ (subindo a partir da DOC/) nem System/requisitos-ears.md — "
                      "cadeia RN->REQ nao verificavel. Use --openspec se a pasta estiver em outro lugar.")
        res = {"efetivos": {}, "todos_ids": set(), "ativas": set(), "arquivadas": set(),
               "n_vigentes": 0, "n_deltas": 0}
        modo = "nenhum"

    # backlog
    f_backlog = achar(doc, "System/backlog-mvp.md")
    if f_backlog:
        texto = f_backlog.read_text(encoding="utf-8")
        for rn in sorted(set(RE_RN.findall(texto)), key=num):
            if rn_definidos and rn not in rn_definidos:
                erros.append(f"backlog-mvp.md cita {rn}, que nao existe na especificacao.")
        # por fatia: REQ de fatia cuja change ainda nao existe e plano, nao furo
        existentes = res["ativas"] | res["arquivadas"]
        for bloco in re.split(r"(?m)^(?=#{2,4}\s)", texto):
            mc = RE_CHANGE_BACKLOG.search(bloco)
            futura = modo == "openspec" and mc and "<" not in mc.group(1) and mc.group(1) not in existentes
            for rid in sorted(set(RE_REQ.findall(bloco)), key=num):
                if res["todos_ids"] and rid not in res["todos_ids"]:
                    if futura:
                        alertas.append(f"backlog-mvp.md cita {rid} na fatia da change '{mc.group(1)}', "
                                       "que ainda nao foi aberta — o REQ nasce nela.")
                    else:
                        erros.append(f"backlog-mvp.md cita {rid}, que nao existe nas specs nem nas changes ativas.")
        if modo == "openspec":
            citadas = set(RE_CHANGE_BACKLOG.findall(texto))
            for c in sorted(citadas):
                if "<" in c:
                    continue  # placeholder do template
                if c not in res["ativas"] and c not in res["arquivadas"]:
                    alertas.append(f"backlog-mvp.md cita a change '{c}', que ainda nao existe em openspec/changes/.")
            for c in sorted(res["ativas"] - citadas):
                alertas.append(f"change ativa '{c}' nao aparece no backlog-mvp.md (ok se for pos-MVP).")
    else:
        avisos.append("backlog-mvp.md nao encontrado — fatias nao verificaveis.")

    # plano de testes (opcional)
    f_plano = achar(doc, "System/plano-de-testes.md")
    if f_plano and res["efetivos"]:
        no_plano = set(RE_TESTE.findall(f_plano.read_text(encoding="utf-8")))
        nas_specs = set()
        for r in res["efetivos"].values():
            nas_specs |= set(r["testes"])
        for t in sorted(nas_specs - no_plano, key=num):
            alertas.append(f"{t} esta nas specs mas nao no plano-de-testes.md.")

    # relatorio
    saida.update(modo=modo, rns=len(rn_definidos), reqs=len(res["efetivos"]),
                 furos=list(erros), alertas=list(alertas), avisos=list(avisos))

    def fim(rc, msg):
        print(f"\nResultado: {msg}")
        saida["resultado"] = msg
        return rc

    print("=== Validacao de rastreabilidade RN -> REQ -> teste ===")
    print(f"Modo: {modo} | RNs definidos: {len(rn_definidos)} | REQ efetivos: {len(res['efetivos'])}"
          + (f" (vigentes: {res['n_vigentes']}, em deltas de changes ativas: {res['n_deltas']})" if modo == "openspec" else ""))
    for lista, titulo in ((erros, "FUROS"), (alertas, "ALERTAS"), (avisos, "PARSE")):
        if lista:
            print(f"\n[{titulo}]")
            for item in lista:
                print(f"- {item}")
    if not erros and not alertas and not avisos:
        print("\nCadeia integra: todo REQ tem RN e cenario de teste; backlog consistente.")
    if erros:
        return fim(1, f"{len(erros)} furo(s). Registre no to-do.md o que nao for corrigir agora.")
    if res["efetivos"] and not rn_definidos:
        return fim(2, "ha REQ mas nenhum RN verificavel (especificacao-negocio.md ausente ou sem RN-n) — "
                      "cadeia NAO verificada; faca a checagem manual e diga isso no relato.")
    if modo == "openspec" and not f_espec and not f_backlog and not res["todos_ids"]:
        # DOC honesta sem RN, REQ nem fatias (template vazio apagado): nada a rastrear, nao e falha de parse
        return fim(0, "nada a rastrear ainda — sem RN, REQ nem fatias.")
    if avisos and not res["efetivos"] and not rn_definidos:
        return fim(2, "nao consegui parsear o essencial — faca a checagem manual e diga isso no relato.")
    return fim(0, f"sem furos bloqueantes ({len(alertas)} alerta(s)).")


if __name__ == "__main__":
    sys.exit(main())
