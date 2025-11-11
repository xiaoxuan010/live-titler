import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import './index.css'
import App from './App.vue'
import ControlPanel from './views/ControlPanel.vue'
import ShowSource from './views/ShowSource.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/control-panel'
    },
    {
      path: '/control-panel',
      name: 'control-panel',
      component: ControlPanel
    },
    {
      path: '/show-source',
      name: 'show-source',
      component: ShowSource
    }
  ]
})

createApp(App).use(router).mount('#app')
