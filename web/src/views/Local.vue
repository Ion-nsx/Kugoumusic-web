<template>
  <div class="page">
    <h1 class="page-title">本地音乐</h1>

    <div class="local-tip">
      从电脑选择本地音频文件即可播放，文件仅在本机播放，不会上传。
    </div>

    <div class="local-actions">
      <button class="btn btn-primary" @click="pickFiles">＋ 添加音乐文件</button>
      <span v-if="player.localFiles.length" class="local-count">{{ player.localFiles.length }} 首</span>
      <input
        ref="fileInput"
        type="file"
        accept="audio/*"
        multiple
        style="display:none;"
        @change="onFilesPicked"
      />
    </div>

    <div v-if="player.localFiles.length" class="song-table" style="margin-top: 20px;">
      <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>歌手</span><span>大小</span><span></span><span></span></div>
      <div
        class="song-row"
        v-for="(song, idx) in player.localFiles"
        :key="song.id"
        :class="{ playing: isCurrent(song) }"
        @click="playLocal(idx)"
      >
        <div class="idx">
          <span class="num">{{ pad(idx + 1) }}</span>
          <svg class="playing" viewBox="0 0 24 24"><rect x="7" y="4" width="3" height="16" rx="1.5" fill="#2CA6F8"/><rect x="12" y="8" width="3" height="12" rx="1.5" fill="#2CA6F8" opacity=".7"/><rect x="17" y="5" width="3" height="15" rx="1.5" fill="#2CA6F8" opacity=".4"/></svg>
        </div>
        <div class="s-cover"><span style="font-size:16px;color:rgba(15,23,42,.3)">♪</span></div>
        <div class="s-name">{{ song.name }}</div>
        <div class="s-album">{{ song.singer }}</div>
        <div class="s-artist">{{ fmtSize(song.size) }}</div>
        <button class="like" title="移除" @click.stop="removeLocal(song)">
          <svg viewBox="0 0 24 24"><path d="M3 6h18M8 6V4h8v2M6 6l1 14h10l1-14M10 11v6M14 11v6"/></svg>
        </button>
        <div class="dur"></div>
      </div>
    </div>

    <div v-else class="empty" style="margin-top: 20px;">暂无本地音乐，点击上方按钮添加</div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { usePlayerStore } from '../stores/player'

const player = usePlayerStore()
const fileInput = ref(null)

function pickFiles() {
  fileInput.value?.click()
}

function onFilesPicked(e) {
  const files = Array.from(e.target.files || [])
  if (files.length) {
    player.addLocalFiles(files)
  }
  e.target.value = ''
}

function playLocal(index) {
  const song = player.localFiles[index]
  if (!song) return
  if (!song.localUrl) {
    alert('该文件已失效，请重新添加')
    return
  }
  player.playList(player.localFiles, index)
}

function removeLocal(song) {
  player.removeLocalFile(song)
}

function isCurrent(song) {
  return player.currentSong && player.currentSong.id === song.id
}

function pad(n) { return String(n).padStart(2, '0') }
function fmtSize(size) {
  if (!size) return ''
  return (size / 1024 / 1024).toFixed(1) + ' MB'
}
</script>

<style scoped>
.local-tip {
  padding: 12px 16px;
  margin-bottom: 14px;
  background: var(--primary-soft);
  color: var(--primary-deep);
  font-size: var(--font-size-sm);
  border-radius: var(--r-md);
}
.local-count {
  font-size: var(--font-size-sm); color: var(--text-3);
}
</style>
