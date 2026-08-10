<template>
  <div class="page detail-page">
    <BackBar title="歌手" />

    <div v-if="loading" class="loading">
      <div class="skeleton" style="width: 160px; height: 160px; border-radius: 50%; margin-bottom: 20px;"></div>
      <div class="skeleton" style="width: 40%; height: 24px;"></div>
    </div>

    <div v-else-if="artist" class="detail-wrap">
      <div class="detail-header">
        <div class="detail-avatar">
          <img v-if="artist.pic" :src="imgUrl(artist.pic)" :alt="artist.name" />
          <div v-else class="img-placeholder"></div>
        </div>
        <div class="detail-info">
          <h1 class="detail-name">{{ artist.name }}</h1>
          <p class="detail-sub" v-if="artist.count">{{ artist.count }} 首歌曲</p>
          <p class="detail-desc" v-if="artist.desc">{{ artist.desc }}</p>
          <div class="detail-actions">
            <button class="btn btn-primary" @click="playAll">
              <svg viewBox="0 0 24 24" style="fill:currentColor;stroke:none"><path d="M8 5v14l11-7Z"/></svg>
              播放全部
            </button>
            <button class="btn btn-ghost" @click="toggleFollow">
              <svg viewBox="0 0 24 24"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M19 8v6M22 11h-6"/></svg>
              {{ followed ? '已关注' : '关注' }}
            </button>
          </div>
        </div>
      </div>

      <!-- 标签切换 -->
      <div class="artist-tabs">
        <button class="tab-btn" :class="{ active: tab === 'songs' }" @click="switchTab('songs')">单曲</button>
        <button class="tab-btn" :class="{ active: tab === 'albums' }" @click="switchTab('albums')">专辑</button>
        <button class="tab-btn" :class="{ active: tab === 'detail' }" @click="switchTab('detail')">详情</button>
      </div>

      <!-- 单曲列表 -->
      <template v-if="tab === 'songs'">
        <div class="song-table" v-if="songs.length">
          <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span></span><span>时长</span></div>
          <div class="song-row" v-for="(song, idx) in songs" :key="song.id" @click="player.playList(songs, idx)">
            <div class="idx"><span class="num">{{ pad((page - 1) * PAGE_SIZE + idx + 1) }}</span></div>
            <div class="s-cover">
              <img v-if="song.img" :src="imgUrl(song.img)" :alt="song.name" loading="lazy" />
            </div>
            <div class="s-name">{{ song.name }}</div>
            <div class="s-album link" @click.stop="goAlbum(song.album_id)">{{ song.album }}</div>
            <div class="s-artist link" @click.stop="goArtist(song.singer_id)">{{ song.singer }}</div>
            <button class="like" :class="{ liked: player.isLiked(song) }" @click.stop="player.toggleLike(song)">
              <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
            </button>
            <div class="dur">{{ formatDuration(song.duration) }}</div>
          </div>
        </div>
        <div class="pagination" v-if="totalPages > 1">
          <button class="pg-btn" :disabled="page === 1" @click="goPage(1)">首页</button>
          <button class="pg-btn" :disabled="page === 1" @click="goPage(page - 1)">‹</button>
          <button class="pg-num" v-for="p in pages" :key="p" :class="{ active: p === page, dots: p === '...' }" :disabled="p === '...'" @click="p !== '...' && goPage(p)">{{ p }}</button>
          <button class="pg-btn" :disabled="page >= totalPages" @click="goPage(page + 1)">›</button>
          <button class="pg-btn" :disabled="page >= totalPages" @click="goPage(totalPages)">末页</button>
          <span class="pg-jump">跳至 <input class="pg-input" v-model="jumpPage" @keyup.enter="goJump" /> 页 <button class="pg-go" @click="goJump">GO</button></span>
        </div>
        <div v-else-if="songsLoaded" class="empty" style="padding:40px">暂无歌曲</div>
      </template>

      <!-- 专辑列表 -->
      <template v-if="tab === 'albums'">
        <div class="album-grid" v-if="albums.length">
          <div class="album-card" v-for="a in albums" :key="a.id" @click="goAlbum(a.id)">
            <div class="ac-cover">
              <img v-if="a.cover" :src="imgUrl(a.cover)" :alt="a.name" loading="lazy" />
              <svg v-else viewBox="0 0 24 24"><path d="M8.5 17.5a3.5 3.5 0 1 1-1-2.47l8-2.2V7.4L6.8 10.4a4.5 4.5 0 1 0 1.2 2.9l8-2.2V4.4a1 1 0 0 1 1.3-.95l3.2 1A1 1 0 0 1 21 5.4v8.1a4.5 4.5 0 1 0-1.3-3.2v3.9l-11.2 3.3Z"/></svg>
            </div>
            <div class="ac-name">{{ a.name }}</div>
            <div class="ac-time">{{ a.time || '' }}</div>
          </div>
        </div>
        <div v-else-if="albumsLoaded" class="empty" style="padding:40px">暂无专辑</div>
      </template>

      <!-- 详情 -->
      <template v-if="tab === 'detail'">
        <div class="artist-detail" v-if="artist">
          <div class="detail-row" v-if="artist.name"><span class="dl">歌手</span><span>{{ artist.name }}</span></div>
          <div class="detail-row" v-if="artist.count"><span class="dl">歌曲数</span><span>{{ artist.count }}</span></div>
          <div class="detail-row" v-if="artist.sex"><span class="dl">性别</span><span>{{ artist.sex }}</span></div>
          <div class="detail-row" v-if="artist.area"><span class="dl">地区</span><span>{{ artist.area }}</span></div>
          <div class="detail-row" v-if="artist.type"><span class="dl">类型</span><span>{{ artist.type }}</span></div>
          <div class="detail-row" v-if="artist.desc"><span class="dl">简介</span><span>{{ artist.desc }}</span></div>
        </div>
        <div v-else class="empty" style="padding:40px">暂无详情</div>
      </template>
    </div>

    <div v-else class="empty">歌手不存在</div>
  </div>
