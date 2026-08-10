<template>
  <div class="page cloud-page">
    <div class="cloud-header">
      <div class="cloud-title-row">
        <h1 class="page-title">音乐云盘</h1>
        <button class="upload-btn" @click="triggerUpload" :disabled="uploading" :title="uploading ? '上传中...' : '上传音乐文件'">
          <svg viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/></svg>
          {{ uploading ? '上传中...' : '上传' }}
        </button>
        <button class="refresh-btn" @click="loadFiles" :disabled="loading || uploading">
          <svg viewBox="0 0 24 24" :class="{ spinning: loading }"><path d="M21 2v5h-5M3.6 6.6A9 9 0 0 1 19.4 6M3 22v-5h5M20.4 17.4A9 9 0 0 1 4.6 18"/></svg>
          刷新
        </button>
      </div>
      <input type="file" ref="fileInput" accept=".mp3,.flac,.wav,.ogg,.aac,.m4a,.wma,.ape" @change="handleFileChange" style="display:none" />
      <p class="page-sub" v-if="total > 0">共 {{ total }} 首 · 当前第 {{ page }} 页</p>
    </div>

    <div v-if="uploading" class="upload-progress">
      <div class="progress-bar">
        <div class="progress-fill" :style="{ width: uploadProgress === -1 ? '100%' : uploadProgress + '%' }"
             :class="{ indeterminate: uploadProgress === -1 }" />
      </div>
      <p v-if="uploadProgress >= 0 && uploadProgress < 100">上传中 {{ uploadProgress }}%</p>
      <p v-else>正在处理，请稍候…</p>
    </div>

    <div v-if="loading && !files.length" class="loading">
      <div class="skeleton" style="width: 100%; height: 200px; border-radius: 12px;"></div>
    </div>

    <div v-else-if="files.length" class="song-table">
      <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span></span><span>时长</span></div>
      <div
        class="song-row"
        v-for="(song, idx) in files"
        :key="song.id"
        :class="{ playing: isCurrent(song) }"
        @click="player.playList(files, idx)"
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
        <div class="s-album">{{ song.album }}</div>
        <div class="s-artist" @click.stop="goArtist(song.singer_id)">{{ song.singer }}</div>
        <div class="s-del" @click.stop="confirmDelete(song, idx)" title="删除">
          <svg viewBox="0 0 24 24"><path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/></svg>
        </div>
        <div class="dur">{{ fmtDuration(song.duration) }}</div>
      </div>
    </div>

    <div v-else-if="!loading" class="empty">暂无上传文件</div>

    <div v-if="total > pageSize" class="pagination">
      <button :disabled="page <= 1" @click="goPage(page - 1)">上一页</button>
      <span class="page-info">{{ page }} / {{ totalPages }}</span>
      <button :disabled="page >= totalPages" @click="goPage(page + 1)">下一页</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, inject } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/player'
import { getCloudList, deleteCloudFile, cloudUpload, imgUrl } from '../utils/api'

const router = useRouter()
const player = usePlayerStore()
const requireLogin = inject('requireLogin', null)

const files = ref([])
const loading = ref(true)
const uploading = ref(false)
const uploadProgress = ref(0) // -1=处理中, 0-100=上传进度
const total = ref(0)
const page = ref(1)
const pageSize = 30
const fileInput = ref(null)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

onMounted(loadFiles)

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

function goArtist(id) { if (id) router.push('/artist/' + id) }

async function loadFiles() {
  loading.value = true
  try {
    const res = await getCloudList(page.value, pageSize)
    if (res.data.status === 1 && res.data.data) {
      const list = res.data.data.list || res.data.data.info || []
      total.value = res.data.data.total || 0
      files.value = list.map(normalizeCloud).filter(Boolean)
    } else {
      files.value = []
    }
  } catch (e) {
    console.error('云盘文件加载失败:', e)
    if (e.response?.status === 401 && requireLogin) {
      requireLogin('查看音乐云盘')
    }
  } finally {
    loading.value = false
  }
}

function goPage(p) {
  page.value = p
  loadFiles()
}

function triggerUpload() {
  if (fileInput.value) fileInput.value.click()
}

