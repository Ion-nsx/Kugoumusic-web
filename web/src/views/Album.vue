<template>
  <div class="page detail-page">
    <BackBar title="专辑" />

    <template v-if="loading">
      <div class="loading">
        <div class="skeleton" style="width: 200px; height: 200px; border-radius: 16px; margin-bottom: 20px;"></div>
        <div class="skeleton" style="width: 60%; height: 24px; margin-bottom: 12px;"></div>
      </div>
    </template>

    <template v-else-if="album">
      <div class="detail-wrap">
      <div class="detail-header">
        <div class="detail-cover">
          <img v-if="album.cover" :src="imgUrl(album.cover)" :alt="album.name" />
          <div v-else class="img-placeholder">♪</div>
        </div>
        <div class="detail-info">
          <h1 class="detail-name">{{ album.name }}</h1>
          <p class="detail-sub">{{ album.singer }}</p>
          <p class="detail-desc" v-if="album.time || album.company">{{ album.time }} {{ album.company }}</p>
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
          <div class="idx"><span class="num">{{ pad(idx + 1) }}</span></div>
          <div class="s-cover">
            <img v-if="song.img" :src="imgUrl(song.img)" :alt="song.name" loading="lazy" />
          </div>
          <div class="s-name">{{ song.name }}</div>
          <div class="s-album">{{ song.album }}</div>
          <div class="s-artist">{{ song.singer }}</div>
          <button class="like" :class="{ liked: player.isLiked(song) }" @click.stop="player.toggleLike(song)">
            <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
          </button>
          <div class="dur">{{ formatDuration(song.duration) }}</div>
        </div>
      </div>
      </div>
    </template>

    <!-- 评论弹层 -->
    <CommentModal
      v-if="showComments"
      resource-type="album"
      :title="album?.name || '专辑评论'"
      :id="String(route.params.id)"
      @close="showComments = false"
    />

    <div v-if="!loading && !album" class="empty">专辑不存在</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { usePlayerStore } from '../stores/player'
import { getAlbumDetail, getAlbumSongs, imgUrl } from '../utils/api'
import BackBar from '../components/BackBar.vue'
import CommentModal from '../components/CommentModal.vue'

const route = useRoute()
const player = usePlayerStore()

const album = ref(null)
const songs = ref([])
const loading = ref(true)
const showComments = ref(false)

onMounted(async () => {
  const id = route.params.id
  try {
    // 专辑信息 + 歌曲列表（detail 接口不返回歌曲，歌曲走 album/songs）
    const [detailRes, songsRes] = await Promise.all([
      getAlbumDetail(id),
      getAlbumSongs(id, 1, 50)
    ])

    if (detailRes.data.status === 1 && detailRes.data.data) {
      album.value = detailRes.data.data
    }

    if (songsRes.data.status === 1 && Array.isArray(songsRes.data.data)) {
      songs.value = songsRes.data.data
      // 歌曲无封面时回退到专辑封面
      const defaultCover = album.value?.cover ? imgUrl(album.value.cover) : ''
      if (defaultCover) {
        songs.value = songs.value.map(s => ({ ...s, img: s.img || defaultCover }))
      }
      if (album.value && !album.value.cover && songs.value.length && songs.value[0].img) {
        album.value.cover = songs.value[0].img
      }
    }
  } catch (e) {
    console.error('获取专辑详情失败:', e)
  } finally {
    loading.value = false
  }
})

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
</script>

<style scoped>
.detail-page { max-width: 960px; }
</style>
