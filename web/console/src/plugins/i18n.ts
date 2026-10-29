import type { App } from 'vue'
import { createI18n } from 'vue-i18n'
import ptBR from '@/locales/pt-BR.json'

// Só pt-BR (RN-12): nenhum texto fica no template, todos vêm daqui.
export const i18n = createI18n({
  legacy: false,
  locale: 'pt-BR',
  fallbackLocale: 'pt-BR',
  messages: { 'pt-BR': ptBR },
})

export default function (app: App) {
  app.use(i18n)
}
