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
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"ominicore/internal/adaptadores/entrada/worker"
	"ominicore/internal/adaptadores/saida/keycloak"
	"ominicore/internal/adaptadores/saida/openbao"
	"ominicore/internal/adaptadores/saida/postgres"
	"ominicore/internal/adaptadores/saida/redis"
	"ominicore/internal/adaptadores/saida/sendgrid"
	"ominicore/internal/adaptadores/saida/twilio"
	"ominicore/internal/adaptadores/saida/ultramsg"
	"ominicore/internal/adaptadores/saida/whatsapp"
	"ominicore/internal/adaptadores/saida/zapi"
	"ominicore/internal/aplicacao"
	"ominicore/internal/dominio"
	"ominicore/internal/logs"
	"ominicore/internal/portas"

	entradahttp "ominicore/internal/adaptadores/entrada/http"
	"ominicore/internal/adaptadores/entrada/webhooks"
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
	cofre, seg, err := carregarSegredos(ctx, cfg)
	if err != nil {
		return err
	}
	pool, err := postgres.Conectar(ctx, seg.PostgresDSN, cfg.Schema)
	if err != nil {
		return err
	}
	defer pool.Close()
	if len(seg.ChaveOptOut) < 32 {
		return fmt.Errorf("%w: chave de opt-out com menos de 32 caracteres", openbao.ErrOpenBao)
	}
	repo := postgres.Repositorio{Pool: pool, ChaveOptOut: []byte(seg.ChaveOptOut)}
	switch args[0] {
	case "migrate":
		return postgres.Migrar(ctx, pool, cfg.Schema)
	case "api", "worker":
		releitura, err := time.ParseDuration(cfg.Releitura)
		if err != nil || releitura < time.Second {
			return fmt.Errorf("OMINICORE_CONFIG_RELEITURA inválido: %q (duração ≥ 1s, ex.: 30s)", cfg.Releitura)
		}
		anunciarCanais(log, seg)
		atual := aplicacao.NovaConfigWhatsAppAtual(seg.ConfigWhatsApp(), func(c dominio.ConfigWhatsApp) portas.MessageSenderPort {
			return remetenteWhatsApp(cfg, c)
		})
		go aplicacao.ReleituraConfig{Cofre: cofre, Atual: atual, Log: log}.Rodar(ctx, releitura)
		if args[0] == "api" {
			return rodarAPI(ctx, cfg, seg, atual, cofre, repo, log)
		}
		return rodarWorker(ctx, cfg, seg, atual, repo, log)
	default:
		return fmt.Errorf("subcomando desconhecido: %s", args[0])
	}
}

