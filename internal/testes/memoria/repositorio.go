// Package memoria é um MessageRepositoryPort em memória, só para testes que
// não precisam do Postgres. A versão real (e a idempotência sob concorrência)
// é provada nos testes de integração do adaptador postgres.
package memoria

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// Repositorio guarda tudo em mapas protegidos por mutex.
type Repositorio struct {
	mu           sync.Mutex
	Templates    map[string]dominio.Template
	Notificacoes map[string]dominio.Notificacao
	Tentativas   []dominio.Tentativa
	disponivel   []time.Time // disponivel_em de cada Tentativa (a fila respeita)
	idProvedor   []string    // id_provedor de cada Tentativa aceita
	chaves       map[string]string
}

var _ portas.MessageRepositoryPort = (*Repositorio)(nil)

// Novo cria o repositório com o Template boas-vindas das migrations 00002/00003.
func Novo() *Repositorio {
	return &Repositorio{
		Templates:    map[string]dominio.Template{"boas-vindas": BoasVindas()},
		Notificacoes: map[string]dominio.Notificacao{},
		chaves:       map[string]string{},
	}
}

// BoasVindas espelha o seed das migrations 00002, 00003 e 00004.
func BoasVindas() dominio.Template {
	return dominio.Template{
		ID:          "boas-vindas",
		OrdemCanais: []dominio.Canal{dominio.CanalEmail, dominio.CanalSMS},
		Conteudos: map[dominio.Canal]dominio.ConteudoCanal{
			dominio.CanalEmail:    {Assunto: "Bem-vindo, {{nome}}", Corpo: "Olá {{nome}}, sua conta está pronta."},
			dominio.CanalSMS:      {Corpo: "Olá {{nome}}, sua conta está pronta."},
			dominio.CanalWhatsApp: {ModeloAprovado: "boas_vindas", Idioma: "pt_BR", Parametros: []string{"nome"}},
		},
	}
}

func (r *Repositorio) BuscarTemplate(_ context.Context, id string) (dominio.Template, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.Templates[id]
	if !ok {
		return dominio.Template{}, fmt.Errorf("%w: %s", dominio.ErrTemplateInexistente, id)
	}
	return t, nil
}

func (r *Repositorio) RegistrarNotificacao(_ context.Context, n portas.NovaNotificacao) (portas.Registro, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	chave := n.SistemaOrigem + "\x00" + n.ChaveIdempotencia
	if id, ok := r.chaves[chave]; ok {
		return portas.Registro{ID: id, Status: r.Notificacoes[id].Status, Reutilizada: true}, nil
	}
	id := fmt.Sprintf("00000000-0000-4000-8000-%012d", len(r.Notificacoes)+1) // formato UUID, como no Postgres
	r.chaves[chave] = id
	r.Notificacoes[id] = dominio.Notificacao{
		ID: id, SistemaOrigem: n.SistemaOrigem, TemplateID: n.TemplateID, Destinatario: n.Destinatario,
		Dados: n.Dados, OrdemCanais: n.OrdemCanais, Status: dominio.NotificacaoPendente,
	}
	r.enfileirar(id, n.PrimeiroCanal, 1, 0)
	return portas.Registro{ID: id, Status: dominio.NotificacaoPendente}, nil
}

// enfileirar exige r.mu travado.
func (r *Repositorio) enfileirar(notificacao string, c dominio.Canal, numero int, espera time.Duration) {
	r.Tentativas = append(r.Tentativas, dominio.Tentativa{
		ID: int64(len(r.Tentativas) + 1), NotificacaoID: notificacao, Canal: c, Numero: numero, Status: dominio.TentativaPendente,
	})
	r.disponivel = append(r.disponivel, time.Now().Add(espera))
	r.idProvedor = append(r.idProvedor, "")
}

func (r *Repositorio) ReservarTentativa(_ context.Context) (portas.TentativaReservada, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, t := range r.Tentativas {
		if t.Status != dominio.TentativaPendente || time.Now().Before(r.disponivel[i]) {
			continue
		}
		i := i
		return portas.TentativaReservada{
			Tentativa:   t,
			Notificacao: r.Notificacoes[t.NotificacaoID],
			Concluir: func(_ context.Context, d dominio.Decisao, res dominio.ResultadoEnvio) error {
				r.mu.Lock()
				defer r.mu.Unlock()
				r.Tentativas[i].Status, r.Tentativas[i].Motivo, r.Tentativas[i].Categoria = d.Tentativa, d.Motivo, d.Categoria
				r.idProvedor[i] = res.IDProvedor
				if p := d.Proxima; p != nil {
					r.enfileirar(t.NotificacaoID, p.Canal, p.Numero, p.Espera)
				}
				n := r.Notificacoes[t.NotificacaoID]
				n.Status = d.Notificacao
				r.Notificacoes[t.NotificacaoID] = n
				return nil
			},
		}, true, nil
	}
	return portas.TentativaReservada{}, false, nil
}

