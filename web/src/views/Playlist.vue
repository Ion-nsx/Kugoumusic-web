<template>
  <div class="page detail-page">
    <BackBar title="歌单" />

    <template v-if="loading">
      <div class="loading">
        <div class="skeleton" style="width: 200px; height: 200px; border-radius: 16px; margin-bottom: 20px;"></div>
        <div class="skeleton" style="width: 60%; height: 24px; margin-bottom: 12px;"></div>
        <div class="skeleton" style="width: 40%; height: 16px;"></div>
      </div>
    </template>

    <template v-else-if="playlist">
      <div class="detail-wrap">
      <div class="detail-header">
        <div class="detail-cover">
          <img v-if="playlist.cover" :src="imgUrl(playlist.cover)" :alt="playlist.name" />
          <div v-else class="img-placeholder"></div>
        </div>
        <div class="detail-info">
          <h1 class="detail-name">{{ playlist.name }}</h1>
          <p class="detail-sub" v-if="playlist.author">{{ playlist.author }}</p>
          <p class="detail-desc" v-if="playlist.desc">{{ playlist.desc }}</p>
          <div class="detail-meta">
            <span v-if="playlist.count">{{ playlist.count }} 首</span>
            <span v-if="playlist.play_count">{{ formatCount(playlist.play_count) }} 次播放</span>
          </div>
          <div class="detail-actions">
            <button class="btn btn-primary" @click="playAll">
              <svg viewBox="0 0 24 24" style="fill:currentColor;stroke:none"><path d="M8 5v14l11-7Z"/></svg>
              播放全部
            </button>
            <button class="btn btn-ghost" @click="showComments = true">
              <svg viewBox="0 0 24 24"><path d="M21 12a8 8 0 0 1-8 8H4l2.2-2.6A8 8 0 1 1 21 12Z"/></svg>
              评论
            </button>
          </div>
        </div>
      </div>

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
      </div>
    </template>

    <!-- 评论弹层 -->
    <CommentModal
      v-if="showComments"
      resource-type="playlist"
      :title="playlist?.name || '歌单评论'"
      :id="playlist?.comment_id || playlist?.id || String(route.params.id)"
      @close="showComments = false"
    />

    <div v-if="!loading && !playlist" class="empty">歌单不存在</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/player'
import { getPlaylistDetail, getPlaylistSongs, imgUrl } from '../utils/api'
import BackBar from '../components/BackBar.vue'
import CommentModal from '../components/CommentModal.vue'

const route = useRoute()
const router = useRouter()
const player = usePlayerStore()

const PAGE_SIZE = 30
const playlist = ref(null)
const songs = ref([])
const loading = ref(true)
const showComments = ref(false)
const page = ref(1)
const totalPages = ref(1)
const switchingPage = ref(false)
const jumpPage = ref('')

const pages = computed(() => {
  const t = totalPages.value
  const p = page.value
  const result = []
  if (t <= 7) { for (let i = 1; i <= t; i++) result.push(i); return result }
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
    const [detailRes, songsRes] = await Promise.all([
      getPlaylistDetail(id),
      getPlaylistSongs(id, 1, PAGE_SIZE)
    ])

    if (detailRes.data.status === 1 && detailRes.data.data) {
      playlist.value = detailRes.data.data
      const cnt = Number(playlist.value.count) || 0
      if (cnt > 0) {
        totalPages.value = Math.ceil(cnt / PAGE_SIZE)
      }
    }

    if (songsRes.data.status === 1 && Array.isArray(songsRes.data.data)) {
      songs.value = songsRes.data.data
      if (playlist.value && !playlist.value.cover && songs.value.length && songs.value[0].img) {
        playlist.value.cover = songs.value[0].img
      }
    }
  } catch (e) {
    console.error('获取歌单详情失败:', e)
  } finally {
    loading.value = false
  }
})

async function goPage(pg) {
  if (switchingPage.value || pg === page.value || pg < 1 || pg > totalPages.value) return
  switchingPage.value = true
  page.value = pg
  jumpPage.value = ''
  try {
    const res = await getPlaylistSongs(route.params.id, pg, PAGE_SIZE)
    if (res.data.status === 1 && Array.isArray(res.data.data)) {
      songs.value = res.data.data
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

function pad(n) { return String(n).padStart(2, '0') }
function formatDuration(seconds) {
  if (!seconds) return '--:--'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${pad(s)}`
}

function formatCount(count) {
  count = Number(count) || 0
  if (count >= 10000) return (count / 10000).toFixed(1) + '万'
  return String(count)
}

function goAlbum(id) { if (id) router.push('/album/' + id) }
function goArtist(id) { if (id) router.push('/artist/' + id) }
</script>

<style scoped>
.detail-page { max-width: 960px; }

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
</style>
