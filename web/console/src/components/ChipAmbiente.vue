<script setup lang="ts">
import { useSessao } from '@/console/sessao'

const sessao = useSessao()
</script>

<template>
  <VChip
    v-if="sessao.ambiente"
    data-teste="chip-ambiente"
    :color="sessao.producao ? 'error' : 'primary'"
    :variant="sessao.producao ? 'elevated' : 'tonal'"
    :prepend-icon="sessao.producao ? 'ri-alarm-warning-line' : 'ri-flask-line'"
    label
  >
    {{ $t(`ambiente.${sessao.ambiente.id}`) }}
  </VChip>
</template>
