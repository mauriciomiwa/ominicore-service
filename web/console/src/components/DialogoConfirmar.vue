<script setup lang="ts">
const aberto = ref(false)
const texto = ref('')

// Pergunta antes de uma troca que mexe nos envios (REQ-31, T-112).
let responder: (sim: boolean) => void = () => {}

function perguntar(pergunta: string): Promise<boolean> {
  texto.value = pergunta
  aberto.value = true

  return new Promise(resolve => {
    responder = resolve
  })
}

function fechar(sim: boolean) {
  aberto.value = false
  responder(sim)
}

defineExpose({ perguntar })
</script>

<template>
  <VDialog
    v-model="aberto"
    max-width="480"
    persistent
  >
    <VCard
      data-teste="dialogo-confirmar"
      :title="$t('confirmar.titulo')"
    >
      <VCardText data-teste="dialogo-texto">
        {{ texto }}
      </VCardText>
      <VCardActions>
        <VSpacer />
        <VBtn
          data-teste="dialogo-nao"
          variant="text"
          @click="fechar(false)"
        >
          {{ $t('confirmar.nao') }}
        </VBtn>
        <VBtn
          data-teste="dialogo-sim"
          color="error"
          variant="elevated"
          @click="fechar(true)"
        >
          {{ $t('confirmar.sim') }}
        </VBtn>
      </VCardActions>
    </VCard>
  </VDialog>
</template>