// Contar devolve quantas Notificações e Tentativas existem.
func (r *Repositorio) Contar() (notificacoes, tentativas int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Notificacoes), len(r.Tentativas)
}

// StatusTentativa devolve o status da i-ésima Tentativa ("" se não existir).
func (r *Repositorio) StatusTentativa(i int) dominio.StatusTentativa {
	r.mu.Lock()
	defer r.mu.Unlock()
	if i >= len(r.Tentativas) {
		return ""
	}
	return r.Tentativas[i].Status
}

// Historico devolve uma cópia das Tentativas, na ordem de criação.
func (r *Repositorio) Historico() []dominio.Tentativa {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]dominio.Tentativa(nil), r.Tentativas...)
}

// StatusNotificacao devolve o status da Notificação.
func (r *Repositorio) StatusNotificacao(id string) dominio.StatusNotificacao {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Notificacoes[id].Status
}

// ReservarPorIDProvedor acha a Tentativa pelo id do Provedor. O mutex é
// liberado entre a reserva e o Aplicar — suficiente para testes de um aviso
// por vez; a concorrência real é provada no Postgres.
func (r *Repositorio) ReservarPorIDProvedor(_ context.Context, id string) (portas.AvisoReservado, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, t := range r.Tentativas {
		if id == "" || r.idProvedor[i] != id {
			continue
		}
		return portas.AvisoReservado{
			Tentativa: t, Notificacao: r.Notificacoes[t.NotificacaoID],
			Aplicar: func(_ context.Context, a dominio.AtualizacaoAviso) error {
				r.mu.Lock()
				defer r.mu.Unlock()
				r.Tentativas[i].Status, r.Tentativas[i].Categoria = a.Tentativa, a.Categoria
				if a.Motivo != "" {
					r.Tentativas[i].Motivo = a.Motivo
				}
				if p := a.Proxima; p != nil {
					r.enfileirar(t.NotificacaoID, p.Canal, p.Numero, p.Espera)
				}
				if a.Notificacao != "" {
					n := r.Notificacoes[t.NotificacaoID]
					n.Status = a.Notificacao
					r.Notificacoes[t.NotificacaoID] = n
				}
				return nil
			},
			Liberar: func(context.Context) {},
		}, true, nil
	}
	return portas.AvisoReservado{}, false, nil
}

// ConsultarNotificacao devolve a Notificação só para o Sistema de origem dono.
func (r *Repositorio) ConsultarNotificacao(_ context.Context, sistemaOrigem, id string) (portas.Consulta, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	n, ok := r.Notificacoes[id]
	if !ok || n.SistemaOrigem != sistemaOrigem {
		return portas.Consulta{}, false, nil
	}
	c := portas.Consulta{ID: n.ID, TemplateID: n.TemplateID, Status: n.Status}
	for _, t := range r.Tentativas {
		if t.NotificacaoID == id {
			c.Tentativas = append(c.Tentativas, portas.TentativaConsultada{Canal: t.Canal, Numero: t.Numero, Status: t.Status, Categoria: t.Categoria})
		}
	}
	return c, true, nil
}

// MarcarAceita simula o envio aceito pelo Provedor com o id dado (testes de aviso).
func (r *Repositorio) MarcarAceita(i int, idProvedor string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Tentativas[i].Status = dominio.TentativaEnviada
	r.idProvedor[i] = idProvedor
	n := r.Notificacoes[r.Tentativas[i].NotificacaoID]
	n.Status = dominio.NotificacaoEnviada
	r.Notificacoes[n.ID] = n
}

// AdicionarAceita põe na Notificação mais uma Tentativa já aceita, sem mexer no
// status da Notificação (cenário de duas Tentativas abertas, T-42).
func (r *Repositorio) AdicionarAceita(notificacao string, c dominio.Canal, idProvedor string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enfileirar(notificacao, c, 1, 0)
	i := len(r.Tentativas) - 1
	r.Tentativas[i].Status = dominio.TentativaEnviada
	r.idProvedor[i] = idProvedor
}
