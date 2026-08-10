<template>
  <header class="topbar">
    <!-- 左侧占位（保持搜索居中，与内容区对齐） -->
    <div class="topbar-left"></div>

    <div class="search">
      <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="7"/><path d="m21 21-4.3-4.3"/></svg>
      <input
        ref="inputEl"
        v-model="keyword"
        placeholder="搜索歌曲、歌手、歌单"
        @keyup.enter="doSearch"
      />
    </div>

    <div class="topbar-right">
      <!-- 设置：默认音质 -->
      <div class="settings-wrap">
        <button class="icon-btn" title="设置" @click="toggleSettings">
          <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1Z"/></svg>
        </button>

        <!-- 设置弹层 -->
        <div class="settings-pop" v-if="showSettings" @click.stop>
          <div class="settings-head">
            <span>设置</span>
            <button class="settings-close" @click="showSettings = false">✕</button>
          </div>
          <div class="settings-body">
            <div class="settings-label">默认音质</div>
            <div class="settings-tip">歌曲无对应音质时自动使用下一级</div>
            <div class="quality-list">
              <div
                v-for="q in qualityOptions"
                :key="q.value"
                class="quality-item"
                :class="{ on: player.quality === q.value }"
                @click="changeQuality(q.value)"
              >
                <span class="q-name">{{ q.label }}</span>
                <span class="q-tip">{{ q.tip }}</span>
                <span v-if="player.quality === q.value" class="q-check">✓</span>
            </div>
            <div class="settings-label" style="margin-top:20px">深色模式</div>
            <div class="theme-toggle" @click="toggleTheme">
              <span :class="{ on: !darkMode }">浅色</span>
              <span :class="{ on: darkMode }">深色</span>
            </div>
            <div class="settings-label" style="margin-top:20px">搜索排序</div>
            <div class="theme-toggle" @click="toggleOriginalFirst">
              <span :class="{ on: originalFirst }">原版优先</span>
              <span :class="{ on: !originalFirst }">原始排序</span>
            </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 用户区：未登录点击登录，已登录进入我的主页 -->
      <button class="user-entry" @click="handleUserClick" :title="auth.isLoggedIn ? '我的主页' : '登录'">
        <span v-if="auth.user?.avatar" class="user-avatar-img">
          <img :src="auth.user.avatar" alt="头像" />
        </span>
        <span v-else class="user-avatar">{{ avatarText }}</span>
        <span class="user-name">{{ displayName }}</span>
      </button>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, inject, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { usePlayerStore } from '../stores/player'

const router = useRouter()
const route = useRoute()
const showLogin = inject('showLogin')
const auth = useAuthStore()
const player = usePlayerStore()

const keyword = ref('')
const showSettings = ref(false)
const darkMode = ref(localStorage.getItem('vibe_dark') === '1')
const originalFirst = ref(localStorage.getItem('vibe_original_first') !== '0')

const qualityOptions = [
  { value: '128', label: '标准', tip: '128Kbps' },
  { value: '320', label: '高品', tip: '320Kbps' },
  { value: 'flac', label: '无损', tip: 'FLAC' },
  { value: 'high', label: 'Hi-Res', tip: '高解析' }
]

if (route.query.keyword) {
  keyword.value = route.query.keyword
}

const avatarText = computed(() => {
  if (auth.isLoggedIn) {
    const name = auth.user?.name || 'V'
    return name.slice(0, 1).toUpperCase()
  }
  return 'KG'
})

const displayName = computed(() => {
  if (auth.isLoggedIn) return auth.user?.name || '音乐用户'
  return '登录'
})

function handleUserClick() {
  if (auth.isLoggedIn) {
    router.push('/me')
  } else {
    showLogin.value = true
  }
}

function doSearch() {
  if (!keyword.value) return
  router.push({ path: '/search', query: { keyword: keyword.value, type: 'complex' } })
}

function toggleSettings() {
  showSettings.value = !showSettings.value
}

function toggleTheme() {
  darkMode.value = !darkMode.value
  localStorage.setItem('vibe_dark', darkMode.value ? '1' : '0')
  applyTheme()
}
function toggleOriginalFirst() {
  originalFirst.value = !originalFirst.value
  localStorage.setItem('vibe_original_first', originalFirst.value ? '1' : '0')
}
function applyTheme() {
  document.documentElement.setAttribute('data-theme', darkMode.value ? 'dark' : '')
}
// 初始化时应用主题
applyTheme()

// 切换默认音质：调用 player.setQuality 生效并持久化；弹层保持显示
async function changeQuality(q) {
  await player.setQuality(q)
}

// 点击弹层外部关闭
function onClickOutside(e) {
  const wrap = document.querySelector('.settings-wrap')
  if (wrap && !wrap.contains(e.target)) {
    showSettings.value = false
  }
}
onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))
</script>

