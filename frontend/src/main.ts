import { createApp } from 'vue'

import App from './App.vue'
import router from './router'
import { pinia } from './stores'
import { installSessionHandlers } from './stores/session'
import './styles/tokens.css'

const app = createApp(App)

app.use(pinia)
app.use(router)
installSessionHandlers(router)
app.mount('#app')
