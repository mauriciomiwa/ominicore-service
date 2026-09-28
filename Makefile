# ominicore — build e verificação. CI: .github/workflows/go.yml.
# Regra: `build` depende de `check` — não sai binário com teste vermelho.

# arm64 = aviva2.0 (t4g, Graviton)
GOARCH ?= arm64
SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
# testcontainers sobre o Podman (VM do podman machine no macOS)
PODMAN_SOCK := $(shell podman machine inspect --format '{{.ConnectionInfo.PodmanSocket.Path}}' 2>/dev/null)

.PHONY: help check fmt vet lint test itest gen build compose-up compose-down contrato

help:
	@echo "make check        - gofmt, go vet, testes unitários (-race)"
	@echo "make lint         - golangci-lint (config em .golangci.yml)"
	@echo "make itest        - testes de integração com testcontainers no Podman"
	@echo "make gen          - regera o código sqlc"
	@echo "make build        - check + binário estático linux/\$$(GOARCH) em dist/"
	@echo "make compose-up   - sobe Postgres, Redis, OpenBao e Keycloak locais (podman compose)"
	@echo "make contrato     - lint do api/openapi.yaml"

check: fmt vet test

fmt:
	@test -z "$$(gofmt -l . | tee /dev/stderr)" || { echo '>> rode: gofmt -w .'; exit 1; }

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test -race ./...

itest:
	@test -n "$(PODMAN_SOCK)" || { echo 'podman machine não encontrada: podman machine start'; exit 1; }
	DOCKER_HOST=unix://$(PODMAN_SOCK) TESTCONTAINERS_RYUK_DISABLED=true go test -race -count=1 ./internal/integracao/ -v

gen:
	sqlc generate

build: check
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -trimpath \
		-ldflags="-s -w -X main.version=$(SHA)" -o dist/ominicore-linux-$(GOARCH) ./cmd/ominicore

compose-up:
	podman compose -f deploy/local/compose.yaml up -d
	./deploy/local/bootstrap.sh

compose-down:
	podman compose -f deploy/local/compose.yaml down

contrato:
	npx --yes @redocly/cli@latest lint api/openapi.yaml
