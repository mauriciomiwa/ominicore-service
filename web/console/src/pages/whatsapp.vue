<script setup lang="ts">
import DialogoConfirmar from '@/components/DialogoConfirmar.vue'
import { criarApi, provedores } from '@/console/api'
import type { ConfigWhatsApp, Resultado } from '@/console/api'
import { useSessao } from '@/console/sessao'

// Consulta, troca e desligamento do WhatsApp pela API do Ambiente (REQ-32).
// Segredos nunca aparecem preenchidos: só "configurado" (REQ-23).
const { t } = useI18n()
const sessao = useSessao()
const api = criarApi(sessao)
const dialogo = ref<InstanceType<typeof DialogoConfirmar>>()

const carregando = ref(true)
const semPermissao = ref(false)
const falha = ref<number | null>(null)
const config = ref<ConfigWhatsApp>({ ligado: false })
const provedor = ref('zapi')
const valores = reactive<Record<string, string>>({})
const recusa = ref('')
const campoRecusado = ref('')
const sucesso = ref('')
const enviando = ref(false)

const campos = computed(() => provedores[provedor.value] ?? [])
const nomeAmbiente = computed(() => sessao.ambiente ? t(`ambiente.${sessao.ambiente.id}`) : '')

function limparSegredos() {
  for (const lista of Object.values(provedores)) {
    for (const c of lista) {
      if (c.segredo)
        valores[c.nome] = ''
    }
  }
}

function mostrar(c: ConfigWhatsApp) {
  config.value = c
  if (c.ligado && c.provedor) {
    provedor.value = c.provedor
    for (const [nome, valor] of Object.entries(c.campos ?? {}))
      valores[nome] = valor
  }
  limparSegredos()
}

// Marca de um novo login já tentado por 401: se a API recusar de novo o token
// recém-emitido (ex.: audiência errada), mostra a falha em vez de entrar num
// laço Console ↔ Keycloak. Só a marca, nunca o token (REQ-29).
const marcaRelogin = 'ominicore-console:relogin'

function jaTentouRelogin(): boolean {
  try {
    const tentou = sessionStorage.getItem(marcaRelogin) !== null

    if (tentou)
      sessionStorage.removeItem(marcaRelogin)
    else
      sessionStorage.setItem(marcaRelogin, '1')

    return tentou
  }
  catch {
    return true
  }
}

// Trata o que não é sucesso; true quando a tela deve parar.
async function tratar<T>(r: Resultado<T>): Promise<boolean> {
  if (r.tipo === 'sem-sessao') {
    if (jaTentouRelogin())
      falha.value = 401
    else if (sessao.ambiente)
      await sessao.entrar(sessao.ambiente)

    return true
  }
  if (r.tipo === 'sem-permissao') {
    semPermissao.value = true

    return true
  }
  if (r.tipo === 'falha') {
    falha.value = r.status

    return true
  }

  return false
}

async function carregar() {
  carregando.value = true
  falha.value = null

  let r: Resultado<ConfigWhatsApp>
  try {
    r = await api.consultar()
  }
  finally {
    // nunca carregando para sempre (REQ-43)
    carregando.value = false
  }
  if (await tratar(r))
    return
  if (r.tipo === 'ok') {
    try {
      sessionStorage.removeItem(marcaRelogin)
    }
    catch {
      // sem sessionStorage (navegação privada): a marca é só conveniência
    }
    mostrar(r.dados)
  }
}

async function gravar() {
  if (sessao.producao && !(await dialogo.value?.perguntar(t('confirmar.gravarProducao'))))
    return
  recusa.value = ''
  campoRecusado.value = ''
  sucesso.value = ''
  enviando.value = true

  const corpo: Record<string, string> = { provedor: provedor.value }
  for (const c of campos.value) {
    if (valores[c.nome])
      corpo[c.nome] = valores[c.nome]
  }
  const r = await api.gravar(corpo)

  enviando.value = false
  if (await tratar(r))
    return
  if (r.tipo === 'recusada') {
    recusa.value = r.detalhe
    campoRecusado.value = r.campo ?? ''

    return
  }
  if (r.tipo === 'ok') {
    mostrar(r.dados)
    sucesso.value = t('whatsapp.gravado')
  }
}

async function desligar() {
  if (!(await dialogo.value?.perguntar(t('confirmar.desligar', { ambiente: nomeAmbiente.value }))))
    return
  sucesso.value = ''

  const r = await api.desligar()

  if (await tratar(r))
    return
  if (r.tipo === 'ok') {
    mostrar({ ligado: false })
    sucesso.value = t('whatsapp.desligadoOk')
  }
}

onMounted(carregar)
</script>

