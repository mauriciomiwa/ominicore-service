import type { App } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { routes } from './routes'
import { useSessao } from '@/console/sessao'
import { store } from '@/plugins/pinia'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// Sem sessão (o token só vive em memória: recarregar a página a perde), volta
// à escolha do Ambiente.
router.beforeEach(to => {
  if (to.meta.sessao && !useSessao(store).autenticada)
    return '/'
})

export default function (app: App) {
  app.use(router)
}

export { router }
