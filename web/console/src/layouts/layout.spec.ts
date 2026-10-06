import { afterEach, expect, it, vi } from 'vitest'
import DefaultLayout from './components/DefaultLayoutWithVerticalNav.vue'
import { assentar, hml, montar, sessaoAberta } from '@/testes/montar'

afterEach(() => {
  vi.restoreAllMocks()
  document.body.innerHTML = ''
})

it('layout do Console mostra o Ambiente, a pessoa e o menu do WhatsApp', async () => {
  const { pinia } = sessaoAberta(hml, 'ana')

  await montar(DefaultLayout, pinia)
  await assentar()

  const texto = document.body.textContent ?? ''

  expect(texto).toContain('Homologação')
  expect(texto).toContain('ana')
  expect(texto).toContain('WhatsApp')
  expect(texto).not.toContain('Materio')
})
