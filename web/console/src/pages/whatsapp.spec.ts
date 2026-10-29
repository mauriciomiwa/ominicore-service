import { afterEach, describe, expect, it, vi } from 'vitest'
import { VSelect } from 'vuetify/components/VSelect'
import Whatsapp from './whatsapp.vue'
import DefaultLayout from '@/layouts/components/DefaultLayoutWithVerticalNav.vue'
import { assentar, hml, montar, prod, resposta, sessaoAberta } from '@/testes/montar'

const zapi = { ligado: true, provedor: 'zapi', campos: { instance_id: 'inst1' }, segredos: { token: true, client_token: true } }

interface Chamada { metodo: string; url: string; corpo?: Record<string, string> }

// fetch de mentira: responde na ordem e guarda as chamadas
function api(...respostas: Response[]) {
  const chamadas: Chamada[] = []

  vi.stubGlobal('fetch', vi.fn(async (url: string, init: RequestInit) => {
    chamadas.push({ metodo: init.method ?? 'GET', url, corpo: init.body ? JSON.parse(String(init.body)) : undefined })

    return respostas.shift() ?? resposta(500)
  }))

  return chamadas
}

const el = (teste: string) => document.body.querySelector<HTMLElement>(`[data-teste="${teste}"]`)
const entrada = (nome: string) => el(`entrada-${nome}`)!.querySelector('input')!

async function digitar(nome: string, valor: string) {
  const input = entrada(nome)

  input.value = valor
  input.dispatchEvent(new Event('input'))
  await assentar()
}

async function clicar(teste: string) {
  el(teste)!.click()
  await assentar()
}

afterEach(() => {
  sessionStorage.clear()
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
  document.body.innerHTML = ''
})

describe('tela do WhatsApp', () => {
  it('T-108 conta sem o papel vê falta de permissão e nenhum formulário', async () => {
    const { pinia } = sessaoAberta(hml)

    api(resposta(403, { status: 403, title: 'sem permissão' }))
    await montar(Whatsapp, pinia)
    await assentar()

    expect(el('sem-permissao')).not.toBeNull()
    expect(el('gravar')).toBeNull()
    expect(el('estado')).toBeNull()
  })

  it('T-110 consulta mostra o Provedor e os segredos só como configurados', async () => {
    const { pinia } = sessaoAberta(hml)
    const chamadas = api(resposta(200, zapi))

    await montar(Whatsapp, pinia)
    await assentar()

    expect(chamadas[0]).toMatchObject({ metodo: 'GET', url: `${hml.api}/v1/config/whatsapp` })
    expect(el('ligado')!.textContent).toContain('Ligado')
    expect(el('provedor-ativo')!.textContent).toContain('Z-API')
    expect(el('campo-instance_id')!.textContent).toContain('inst1')
    expect(el('segredo-token')!.textContent).toMatch(/configurado$/)
    expect(el('segredo-token')!.textContent).not.toContain('não')
    expect(el('segredo-client_token')!.textContent).not.toContain('não')
    expect(entrada('token').value).toBe('')
    expect(entrada('client_token').value).toBe('')
    expect(entrada('token').type).toBe('password')
  })

  it('T-111 recusa 422 cita o campo sem limpar o formulário; depois confirma, com os segredos vazios', async () => {
    const { pinia } = sessaoAberta(hml)
    const ultra = { ligado: true, provedor: 'ultramsg', campos: { instance_id: 'instance9' }, segredos: { token: true } }

    const chamadas = api(
      resposta(200, { ligado: false }),
      resposta(422, { status: 422, title: 'configuração recusada', detail: 'configuração whatsapp incompleta: sem o campo token' }),
      resposta(200, ultra),
    )

    const tela = await montar(Whatsapp, pinia)

    await assentar()
    tela.findComponent(VSelect).vm.$emit('update:modelValue', 'ultramsg')
    await assentar()
    await digitar('instance_id', 'instance9')
    await clicar('gravar')

    expect(chamadas[1]).toMatchObject({ metodo: 'PUT', corpo: { provedor: 'ultramsg', instance_id: 'instance9' } })
    expect(el('recusa')!.textContent).toContain('Token')
    expect(entrada('instance_id').value).toBe('instance9')

    await digitar('token', 'tk-novo-secreto')
    await clicar('gravar')

    expect(chamadas[2]).toMatchObject({ metodo: 'PUT', corpo: { provedor: 'ultramsg', instance_id: 'instance9', token: 'tk-novo-secreto' } })
    expect(el('sucesso')).not.toBeNull()
    expect(el('recusa')).toBeNull()
    expect(el('provedor-ativo')!.textContent).toContain('UltraMsg')
    expect(entrada('token').value).toBe('')
    expect(document.body.innerHTML).not.toContain('tk-novo-secreto')
  })

  it('T-112 desligar: cancelar não chama a API; confirmar chama e mostra desligado', async () => {
    const { pinia } = sessaoAberta(hml)
    const chamadas = api(resposta(200, zapi), resposta(204))

    await montar(Whatsapp, pinia)
    await assentar()

    await clicar('desligar')
    expect(el('dialogo-confirmar')).not.toBeNull()
    await clicar('dialogo-nao')
    expect(chamadas).toHaveLength(1)

    await clicar('desligar')
    await clicar('dialogo-sim')
    expect(chamadas[1]).toMatchObject({ metodo: 'DELETE', url: `${hml.api}/v1/config/whatsapp` })
    expect(el('ligado')!.textContent).toContain('Desligado')
  })

  it('T-109 produção destacada e confirmação antes de gravar; homologação sem destaque', async () => {
    const { pinia } = sessaoAberta(prod)
    const chamadas = api(resposta(200, zapi), resposta(200, zapi))

    await montar(Whatsapp, pinia)
    await assentar()
    expect(el('faixa-producao')).not.toBeNull()

    await digitar('token', 'tk')
    await clicar('gravar')
    expect(el('dialogo-texto')!.textContent).toContain('PRODUÇÃO')
    expect(chamadas).toHaveLength(1)
    await clicar('dialogo-sim')
    expect(chamadas[1]).toMatchObject({ metodo: 'PUT', url: `${prod.api}/v1/config/whatsapp` })
    document.body.innerHTML = ''

    // o chip do Ambiente fica na barra de toda tela autenticada
    await montar(DefaultLayout, pinia)
    await assentar()
    expect(el('chip-ambiente')!.textContent).toContain('Produção')
    expect(el('chip-ambiente')!.className).toContain('bg-error')
    document.body.innerHTML = ''

    const hmlSessao = sessaoAberta(hml)
    const emHml = api(resposta(200, zapi), resposta(200, zapi))

    await montar(Whatsapp, hmlSessao.pinia)
    await assentar()
    expect(el('faixa-producao')).toBeNull()
    await digitar('token', 'tk')
    await clicar('gravar')
    expect(el('dialogo-texto')?.textContent ?? '').not.toContain('PRODUÇÃO')
    expect(emHml[1]).toMatchObject({ metodo: 'PUT', url: `${hml.api}/v1/config/whatsapp` })
  })
})

