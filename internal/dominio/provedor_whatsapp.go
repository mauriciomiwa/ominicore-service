package dominio

import (
	"errors"
	"fmt"
)

// ProvedorWhatsApp é o Provedor do Canal WhatsApp ativo no ambiente (RN-4,
// ADR 0004): a Meta envia o Modelo aprovado; o UltraMsg, o texto (Corpo) da
// versão WhatsApp do Template.
type ProvedorWhatsApp string

const (
	WhatsAppMeta     ProvedorWhatsApp = "meta"
	WhatsAppUltraMsg ProvedorWhatsApp = "ultramsg"
)

// ErrProvedorWhatsAppDesconhecido: o cofre indica um Provedor que não é Meta
// nem UltraMsg (REQ-6).
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
	case WhatsAppMeta, WhatsAppUltraMsg:
		return p, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrProvedorWhatsAppDesconhecido, s)
	}
}

// VersaoWhatsAppSemForma recusa a versão sem a forma do Provedor; errors.Is
// casa com ErrVersaoWhatsAppSemForma.
type VersaoWhatsAppSemForma struct{ Provedor ProvedorWhatsApp }

func (e VersaoWhatsAppSemForma) Error() string {
	if e.Provedor == WhatsAppUltraMsg {
		return "versão whatsapp sem texto"
	}
	return "versão whatsapp sem modelo aprovado"
}

// Is liga o erro ao sentinela ErrVersaoWhatsAppSemForma.
func (e VersaoWhatsAppSemForma) Is(alvo error) bool { return alvo == ErrVersaoWhatsAppSemForma }

// Servir diz se a versão WhatsApp tem a forma que este Provedor envia.
func (p ProvedorWhatsApp) Servir(c ConteudoCanal) error {
	if (p == WhatsAppUltraMsg && c.Corpo == "") || (p != WhatsAppUltraMsg && c.ModeloAprovado == "") {
		return VersaoWhatsAppSemForma{Provedor: p}
	}
	return nil
}