type config struct {
	Endereco      string
	KCEmissor     string // `iss` conferido no token (host público do Keycloak)
	KCURL         string // backchannel para JWKS e token; vazio = o próprio emissor
	KCPublico     string
	KCClientID    string
	KCSecret      string
	Bao           openbao.Config
	RedisEndereco string
	RedisSenha    string // podman secret redis-password no aviva2.0; vazio no local
	SendGridURL   string
	TwilioURL     string
	WhatsAppURL   string
	UltraMsgURL   string
	ZAPIURL       string
	Releitura     string // OMINICORE_CONFIG_RELEITURA: intervalo da releitura da configuração (REQ-26)
	Schema        string // search_path no Postgres: ominicore (prod) ou ominicore_hml
	WebhookURL    string // origem pública dos webhooks (https://notify-hml.flipapi.com.br); vazia no local
	RetencaoDias  string // OMINICORE_RETENCAO_DIAS, padrão 30 (RN-9)
	ConsoleOrigem string // OMINICORE_CONSOLE_ORIGEM: única origem com CORS em /v1/config/* (REQ-34); vazia = sem CORS
	ConsoleClient string // OMINICORE_CONSOLE_CLIENT: client público do Console (REQ-36, REQ-27)
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
		KCURL:      env("OMINICORE_KC_URL"),
		KCPublico:  padrao("OMINICORE_KC_AUDIENCE", "ominicore"),
		KCClientID: padrao("OMINICORE_KC_CLIENT_ID", "ominicore-app"),
		KCSecret:   env("OMINICORE_KC_CLIENT_SECRET"),
		Bao: openbao.Config{
			Endereco:  env("OMINICORE_BAO_ADDR"),
			Mount:     padrao("OMINICORE_BAO_MOUNT", "kv"),
			Prefixo:   padrao("OMINICORE_BAO_PREFIXO", "ominicore"),
			Token:     env("OMINICORE_BAO_TOKEN"),
			Role:      padrao("OMINICORE_BAO_ROLE", "ominicore"),
			AuthMount: padrao("OMINICORE_BAO_AUTH_MOUNT", "jwt"),
		},
		RedisEndereco: padrao("OMINICORE_REDIS_ADDR", "redis:6379"),
		RedisSenha:    env("OMINICORE_REDIS_PASSWORD"),
		SendGridURL:   env("OMINICORE_SENDGRID_URL"),
		TwilioURL:     env("OMINICORE_TWILIO_URL"),
		WhatsAppURL:   env("OMINICORE_WHATSAPP_URL"),
		UltraMsgURL:   env("OMINICORE_ULTRAMSG_URL"),
		ZAPIURL:       env("OMINICORE_ZAPI_URL"),
		Releitura:     padrao("OMINICORE_CONFIG_RELEITURA", "30s"),
		Schema:        padrao("OMINICORE_DB_SCHEMA", postgres.SchemaPadrao),
		WebhookURL:    strings.TrimRight(env("OMINICORE_WEBHOOK_URL"), "/"),
		RetencaoDias:  padrao("OMINICORE_RETENCAO_DIAS", "30"),
		ConsoleOrigem: strings.TrimRight(env("OMINICORE_CONSOLE_ORIGEM"), "/"),
		ConsoleClient: padrao("OMINICORE_CONSOLE_CLIENT", "ominicore-console"),
	}
}

// realmInterno é o realm pelo backchannel: OMINICORE_KC_URL + o caminho do
// emissor (https://auth.flipapi.com.br/realms/x → http://keycloak:8080/realms/x).
// O host público do Keycloak tem allowlist por IP no Traefik.
func (c config) realmInterno() string {
	if c.KCURL == "" {
		return c.KCEmissor
	}
	u, err := url.Parse(c.KCEmissor)
	if err != nil {
		return c.KCEmissor
	}
	return strings.TrimRight(c.KCURL, "/") + u.Path
}

// carregarSegredos entra no OpenBao (token local, ou JWT do Keycloak em hml e
// prod — ADR 0002) e falha fechado: sem segredos o processo não sobe (T-13).
// O Cofre fica aberto para a releitura e a gravação da configuração do
// WhatsApp, renovando o JWT quando a sessão expira (REQ-26).
func carregarSegredos(ctx context.Context, cfg config) (*openbao.Cofre, openbao.Segredos, error) {
	if cfg.Bao.Endereco == "" {
		return nil, openbao.Segredos{}, fmt.Errorf("%w: OMINICORE_BAO_ADDR vazio", openbao.ErrOpenBao)
	}
	if cfg.Bao.Token == "" {
		cfg.Bao.RenovarJWT = func(ctx context.Context) (string, error) {
			return keycloak.TokenDeServico(ctx, cfg.realmInterno(), cfg.KCClientID, cfg.KCSecret)
		}
	}
	ctx, cancelar := context.WithTimeout(ctx, 15*time.Second)
	defer cancelar()
	cofre, err := openbao.Abrir(ctx, cfg.Bao)
	if err != nil {
		return nil, openbao.Segredos{}, err
	}
	seg, err := cofre.Carregar(ctx)
	return cofre, seg, err
}

