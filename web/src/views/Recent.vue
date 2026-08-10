<template>
  <div class="page">
    <h1 class="page-title">最近播放</h1>

    <div class="section-head" v-if="player.history.length">
      <div class="section-title">{{ player.history.length }} 首</div>
      <div class="section-more" @click="clearHistory">清空</div>
    </div>

    <div v-if="player.history.length" class="song-table">
      <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span></span><span>时长</span></div>
      <div
        class="song-row"
        v-for="(song, idx) in player.history"
        :key="song.id"
        :class="{ playing: isCurrent(song) }"
        @click="player.playList(player.history, idx)"
      >
        <div class="idx">
          <span class="num">{{ pad(idx + 1) }}</span>
          <svg class="playing" viewBox="0 0 24 24"><rect x="7" y="4" width="3" height="16" rx="1.5" fill="#2CA6F8"/><rect x="12" y="8" width="3" height="12" rx="1.5" fill="#2CA6F8" opacity=".7"/><rect x="17" y="5" width="3" height="15" rx="1.5" fill="#2CA6F8" opacity=".4"/></svg>
        </div>
        <div class="s-cover">
          <img v-if="song.img" :src="imgUrl(song.img)" :alt="song.name" loading="lazy" />
        </div>
        <div class="s-name">{{ song.name }}</div>
        <div class="s-album link" @click.stop="goAlbum(song.album_id)">{{ song.album || '单曲' }}</div>
        <div class="s-artist link" @click.stop="goArtist(song.singer_id)">{{ song.singer }}</div>
        <button class="like" :class="{ liked: player.isLiked(song) }" title="喜欢" @click.stop="player.toggleLike(song)">
          <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
        </button>
        <div class="dur">{{ fmtLocalTime(song.played_at) }}</div>
      </div>
    </div>

    <div v-else class="empty">暂无播放记录</div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
const router = useRouter()
import { usePlayerStore } from '../stores/player'
import { imgUrl } from '../utils/api'

const player = usePlayerStore()

function isCurrent(song) {
  return player.currentSong && player.currentSong.id === song.id
}

function clearHistory() {
  if (confirm('确定清空播放历史吗？')) {
    player.history.splice(0)
    localStorage.removeItem('vibe_history')
  }
}

function pad(n) { return String(n).padStart(2, '0') }
function fmtLocalTime(ts) {
  if (!ts) return ''
  const d = new Date(ts)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) {
    return `${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
  return `${d.getMonth() + 1}/${d.getDate()}`
}

function goAlbum(id) { if (id) router.push('/album/' + id) }
function goArtist(id) { if (id) router.push('/artist/' + id) }
</script>