</template>

<script setup>
import { ref, inject, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/player'
import { useAuthStore } from '../stores/auth'
import { getArtistDetail, getArtistAudios, getArtistAlbums, followSinger, imgUrl } from '../utils/api'
import BackBar from '../components/BackBar.vue'

const route = useRoute()
const router = useRouter()
const player = usePlayerStore()
const auth = useAuthStore()
const requireLogin = inject('requireLogin', () => {})

const PAGE_SIZE = 30
const artist = ref(null)
const songs = ref([])
const albums = ref([])
const loading = ref(true)
const followed = ref(false)
const following = ref(false)
const page = ref(1)
const totalPages = ref(1)
const switchingPage = ref(false)
const jumpPage = ref('')
const tab = ref('songs')
const songsLoaded = ref(false)
const albumsLoaded = ref(false)

const pages = computed(() => {
  const t = totalPages.value
  const p = page.value
  const result = []
  if (t <= 7) {
    for (let i = 1; i <= t; i++) result.push(i)
    return result
  }
  result.push(1)
  if (p > 3) result.push('...')
  for (let i = Math.max(2, p - 1); i <= Math.min(t - 1, p + 1); i++) result.push(i)
  if (p < t - 2) result.push('...')
  result.push(t)
  return result
})

onMounted(async () => {
  const id = route.params.id
  try {
    const [detailRes, audiosRes] = await Promise.all([
      getArtistDetail(id),
      getArtistAudios(id, 1, PAGE_SIZE)
    ])

    if (detailRes.data.status === 1 && detailRes.data.data) {
      artist.value = detailRes.data.data
    }

    if (audiosRes.data.status === 1 && audiosRes.data.data) {
      const d = audiosRes.data.data
      const list = Array.isArray(d) ? d : (d.lists || d.songs || [])
      setSongs(list, 1)
    }
    songsLoaded.value = true
  } catch (e) {
    console.error('获取歌手信息失败:', e)
    songsLoaded.value = true
  } finally {
    loading.value = false
  }
})

async function switchTab(t) {
  tab.value = t
  if (t === 'albums' && !albumsLoaded.value) {
    loadAlbums()
  }
}

async function loadAlbums() {
  try {
    const res = await getArtistAlbums(route.params.id, 1, 50)
    if (res.data.status === 1 && res.data.data) {
      const list = Array.isArray(res.data.data) ? res.data.data : (res.data.data.lists || res.data.data.albums || [])
      albums.value = list.map(a => ({
        id: a.ID || a.id || a.albumid || a.album_id || '',
        name: a.Name || a.name || a.albumname || a.title || '',
        cover: a.Cover || a.sizable_cover || a.imgurl || a.cover || a.image || a.pic || '',
        time: a.Time || a.publish_time || a.time || a.publishtime || ''
      }))
    }
  } catch (e) {
    console.error('加载专辑失败:', e)
  } finally {
    albumsLoaded.value = true
  }
}

function enrichSongs(list) {
  return list.map(s => ({ ...s, img: s.img || artist.value?.pic || '' }))
}

function setSongs(list, pg) {
  songs.value = enrichSongs(list)
  page.value = pg
  if (list.length < PAGE_SIZE) {
    totalPages.value = pg
  } else if (artist.value?.count) {
    totalPages.value = Math.max(pg, Math.ceil(artist.value.count / PAGE_SIZE))
  } else {
    totalPages.value = pg + 1
  }
}

async function goPage(pg) {
  if (switchingPage.value || pg === page.value || pg < 1 || pg > totalPages.value) return
  switchingPage.value = true
  page.value = pg
  jumpPage.value = ''
  try {
    const res = await getArtistAudios(route.params.id, pg, PAGE_SIZE)
    if (res.data.status === 1 && res.data.data) {
      const d = res.data.data
      const list = Array.isArray(d) ? d : (d.lists || d.songs || [])
      if (list.length > 0) {
        setSongs(list, pg)
      } else {
        totalPages.value = pg - 1
      }
    }
  } catch (e) {
    console.error('切换页面失败:', e)
  } finally {
    switchingPage.value = false
  }
}

function goJump() {
  const n = parseInt(jumpPage.value)
  if (n >= 1 && n <= totalPages.value) goPage(n)
}

function playAll() {
  player.playList(songs.value)
}

async function toggleFollow() {
  if (!auth.isLoggedIn) {
    requireLogin('关注歌手需要登录')
    return
  }
  if (following.value || !artist.value?.id) return
  following.value = true
  try {
    const target = !followed.value
    const res = await followSinger(artist.value.id, target)
    if (res.data.status === 1) {
      followed.value = target
    } else {
      alert(res.data.errmsg || res.data.msg || '操作失败')
    }
  } catch (e) {
    if (e.response?.status !== 401) alert('网络错误')
  } finally {
    following.value = false
  }
}

function pad(n) { return String(n).padStart(2, '0') }
function formatDuration(seconds) {
  if (!seconds) return '--:--'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${pad(s)}`
}

function goAlbum(id) { if (id) router.push('/album/' + id) }
function goArtist(id) { if (id) router.push('/artist/' + id) }
</script>

<style scoped>
.detail-page { max-width: 960px; }

.artist-tabs {
  display: flex; gap: 4px; margin: 20px 0 16px;
  border-bottom: 1px solid var(--border); padding-bottom: 6px;
}
.tab-btn {
  padding: 8px 20px; border-radius: 8px;
  border: none; background: transparent;
  font-size: 14px; font-weight: 600;
  color: var(--text-3); cursor: pointer;
  transition: all .15s;
}
.tab-btn:hover { color: var(--text-1); background: var(--surface-3); }
.tab-btn.active { color: var(--primary); background: var(--primary-soft); }

.pagination {
  display: flex; align-items: center; justify-content: center;
  gap: 4px; padding: 24px 0;
}
.pg-btn, .pg-num {
  min-width: 34px; height: 34px; border-radius: 8px;
  border: 1px solid var(--border); background: var(--surface);
  font-size: 13px; color: var(--text-2); cursor: pointer;
  display: grid; place-items: center; transition: all .15s;
}
.pg-btn:disabled { opacity: .35; cursor: default; }
.pg-btn:hover:not(:disabled) { background: var(--surface-3); color: var(--text-1); }
.pg-num.active { background: var(--primary); color: #fff; border-color: var(--primary); }
.pg-num.dots { border: none; background: transparent; cursor: default; color: var(--text-4); }
.pg-num:hover:not(.active):not(.dots) { background: var(--surface-3); color: var(--text-1); }
.pg-jump { font-size: 13px; color: var(--text-2); margin-left: 12px; display: flex; align-items: center; gap: 4px; }
.pg-input { width: 44px; height: 28px; border: 1px solid var(--border); border-radius: 6px; text-align: center; font-size: 13px; background: var(--surface); }
.pg-go { height: 28px; padding: 0 10px; border: 1px solid var(--border); border-radius: 6px; background: var(--surface); font-size: 12px; cursor: pointer; color: var(--text-2); }
.pg-go:hover { border-color: var(--primary); color: var(--primary); }

/* 专辑网格 */
.album-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 16px; }
.album-card { cursor: pointer; transition: transform .15s; }
.album-card:hover { transform: translateY(-2px); }
.ac-cover {
  aspect-ratio: 1; border-radius: 12px; overflow: hidden;
  background: var(--surface-3); display: grid; place-items: center;
  box-shadow: var(--shadow-md);
}
.ac-cover img { width: 100%; height: 100%; object-fit: cover; }
.ac-cover svg { width: 30px; height: 30px; fill: var(--text-3); }
.ac-name {
  margin-top: 8px; font-size: 13px; font-weight: 600;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.ac-time { font-size: 12px; color: var(--text-3); margin-top: 2px; }

/* 详情行 */
.artist-detail { padding: 8px 0; }
.detail-row { display: flex; padding: 12px 0; border-bottom: 1px solid var(--border); font-size: 14px; }
.detail-row:last-child { border-bottom: none; }
.dl { width: 80px; color: var(--text-3); flex-shrink: 0; }
</style>
