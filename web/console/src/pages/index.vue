<script setup lang="ts">
import { ambientes } from '@/console/ambientes'
import { useSessao } from '@/console/sessao'

// Escolha do Ambiente: cada um tem o próprio realm e a própria API (REQ-29).
const sessao = useSessao()
</script>

<template>
  <div class="d-flex align-center justify-center pa-4 min-h-screen">
    <VCard
      max-width="460"
      class="pa-4"
    >
      <VCardItem>
        <VCardTitle class="text-h5">
          {{ $t('app.nome') }}
        </VCardTitle>
        <VCardSubtitle>{{ $t('app.subtitulo') }}</VCardSubtitle>
      </VCardItem>
      <VCardText>
        <h2 class="text-h6 mb-2">
          {{ $t('ambiente.titulo') }}
        </h2>
        <p class="mb-6">
          {{ $t('ambiente.explicacao') }}
        </p>
        <VBtn
          v-for="amb in ambientes"
          :key="amb.id"
          :data-teste="`entrar-${amb.id}`"
          block
          class="mb-3"
          :color="amb.producao ? 'error' : 'primary'"
          :variant="amb.producao ? 'outlined' : 'elevated'"
          @click="sessao.entrar(amb)"
        >
          {{ $t('ambiente.entrar', { ambiente: $t(`ambiente.${amb.id}`) }) }}
        </VBtn>
      </VCardText>
    </VCard>
  </div>
</template>
