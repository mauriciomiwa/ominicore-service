import { InMemoryWebStorage, UserManager, WebStorageStateStore } from 'oidc-client-ts'
import type { User, UserManagerSettings } from 'oidc-client-ts'
import { defineStore } from 'pinia'
import { ambientePorId } from './ambientes'
import type { Ambiente } from './ambientes'

// O que a sessão usa do UserManager; nos testes, um falso.
export type GerenciadorOIDC = Pick<UserManager, 'signinRedirect' | 'signinRedirectCallback' | 'getUser' | 'removeUser' | 'signoutRedirect'>
export type FabricaOIDC = (s: UserManagerSettings) => GerenciadorOIDC

let fabrica: FabricaOIDC = s => new UserManager(s)

/** Troca a fábrica do cliente OIDC (testes). */
export function definirFabricaOIDC(f: FabricaOIDC) {
  fabrica = f
}

/**
 * Configuração do login no realm do Ambiente (REQ-29): Authorization Code com
 * PKCE, client público sem segredo e token só em memória. O Ambiente vai na
 * redirect_uri, para o callback saber em que realm concluir.
 */
export function configuracaoOIDC(amb: Ambiente, origem = window.location.origin): UserManagerSettings {
  const oidc = `${amb.issuer}/protocol/openid-connect`

  return {
    authority: amb.issuer,
    client_id: amb.clientId,
    redirect_uri: `${origem}/callback/${amb.id}`,
    post_logout_redirect_uri: `${origem}/`,
    response_type: 'code',
    scope: 'openid',
    disablePKCE: false,
    loadUserInfo: false,
    automaticSilentRenew: true,
    userStore: new WebStorageStateStore({ store: new InMemoryWebStorage() }),
    metadata: {
      issuer: amb.issuer,
      authorization_endpoint: `${oidc}/auth`,
      token_endpoint: `${oidc}/token`,
      end_session_endpoint: `${oidc}/logout`,
    },
  }
}

// O gerenciador fica fora do estado reativo: ele guarda o token, que não deve
// aparecer em devtools nem ser serializado.
let gerenciador: GerenciadorOIDC | null = null

export const useSessao = defineStore('sessao', {
  state: () => ({
    ambiente: null as Ambiente | null,
    usuario: '',
  }),
  getters: {
    autenticada: s => s.ambiente !== null && s.usuario !== '',
    producao: s => s.ambiente?.producao === true,
  },
  actions: {
    /** Vai ao login do Ambiente; trocar de Ambiente descarta a sessão atual (T-107). */
    async entrar(amb: Ambiente) {
      await this.descartar()
      this.ambiente = amb
      gerenciador = fabrica(configuracaoOIDC(amb))
      await gerenciador.signinRedirect()
    },

    /** Conclui o login na volta do Keycloak. */
    async concluir(id: string): Promise<boolean> {
      const amb = ambientePorId(id)
      if (!amb)
        return false
      gerenciador = fabrica(configuracaoOIDC(amb))

      const usuario: User = await gerenciador.signinRedirectCallback()

      this.ambiente = amb
      this.usuario = String(usuario.profile.preferred_username ?? usuario.profile.sub)

      return true
    },

    /** Token de acesso do Ambiente atual, ou null sem sessão válida. */
    async token(): Promise<string | null> {
      if (!gerenciador || !this.ambiente)
        return null
      const usuario = await gerenciador.getUser()

      return usuario && !usuario.expired ? usuario.access_token : null
    },

    async descartar() {
      if (gerenciador)
        await gerenciador.removeUser()
      gerenciador = null
      this.ambiente = null
      this.usuario = ''
    },

    async sair() {
      const atual = gerenciador

      await this.descartar()
      if (atual)
        await atual.signoutRedirect()
    },
  },
})
