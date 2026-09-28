// Comando ominicore: um binário, subcomandos `api`, `worker` e `migrate`.
// Configuração por variáveis OMINICORE_* (ver deploy/local/README.md); as
// credenciais vêm só do OpenBao (REQ-6).
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"ominicore/internal/adaptadores/entrada/worker"
	"ominicore/internal/adaptadores/saida/keycloak"
	"ominicore/internal/adaptadores/saida/openbao"
	"ominicore/internal/adaptadores/saida/postgres"
	"ominicore/internal/adaptadores/saida/redis"
	"ominicore/internal/adaptadores/saida/sendgrid"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/logs"
	"ominicore/internal/portas"

	entradahttp "ominicore/internal/adaptadores/entrada/http"
)

var version = "dev"

func main() {
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()
	if err := run(ctx, os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "ominicore:", err)
		parar()
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, env func(string) string, saida io.Writer) error {
	if len(args) == 0 {
		return errors.New("uso: ominicore api|worker|migrate|--version")
	}
	if args[0] == "--version" {
		_, err := fmt.Fprintln(saida, version)
		return err
	}
	log := logs.Novo(saida, slog.LevelInfo)
	cfg := lerConfig(env)
	seg, err := carregarSegredos(ctx, cfg)
	if err != nil {
		return err
	}
	pool, err := postgres.Conectar(ctx, seg.PostgresDSN)
	if err != nil {
		return err
	}
	defer pool.Close()
	repo := postgres.Repositorio{Pool: pool}
	switch args[0] {
	case "migrate":
		return postgres.Migrar(ctx, pool)
	case "api":
		return rodarAPI(ctx, cfg, repo, log)
	case "worker":
		return rodarWorker(ctx, cfg, seg, repo, log)
	default:
		return fmt.Errorf("subcomando desconhecido: %s", args[0])
	}
}

type config struct {
	Endereco      string
	KCEmissor     string
	KCPublico     string
	KCClientID    string
	KCSecret      string
	Bao           openbao.Config
	RedisEndereco string
	SendGridURL   string
}

func lerConfig(env func(string) string) config {
	padrao := func(k, p string) string {
		if v := env(k); v != "" {
			return v
		}
		return p
	}
	return config{
		Endereco:   padrao("OMINICORE_ADDR", "0.0.0.0:8080"),
		KCEmissor:  env("OMINICORE_KC_ISSUER"),
		KCPublico:  padrao("OMINICORE_KC_AUDIENCE", "ominicore"),
		KCClientID: padrao("OMINICORE_KC_CLIENT_ID", "ominicore-app"),
		KCSecret:   env("OMINICORE_KC_CLIENT_SECRET"),
		Bao: openbao.Config{
			Endereco: env("OMINICORE_BAO_ADDR"),
			Mount:    padrao("OMINICORE_BAO_MOUNT", "kv"),
			Token:    env("OMINICORE_BAO_TOKEN"),
			Role:     padrao("OMINICORE_BAO_ROLE", "ominicore"),
		},
		RedisEndereco: padrao("OMINICORE_REDIS_ADDR", "redis:6379"),
		SendGridURL:   env("OMINICORE_SENDGRID_URL"),
	}
}

// carregarSegredos entra no OpenBao (token local, ou JWT do Keycloak em hml e
// prod — ADR 0002) e falha fechado: sem segredos o processo não sobe (T-13).
func carregarSegredos(ctx context.Context, cfg config) (openbao.Segredos, error) {
	if cfg.Bao.Endereco == "" {
		return openbao.Segredos{}, fmt.Errorf("%w: OMINICORE_BAO_ADDR vazio", openbao.ErrOpenBao)
	}
	if cfg.Bao.Token == "" {
		jwt, err := keycloak.TokenDeServico(ctx, cfg.KCEmissor, cfg.KCClientID, cfg.KCSecret)
		if err != nil {
			return openbao.Segredos{}, fmt.Errorf("%w: %w", openbao.ErrOpenBao, err)
		}
		cfg.Bao.JWT = jwt
	}
	ctx, cancelar := context.WithTimeout(ctx, 15*time.Second)
	defer cancelar()
	return openbao.Carregar(ctx, cfg.Bao)
}

func rodarAPI(ctx context.Context, cfg config, repo postgres.Repositorio, log *slog.Logger) error {
	if cfg.KCEmissor == "" {
		return errors.New("OMINICORE_KC_ISSUER vazio")
	}
	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.RedisEndereco})
	defer func() { _ = rdb.Close() }()
	srv := &http.Server{
		Addr: cfg.Endereco,
		Handler: entradahttp.Servidor{
			Auth:   keycloak.Novo(cfg.KCEmissor, cfg.KCPublico, redis.Cache{Cliente: rdb}),
			Enviar: aplicacao.SendMessageUseCase{Repo: repo},
			Log:    log,
		}.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		// o ctx já foi cancelado: o prazo do desligamento parte de um contexto sem cancelamento
		desligar, cancelar := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancelar()
		_ = srv.Shutdown(desligar)
	}()
	log.Info("api no ar", "endereco", cfg.Endereco, "versao", version)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func rodarWorker(ctx context.Context, cfg config, seg openbao.Segredos, repo postgres.Repositorio, log *slog.Logger) error {
	uc := aplicacao.DespacharUseCase{
		Repo: repo,
		Remetentes: map[dominio.Canal]portas.MessageSenderPort{
			dominio.CanalEmail: sendgrid.Novo(cfg.SendGridURL, seg.SendGridAPIKey, seg.SendGridRemetente),
		},
		Log: log,
	}
	log.Info("worker no ar", "versao", version)
	worker.Worker{Despachante: uc, Log: log}.Rodar(ctx)
	return nil
}
