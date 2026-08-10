<template>
  <div class="page fm-page">
    <div class="fm-header">
      <div class="fm-title-row">
        <h1 class="page-title">猜你喜欢</h1>
        <button class="refresh-btn" @click="loadFM" :disabled="loading">
          <svg viewBox="0 0 24 24" :class="{ spinning: loading }"><path d="M21 2v5h-5M3.6 6.6A9 9 0 0 1 19.4 6M3 22v-5h5M20.4 17.4A9 9 0 0 1 4.6 18"/></svg>
          换一批
        </button>
      </div>
      <p class="page-sub" v-if="songs.length">为你推荐 {{ songs.length }} 首歌 · 基于你的听歌口味</p>
    </div>

    <div v-if="loading && !songs.length" class="loading">
      <div class="skeleton" style="width: 100%; height: 200px; border-radius: 12px;"></div>
    </div>

    <div v-else-if="songs.length" class="song-table">
      <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span></span><span>时长</span></div>
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
          <img v-if="song.img" :src="song.img" :alt="song.name" loading="lazy" />
          <svg v-else viewBox="0 0 24 24"><path d="M8.5 17.5a3.5 3.5 0 1 1-1-2.47l8-2.2V7.4L6.8 10.4a4.5 4.5 0 1 0 1.2 2.9l8-2.2V4.4a1 1 0 0 1 1.3-.95l3.2 1A1 1 0 0 1 21 5.4v8.1a4.5 4.5 0 1 0-1.3-3.2v3.9l-11.2 3.3Z"/></svg>
        </div>
        <div class="s-name">{{ song.name }}</div>
        <div class="s-album link" @click.stop="goAlbum(song.album_id)">{{ song.album }}</div>
        <div class="s-artist link" @click.stop="goArtist(song.singer_id)">{{ song.singer }}</div>
        <button class="like" :class="{ liked: player.isLiked(song) }" @click.stop="player.toggleLike(song)">
          <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
        </button>
        <div class="dur">{{ fmtDuration(song.duration) }}</div>
      </div>
    </div>

    <div v-else class="empty">暂无推荐，请稍后重试</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/player'
import { getFMSongs, imgUrl } from '../utils/api'

const router = useRouter()
const player = usePlayerStore()

const songs = ref([])
const loading = ref(true)

onMounted(loadFM)

function pad(n) { return String(n).padStart(2, '0') }
function fmtDuration(seconds) {
  if (!seconds) return '--:--'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${pad(m)}:${pad(s)}`
}

function isCurrent(song) {
  return player.currentSong && player.currentSong.id === song.id
}

function goAlbum(id) { if (id) router.push('/album/' + id) }
function goArtist(id) { if (id) router.push('/artist/' + id) }

async function loadFM() {
  loading.value = true
  try {
    const map = new Map()
    // 每次 API 返回 5 首，调 2 次凑够 10 首
    for (let i = 0; i < 2; i++) {
      const res = await getFMSongs('play', '0', 0)
      if (res.data.status === 1 && res.data.data) {
        const list = res.data.data.song_list || []
        for (const item of list) {
          const song = normalizeFM(item)
          if (song && !map.has(song.id)) map.set(song.id, song)
        }
      }
    }
    songs.value = [...map.values()]
  } catch (e) {
    console.error('推荐加载失败:', e)
  } finally {
    loading.value = false
  }
}

function normalizeFM(item) {
  const hash = item.hash != null ? String(item.hash) : ''
  if (!hash) return null
  const raw = Number(item.time_length) || 0
  const duration = raw > 1000 ? Math.round(raw / 1000) : raw
  const cover = (item.trans_param && item.trans_param.union_cover)
    || (item.relate_goods && item.relate_goods[0] && item.relate_goods[0].info && item.relate_goods[0].info.image)
    || item.sizable_cover || ''
  const singerName = (item.singerinfo && item.singerinfo[0] && item.singerinfo[0].name) || item.author_name || ''
  const singerId = (item.singerinfo && item.singerinfo[0] && item.singerinfo[0].id) || item.author_id || ''
  return {
    id: hash,
    mixsongid: item.mixsongid != null ? String(item.mixsongid) : '',
    name: item.ori_audio_name || item.songname || (item.filename || '').split(' - ').slice(1).join(' - ') || '',
    singer: singerName,
    singer_id: String(singerId),
    album: (item.relate_goods && item.relate_goods[0] && item.relate_goods[0].albumname) || item.album_name || '',
    album_id: item.album_id != null ? String(item.album_id) : '',
    duration,
    img: imgUrl(cover, 240),
    hash,
    hash_320: item.hash_320,
    hash_flac: item.hash_flac,
    pay_type: item.pay_type,
    privilege: item.privilege,
  }
}
</script>

<style scoped>
.fm-page { max-width: 1060px; }

.fm-header { margin-bottom: var(--spacing-xl); }
.fm-title-row {
  display: flex; align-items: center; gap: 16px;
}
.fm-title-row .page-title { margin-bottom: 0; }

.page-sub { color: var(--text-3); font-size: var(--font-size-sm); margin-top: 4px; }

.refresh-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 16px;
  border: 1px solid var(--border);
  border-radius: 20px;
  background: var(--glass-2);
  color: var(--text-2);
  font-size: 13px;
  cursor: pointer;
  transition: all .2s;
  white-space: nowrap;
}
.refresh-btn:hover { background: var(--glass-3); color: var(--text-1); border-color: var(--text-3); }
.refresh-btn:active { transform: scale(.95); }
.refresh-btn:disabled { opacity: .5; cursor: default; }
.refresh-btn svg { width: 16px; height: 16px; stroke: currentColor; fill: none; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }
.refresh-btn svg.spinning { animation: spin .8s linear infinite; }

@keyframes spin { to { transform: rotate(360deg); } }
</style>
