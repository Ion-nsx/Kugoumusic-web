<template>
  <div class="app">
    <!-- 左侧导航 -->
    <Sidebar />

    <!-- 主区域：顶栏 + 滚动内容 -->
    <div class="main">
      <TopBar />

      <div class="content">
        <router-view />
      </div>
    </div>

    <!-- 底部播放器 -->
    <PlayerBar v-if="player.currentSong" />

    <LoginModal v-if="showLogin" :reason="loginReason" @close="showLogin = false" />
  </div>
</template>

<script setup>
import { ref, provide, onMounted, watch } from 'vue'
import Sidebar from './components/Sidebar.vue'
import TopBar from './components/TopBar.vue'
import PlayerBar from './components/PlayerBar.vue'
import LoginModal from './components/LoginModal.vue'
import { useAuthStore } from './stores/auth'
import { usePlayerStore } from './stores/player'
import { registerDevice } from './utils/api'

const auth = useAuthStore()
const player = usePlayerStore()
const showLogin = ref(false)
const loginReason = ref('')

provide('showLogin', showLogin)
provide('loginReason', loginReason)

function openLogin() {
  loginReason.value = ''
  showLogin.value = true
}
provide('openLogin', openLogin)

// 需要登录的功能入口统一调用：提示后弹出登录框
function requireLogin(reason) {
  loginReason.value = reason || '请先登录'
  showLogin.value = true
}
provide('requireLogin', requireLogin)

// 监听后端 401「请先登录」事件，统一弹出登录框并提示原因
function openLoginWithReason(reason) {
  loginReason.value = reason || '请先登录'
  showLogin.value = true
}
window.addEventListener('vibe:login-required', () => openLoginWithReason('请先登录'))

// 登录态变化时同步加载/清空云端「我喜欢」收藏
watch(() => auth.isLoggedIn, (logged) => {
  if (logged) {
    player.loadLiked()
  } else {
    player.clearLiked()
  }
}, { immediate: true })

onMounted(async () => {
  if (!auth.mid) {
    try {
      const res = await registerDevice()
      if (res.data.status === 1 && res.data.data) {
        auth.setDeviceInfo(res.data.data)
      }
    } catch (e) {
      console.error('设备注册失败:', e)
    }
  } else {
    auth.applyAuth()
  }
})
</script>

<style scoped>
/* 布局骨架来自全局 .app / .main / .content */
</style>
