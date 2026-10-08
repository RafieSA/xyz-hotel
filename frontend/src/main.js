import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import App from './App.vue'
import router from './router'
import { WarmAura } from './theme/aura'
import './style.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(PrimeVue, {
  theme: {
    preset: WarmAura,
    options: { darkModeSelector: '.dark', cssLayer: false },
  },
})
app.use(ToastService)
app.mount('#app')