<template>
  <div>
    <FaixaAmbiente />
    <h1 class="text-h4 mb-6">
      {{ $t('whatsapp.titulo') }}
    </h1>

    <VProgressLinear
      v-if="carregando"
      indeterminate
    />

    <VAlert
      v-else-if="semPermissao"
      data-teste="sem-permissao"
      type="warning"
      :text="$t('whatsapp.semPermissao')"
    />

    <VAlert
      v-else-if="falha !== null"
      data-teste="falha"
      type="error"
      :text="falha === 0 ? $t('whatsapp.semResposta', { ambiente: nomeAmbiente }) : $t('whatsapp.falha', { status: falha })"
    >
      <template #append>
        <VBtn
          data-teste="tentar-de-novo"
          variant="outlined"
          @click="carregar"
        >
          {{ $t('whatsapp.tentarDeNovo') }}
        </VBtn>
      </template>
    </VAlert>

    <template v-else>
      <VAlert
        v-if="sucesso"
        data-teste="sucesso"
        type="success"
        class="mb-6"
        :text="sucesso"
      />

      <VCard
        class="mb-6"
        :title="$t('whatsapp.estado')"
      >
        <VCardText data-teste="estado">
          <VChip
            data-teste="ligado"
            :color="config.ligado ? 'success' : 'secondary'"
            class="me-3"
          >
            {{ config.ligado ? $t('whatsapp.ligado') : $t('whatsapp.desligado') }}
          </VChip>
          <span v-if="config.ligado && config.provedor">
            {{ $t('whatsapp.provedorAtivo') }}:
            <strong data-teste="provedor-ativo">{{ $t(`whatsapp.provedores.${config.provedor}`) }}</strong>
          </span>
          <span v-else>{{ $t('whatsapp.semProvedor') }}</span>

          <VTable
            v-if="config.ligado"
            density="compact"
            class="mt-4"
          >
            <tbody>
              <tr
                v-for="(valor, nome) in config.campos"
                :key="nome"
                :data-teste="`campo-${nome}`"
              >
                <td>{{ $t(`whatsapp.campos.${nome}`) }}</td>
                <td>{{ valor }}</td>
              </tr>
              <tr
                v-for="(configurado, nome) in config.segredos"
                :key="nome"
                :data-teste="`segredo-${nome}`"
              >
                <td>{{ $t(`whatsapp.campos.${nome}`) }}</td>
                <td>{{ configurado ? $t('whatsapp.configurado') : $t('whatsapp.naoConfigurado') }}</td>
              </tr>
            </tbody>
          </VTable>
        </VCardText>
        <VCardActions v-if="config.ligado">
          <VBtn
            data-teste="desligar"
            color="error"
            variant="outlined"
            @click="desligar"
          >
            {{ $t('whatsapp.desligar') }}
          </VBtn>
        </VCardActions>
      </VCard>

      <VCard :title="$t('whatsapp.trocar')">
        <VCardText>
          <p class="mb-4">
            {{ $t('whatsapp.explicacaoSegredos') }}
          </p>
          <VAlert
            v-if="recusa"
            data-teste="recusa"
            type="error"
            class="mb-4"
            :text="campoRecusado
              ? $t('whatsapp.campoFaltando', { campo: $t(`whatsapp.campos.${campoRecusado}`) })
              : $t('whatsapp.recusada', { detalhe: recusa })"
          />
          <VForm @submit.prevent="gravar">
            <VSelect
              v-model="provedor"
              data-teste="provedor"
              class="mb-4"
              :label="$t('whatsapp.provedor')"
              :items="Object.keys(provedores).map(p => ({ value: p, title: $t(`whatsapp.provedores.${p}`) }))"
            />
            <VTextField
              v-for="c in campos"
              :key="`${provedor}-${c.nome}`"
              v-model="valores[c.nome]"
              :data-teste="`entrada-${c.nome}`"
              class="mb-4"
              :label="$t(`whatsapp.campos.${c.nome}`)"
              :hint="c.obrigatorio ? $t('whatsapp.obrigatorio') : $t('whatsapp.opcional')"
              persistent-hint
              :type="c.segredo ? 'password' : 'text'"
              :autocomplete="c.segredo ? 'new-password' : 'off'"
              :error="campoRecusado === c.nome"
            />
            <VBtn
              data-teste="gravar"
              type="submit"
              :loading="enviando"
              :color="sessao.producao ? 'error' : 'primary'"
            >
              {{ $t('whatsapp.gravar') }}
            </VBtn>
          </VForm>
        </VCardText>
      </VCard>
    </template>

    <DialogoConfirmar ref="dialogo" />
  </div>
</template>
