import { createApp } from 'vue'
import { createPinia } from 'pinia'
import './assets/style/global.css'
import App from './App.vue'
import router from './router'
import SongMenu from './components/SongMenu.vue'

const app = createApp(App)
app.use(createPinia())
app.use(router)
// 全局注册：各歌曲列表可直接用 <SongMenu :song="..." />，无需逐个 import
app.component('SongMenu', SongMenu)
app.mount('#app')