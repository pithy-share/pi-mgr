import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import App from './App.vue'
import ConfigPage from './views/ConfigPage.vue'
import SshSync from './views/SshSync.vue'
import './style.css'

const routes = [
  { path: '/', component: ConfigPage },
  { path: '/ssh-sync', component: SshSync },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

const app = createApp(App)
app.use(router)
app.mount('#app')