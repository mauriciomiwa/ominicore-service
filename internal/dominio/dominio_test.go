package dominio

import (
	"errors"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

func boasVindas() Template {
	return Template{
		ID:          "boas-vindas",
		OrdemCanais: []Canal{CanalEmail},
		Conteudos:   map[Canal]ConteudoCanal{CanalEmail: {Assunto: "Bem-vindo, {{nome}}", Corpo: "Olá {{ nome }}, código {{codigo}}."}},
	}
}

func TestTemplateVariaveisSemRepeticao(t *testing.T) {
	got := boasVindas().Variaveis()
	if len(got) != 2 || got[0] != "codigo" || got[1] != "nome" {
		t.Fatalf("variáveis = %v, quero [codigo nome]", got)
	}
}

func TestTemplateValidarDadosCitaAsFaltantes(t *testing.T) {
	err := boasVindas().ValidarDados(map[string]string{})
	if !errors.Is(err, ErrVariavelFaltando) || err.Error() != "variável faltando: codigo, nome" {
		t.Fatalf("erro = %v", err)
	}
}

func TestTemplatePreencher(t *testing.T) {
	c, err := boasVindas().Preencher(CanalEmail, map[string]string{"nome": "Ana", "codigo": "42"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Assunto != "Bem-vindo, Ana" || c.Corpo != "Olá Ana, código 42." {
		t.Fatalf("conteúdo = %+v", c)
	}
}

// notificacao com ordem Email → SMS; telefone opcional.
func notificacao(telefone string) Notificacao {
	return Notificacao{
		OrdemCanais:  []Canal{CanalEmail, CanalSMS},
		Destinatario: Destinatario{Email: "ana@exemplo.com", Telefone: telefone},
	}
}

var (
	temporaria = ResultadoEnvio{Tipo: FalhaTemporaria, Motivo: "sendgrid 503"}
	definitiva = ResultadoEnvio{Tipo: FalhaDefinitiva, Motivo: "sendgrid 400: endereço inválido"}
	meio       = NotificationRouter{Sorteio: func() float64 { return 0.5 }} // jitter zero
)

func TestT12_RecusaDoProvedorEncerraSemOutroCanal(t *testing.T) {
	n := Notificacao{OrdemCanais: []Canal{CanalEmail}, Destinatario: Destinatario{Email: "ana@exemplo.com"}}
	d := meio.Decidir(Tentativa{Canal: CanalEmail, Numero: 1}, n, definitiva)
	if d.Tentativa != TentativaFalhou || d.Notificacao != NotificacaoFalhou || d.Motivo != definitiva.Motivo || d.Proxima != nil {
		t.Fatalf("decisão = %+v", d)
	}
}

func TestT21_FalhaTemporariaAgendaRetentativaNoMesmoCanal(t *testing.T) {
	for _, sorteio := range []float64{0, 0.5, 0.999} {
		r := NotificationRouter{Sorteio: func() float64 { return sorteio }}
		d := r.Decidir(Tentativa{Canal: CanalEmail, Numero: 1}, notificacao("+5511999998888"), temporaria)
		p := d.Proxima
		if d.Notificacao != NotificacaoPendente || p == nil || p.Canal != CanalEmail || p.Numero != 2 ||
			p.Espera < 4*time.Second || p.Espera > 6*time.Second {
			t.Fatalf("sorteio %v: decisão = %+v, próxima = %+v", sorteio, d, p)
		}
	}
}

func TestT22_EsperaCresceACadaRetentativa(t *testing.T) {
	bases := []time.Duration{5 * time.Second, 30 * time.Second, 2 * time.Minute}
	fonte := rand.New(rand.NewPCG(1, 2))
	r := NotificationRouter{Sorteio: fonte.Float64}
	for i := range 1000 {
		numero := i%3 + 1
		d := r.Decidir(Tentativa{Canal: CanalEmail, Numero: numero}, notificacao(""), temporaria)
		base := bases[numero-1]
		minimo, maximo := time.Duration(float64(base)*0.8), time.Duration(float64(base)*1.2)
		if d.Proxima == nil || d.Proxima.Numero != numero+1 || d.Proxima.Espera < minimo || d.Proxima.Espera > maximo {
			t.Fatalf("Tentativa %d: próxima = %+v, quero espera em [%s, %s]", numero, d.Proxima, minimo, maximo)
		}
	}
	if e := meio.espera(1); e != 5*time.Second {
		t.Fatalf("sem jitter a 1ª espera é 5 s, veio %s", e)
	}
}

func TestT23_RetentativasEsgotadasEncerramOCanal(t *testing.T) {
	d := meio.Decidir(Tentativa{Canal: CanalEmail, Numero: 4}, notificacao("+5511999998888"), temporaria)
	if d.Proxima == nil || d.Proxima.Canal != CanalSMS || d.Proxima.Numero != 1 || d.Proxima.Espera != 0 {
		t.Fatalf("depois da 4ª Tentativa quero o SMS, veio %+v", d.Proxima)
	}
	d = meio.Decidir(Tentativa{Canal: CanalSMS, Numero: 4}, notificacao("+5511999998888"), temporaria)
	if d.Proxima != nil || d.Notificacao != NotificacaoFalhou {
		t.Fatalf("SMS é o último Canal: decisão = %+v", d)
	}
}

func TestT24_FalhaDefinitivaNoEmailPassaAoSMS(t *testing.T) {
	d := meio.Decidir(Tentativa{Canal: CanalEmail, Numero: 1}, notificacao("+5511999998888"), definitiva)
	if d.Tentativa != TentativaFalhou || d.Notificacao != NotificacaoPendente ||
		d.Proxima == nil || d.Proxima.Canal != CanalSMS || d.Proxima.Numero != 1 {
		t.Fatalf("decisão = %+v, próxima = %+v", d, d.Proxima)
	}
}

func TestT25_CanalSemContatoEPulado(t *testing.T) {
	d := meio.Decidir(Tentativa{Canal: CanalEmail, Numero: 1}, notificacao(""), definitiva)
	if d.Proxima != nil || d.Notificacao != NotificacaoFalhou {
		t.Fatalf("sem telefone não há SMS: decisão = %+v", d)
	}
}

func TestT26_TodosOsCanaisFalham(t *testing.T) {
	d := meio.Decidir(Tentativa{Canal: CanalSMS, Numero: 1}, notificacao("+5511999998888"), ResultadoEnvio{Tipo: FalhaDefinitiva, Motivo: "twilio 400/21211"})
	if d.Proxima != nil || d.Notificacao != NotificacaoFalhou || d.Tentativa != TentativaFalhou {
		t.Fatalf("decisão = %+v", d)
	}
}

func TestT27_SucessoEncerraOFallback(t *testing.T) {
	d := meio.Decidir(Tentativa{Canal: CanalEmail, Numero: 1}, notificacao("+5511999998888"), ResultadoEnvio{Tipo: Aceito})
	if d.Tentativa != TentativaEnviada || d.Notificacao != NotificacaoEnviada || d.Proxima != nil {
		t.Fatalf("decisão = %+v", d)
	}
}

func TestPrimeiroComContatoPulaCanalSemContato(t *testing.T) {
	d := Destinatario{Telefone: "+5511999998888"}
	if c, ok := d.PrimeiroComContato([]Canal{CanalEmail, CanalSMS}, 0); !ok || c != CanalSMS {
		t.Fatalf("canal = %q ok=%v", c, ok)
	}
	if _, ok := d.PrimeiroComContato([]Canal{CanalEmail, CanalSMS}, 2); ok {
		t.Fatal("além do fim da ordem não há Canal")
	}
}

func TestPreencherWhatsAppDevolveValoresNaOrdemDosParametros(t *testing.T) {
	tpl := Template{ID: "x", Conteudos: map[Canal]ConteudoCanal{
		CanalWhatsApp: {ModeloAprovado: "m", Idioma: "pt_BR", Parametros: []string{"nome", "codigo"}},
	}}
	c, err := tpl.Preencher(CanalWhatsApp, map[string]string{"codigo": "42", "nome": "Ana"})
	if err != nil || c.ModeloAprovado != "m" || c.Idioma != "pt_BR" || len(c.Parametros) != 2 || c.Parametros[0] != "Ana" || c.Parametros[1] != "42" {
		t.Fatalf("conteúdo = %+v err=%v", c, err)
	}
	if got := tpl.Variaveis(); len(got) != 2 {
		t.Fatalf("variáveis = %v", got)
	}
}

func TestAvancaSoParaFrente(t *testing.T) {
	casos := []struct {
		de, para StatusTentativa
		quero    bool
	}{
		{TentativaEnviada, TentativaEntregue, true},
		{TentativaEnviada, TentativaLida, true},
		{TentativaEntregue, TentativaLida, true},
		{TentativaLida, TentativaEntregue, false},
		{TentativaEntregue, TentativaEntregue, false},
		{TentativaEnviada, TentativaFalhou, true},
		{TentativaEntregue, TentativaFalhou, false},
		{TentativaFalhou, TentativaEntregue, false},
		{TentativaPendente, TentativaEntregue, false},
	}
	for _, c := range casos {
		if got := c.de.Avanca(c.para); got != c.quero {
			t.Errorf("%s → %s: %v, quero %v", c.de, c.para, got, c.quero)
		}
	}
}

func TestCategoriaSoNoCanalFalho(t *testing.T) {
	d := meio.Decidir(Tentativa{Canal: CanalEmail, Numero: 1}, notificacao(""), temporaria)
	if d.Categoria != "" {
		t.Fatalf("Retentativa não é Canal falho: categoria %q", d.Categoria)
	}
	if d = meio.Decidir(Tentativa{Canal: CanalEmail, Numero: 4}, notificacao(""), temporaria); d.Categoria != FalhaRetentativasEsgotadas {
		t.Fatalf("4ª Tentativa temporária: categoria %q", d.Categoria)
	}
	if d = meio.Decidir(Tentativa{Canal: CanalEmail, Numero: 1}, notificacao(""), definitiva); d.Categoria != FalhaRecusadoPeloProvedor {
		t.Fatalf("definitiva sem categoria do adaptador: %q", d.Categoria)
	}
}

func TestDecidirAvisoFalhaTardia(t *testing.T) {
	n := notificacao("+5511999998888")
	n.Status = NotificacaoEnviada
	a := meio.DecidirAviso(Tentativa{Canal: CanalEmail, Numero: 1}, n, AvisoStatus{Status: TentativaFalhou, Categoria: FalhaContatoInvalido})
	if a.Tentativa != TentativaFalhou || a.Proxima == nil || a.Proxima.Canal != CanalSMS || a.Categoria != FalhaContatoInvalido {
		t.Fatalf("atualização = %+v", a)
	}
	n.Status = NotificacaoEntregue
	a = meio.DecidirAviso(Tentativa{Canal: CanalEmail, Numero: 1}, n, AvisoStatus{Status: TentativaFalhou})
	if a.Proxima != nil || a.Notificacao != "" {
		t.Fatalf("Notificação entregue não reabre: %+v", a)
	}
}

func TestOptOutTornaCanalIndisponivel(t *testing.T) {
	d := Destinatario{Email: "ana@exemplo.com", Telefone: "+5511999998888", OptOut: map[Canal]bool{CanalSMS: true}}
	if d.Disponivel(CanalSMS) || !d.Disponivel(CanalEmail) || !d.Disponivel(CanalWhatsApp) {
		t.Fatalf("disponível: sms=%v email=%v whatsapp=%v", d.Disponivel(CanalSMS), d.Disponivel(CanalEmail), d.Disponivel(CanalWhatsApp))
	}
	if c, ok := d.PrimeiroComContato([]Canal{CanalSMS, CanalEmail}, 0); !ok || c != CanalEmail {
		t.Fatalf("primeiro = %s", c)
	}
}

func TestNormalizarContato(t *testing.T) {
	if got, err := NormalizarContato(CanalEmail, "  Ana@Exemplo.COM "); err != nil || got != "ana@exemplo.com" {
		t.Fatalf("e-mail: %q %v", got, err)
	}
	for _, c := range []struct {
		canal   Canal
		contato string
	}{{CanalEmail, "sem-arroba"}, {CanalEmail, "@x.com"}, {CanalSMS, "11 99999-8888"}, {"telegram", "x"}} {
		if _, err := NormalizarContato(c.canal, c.contato); !errors.Is(err, ErrContatoInvalido) {
			t.Errorf("%s %q aceito", c.canal, c.contato)
		}
	}
}

func TestProvedorWhatsAppServeSoASuaForma(t *testing.T) {
	soModelo := ConteudoCanal{ModeloAprovado: "boas_vindas", Idioma: "pt_BR"}
	soTexto := ConteudoCanal{Corpo: "Olá {{nome}}"}
	if WhatsAppMeta.Servir(soModelo) != nil || WhatsAppUltraMsg.Servir(soTexto) != nil || WhatsAppZAPI.Servir(soTexto) != nil {
		t.Fatal("recusou a forma do próprio Provedor")
	}
	for _, p := range []ProvedorWhatsApp{WhatsAppUltraMsg, WhatsAppZAPI} {
		if err := p.Servir(soModelo); !errors.Is(err, ErrVersaoWhatsAppSemForma) || err.Error() != "versão whatsapp sem texto" {
			t.Fatalf("%s sem texto: %v", p, err)
		}
	}
	if err := WhatsAppMeta.Servir(soTexto); !errors.Is(err, ErrVersaoWhatsAppSemForma) || err.Error() != "versão whatsapp sem modelo aprovado" {
		t.Fatalf("meta sem modelo: %v", err)
	}
}

func TestLerProvedorWhatsApp(t *testing.T) {
	for s, quero := range map[string]ProvedorWhatsApp{"": WhatsAppMeta, "meta": WhatsAppMeta, "ultramsg": WhatsAppUltraMsg, "zapi": WhatsAppZAPI} {
		if p, err := LerProvedorWhatsApp(s); err != nil || p != quero {
			t.Fatalf("%q → %q, %v", s, p, err)
		}
	}
	if _, err := LerProvedorWhatsApp("twilio"); !errors.Is(err, ErrProvedorWhatsAppDesconhecido) {
		t.Fatalf("erro = %v", err)
	}
}

func TestServirTextoLimitaEmCaracteres(t *testing.T) {
	if err := WhatsAppZAPI.ServirTexto(ConteudoCanal{Corpo: strings.Repeat("ç", LimiteTextoWhatsApp)}); err != nil {
		t.Fatalf("4096 caracteres multibyte recusados: %v", err)
	}
	if err := WhatsAppZAPI.ServirTexto(ConteudoCanal{Corpo: strings.Repeat("ç", LimiteTextoWhatsApp+1)}); err == nil || !strings.Contains(err.Error(), "limite 4096") {
		t.Fatalf("erro = %v", err)
	}
	if err := WhatsAppZAPI.ServirTexto(ConteudoCanal{ModeloAprovado: "m"}); !errors.Is(err, ErrVersaoWhatsAppSemForma) {
		t.Fatalf("erro = %v", err)
	}
}
