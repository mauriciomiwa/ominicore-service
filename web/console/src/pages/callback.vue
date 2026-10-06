<script setup lang="ts">
import { useSessao } from '@/console/sessao'

// Volta do Keycloak: conclui o login no realm do Ambiente da URL.
const route = useRoute()
const router = useRouter()
const sessao = useSessao()
const falhou = ref(false)

onMounted(async () => {
  try {
    if (await sessao.concluir(String(route.params.ambiente)))
      await router.replace('/whatsapp')
    else
      falhou.value = true
  }
  catch {
    falhou.value = true
  }
})
</script>

<template>
  <div class="d-flex align-center justify-center pa-4 min-h-screen">
    <VCard
      max-width="460"
      class="pa-4"
    >
      <VCardText v-if="!falhou">
        <VProgressCircular
          indeterminate
          class="me-3"
        />
        {{ $t('sessao.concluindo') }}
      </VCardText>
      <VCardText v-else>
        <p>{{ $t('sessao.falhou') }}</p>
        <VBtn to="/">
          {{ $t('sessao.voltar') }}
        </VBtn>
      </VCardText>
    </VCard>
  </div>
</template>