async function handleFileChange(e) {
  const file = e.target.files?.[0]
  if (!file) return

  uploading.value = true
  uploadProgress.value = 0
  try {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('auto_match', '1')

    const res = await cloudUpload(formData, (pct) => {
      uploadProgress.value = pct
    })
    // 上传完成，进入处理阶段
    uploadProgress.value = -1
    if (res.data.status === 1) {
      alert('上传成功！')
      // 跳回第一页刷新列表
      page.value = 1
      await loadFiles()
    } else {
      alert(res.data.error || '上传失败')
    }
  } catch (e) {
    console.error('上传失败:', e)
    if (e.response?.status === 401 && requireLogin) {
      requireLogin('上传文件到云盘')
    } else {
      alert(e.response?.data?.error || '上传失败，请重试')
    }
  } finally {
    uploading.value = false
    // 重置 file input 以便重复选同一个文件
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function confirmDelete(song, idx) {
  if (!confirm(`确定删除「${song.name}」吗？`)) return
  try {
    const res = await deleteCloudFile(String(song.kv_id), String(song.cloudAlbumAudioID))
    if (res.data.status === 1) {
      files.value.splice(idx, 1)
      total.value = Math.max(0, total.value - 1)
    } else {
      alert(res.data.error || '删除失败')
    }
  } catch (e) {
    console.error('删除失败:', e)
    alert('删除失败')
  }
}

function normalizeCloud(item) {
  const hash = item.hash || ''
  const rawName = item.name || ''
  let singer = item.author_name || ''
  let displayName = rawName
  if (rawName.includes(' - ')) {
    const parts = rawName.split(' - ')
    if (parts.length >= 2) {
      singer = parts[0].trim() || singer
      displayName = parts.slice(1).join(' - ').replace(/\.\w+$/, '')
    }
  }
  const duration = Math.round((Number(item.timelen) || 0) / 1000)
  const cover = item.album_info?.sizable_cover || item.authors?.[0]?.sizable_avatar || ''
  return {
    id: hash,
    kv_id: item.kv_id,
    cloudAlbumAudioID: item.album_audio_id || 0,
    cloudAudioID: item.audio_id || 0,
    name: displayName || rawName,
    singer,
    singer_id: '',
    album: item.album_name || '',
    album_id: item.album_id || '',
    duration,
    img: imgUrl(cover, 240),
    hash,
    cloudFile: true,
  }
}
</script>

<style scoped>
.cloud-page { max-width: 1060px; }

.cloud-header { margin-bottom: var(--spacing-xl); }
.cloud-title-row {
  display: flex; align-items: center; gap: 16px;
}
.cloud-title-row .page-title { margin-bottom: 0; }

.page-sub { color: var(--text-3); font-size: var(--font-size-sm); margin-top: 4px; }

.upload-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 16px;
  border: 1px solid var(--primary, #2CA6F8);
  border-radius: 20px;
  background: var(--primary, #2CA6F8);
  color: #fff;
  font-size: 13px;
  cursor: pointer;
  transition: all .2s;
  white-space: nowrap;
}
.upload-btn:hover { opacity: .85; }
.upload-btn:active { transform: scale(.95); }
.upload-btn:disabled { opacity: .5; cursor: default; }
.upload-btn svg { width: 16px; height: 16px; stroke: currentColor; fill: none; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; }

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

.upload-progress {
  text-align: center; padding: 32px 0;
}
.upload-progress p {
  color: var(--text-3); font-size: 13px; margin-top: 12px;
}
.progress-bar {
  width: 280px; height: 4px; border-radius: 2px;
  background: var(--glass-2); margin: 0 auto; overflow: hidden;
}
.progress-fill {
  height: 100%;
  background: var(--primary, #2CA6F8);
  border-radius: 2px;
  transition: width .3s ease;
}
.progress-fill.indeterminate {
  width: 40% !important;
  animation: progressAnim 1.5s ease-in-out infinite;
}
@keyframes progressAnim {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(350%); }
}

.s-del {
  width: 40px; display: flex; align-items: center; justify-content: center;
  cursor: pointer; opacity: 0; transition: opacity .18s;
}
.song-row:hover .s-del { opacity: .5; }
.s-del:hover { opacity: 1 !important; }
.s-del svg { width: 16px; height: 16px; stroke: var(--text-3); fill: none; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
.s-del:hover svg { stroke: #e74c3c; }

.pagination {
  display: flex; align-items: center; justify-content: center; gap: 16px;
  margin-top: var(--spacing-xl); padding: 12px 0;
}
.pagination button {
  padding: 6px 20px; border: 1px solid var(--border); border-radius: var(--r-md);
  background: var(--glass-2); color: var(--text-2); font-size: 13px;
  cursor: pointer; transition: all .18s;
}
.pagination button:hover:not(:disabled) { background: var(--glass-3); color: var(--text-1); }
.pagination button:disabled { opacity: .35; cursor: default; }
.page-info { font-size: 13px; color: var(--text-3); }

@keyframes spin { to { transform: rotate(360deg); } }
</style>
