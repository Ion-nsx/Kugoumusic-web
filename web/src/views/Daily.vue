<template>
  <div class="page daily-page">
    <h1 class="page-title">每日推荐</h1>
    <p class="page-sub" v-if="dailySongs.length">{{ dailySongs.length }} 首今日推荐 · 根据你的口味每日更新</p>

    <div v-if="loading" class="loading">
      <div class="skeleton" style="width: 100%; height: 200px; border-radius: 12px;"></div>
    </div>

    <div v-else-if="dailySongs.length" class="song-table">
      <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span></span><span>时长</span></div>
      <div
        class="song-row"
        v-for="(song, idx) in dailySongs"
        :key="song.id"
        :class="{ playing: isCurrent(song) }"
        @click="player.playList(dailySongs, idx)"
      >
        <div class="idx">
          <span class="num">{{ pad(idx + 1) }}</span>
          <svg class="playing" viewBox="0 0 24 24"><rect x="7" y="4" width="3" height="16" rx="1.5" fill="#2CA6F8"/><rect x="12" y="8" width="3" height="12" rx="1.5" fill="#2CA6F8" opacity=".7"/><rect x="17" y="5" width="3" height="15" rx="1.5" fill="#2CA6F8" opacity=".4"/></svg>
        </div>
        <div class="s-cover">
          <img v-if="song.img" :src="song.img" :alt="song.name" loading="lazy" />
          <svg v-else viewBox="0 0 24 24"><path d="M8.5 17.5a3.5 3.5 0 1 1-1-2.47l8-2.2V7.4L6.8 10.4a4.5 4.5 0 1 0 1.2 2.9l8-2.2V4.4a1 1 0 0 1 1.3-.95l3.2 1A1 1 0 0 1 21 5.4v8.1a4.5 4.5 0 1 0-1.3-3.2v3.9l-11.2 3.3Z"/></svg>
        </div>
        <div class="s-name">{{ song.name }}<span v-if="idx === 0" class="tag">独家</span></div>
        <div class="s-album link" @click.stop="goAlbum(song.album_id)">{{ song.album }}</div>
        <div class="s-artist link" @click.stop="goArtist(song.singer_id)">{{ song.singer }}</div>
        <button class="like" :class="{ liked: player.isLiked(song) }" @click.stop="player.toggleLike(song)">
          <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
        </button>
        <div class="dur">{{ fmtDuration(song.duration) }}</div>
      </div>
    </div>

    <div v-else class="empty">暂无每日推荐，请稍后重试</div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
const router = useRouter()
import { ref, onMounted } from 'vue'
import { usePlayerStore } from '../stores/player'
import { getEverydayRecommend, imgUrl } from '../utils/api'

const player = usePlayerStore()

const dailySongs = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getEverydayRecommend()
    if (res.data.status === 1 && res.data.data && res.data.data.song_list) {
      dailySongs.value = (res.data.data.song_list || []).map(normalizeDaily).filter(Boolean)
    }
  } catch (e) {
    console.error('获取每日推荐失败:', e)
  } finally {
    loading.value = false
  }
})

// 每日推荐字段归一化
function normalizeDaily(item) {
  const hash = item.hash != null ? String(item.hash) : ''
  if (!hash) return null
  // time_length 单位可能是秒或毫秒，>1000 视为毫秒
  const raw = Number(item.time_length) || 0
  const duration = raw > 1000 ? Math.round(raw / 1000) : raw
  return {
    id: hash,
    mixsongid: item.mixsongid != null ? String(item.mixsongid) : '',
    name: item.ori_audio_name || item.songname || '',
    singer: item.author_name || '',
    singer_id: item.author_id != null ? String(item.author_id) : '',
    album: item.album_name || '',
    album_id: item.album_id != null ? String(item.album_id) : '',
    duration,
    img: imgUrl(item.sizable_cover, 240)
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
.daily-page { max-width: 1060px; }
.page-title { margin-bottom: 4px; }
.page-sub { color: var(--text-3); font-size: var(--font-size-sm); margin-bottom: var(--spacing-xl); }
</style>
