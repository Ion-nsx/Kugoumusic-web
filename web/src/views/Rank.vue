<template>
  <div class="page">
    <h1 class="page-title">听歌排行</h1>

    <template v-if="auth.isLoggedIn">
      <p class="page-sub" v-if="songs.length">累计播放 · 共 {{ songs.length }} 首</p>

      <div v-if="loading" class="loading">
        <div class="skeleton" style="width: 100%; height: 200px; border-radius: 12px;"></div>
      </div>

      <div v-else-if="songs.length" class="song-table fm-table">
        <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span>次数</span><span>时长</span></div>
        <div
          class="song-row"
          v-for="(song, idx) in songs"
          :key="song.id"
          :class="{ playing: isCurrent(song) }"
          @click="player.playList(songs, idx)"
        >
          <div class="idx">
            <span class="num">{{ pad(idx + 1) }}</span>
            <svg class="playing" viewBox="0 0 24 24"><rect x="7" y="4" width="3" height="16" rx="1.5" fill="#2CA6F8"/><rect x="12" y="8" width="3" height="12" rx="1.5" fill="#2CA6F8" opacity=".7"/><rect x="17" y="5" width="3" height="15" rx="1.5" fill="#2CA6F8" opacity=".4"/></svg>
          </div>
          <div class="s-cover">
            <img v-if="song.img" :src="imgUrl(song.img)" :alt="song.name" loading="lazy" />
            <svg v-else viewBox="0 0 24 24"><path d="M8.5 17.5a3.5 3.5 0 1 1-1-2.47l8-2.2V7.4L6.8 10.4a4.5 4.5 0 1 0 1.2 2.9l8-2.2V4.4a1 1 0 0 1 1.3-.95l3.2 1A1 1 0 0 1 21 5.4v8.1a4.5 4.5 0 1 0-1.3-3.2v3.9l-11.2 3.3Z"/></svg>
          </div>
          <div class="s-name">{{ song.name }}</div>
          <div class="s-album link" @click.stop="goAlbum(song.album_id)">{{ song.album || '单曲' }}</div>
          <div class="s-artist link" @click.stop="goArtist(song.singer_id)">{{ song.singer }}</div>
          <span class="col-play-count">{{ song.play_count || '' }}</span>
          <div class="dur">{{ fmtDuration(song.duration) }}</div>
        </div>
      </div>

      <div v-else class="empty">暂无听歌排行</div>
    </template>

    <div v-else class="empty">请先登录</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/player'
import { useAuthStore } from '../stores/auth'
import { getUserListen, imgUrl } from '../utils/api'

const router = useRouter()
const player = usePlayerStore()
const auth = useAuthStore()

const songs = ref([])
const loading = ref(false)

onMounted(() => {
  if (auth.isLoggedIn) load()
})

async function load() {
  loading.value = true
  try {
    const res = await getUserListen()
    if (res.data.status === 1) {
      const payload = res.data.data || res.data
      const list = (payload.lists || payload.list || payload.songs || []).map(normalize).filter(Boolean)
      songs.value = list
    }
  } catch (e) {
    console.error('加载排行失败:', e)
  } finally {
    loading.value = false
  }
}

function normalize(item) {
  const hash = item.hash != null ? String(item.hash) : ''
  if (!hash) return null
  const raw = Number(item.duration) || 0
  const duration = raw > 1000 ? Math.round(raw / 1000) : raw
  const singer = item.singername || item.singer || item.author_name || ''
  // 去掉 name 里的 "歌手 - " 前缀
  let name = item.name || item.ori_audio_name || item.songname || ''
  if (singer && name.startsWith(singer + ' - ')) {
    name = name.slice(singer.length + 3)
  }
  return {
    id: hash,
    mixsongid: item.mixsongid != null ? String(item.mixsongid) : String(item.album_audio_id || item.id || ''),
    name,
    singer,
    singer_id: '',
    album: item.album_name || item.album || '',
    album_id: item.album_audio_id != null ? String(item.album_audio_id) : (item.album_id != null ? String(item.album_id) : ''),
    duration,
    img: imgUrl(item.union_cover || item.image || item.cover || item.img || '', 240),
    hash,
    hash_320: item.hash_320,
    hash_flac: item.hash_flac || item.hash_sq,
    pay_type: item.pay_type || item.privilege,
    play_count: item.listen_count || item.play_count || '',
  }
}

function isCurrent(song) {
  return player.currentSong && player.currentSong.id === song.id
}

function pad(n) { return String(n).padStart(2, '0') }
function fmtDuration(seconds) {
  if (!seconds) return '--:--'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${pad(m)}:${pad(s)}`
}

function goAlbum(id) { if (id) router.push('/album/' + id) }
function goArtist(id) { if (id) router.push('/artist/' + id) }
</script>

<style scoped>
.page-sub { color: var(--text-3); font-size: var(--font-size-sm); margin-bottom: var(--spacing-lg); }

.fm-table :deep(.song-head),
.fm-table :deep(.song-row) {
  grid-template-columns: 36px 52px 2fr 1fr 1fr 52px 52px;
}
.col-play-count {
  color: var(--text-3); font-size: 12px; text-align: right;
  font-variant-numeric: tabular-nums;
}
</style>
