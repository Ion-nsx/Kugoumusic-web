<template>
  <div class="page me-page">
    <!-- 头部用户卡 -->
    <div class="me-hero">
      <div class="me-avatar" :class="{ placeholder: !auth.isLoggedIn }">
        <img v-if="auth.isLoggedIn && auth.user?.avatar" class="me-avatar-img" :src="auth.user.avatar" alt="头像" />
        <template v-else-if="auth.isLoggedIn">{{ avatarText }}</template>
        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"><circle cx="12" cy="8" r="4"/><path d="M4 21c0-3.3 3.6-6 8-6s8 2.7 8 6"/></svg>
      </div>

      <div class="me-info">
        <div class="me-name">{{ auth.isLoggedIn ? (auth.user?.name || '酷狗用户') : '未登录' }}</div>
        <div class="me-sub">{{ auth.isLoggedIn ? '欢迎回来' : '登录后同步你的收藏与歌单' }}</div>
      </div>

      <button v-if="!auth.isLoggedIn" class="btn btn-primary btn-sm" @click="openLogin">立即登录</button>
      <button v-else class="btn btn-ghost btn-sm" @click="auth.logout()">退出登录</button>
    </div>

    <!-- 我的歌单列表 -->
    <section v-if="auth.isLoggedIn" class="me-section">
      <div class="section-head">
        <div class="section-title">我的歌单</div>
        <div class="section-more">
          <button class="btn btn-primary btn-sm" @click="showCreate = !showCreate">＋ 创建歌单</button>
        </div>
      </div>

      <!-- 创建歌单 -->
      <div class="create-box" v-if="showCreate">
        <input v-model="newPlaylistName" class="create-input" placeholder="输入歌单名称" @keyup.enter="doCreatePlaylist" />
        <button class="btn btn-primary btn-sm" @click="doCreatePlaylist" :disabled="creating">{{ creating ? '创建中...' : '创建' }}</button>
        <button class="btn btn-ghost btn-sm" @click="showCreate = false">取消</button>
      </div>

      <div v-if="loadingPlaylists" class="skeleton pl-skeleton"></div>
      <div v-else-if="myPlaylists.length" class="pl-list">
        <div
          class="pl-item"
          v-for="(pl, idx) in myPlaylists"
          :key="pl.id"
          @click="goPlaylist(pl.id)"
        >
          <img v-if="pl.cover" class="pl-cover" :src="imgUrl(pl.cover)" :alt="pl.name" loading="lazy" />
          <div v-else class="pl-cover placeholder">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"><rect x="3" y="5" width="18" height="14" rx="2"/><path d="M8 9.5h8M8 12.5h8M8 15.5h4"/></svg>
          </div>
          <div class="pl-meta">
            <div class="pl-name">{{ pl.name }}</div>
            <div class="pl-count">{{ pl.count || 0 }} 首</div>
          </div>
          <span class="pl-del" title="删除歌单" @click.stop="doDeletePlaylist(pl)">🗑</span>
          <span class="pl-arrow">›</span>
        </div>
      </div>
      <div v-else class="pl-empty">暂无歌单，去首页逛逛吧</div>
    </section>

    <!-- 关注歌手 -->
    <section v-if="auth.isLoggedIn" class="me-section">
      <div class="section-head">
        <div class="section-title">关注歌手</div>
      </div>

      <div v-if="loadingFollows" class="skeleton pl-skeleton"></div>
      <div v-else-if="followArtists.length" class="pl-list">
        <div
          class="pl-item"
          v-for="(s, idx) in followArtists"
          :key="s.singerid || idx"
          @click="goArtist(s.singerid)"
        >
          <img v-if="s.pic" class="pl-cover" :src="imgUrl(s.pic)" :alt="s.nickname" loading="lazy" />
          <div v-else class="pl-cover placeholder">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"><circle cx="12" cy="8" r="5"/><path d="M3 21c0-2.8 4-5 9-5s9 2.2 9 5"/></svg>
          </div>
          <div class="pl-meta">
            <div class="pl-name">{{ s.nickname }}</div>
            <div class="pl-count">{{ s.score ? s.score + ' 分' : '' }}</div>
          </div>
          <button class="btn btn-ghost btn-sm" @click.stop="doUnfollow(s)" :disabled="s._unfollowing">
            {{ s._unfollowing ? '...' : '取消关注' }}
          </button>
          <span class="pl-arrow">›</span>
        </div>
      </div>
      <div v-else class="pl-empty">还没有关注的歌手，去发现页看看吧</div>
    </section>

    <!-- 设置 -->
    <div class="me-settings">
      <div class="set-item" @click="player.clearQueue()">
        <svg class="set-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M3 6h18M8 6V4h8v2M6 6l1 14h10l1-14M10 11v6M14 11v6"/></svg>
        <span class="set-text">清空播放队列</span>
        <span class="pl-arrow">›</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, inject, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { usePlayerStore } from '../stores/player'
