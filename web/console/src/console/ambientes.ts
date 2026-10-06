import publicados from './ambientes-publicados.json'
import locais from './ambientes-locais.json'

// Ambiente é uma instalação do OminiCore (homologação ou produção) com realm,
// cofre e API próprios (glossário). Nenhum segredo aqui: o client é público.
export interface Ambiente {
  id: string
  producao: boolean
  issuer: string
  clientId: string
  api: string
}

// `bun run dev` fala só com o ambiente local (deploy/local); o build publicado
// só com homologação e produção.
export const ambientes: Ambiente[] = import.meta.env.DEV ? locais : publicados

export function ambientePorId(id: string, lista: Ambiente[] = ambientes): Ambiente | undefined {
  return lista.find(a => a.id === id)
}