it('401 com token novo leva ao login uma vez e, repetindo, mostra a falha sem laço', async () => {
  const primeira = sessaoAberta(hml)

  api(resposta(401))
  await montar(Whatsapp, primeira.pinia)
  await assentar()
  expect(primeira.sessao.entrar).toHaveBeenCalledTimes(1)
  document.body.innerHTML = ''

  // volta do Keycloak e a API recusa de novo
  const segunda = sessaoAberta(hml)

  api(resposta(401))
  await montar(Whatsapp, segunda.pinia)
  await assentar()
  expect(segunda.sessao.entrar).not.toHaveBeenCalled()
  expect(el('falha')!.textContent).toContain('401')
})

describe('API do Ambiente sem resposta (REQ-43)', () => {
  it('T-137 API fora do ar e nova tentativa', async () => {
    const { pinia } = sessaoAberta(prod)
    let chamadas = 0

    vi.stubGlobal('fetch', vi.fn(async () => {
      chamadas++
      if (chamadas === 1)
        throw new TypeError('Failed to fetch')

      return resposta(200, zapi)
    }))

    await montar(Whatsapp, pinia)
    await assentar()

    expect(el('falha')!.textContent).toContain('A API de Produção não respondeu')
    expect(document.body.querySelector('.v-progress-linear')).toBeNull()
    expect(el('estado')).toBeNull()

    await clicar('tentar-de-novo')

    expect(chamadas).toBe(2)
    expect(el('falha')).toBeNull()
    expect(el('provedor-ativo')!.textContent).toContain('Z-API')
  })

  it('T-138 Erro do servidor mostra o status', async () => {
    const { pinia } = sessaoAberta(prod)

    api(resposta(503, { status: 503, title: 'cofre indisponível' }))
    await montar(Whatsapp, pinia)
    await assentar()

    expect(el('falha')!.textContent).toContain('503')
    expect(el('tentar-de-novo')).not.toBeNull()
  })
})

describe('webhooks da Z-API (REQ-42)', () => {
  it('T-136 Webhooks não cadastrados e recadastro', async () => {
    const { pinia } = sessaoAberta(hml)
    const base = { ligado: true, provedor: 'zapi', segredos: { token: true, client_token: false, webhook_segredo: true } }

    const chamadas = api(
      resposta(200, { ...base, campos: { instance_id: 'inst1', webhooks_estado: 'falhou', webhooks_em: '2026-10-29T12:00:00Z' } }),
      resposta(200, { ...base, campos: { instance_id: 'inst1', webhooks_estado: 'cadastrados', webhooks_em: '2026-10-29T12:05:00Z' } }),
    )

    await montar(Whatsapp, pinia)
    await assentar()
    expect(el('webhooks-nao-cadastrados')).not.toBeNull()
    expect(el('webhooks-cadastrados')).toBeNull()

    await clicar('cadastrar-de-novo')

    expect(chamadas[1]).toMatchObject({ metodo: 'POST', url: `${hml.api}/v1/config/whatsapp/webhooks` })
    expect(el('webhooks-cadastrados')).not.toBeNull()
    expect(el('webhooks-nao-cadastrados')).toBeNull()

    // nem segredo nem URL do webhook na tela; o estado não vira "campo"
    const tela = document.body.innerHTML

    expect(tela).not.toContain('/webhooks/zapi')
    expect(el('campo-webhooks_estado')).toBeNull()
    expect(el('segredo-webhook_segredo')).toBeNull()
  })

  it('sem a Z-API ativa não mostra os webhooks', async () => {
    const { pinia } = sessaoAberta(hml)

    api(resposta(200, { ligado: true, provedor: 'ultramsg', campos: { instance_id: 'i' }, segredos: { token: true } }))
    await montar(Whatsapp, pinia)
    await assentar()
    expect(el('webhooks')).toBeNull()
  })
})
