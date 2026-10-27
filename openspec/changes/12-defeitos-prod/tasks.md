# Tasks

## 1. Pré-requisitos
- [ ] 1.1 Conferir que a 08, a 09, a 10 e a 11 estão arquivadas (`console-operador` vigente) — `openspec validate 12-defeitos-prod --strict` sem o aviso de "Archive would refuse"

## 2. Desligar o WhatsApp (REQ-24)
- [x] 2.1 Teste T-98 contra o OpenBao 2.6.3 real (testcontainers, modo dev) em `internal/integracao/`: grava a Z-API, desliga pela API, a consulta mostra desligado e a releitura não dá erro — falha antes da correção (`make itest`)
- [x] 2.2 `LerWhatsApp` trata versão com `Data` vazio ou `DeletionTime` preenchido como desligado — `go test ./internal/adaptadores/saida/openbao/` e o T-98 de integração passando
- [x] 2.3 Rodar de novo T-92 a T-105 — `go test -race ./...`

## 3. Reinício dos serviços
- [x] 3.1 `deploy/aviva2/quadlet/ominicore.container.tmpl`: `StartLimitIntervalSec=0` em `[Unit]` e `RestartSec=10s` em `[Service]`; runbook `deploy/aviva2/README.md` (conferência com `systemctl show`) — arquivo gerado pelo `publicar.sh` conferido com o `perl` do script

- [x] 3.2 Healthcheck na forma de array (`HealthCmd=["/ominicore","--version"]`): a forma de texto vira `CMD-SHELL`, e a imagem distroless não tem `/bin/sh`, então o container ficava sempre "unhealthy" (visto em prod em 2026-10-27; incluído na change a pedido do dono) — arquivo gerado conferido com o `perl` do `publicar.sh`
- [ ] 3.3 Depois do `publicar.sh prod`, conferir no host `podman healthcheck run ominicore-api` com código 0 e status "healthy" nos dois containers

## 4. Console com a API fora (REQ-43)
- [x] 4.1 `api.ts` devolve `falha` com status 0 quando o `fetch` lança; a tela mostra `semResposta`/`falha` com "Tentar de novo" e encerra o carregamento em `finally`; textos em `pt-BR.json` — `bun run lint` e `bun run typecheck`
- [x] 4.2 Teste T-137 (rede falha, depois "Tentar de novo" mostra a configuração; sem carregamento infinito)
- [x] 4.3 Teste T-138 (503 mostra o status e "Tentar de novo")

## 5. Documentação e checagens
- [x] 5.1 `DOC/System/plano-de-testes.md` (T-137, T-138 e o T-98 de integração), `DOC/to-do.md` (os três itens fechados) e `DOC/log.md` — linhas presentes
- [x] 5.2 `go test -race ./...`, `make itest`, `golangci-lint run`, `bun run lint`, `bun run test` e `bun run build` passando

## 6. Fechamento
- [x] 6.1 Rodar `openspec validate --all --strict` e `python3 DOC/tools/valida_rastreabilidade.py DOC` sem furo
- [x] 6.2 Auditar o diff com a skill sentinel
