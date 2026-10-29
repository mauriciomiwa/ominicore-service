<script lang="ts" setup>
import VerticalNavSectionTitle from '@/@layouts/components/VerticalNavSectionTitle.vue'
import VerticalNavLink from '@layouts/components/VerticalNavLink.vue'

const { t } = useI18n()
</script>

<template>
  <VerticalNavSectionTitle :item="{ heading: t('menu.canais') }" />
  <VerticalNavLink
    :item="{
      title: t('menu.whatsapp'),
      icon: 'ri-whatsapp-line',
      to: '/whatsapp',
    }"
  />
</template>
