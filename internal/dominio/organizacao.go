package dominio

import "errors"

// Organizacao é a empresa ou marca do grupo a que pertencem um Sistema de
// origem e as suas Notificações (RN-15). Uma não enxerga os dados da outra.
type Organizacao string

// OrganizacaoInicial recebe tudo o que existia antes das Organizações (REQ-51).
const OrganizacaoInicial Organizacao = "wbem"

// ErrSemOrganizacao: o client do token não pertence a nenhuma Organização (REQ-50).
var ErrSemOrganizacao = errors.New("sistema de origem sem organização")