func rodarAPI(ctx context.Context, cfg config, seg openbao.Segredos, atual *aplicacao.ConfigWhatsAppAtual, cofre *openbao.Cofre, repo postgres.Repositorio, log *slog.Logger) error {
	if cfg.KCEmissor == "" {
		return errors.New("OMINICORE_KC_ISSUER vazio")
	}
	chaveSG, err := webhooks.ChavePublicaSendGrid(seg.Webhooks["sendgrid_chave_publica"])
	if err != nil {
		return err
	}
	rdb := goredis.NewClient(&goredis.Options{Addr: cfg.RedisEndereco, Password: cfg.RedisSenha})
	defer func() { _ = rdb.Close() }()
	srv := &http.Server{
		Addr: cfg.Endereco,
		Handler: entradahttp.Servidor{
			Auth:      keycloak.NovoComBackchannel(cfg.KCEmissor, cfg.realmInterno(), cfg.KCPublico, redis.Cache{Cliente: rdb}),
			Enviar:    aplicacao.SendMessageUseCase{Repo: repo, Estado: atual},
			Consultar: aplicacao.ConsultarUseCase{Repo: repo},
			OptOut:    aplicacao.OptOutUseCase{Repo: repo},
			Config: &aplicacao.ConfigurarUseCase{
				Cofre: cofre, Registro: repo, Atual: atual, Log: log,
				Cadastro: zapi.NovoWebhooks(cfg.ZAPIURL), URLPublica: cfg.WebhookURL,
			},
			Webhooks: webhooks.Webhooks{
				Cred: webhooks.Credenciais{
					SendGridChavePublica: chaveSG, TwilioAuthToken: seg.Webhooks["twilio_auth_token"],
					URLPublica: cfg.WebhookURL,
				},
				Meta:     metaAtual(atual),
				ZAPI:     zapiAtual(atual),
				Receptor: aplicacao.ReceiveMessageUseCase{Repo: repo, Estado: atual, Log: log},
				OptOut:   aplicacao.OptOutUseCase{Repo: repo},
				Log:      log,
			},
			ConsoleOrigem: cfg.ConsoleOrigem,
			ConsoleClient: cfg.ConsoleClient,
			Log:           log,
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

func rodarWorker(ctx context.Context, cfg config, seg openbao.Segredos, atual *aplicacao.ConfigWhatsAppAtual, repo postgres.Repositorio, log *slog.Logger) error {
	rem := remetentes(cfg, seg)
	rem[dominio.CanalWhatsApp] = atual // segue a configuração em uso (REQ-26)
	uc := aplicacao.DespacharUseCase{
		Repo:       repo,
		Remetentes: rem,
		Estado:     atual,
		Log:        log,
	}
	dias, err := strconv.Atoi(cfg.RetencaoDias)
	if err != nil || dias < 1 {
		return fmt.Errorf("OMINICORE_RETENCAO_DIAS inválido: %q (inteiro ≥ 1)", cfg.RetencaoDias)
	}
	retencao := aplicacao.RetencaoUseCase{Repo: repo, Prazo: time.Duration(dias) * 24 * time.Hour, Log: log}
	go worker.Retencao{Executor: retencao, Log: log}.Rodar(ctx)
	log.Info("worker no ar", "versao", version, "retencao_dias", dias)
	worker.Worker{Despachante: uc, Log: log}.Rodar(ctx)
	return nil
}

// canaisDesligados são os Canais sem credenciais no cofre (REQ-20): hoje só o
// WhatsApp pode faltar.
func canaisDesligados(seg openbao.Segredos) map[dominio.Canal]bool {
	if seg.WhatsAppLigado {
		return nil
	}
	return map[dominio.Canal]bool{dominio.CanalWhatsApp: true}
}

// anunciarCanais registra no boot quais Canais ficaram desligados (T-63) e o
// Provedor do WhatsApp ligado, sem credencial (T-70).
func anunciarCanais(log *slog.Logger, seg openbao.Segredos) {
	if seg.WhatsAppLigado {
		log.Info("whatsapp ligado: "+string(seg.ProvedorWhatsApp), "canal", dominio.CanalWhatsApp, "provedor", seg.ProvedorWhatsApp)
	}
	for c := range canaisDesligados(seg) {
		log.Info(string(c)+" desligado", "canal", c, "motivo", "sem credenciais no openbao")
	}
}

// remetentes liga cada Canal ao seu Provedor; Canal desligado fica sem
// Remetente e a Tentativa pendente nele vira Falha definitiva (T-68).
func remetentes(cfg config, seg openbao.Segredos) map[dominio.Canal]portas.MessageSenderPort {
	r := map[dominio.Canal]portas.MessageSenderPort{
		dominio.CanalEmail: sendgrid.Novo(cfg.SendGridURL, seg.SendGridAPIKey, seg.SendGridRemetente),
		dominio.CanalSMS:   remetenteSMS(cfg, seg),
	}
	if seg.WhatsAppLigado {
		r[dominio.CanalWhatsApp] = remetenteWhatsApp(cfg, seg.ConfigWhatsApp())
	}
	return r
}

// remetenteWhatsApp liga a configuração do WhatsApp ao Remetente do Provedor
// dela; usado no boot e a cada troca de configuração (REQ-26).
func remetenteWhatsApp(cfg config, c dominio.ConfigWhatsApp) portas.MessageSenderPort {
	switch c.Provedor {
	case dominio.WhatsAppUltraMsg:
		return ultramsg.Novo(cfg.UltraMsgURL, ultramsg.Credenciais{InstanceID: c.Campos["instance_id"], Token: c.Campos["token"]})
	case dominio.WhatsAppZAPI:
		return zapi.Novo(cfg.ZAPIURL, zapi.Credenciais{
			InstanceID: c.Campos["instance_id"], Token: c.Campos["token"], ClientToken: c.Campos["client_token"],
		})
	default:
		return whatsapp.Novo(cfg.WhatsAppURL, whatsapp.Credenciais{Token: c.Campos["token"], PhoneNumberID: c.Campos["phone_number_id"]})
	}
}

// metaAtual: as rotas da Meta acompanham o Provedor em uso (REQ-21, T-103).
func metaAtual(a *aplicacao.ConfigWhatsAppAtual) func() webhooks.MetaAtual {
	return func() webhooks.MetaAtual {
		c := a.Config()
		return webhooks.MetaAtual{
			Ligada:    c.Ligado && c.Provedor == dominio.WhatsAppMeta,
			AppSecret: c.Campos["meta_app_secret"], VerifyToken: c.Campos["meta_verify_token"],
		}
	}
}

// zapiAtual dá à rota /webhooks/zapi a instância e o segredo atuais (REQ-37).
func zapiAtual(a *aplicacao.ConfigWhatsAppAtual) func() webhooks.ZAPIAtual {
	return func() webhooks.ZAPIAtual {
		c := a.Config()
		return webhooks.ZAPIAtual{
			Ligada:     c.Ligado && c.Provedor == dominio.WhatsAppZAPI,
			InstanceID: c.Campos["instance_id"], Segredo: c.Campos[dominio.CampoWebhookSegredo],
		}
	}
}

// remetenteSMS pede o Aviso de status ao Twilio quando há URL pública (REQ-12).
func remetenteSMS(cfg config, seg openbao.Segredos) twilio.Remetente {
	r := twilio.Novo(cfg.TwilioURL, twilio.Credenciais{
		AccountSID: seg.Twilio["account_sid"], APIKeySID: seg.Twilio["api_key_sid"],
		APIKeySecret: seg.Twilio["api_key_secret"], MessagingServiceSID: seg.Twilio["messaging_service_sid"],
	})
	if cfg.WebhookURL != "" {
		r.StatusCallback = cfg.WebhookURL + "/webhooks/twilio"
	}
	return r
}