import { getUserPlaylist, createPlaylist, deletePlaylist, getUserFollow, followSinger, imgUrl } from '../utils/api'

const auth = useAuthStore()
const player = usePlayerStore()
const router = useRouter()
const showLogin = inject('showLogin')

const myPlaylists = ref([])
const loadingPlaylists = ref(false)
const showCreate = ref(false)
const creating = ref(false)
const newPlaylistName = ref('')
const followArtists = ref([])
const loadingFollows = ref(false)

const avatarText = computed(() => {
  if (!auth.isLoggedIn) return ''
  const name = auth.user?.name || 'K'
  return name.slice(0, 1).toUpperCase()
})

watch(() => auth.isLoggedIn, (logged) => {
  if (logged) {
    loadPlaylists()
    loadFollows()
  } else {
    myPlaylists.value = []
    followArtists.value = []
  }
}, { immediate: true })

async function loadPlaylists() {
  loadingPlaylists.value = true
  try {
    const res = await getUserPlaylist()
    if (res.data.status === 1 && res.data.data) {
      myPlaylists.value = res.data.data
    }
  } catch (e) {
    console.error('获取我的歌单失败:', e)
  } finally {
    loadingPlaylists.value = false
  }
}

async function loadFollows() {
  loadingFollows.value = true
  try {
    const res = await getUserFollow()
    const data = res.data
    // 兼容多种响应格式：酷狗可能把列表放在 data.data.info / data.data.list / data.info 等
    const inner = data?.data || data
    const list = inner?.lists || inner?.info || inner?.list || []
    followArtists.value = Array.isArray(list) ? list : []
  } catch (e) {
    console.error('获取关注列表失败:', e)
  } finally {
    loadingFollows.value = false
  }
}

async function doUnfollow(s) {
  if (s._unfollowing) return
  s._unfollowing = true
  try {
    await followSinger(s.singerid, false)
    followArtists.value = followArtists.value.filter(a => a.singerid !== s.singerid)
  } catch (e) {
    if (e.response?.status !== 401) alert('操作失败')
  } finally {
    s._unfollowing = false
  }
}

function goArtist(id) {
  if (id) router.push(`/artist/${id}`)
}

async function doCreatePlaylist() {
  const name = newPlaylistName.value.trim()
  if (!name) return
  creating.value = true
  try {
    const res = await createPlaylist(name)
    if (res.data.status === 1 && (res.data.err_code === 0 || res.data.error_code === 0 || res.data.data)) {
      newPlaylistName.value = ''
      showCreate.value = false
      loadPlaylists()
    } else {
      alert(res.data.errmsg || res.data.msg || '创建失败')
    }
  } catch (e) {
    if (e.response?.status !== 401) alert('网络错误')
  } finally {
    creating.value = false
  }
}

async function doDeletePlaylist(pl) {
  if (!window.confirm(`确定删除歌单「${pl.name}」吗？`)) return
  try {
    const res = await deletePlaylist(pl.id)
    if (res.data.status === 1) {
      loadPlaylists()
    } else {
      alert(res.data.errmsg || res.data.msg || '删除失败')
    }
  } catch (e) {
    if (e.response?.status !== 401) alert('网络错误')
  }
}

function openLogin() {
  if (showLogin) showLogin.value = true
}

function goPlaylist(id) {
  router.push(`/playlist/${id}`)
}
</script>

<style scoped>
.me-page { max-width: 820px; }

