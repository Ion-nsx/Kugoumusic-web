<template>
  <div class="page">
    <h1 class="page-title">我喜欢</h1>

    <div class="section-head" v-if="player.likedSongs.length">
      <div class="section-title">{{ player.likedSongs.length }} 首</div>
    </div>

    <div v-if="player.likeLoading" class="skeleton" style="height: 200px; border-radius: 12px;"></div>

    <div v-else-if="player.likedSongs.length" class="song-table">
      <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span></span><span>时长</span></div>
      <div
        class="song-row"
        v-for="(song, idx) in player.likedSongs"
        :key="song.id"
        :class="{ playing: isCurrent(song) }"
        @click="player.playList(player.likedSongs, idx)"
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
        <button class="like liked" @click.stop="unlike(song)">
          <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
        </button>
        <div class="dur">{{ fmtDuration(song.duration) }}</div>
      </div>
    </div>

    <div v-else class="empty">{{ auth.isLoggedIn ? '还没有喜欢的歌曲，点击歌曲旁的红心收藏' : '登录后，喜欢的歌曲会自动同步到云端' }}</div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
const router = useRouter()
import { useAuthStore } from '../stores/auth'
import { usePlayerStore } from '../stores/player'
import { imgUrl } from '../utils/api'

const auth = useAuthStore()
const player = usePlayerStore()

function isCurrent(song) {
  return player.currentSong && player.currentSong.id === song.id
}

function unlike(song) {
  player.toggleLike(song)
}

function pad(n) { return String(n).padStart(2, '0') }
function fmtDuration(sec) {
  if (!sec) return '--:--'
  const m = Math.floor(sec / 60)
  const s = Math.floor(sec % 60)
  return `${m}:${pad(s)}`
}

function goAlbum(id) { if (id) router.push('/album/' + id) }
function goArtist(id) { if (id) router.push('/artist/' + id) }
</script>
