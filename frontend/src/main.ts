import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import App from './App.vue'
import SchemeList from './views/SchemeList.vue'
import SchemeEditor from './views/SchemeEditor.vue'
import SshSync from './views/SshSync.vue'
import './style.css'

const routes = [
  { path: '/', component: SchemeList },
  { path: '/scheme/:id', component: SchemeEditor, props: true },
  { path: '/ssh-sync', component: SshSync },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

const app = createApp(App)
app.use(router)
app.mount('#app')