/* ===== 头部用户卡 ===== */
.me-hero {
  position: relative;
  display: flex; align-items: center; gap: var(--spacing-xl);
  padding: var(--spacing-3xl);
  margin-bottom: var(--spacing-xl);
  border-radius: var(--r-lg);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-md);
  overflow: hidden;
  background:
    radial-gradient(420px 200px at 90% 0%, rgba(124,107,255,.12), transparent 70%),
    radial-gradient(420px 200px at 0% 100%, rgba(44,166,248,.14), transparent 70%),
    var(--surface);
}
.me-avatar {
  width: 72px; height: 72px; flex-shrink: 0;
  border-radius: 50%;
  background: linear-gradient(135deg, #2CA6F8, #7C6BFF);
  display: grid; place-items: center;
  font-size: 28px; font-weight: 700; color: #fff;
  box-shadow: 0 6px 20px rgba(44,166,248,.35);
}
.me-avatar.placeholder { background: var(--surface-3); color: var(--text-3); box-shadow: none; }
.me-avatar svg { width: 30px; height: 30px; }
.me-avatar-img { width: 100%; height: 100%; border-radius: 50%; object-fit: cover; display: block; }
.me-info { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px; }
.me-name { font-size: 22px; font-weight: var(--font-weight-bold); letter-spacing: -0.02em; }
.me-sub { font-size: var(--font-size-sm); color: var(--text-3); }

/* ===== 我的歌单 ===== */
.me-section { margin-bottom: var(--spacing-3xl); }
.pl-skeleton { width: 100%; height: 120px; border-radius: var(--r-lg); }
.pl-list {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-md);
  padding: var(--spacing-sm) var(--spacing-lg);
  box-shadow: var(--shadow-sm);
}
.pl-item {
  display: flex; align-items: center; gap: var(--spacing-lg);
  padding: var(--spacing-md); cursor: pointer; border-radius: var(--r-sm);
  transition: background .15s;
}
.pl-item:hover { background: var(--surface-3); }
.pl-item + .pl-item { border-top: 1px solid var(--border); }
.pl-cover {
  width: 48px; height: 48px; flex-shrink: 0;
  border-radius: var(--r-sm); object-fit: cover; background: var(--surface-3);
}
.pl-cover.placeholder { display: grid; place-items: center; color: var(--text-3); }
.pl-cover.placeholder svg { width: 20px; height: 20px; }
.pl-meta { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.pl-name { font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pl-count { font-size: var(--font-size-xs); color: var(--text-3); }
.pl-arrow { color: var(--text-3); font-size: 20px; }
.pl-del {
  color: var(--text-3); font-size: 14px; padding: 4px 6px;
  border-radius: var(--r-sm); transition: all var(--transition-fast); flex-shrink: 0;
}
.pl-del:hover { color: var(--like); background: var(--surface-3); }

/* 创建歌单 */
.create-box {
  display: flex; align-items: center; gap: var(--spacing-sm);
  margin-bottom: var(--spacing-lg); padding: var(--spacing-sm);
  background: var(--surface); border: 1px solid var(--border);
  border-radius: var(--r-md);
}
.create-input {
  flex: 1; padding: 8px 14px;
  border: 1px solid var(--border-strong); border-radius: var(--r-full);
  background: var(--surface-2); font-size: var(--font-size-sm); outline: none;
}
.create-input:focus { border-color: var(--primary); }
.pl-empty {
  padding: var(--spacing-3xl); text-align: center;
  font-size: var(--font-size-sm); color: var(--text-3);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-md);
}

/* ===== 设置 ===== */
.me-settings {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-md);
  padding: var(--spacing-sm) var(--spacing-lg);
}
.set-item {
  display: flex; align-items: center; gap: var(--spacing-lg);
  padding: var(--spacing-md); cursor: pointer; border-radius: var(--r-sm);
  transition: background .15s;
}
.set-item:hover { background: var(--surface-3); }
.set-icon { width: 18px; height: 18px; color: var(--text-3); flex-shrink: 0; }
.set-text { flex: 1; font-size: var(--font-size-sm); font-weight: var(--font-weight-medium); color: var(--text-2); }
</style>
