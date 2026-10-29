export const routes = [
  { path: '/', component: () => import('@/layouts/blank.vue'), children: [{ path: '', component: () => import('@/pages/index.vue') }] },
  { path: '/callback/:ambiente', component: () => import('@/layouts/blank.vue'), children: [{ path: '', component: () => import('@/pages/callback.vue') }] },
  {
    path: '/',
    component: () => import('@/layouts/default.vue'),
    meta: { sessao: true },
    children: [
      { path: 'whatsapp', component: () => import('@/pages/whatsapp.vue'), meta: { sessao: true } },
    ],
  },
  { path: '/:pathMatch(.*)*', component: () => import('@/layouts/blank.vue'), children: [{ path: '', component: () => import('@/pages/erro.vue') }] },
]
