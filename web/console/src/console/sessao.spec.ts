import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { criarApi } from './api'
import { useSessao } from './sessao'
import { hml, prod, resposta } from '@/testes/montar'

// o build publicado lista homologação e produção; em teste (DEV) só haveria o local
vi.mock('./ambientes', async () => {
  const { hml: h, prod: p } = await import('@/testes/ambientes')

  return { ambientes: [h, p], ambientePorId: (id: string) => [h, p].find(a => a.id === id) }
})

// id_token sem assinatura: o oidc-client-ts não confere a assinatura (quem confere é a API).
function idToken(claims: Record<string, unknown>) {
  const b64 = (o: unknown) => btoa(JSON.stringify(o)).replace(/=+$/, '').replace(/\+/g, '-').replace(/\//g, '_')

  return `${b64({ alg: 'none', typ: 'JWT' })}.${b64(claims)}.`
}

let redirecionamentos: URL[]

beforeEach(() => {
  setActivePinia(createPinia())
  localStorage.clear()
  sessionStorage.clear()
  redirecionamentos = []
  vi.spyOn(window.location, 'assign').mockImplementation(url => {
    redirecionamentos.push(new URL(String(url)))

    // o oidc-client-ts espera a página voltar (pageshow) para resolver o signinRedirect
    setTimeout(() => window.dispatchEvent(new Event('pageshow')), 0)
  })
})

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})

// Simula a volta do Keycloak com o code e o state do último redirecionamento.
async function voltarDoKeycloak(amb: typeof hml, tokens: Record<string, string>) {
  const ida = redirecionamentos.at(-1)!
  const pedidosDeToken: Request[] = []

  vi.stubGlobal('fetch', vi.fn(async (entrada: RequestInfo | URL, init?: RequestInit) => {
    pedidosDeToken.push(new Request(entrada, init))

    return resposta(200, { token_type: 'Bearer', expires_in: 300, scope: 'openid', ...tokens })
  }))
  window.history.replaceState({}, '', `/callback/${amb.id}?code=codigo-1&state=${ida.searchParams.get('state')}`)
  await useSessao().concluir(amb.id)

  return pedidosDeToken
}

describe('sessão por Ambiente', () => {
  it('T-106 escolher homologação leva ao realm hml com PKCE e nenhum token fica gravado no navegador', async () => {
    const sessao = useSessao()

    await sessao.entrar(hml)

    const ida = redirecionamentos.at(-1)!

    expect(`${ida.origin}${ida.pathname}`).toBe(`${hml.issuer}/protocol/openid-connect/auth`)
    expect(ida.searchParams.get('client_id')).toBe('ominicore-console')
    expect(ida.searchParams.get('response_type')).toBe('code')
    expect(ida.searchParams.get('code_challenge_method')).toBe('S256')
    expect(ida.searchParams.get('code_challenge')).toBeTruthy()
    expect(ida.searchParams.get('redirect_uri')).toBe(`${window.location.origin}/callback/hml`)
    expect(ida.searchParams.has('client_secret')).toBe(false)

    const tokens = {
      access_token: 'acesso-secreto-hml',
      refresh_token: 'atualizacao-secreta-hml',
      id_token: idToken({ sub: 'u1', preferred_username: 'ana', iss: hml.issuer, aud: 'ominicore-console' }),
    }

    const pedidos = await voltarDoKeycloak(hml, tokens)

    expect(pedidos[0].url).toBe(`${hml.issuer}/protocol/openid-connect/token`)

    const corpo = new URLSearchParams(await pedidos[0].text())

    expect(corpo.get('code_verifier')).toBeTruthy()
    expect(corpo.has('client_secret')).toBe(false)
    expect(sessao.usuario).toBe('ana')

    const gravado = JSON.stringify({ ...localStorage }) + JSON.stringify({ ...sessionStorage }) + document.cookie
    for (const valor of Object.values(tokens))
      expect(gravado).not.toContain(valor)

    const buscar = vi.fn(async () => resposta(200, { ligado: false }))

    await criarApi(sessao, buscar).consultar()
    expect(buscar).toHaveBeenCalledWith(`${hml.api}/v1/config/whatsapp`, expect.objectContaining({
      headers: expect.objectContaining({ Authorization: 'Bearer acesso-secreto-hml' }),
    }))
  })

  it('T-107 trocar de Ambiente descarta o token e exige novo login antes de chamar a outra API', async () => {
    const sessao = useSessao()

    await sessao.entrar(hml)
    await voltarDoKeycloak(hml, { access_token: 'acesso-hml', id_token: idToken({ sub: 'u1', preferred_username: 'ana' }) })
    expect(await sessao.token()).toBe('acesso-hml')

    await sessao.entrar(prod)

    const ida = redirecionamentos.at(-1)!

    expect(`${ida.origin}${ida.pathname}`).toBe(`${prod.issuer}/protocol/openid-connect/auth`)
    expect(sessao.ambiente?.id).toBe('prod')
    expect(sessao.usuario).toBe('')
    expect(await sessao.token()).toBeNull()

    const buscar = vi.fn(async () => resposta(200, { ligado: false }))
    const r = await criarApi(sessao, buscar).consultar()

    expect(r.tipo).toBe('sem-sessao')
    expect(buscar).not.toHaveBeenCalled()
  })
})
