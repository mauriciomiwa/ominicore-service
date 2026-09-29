package dominio

import (
	"errors"
	"math/rand/v2"
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
