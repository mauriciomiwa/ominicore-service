import type { Ambiente } from './ambientes'

// Visão da configuração sem segredos (REQ-23): de cada segredo, só se está configurado.
export interface ConfigWhatsApp {
  ligado: boolean
  provedor?: string
  campos?: Record<string, string>
  segredos?: Record<string, boolean>
}

export type Resultado<T> =
  | { tipo: 'ok'; dados: T }
  | { tipo: 'sem-sessao' }
  | { tipo: 'sem-permissao' }
  | { tipo: 'recusada'; detalhe: string; campo?: string }
  | { tipo: 'falha'; status: number }

export interface FonteDeSessao {
  ambiente: Ambiente | null
  token: () => Promise<string | null>
}

const caminho = '/v1/config/whatsapp'

/**
 * Cliente da API de configuração do Ambiente da sessão. Sem token, nenhuma
 * chamada sai (T-107); 401 vira "sem sessão" para a tela refazer o login.
 */
export function criarApi(sessao: FonteDeSessao, buscar: typeof fetch = (...a) => fetch(...a)) {
  async function chamar<T>(metodo: string, corpo?: Record<string, string>): Promise<Resultado<T>> {
    const token = await sessao.token()
    if (!token || !sessao.ambiente)
      return { tipo: 'sem-sessao' }

    const headers: Record<string, string> = { Authorization: `Bearer ${token}` }
    if (corpo)
      headers['Content-Type'] = 'application/json'

    const r = await buscar(`${sessao.ambiente.api}${caminho}`, {
      method: metodo,
      headers,
      body: corpo ? JSON.stringify(corpo) : undefined,
    })

    if (r.status === 401)
      return { tipo: 'sem-sessao' }
    if (r.status === 403)
      return { tipo: 'sem-permissao' }
    if (r.status === 422) {
      const problema = await r.json().catch(() => ({})) as { detail?: string }
      const detalhe = problema.detail ?? ''

      return { tipo: 'recusada', detalhe, campo: /sem o campo (\w+)/.exec(detalhe)?.[1] }
    }
    if (!r.ok)
      return { tipo: 'falha', status: r.status }
    if (r.status === 204)
      return { tipo: 'ok', dados: undefined as T }

    return { tipo: 'ok', dados: await r.json() as T }
  }

  return {
    consultar: () => chamar<ConfigWhatsApp>('GET'),
    gravar: (corpo: Record<string, string>) => chamar<ConfigWhatsApp>('PUT', corpo),
    desligar: () => chamar<void>('DELETE'),
  }
}

// Campos de cada Provedor, como o domínio exige (internal/dominio/config_whatsapp.go).
export interface Campo { nome: string; segredo: boolean; obrigatorio: boolean }

export const provedores: Record<string, Campo[]> = {
  meta: [
    { nome: 'phone_number_id', segredo: false, obrigatorio: true },
    { nome: 'token', segredo: true, obrigatorio: true },
    { nome: 'meta_app_secret', segredo: true, obrigatorio: true },
    { nome: 'meta_verify_token', segredo: true, obrigatorio: true },
  ],
  ultramsg: [
    { nome: 'instance_id', segredo: false, obrigatorio: true },
    { nome: 'token', segredo: true, obrigatorio: true },
  ],
  zapi: [
    { nome: 'instance_id', segredo: false, obrigatorio: true },
    { nome: 'token', segredo: true, obrigatorio: true },
    { nome: 'client_token', segredo: true, obrigatorio: false },
  ],
}
