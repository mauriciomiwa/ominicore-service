package aplicacao

import (
	"context"

	"ominicore/internal/dominio"
	"ominicore/internal/portas"
)

// OptOutUseCase registra e remove Opt-out (REQ-15, REQ-16). O Opt-out vale
// para toda a casa; "por" só identifica quem fez, para o histórico.
type OptOutUseCase struct {
	Repo portas.MessageRepositoryPort
}

// Registrar pelo Sistema de origem (API).
func (uc OptOutUseCase) Registrar(ctx context.Context, c dominio.Canal, contato, sistemaOrigem string) error {
	if _, err := dominio.NormalizarContato(c, contato); err != nil {
		return err
	}
	return uc.Repo.RegistrarOptOut(ctx, c, contato, "api:"+sistemaOrigem)
}

// Remover pelo Sistema de origem (API) — a única forma de voltar a receber.
func (uc OptOutUseCase) Remover(ctx context.Context, c dominio.Canal, contato, sistemaOrigem string) error {
	if _, err := dominio.NormalizarContato(c, contato); err != nil {
		return err
	}
	return uc.Repo.RemoverOptOut(ctx, c, contato, "api:"+sistemaOrigem)
}

// RegistrarPeloProvedor: descadastro avisado pelo Provedor (SendGrid, Twilio).
func (uc OptOutUseCase) RegistrarPeloProvedor(ctx context.Context, c dominio.Canal, contato, provedor string) error {
	if _, err := dominio.NormalizarContato(c, contato); err != nil {
		return nil // aviso sem contato utilizável: nada a registrar
	}
	return uc.Repo.RegistrarOptOut(ctx, c, contato, provedor)
}