<style scoped>
.topbar {
  display: grid;
  grid-template-columns: 1fr minmax(0, 560px) 1fr;
  align-items: center;
  padding: 13px 28px;
  background: var(--bg);
  backdrop-filter: blur(22px) saturate(1.5);
  -webkit-backdrop-filter: blur(22px) saturate(1.5);
  border-bottom: 1px solid var(--border);
  z-index: 4;
  flex-shrink: 0;
  gap: 16px;
}

.topbar-left { justify-self: start; }

.search {
  width: 100%; display: flex; align-items: center; gap: 9px;
  height: 42px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-full);
  padding: 0 16px;
  box-shadow: var(--shadow-sm);
  transition: border-color .2s, box-shadow .2s;
}
.search:focus-within {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px var(--primary-soft);
}
.search svg { width: 15px; height: 15px; stroke: var(--text-3); fill: none; stroke-width: 2; flex-shrink: 0; }
.search input {
  border: none; outline: none; background: transparent;
  font-size: 13.5px; color: var(--text-1); width: 100%;
  font-family: inherit;
}
.search input::placeholder { color: var(--text-3); }

.topbar-right { justify-self: end; display: flex; align-items: center; gap: 10px; }

/* 设置按钮 */
.icon-btn {
  width: 36px; height: 36px; border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--surface);
  display: grid; place-items: center; cursor: pointer;
  color: var(--text-2);
  box-shadow: var(--shadow-sm);
  transition: all .15s var(--ease);
}
.icon-btn:hover { background: var(--surface-3); color: var(--text-1); transform: translateY(-1px); box-shadow: var(--shadow-md); }
.icon-btn svg { width: 16px; height: 16px; stroke: currentColor; fill: none; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }

/* 设置弹层 */
.settings-wrap { position: relative; }
.settings-pop {
  position: absolute; right: 0; top: calc(100% + 8px);
  width: 220px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-md);
  box-shadow: var(--shadow-lg);
  z-index: 30;
  overflow: hidden;
  animation: pop-in .18s var(--ease);
}
@keyframes pop-in {
  from { opacity: 0; transform: translateY(-6px) scale(.97); }
  to { opacity: 1; transform: none; }
}
.settings-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 16px; font-size: 13px; font-weight: 600;
  border-bottom: 1px solid var(--border);
}
.settings-close { color: var(--text-3); font-size: 13px; }
.settings-close:hover { color: var(--text-1); }
.settings-body { padding: 6px; }
.settings-label { font-size: 12px; font-weight: 600; color: var(--text-2); padding: 8px 10px 0; }
.settings-tip { font-size: 11px; color: var(--text-3); padding: 2px 10px 6px; }
.quality-list { padding: 4px; }
.quality-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 12px; border-radius: var(--r-sm);
  cursor: pointer; transition: background .15s;
}
.quality-item:hover { background: var(--surface-3); }
.quality-item.on { background: var(--primary-soft); }
.quality-item.on .q-name { color: var(--primary-deep); font-weight: 700; }
.quality-item .q-tip { font-size: 11px; color: var(--text-3); }
.quality-item .q-check { margin-left: auto; color: var(--primary); font-weight: 700; }

.theme-toggle {
  display: flex; gap: 0;
  margin: 4px 10px 8px;
  background: var(--surface-3);
  border-radius: var(--r-sm);
  padding: 2px;
}
.theme-toggle span {
  flex: 1; text-align: center;
  padding: 6px 12px; border-radius: 6px;
  font-size: 12px; color: var(--text-3);
  cursor: pointer; transition: all .15s;
}
.theme-toggle span.on {
  background: var(--surface);
  color: var(--text-1);
  font-weight: 600;
  box-shadow: var(--shadow-sm);
}

/* 用户区（右上角） */
.user-entry {
  display: flex; align-items: center; gap: 8px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-full);
  padding: 4px 14px 4px 4px;
  cursor: pointer;
  transition: all .15s var(--ease);
}
.user-entry:hover { background: var(--surface-3); box-shadow: var(--shadow-sm); }
.user-avatar,
.user-avatar-img {
  width: 26px; height: 26px; border-radius: 50%;
  background: linear-gradient(135deg, #2CA6F8, #7C6BFF);
  display: grid; place-items: center;
  color: #fff; font-size: 11px; font-weight: 700; flex-shrink: 0;
}
.user-avatar-img { overflow: hidden; }
.user-avatar-img img { width: 100%; height: 100%; object-fit: cover; display: block; }
.user-name {
  font-size: 12.5px; font-weight: 600; color: var(--text-1);
  max-width: 96px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}

@media (max-width: 860px) {
  .topbar { padding: 12px 16px; }
}
</style>
