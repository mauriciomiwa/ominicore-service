import type { Ambiente } from '@/console/ambientes'

// Ambientes publicados, para os testes (o build de teste roda em DEV, só com o local).
export const hml: Ambiente = { id: 'hml', producao: false, issuer: 'https://auth.flipapi.com.br/realms/ominicore-hml', clientId: 'ominicore-console', api: 'https://notify-hml.flipapi.com.br' }
export const prod: Ambiente = { id: 'prod', producao: true, issuer: 'https://auth.flipapi.com.br/realms/ominicore', clientId: 'ominicore-console', api: 'https://notify.flipapi.com.br' }
