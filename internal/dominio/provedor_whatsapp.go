package dominio

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// ProvedorWhatsApp é o Provedor do Canal WhatsApp ativo no ambiente (RN-4,
// ADR 0004): a Meta envia o Modelo aprovado; UltraMsg e Z-API, o texto (Corpo)
// da versão WhatsApp do Template.
type ProvedorWhatsApp string

const (
	WhatsAppMeta     ProvedorWhatsApp = "meta"
	WhatsAppUltraMsg ProvedorWhatsApp = "ultramsg"
	WhatsAppZAPI     ProvedorWhatsApp = "zapi"
)

// LimiteTextoWhatsApp é o tamanho máximo, em caracteres, do texto preenchido
// enviado por um Provedor de texto (REQ-10).
const LimiteTextoWhatsApp = 4096

// ErrProvedorWhatsAppDesconhecido: o cofre indica um Provedor que não é Meta
// nem um dos gateways de texto (REQ-6).
var ErrProvedorWhatsAppDesconhecido = errors.New("provedor whatsapp desconhecido")

// ErrVersaoWhatsAppSemForma: a versão WhatsApp não tem a forma que o Provedor
// ativo envia (REQ-10, REQ-11).
var ErrVersaoWhatsAppSemForma = errors.New("versão whatsapp sem a forma do provedor")

// LerProvedorWhatsApp interpreta o campo `provedor` do cofre; vazio é Meta,
// para quem já tinha a Meta configurada antes do campo existir.
func LerProvedorWhatsApp(s string) (ProvedorWhatsApp, error) {
	switch p := ProvedorWhatsApp(s); p {
	case "":
		return WhatsAppMeta, nil
	case WhatsAppMeta, WhatsAppUltraMsg, WhatsAppZAPI:
		return p, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrProvedorWhatsAppDesconhecido, s)
	}
}

// VersaoWhatsAppSemForma recusa a versão sem a forma do Provedor; errors.Is
// casa com ErrVersaoWhatsAppSemForma.
type VersaoWhatsAppSemForma struct{ Provedor ProvedorWhatsApp }

func (e VersaoWhatsAppSemForma) Error() string {
	if e.Provedor.EnviaTexto() {
		return "versão whatsapp sem texto"
	}
	return "versão whatsapp sem modelo aprovado"
}

// Is liga o erro ao sentinela ErrVersaoWhatsAppSemForma.
func (e VersaoWhatsAppSemForma) Is(alvo error) bool { return alvo == ErrVersaoWhatsAppSemForma }

// EnviaTexto diz se o Provedor envia o texto da versão WhatsApp (UltraMsg e
// Z-API) em vez do Modelo aprovado (Meta).
func (p ProvedorWhatsApp) EnviaTexto() bool {
	return p == WhatsAppUltraMsg || p == WhatsAppZAPI
}

// Servir diz se a versão WhatsApp tem a forma que este Provedor envia.
func (p ProvedorWhatsApp) Servir(c ConteudoCanal) error {
	if (p.EnviaTexto() && c.Corpo == "") || (!p.EnviaTexto() && c.ModeloAprovado == "") {
		return VersaoWhatsAppSemForma{Provedor: p}
	}
	return nil
}

// ServirTexto checa, antes de chamar um Provedor de texto, a forma e o limite
// de caracteres do texto preenchido; o erro é o motivo da Falha definitiva
// (T-76, T-77, T-89).
func (p ProvedorWhatsApp) ServirTexto(c ConteudoCanal) error {
	if err := p.Servir(c); err != nil {
		return err
	}
	if n := utf8.RuneCountInString(c.Corpo); n > LimiteTextoWhatsApp {
		return fmt.Errorf("texto whatsapp com %d caracteres (limite %d)", n, LimiteTextoWhatsApp)
	}
	return nil
}
