// O happy-dom não tem visualViewport, que o VDialog do Vuetify usa para posicionar.
if (!('visualViewport' in globalThis)) {
  Object.defineProperty(globalThis, 'visualViewport', {
    value: { width: 1024, height: 768, offsetLeft: 0, offsetTop: 0, scale: 1, addEventListener() {}, removeEventListener() {} },
  })
}
