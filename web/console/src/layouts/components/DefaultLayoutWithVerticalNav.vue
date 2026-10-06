<script lang="ts" setup>
import NavItems from '@/layouts/components/NavItems.vue'
import VerticalNavLayout from '@layouts/components/VerticalNavLayout.vue'
import NavbarThemeSwitcher from '@/layouts/components/NavbarThemeSwitcher.vue'
import { useSessao } from '@/console/sessao'

const sessao = useSessao()
const router = useRouter()

async function trocarAmbiente() {
  await sessao.descartar()
  await router.push('/')
}
</script>

<template>
  <VerticalNavLayout>
    <!-- 👉 navbar -->
    <template #navbar="{ toggleVerticalOverlayNavActive }">
      <div class="d-flex h-100 align-center ga-3">
        <IconBtn
          class="ms-n3 d-lg-none"
          @click="toggleVerticalOverlayNavActive(true)"
        >
          <VIcon icon="ri-menu-line" />
        </IconBtn>

        <ChipAmbiente />

        <VSpacer />

        <span
          class="text-body-1 d-none d-sm-inline"
          data-teste="usuario"
        >{{ sessao.usuario }}</span>

        <VBtn
          variant="text"
          prepend-icon="ri-swap-line"
          data-teste="trocar-ambiente"
          @click="trocarAmbiente"
        >
          {{ $t('ambiente.trocar') }}
        </VBtn>

        <NavbarThemeSwitcher />

        <VBtn
          variant="text"
          prepend-icon="ri-logout-box-r-line"
          data-teste="sair"
          @click="sessao.sair()"
        >
          {{ $t('sessao.sair') }}
        </VBtn>
      </div>
    </template>

    <template #vertical-nav-header="{ toggleIsOverlayNavActive }">
      <RouterLink
        to="/whatsapp"
        class="app-logo app-title-wrapper"
      >
        <VIcon
          icon="ri-notification-3-line"
          color="primary"
          size="28"
        />
        <h1 class="font-weight-medium leading-normal text-xl">
          {{ $t('app.nome') }}
        </h1>
      </RouterLink>

      <IconBtn
        class="d-block d-lg-none"
        @click="toggleIsOverlayNavActive(false)"
      >
        <VIcon icon="ri-close-line" />
      </IconBtn>
    </template>

    <template #vertical-nav-content>
      <NavItems />
    </template>

    <!-- 👉 Pages -->
    <slot />
  </VerticalNavLayout>
</template>

<style lang="scss" scoped>
.app-logo {
  display: flex;
  align-items: center;
  column-gap: 0.75rem;
}
</style>
