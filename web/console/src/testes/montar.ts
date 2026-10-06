import { mount } from '@vue/test-utils'
import type { Component } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { createVuetify } from 'vuetify'
import { VApp } from 'vuetify/components/VApp'
import { vi } from 'vitest'
import type { Ambiente } from '@/console/ambientes'
import { useSessao } from '@/console/sessao'
import { i18n } from '@/plugins/i18n'

export { hml, prod } from './ambientes'

/** Pinia nova com a sessão já aberta no Ambiente, token "tok". */
export function sessaoAberta(amb: Ambiente, usuario = 'ana') {
  const pinia = createPinia()

  setActivePinia(pinia)

  const sessao = useSessao()

  sessao.$patch({ ambiente: amb, usuario })
  vi.spyOn(sessao, 'token').mockResolvedValue('tok')
  vi.spyOn(sessao, 'entrar').mockResolvedValue()

  return { pinia, sessao }
}

/** Monta o componente dentro de um VApp, com Vuetify, i18n e a Pinia dada. */
export async function montar(componente: Component, pinia = createPinia()) {
  const router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/:p(.*)*', component: { render: () => null } }] })

  await router.push('/whatsapp')

  return mount({ render: () => h(VApp, () => h(componente)) }, {
    global: { plugins: [createVuetify(), pinia, i18n, router] },
    attachTo: document.body,
  })
}

/** Resposta JSON de mentira para o fetch. */
export function resposta(status: number, corpo?: unknown): Response {
  return new Response(corpo === undefined ? null : JSON.stringify(corpo), {
    status,
    headers: { 'Content-Type': status >= 400 ? 'application/problem+json' : 'application/json' },
  })
}

/** Espera as promessas e a renderização assentarem. */
export async function assentar() {
  for (let i = 0; i < 5; i++)
    await new Promise(resolve => setTimeout(resolve, 0))
}
