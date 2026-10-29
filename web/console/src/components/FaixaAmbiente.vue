<script setup lang="ts">
import { useSessao } from '@/console/sessao'

// Nome do Ambiente em toda tela; em produção, com destaque de alerta (REQ-31).
const sessao = useSessao()
</script>

<template>
  <VAlert
    v-if="sessao.producao"
    data-teste="faixa-producao"
    color="error"
    variant="elevated"
    density="compact"
    icon="ri-error-warning-line"
    class="mb-6"
    :text="$t('ambiente.faixaProducao')"
  />
</template>
