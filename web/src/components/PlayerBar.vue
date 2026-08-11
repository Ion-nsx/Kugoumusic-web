<template>
  <footer class="player" :class="{ playing: player.isPlaying }" v-if="player.currentSong">
    <div class="now-playing">
      <div class="now-cover">
        <img v-if="player.currentSong.img" :src="player.currentSong.img" :alt="player.currentSong.name" />
        <svg v-else viewBox="0 0 24 24"><path d="M8.5 17.5a3.5 3.5 0 1 1-1-2.47l8-2.2V7.4L6.8 10.4a4.5 4.5 0 1 0 1.2 2.9l8-2.2V4.4a1 1 0 0 1 1.3-.95l3.2 1A1 1 0 0 1 21 5.4v8.1a4.5 4.5 0 1 0-1.3-3.2v3.9l-11.2 3.3Z"/></svg>
        <div class="eq"><i></i><i></i><i></i></div>
      </div>
      <div class="now-info">
        <div class="now-title">{{ player.currentSong.name || '未在播放' }}</div>
        <div class="now-artist">{{ artistText }}</div>
      </div>
      <div class="now-actions">
        <button class="mini-btn" :class="{ liked: player.isLiked(player.currentSong) }" title="喜欢" @click="toggleLike">
          <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
        </button>
        <button class="mini-btn" title="评论" @click="openComments">
          <svg viewBox="0 0 24 24"><path d="M21 12a8 8 0 0 1-8 8H4l2.2-2.6A8 8 0 1 1 21 12Z"/></svg>
        </button>
      </div>
    </div>

    <div class="player-center">
      <div class="player-controls-row">
        <div class="player-btns">
          <button class="pbtn" title="上一首" @click="player.prev()">
            <svg viewBox="0 0 24 24"><path d="M6 5v14M18 6l-8 6 8 6Z"/></svg>
          </button>
          <button class="pbtn-main" title="播放/暂停" @click="player.togglePlay()">
            <svg v-if="player.isPlaying" viewBox="0 0 24 24"><path d="M7 5h4v14H7ZM13 5h4v14h-4Z"/></svg>
            <svg v-else viewBox="0 0 24 24"><path d="M8 5v14l11-7Z"/></svg>
          </button>
          <button class="pbtn" title="下一首" @click="player.next()">
            <svg viewBox="0 0 24 24"><path d="M18 5v14M6 6l8 6-8 6Z"/></svg>
          </button>
          <button class="pbtn" :title="modeTitle" @click="toggleMode">
            <svg v-if="player.playMode === 'one'" viewBox="0 0 24 24"><path d="M4 12a8 8 0 0 1 13.7-5.7L21 9.5"/><path d="M21 4.5V9.5H16"/><path d="M20 12a8 8 0 0 1-13.7 5.7L3 14.5"/><path d="M3 19.5V14.5H8"/><text x="13.4" y="17" font-size="6" font-weight="700" fill="currentColor">1</text></svg>
            <svg v-else-if="player.playMode === 'shuffle'" viewBox="0 0 24 24"><path d="M16 3h5v5M4 20l16-16M21 16v5h-5M15 15l6 6M4 4l5 5"/></svg>
            <svg v-else viewBox="0 0 24 24"><path d="M4 12a8 8 0 0 1 13.7-5.7L21 9.5"/><path d="M21 4.5V9.5H16"/><path d="M20 12a8 8 0 0 1-13.7 5.7L3 14.5"/><path d="M3 19.5V14.5H8"/></svg>
          </button>
        </div>
        <div class="mini-lyric" @click="toggleLyric" :class="{ active: player.lyrics.length > 0 }">
          {{ currentLyricText }}
        </div>
      </div>

      <div class="progress">
        <span class="time" id="curTime">{{ formatTime(player.currentTime) }}</span>
        <div class="bar" :class="{ dragging }" ref="seekBar" @click="seek" @mousedown="startDrag" @touchstart.prevent="startDrag">
          <div class="bar-track">
            <div class="bar-buffer" :style="{ width: bufferPercent + '%' }"></div>
            <div class="bar-fill" :style="{ width: progressPercent + '%' }">
              <i class="bar-glow"></i>
            </div>
            <div v-if="climax" class="climax-zone" :style="climaxStyle"></div>
          </div>
          <div class="bar-dot" :style="{ left: progressPercent + '%' }"><i></i></div>
          <div class="bar-hover-time" v-if="hoverTime !== null && dragging" :style="{ left: hoverPos + '%' }">{{ formatTime(hoverTime) }}</div>
        </div>
        <span class="time r">{{ formatTime(player.duration) }}</span>
      </div>
    </div>

    <div class="player-right">
      <div class="vol">
        <svg viewBox="0 0 24 24"><path d="M4 9v6h4l5 4V5L8 9H4Z"/><path d="M16.5 8.5a5 5 0 0 1 0 7"/><path d="M19 6a8.5 8.5 0 0 1 0 12"/></svg>
        <div class="bar" ref="volBar" @click="seekVolume" @mousedown="startVolDrag" @touchstart.prevent="startVolDrag">
          <div class="bar-fill" :style="{ width: (player.volume * 100) + '%' }"></div>
          <div class="bar-dot" :style="{ left: (player.volume * 100) + '%' }"></div>
        </div>
      </div>
      <button class="pbtn quality-btn" :title="'音质：' + qualityLabel + (realSource ? ' | 实际源：' + realSource : '')" @click="showQuality = !showQuality; showQueue = false">
        <span class="quality-dot" :class="{ ok: sourceReady }"></span>
        <span class="quality-text">{{ qualityLabel }}</span>
      </button>
      <button class="pbtn" title="播放队列" @click="toggleQueue">
        <svg viewBox="0 0 24 24"><path d="M3 6h13M3 12h10M3 18h15"/><path d="M19 6v12l5-6z"/></svg>
      </button>
    </div>

    <!-- 音质选择弹层 -->
    <div class="queue-pop quality-pop" v-if="showQuality" @click.stop>
      <div class="queue-head">
        <span>音质选择</span>
        <button class="queue-close" @click="showQuality = false">✕</button>
      </div>
      <div class="quality-list">
        <div
          class="quality-item"
          v-for="q in qualityOptions"
          :key="q.value"
          :class="{ on: (player.actualQuality || player.quality) === q.value }"
          @click="changeQuality(q.value)"
        >
          <span class="q-name">{{ q.label }}</span>
          <span v-if="q.tip" class="q-tip">{{ q.tip }}</span>
          <span v-if="(player.actualQuality || player.quality) === q.value" class="q-check">✓</span>
        </div>
        <div class="quality-msg" v-if="qualityMsg">{{ qualityMsg }}</div>
      </div>
    </div>

    <!-- 评论弹层 -->
    <CommentModal
      v-if="showComments && player.currentSong"
      resource-type="song"
      :title="player.currentSong.name || '评论'"
      :mixsongid="player.currentSong.mixsongid"
      :hash="player.currentSong.id"
      @close="showComments = false"
    />

    <!-- 轻提示 -->
    <div class="toast" v-if="toast">{{ toast }}</div>

    <!-- 播放队列弹层 -->
    <div class="queue-pop" v-if="showQueue">
      <div class="queue-head">
        <span>播放队列</span>
        <button class="queue-close" @click="showQueue = false">✕</button>
      </div>
      <div class="queue-list">
        <div
          class="queue-item"
          v-for="(s, i) in player.queue"
          :key="i"
          :class="{ on: i === player.queueIndex }"
          @click="playAt(i)"
        >
          <span class="q-idx">{{ i + 1 }}</span>
          <span class="q-name">{{ s.name }}</span>
          <span class="q-singer">{{ s.singer }}</span>
        </div>
        <div v-if="!player.queue.length" class="queue-empty">队列为空</div>
      </div>
    </div>

    <!-- 音频元素
        注意：不要绑定 @play/@pause 到 player.play()/pause()，
        切歌时 src 清空会触发 DOM pause，把 isPlaying 置 false 导致新地址就绪后不续播。
        播放状态完全由 store 控制（isPlaying watch 调 play/pause）。 -->
    <audio
      ref="audioEl"
      :src="player.audioUrl"
      @timeupdate="onTimeUpdate"
      @loadedmetadata="onLoadedMetadata"
      @ended="onEnded"
      @playing="onSourcePlaying"
      :volume="player.isMuted ? 0 : player.volume"
    ></audio>
  </footer>
</template>

<script setup>
import { ref, computed, watch, nextTick, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/player'
import CommentModal from './CommentModal.vue'
import { getSongClimax } from '../utils/api'

const router = useRouter()
const player = usePlayerStore()
const audioEl = ref(null)
const seekBar = ref(null)
const volBar = ref(null)
const showQueue = ref(false)
const showQuality = ref(false)
const qualityMsg = ref('')
const showComments = ref(false)
const toast = ref('')
const climax = ref(null) // { start: ms, end: ms }
// 音质源状态确认：切换音质后等待新源真正加载完成
const realSource = ref('')     // audio 实际加载的源 URL（与选择的音质对应）
const sourcePending = ref(false) // 切换音质后是否在等待新源
const sourceReady = ref(true)    // 新源是否已加载就绪（绿点）
let toastTimer = null

const qualityOptions = [
  { value: '128', label: '标准', tip: '128Kbps' },
  { value: '320', label: '高品', tip: '320Kbps' },
  { value: 'flac', label: '无损', tip: 'FLAC' },
  { value: 'high', label: 'Hi-Res', tip: '高解析' }
]

const qualityLabel = computed(() => {
  // 显示实际播放音质（自动降级后的）
  const useQ = player.actualQuality || player.quality
  const q = qualityOptions.find(o => o.value === useQ)
  return q ? q.label : '标准'
})

const artistText = computed(() => {
  const s = player.currentSong
  if (!s) return ''
  return [s.singer, s.album].filter(Boolean).join(' · ')
})

const currentLyricText = computed(() => {
  if (!player.lyrics.length) return '暂无歌词'
  return player.lyrics[player.currentLyricIndex]?.text || ''
})

const progressPercent = computed(() => {
  if (player.duration === 0) return 0
  return (player.currentTime / player.duration) * 100
})

const climaxStyle = computed(() => {
  if (!climax.value || !player.duration) return { display: 'none' }
  const left = (climax.value.start / player.duration) * 100
  const width = ((climax.value.end - climax.value.start) / player.duration) * 100
  return { left: left + '%', width: Math.min(width, 100 - left) + '%' }
})

const modeTitle = computed(() => ({ loop: '列表循环', one: '单曲循环', shuffle: '随机播放' })[player.playMode] || '列表循环')
const modeList = ['loop', 'one', 'shuffle']

function toggleMode() {
  const i = modeList.indexOf(player.playMode)
  player.playMode = modeList[(i + 1) % modeList.length]
}

function formatTime(seconds) {
  if (!seconds || isNaN(seconds)) return '00:00'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${pad(m)}:${pad(s)}`
}
function pad(n) { return String(n).padStart(2, '0') }

function onTimeUpdate() {
  if (audioEl.value) player.updateCurrentTime(audioEl.value.currentTime)
  updateBuffer()
}
function onLoadedMetadata() {
  if (audioEl.value) player.duration = audioEl.value.duration
  // 新源 metadata 就绪 = 真的加载到了目标音质源
  const src = audioEl.value?.currentSrc || player.audioUrl || ''
  if (src && src !== realSource.value) {
    realSource.value = src
  }
  if (sourcePending.value) {
    sourcePending.value = false
    sourceReady.value = true
    showToast(`已切换音质并加载新源：${qualityLabel.value}`)
  }
  // 换源后 audio 的 playbackRate 会重置为 1，这里重新应用用户设置的倍速
  if (player.playbackRate !== 1) audioEl.value.playbackRate = player.playbackRate
}
// 开始出声时也确认源已就绪（兼容某些浏览器 metadata 回调时序）
function onSourcePlaying() {
  if (sourcePending.value) {
    sourcePending.value = false
    sourceReady.value = true
    const src = audioEl.value?.currentSrc || player.audioUrl || ''
    if (src) realSource.value = src
    showToast(`已切换音质并开始播放：${qualityLabel.value}`)
  }
}
function onEnded() {
  if (player.playMode === 'one') {
    audioEl.value.currentTime = 0
    audioEl.value.play()
  } else {
    player.next()
  }
}

function seek(e) {
  const time = seekByEvent(e)
  if (time !== null && audioEl.value) audioEl.value.currentTime = time
}

function seekByEvent(e) {
  if (!seekBar.value || player.duration === 0) return null
  const rect = seekBar.value.getBoundingClientRect()
  const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width))
  return ratio * player.duration
}

// 缓冲进度：取 audio 已缓冲范围，用于进度条"已加载"显示
const bufferPercent = ref(0)
function updateBuffer() {
  const el = audioEl.value
  if (!el || !el.buffered || el.buffered.length === 0 || player.duration === 0) {
    bufferPercent.value = 0
    return
  }
  const end = el.buffered.end(el.buffered.length - 1)
  bufferPercent.value = Math.min(100, (end / player.duration) * 100)
}

// 悬停时间 tooltip：拖动时显示拖动位置对应的时间
const hoverTime = ref(null)
const hoverPos = ref(0)
function seekHover(e) {
  if (!seekBar.value || player.duration === 0) return
  const rect = seekBar.value.getBoundingClientRect()
  const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width))
  hoverPos.value = ratio * 100
  hoverTime.value = ratio * player.duration
}

// 进度条拖动
const dragging = ref(false)
let dragCleanup = null
let wasPlayingBeforeDrag = false

function startDrag(e) {
  if (player.duration === 0) return
  dragging.value = true
  // 拖动时暂停播放
  wasPlayingBeforeDrag = player.isPlaying
  if (audioEl.value) audioEl.value.pause()
  seekHover(e)

  const time = seekByEvent(e)
  if (time !== null && audioEl.value) audioEl.value.currentTime = time

  const onMove = (ev) => {
    seekHover(ev)
    const t = seekByEvent(ev)
    if (t !== null && audioEl.value) audioEl.value.currentTime = t
  }
  const onUp = () => {
    dragging.value = false
    hoverTime.value = null
    // 拖动结束后恢复播放
    if (wasPlayingBeforeDrag && audioEl.value) {
      audioEl.value.play().catch(() => {})
    }
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    document.removeEventListener('touchmove', onMove)
    document.removeEventListener('touchend', onUp)
    dragCleanup = null
  }

  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
  document.addEventListener('touchmove', onMove)
  document.addEventListener('touchend', onUp)
  dragCleanup = onUp
}

onBeforeUnmount(() => {
  if (dragCleanup) dragCleanup()
})

function seekVolume(e) {
  if (!volBar.value) return
  const rect = volBar.value.getBoundingClientRect()
  const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width))
  player.setVolume(ratio)
}

// 音量拖动
function startVolDrag(e) {
  if (!volBar.value) return
  const rect = volBar.value.getBoundingClientRect()
  const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width))
  player.setVolume(ratio)

  const onMove = (ev) => {
    const r = volBar.value.getBoundingClientRect()
    const rr = Math.min(1, Math.max(0, (ev.clientX - r.left) / r.width))
    player.setVolume(rr)
  }
  const onUp = () => {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    document.removeEventListener('touchmove', onMove)
    document.removeEventListener('touchend', onUp)
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
  document.addEventListener('touchmove', onMove)
  document.addEventListener('touchend', onUp)
}

function toggleLike() {
  if (player.currentSong) player.toggleLike(player.currentSong)
}

function toggleLyric() {
  router.push('/lyric')
}

function toggleQueue() {
  showQueue.value = !showQueue.value
  showQuality.value = false
}

async function changeQuality(q) {
  qualityMsg.value = ''
  const oldUrl = player.audioUrl
  const ok = await player.switchQuality(q)
  if (ok) {
    showQuality.value = false
    if (player.audioUrl && player.audioUrl !== oldUrl) {
      sourcePending.value = true
      sourceReady.value = false
      realSource.value = ''
    } else {
      sourcePending.value = false
      sourceReady.value = true
    }
    if (player.actualQuality !== q) {
      showToast(`${qualityLabel.value}（${q} 不可用，已自动降级）`)
    } else {
      showToast(`已切换至 ${qualityLabel.value}`)
    }
  } else {
    qualityMsg.value = '当前歌曲不支持该音质'
  }
}

function openComments() {
  showQueue.value = false
  showQuality.value = false
  if (!player.currentSong || !player.currentSong.mixsongid) {
    showToast('该歌曲暂不支持评论')
    return
  }
  showComments.value = true
}

function showToast(msg) {
  toast.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toast.value = '' }, 2000)
}

function playAt(i) {
  player.playAt(i)
  showQueue.value = false
}

watch(() => player.isPlaying, async (playing) => {
  if (!audioEl.value) return
  if (playing) {
    // 等 DOM 将最新 src 绑定到 audio 元素后再播放
    await nextTick()
    if (audioEl.value.src) audioEl.value.play().catch(() => {})
  } else {
    audioEl.value.pause()
  }
})

// 将 audio 元素注册到 player store，供 seekTo 等使用
watch(audioEl, (el) => { if (el) player.setAudioEl(el) }, { immediate: true })

// 倍速：换源后 audio 的 playbackRate 会重置为 1，需在 src 变化时重新应用
watch(() => player.audioUrl, async () => {
  await nextTick()
  if (audioEl.value && player.playbackRate !== 1) {
    audioEl.value.playbackRate = player.playbackRate
  }
})

watch(() => player.currentSong, async () => {
  if (audioEl.value && player.isPlaying) {
    await nextTick()
    if (audioEl.value.src) audioEl.value.play().catch(() => {})
  }
})

// 播放地址异步就绪后自动续播（切歌/首次加载）
// 注意：audioUrl 变化时 DOM 的 :src 尚未更新，需 nextTick 后 play 才能生效
watch(() => player.audioUrl, async () => {
  if (!audioEl.value || !player.audioUrl || !player.isPlaying) return
  await nextTick()
  if (audioEl.value.src) audioEl.value.play().catch(() => {})
})

// 播放失败（VIP/版权限制等）时提示用户，3 秒后自动跳下一首
let skipTimer = null
function clearSkipTimer() {
  if (skipTimer) { clearTimeout(skipTimer); skipTimer = null }
}
function onPlayError(msg) {
  if (!msg) { clearSkipTimer(); return }
  player.pause()
  showToast(msg)
  clearSkipTimer()
  // 队列有多首歌才自动跳，避免单曲死循环
  if (player.queue.length > 1) {
    skipTimer = setTimeout(() => {
      skipTimer = null
      player.next()
    }, 3000)
  }
}
watch(() => player.playError, onPlayError)

// 手动切歌/上一首/下一首时取消待跳转计时器
watch(() => player.currentSong, clearSkipTimer)

// 切歌时获取高潮片段区间
watch(() => player.currentSong?.id, (hash) => {
  climax.value = null
  if (!hash) return
  getSongClimax(hash).then(res => {
    const list = res.data?.data
    if (!Array.isArray(list) || !list.length) return
    const item = list[0]
    const start = Number(item.start_time)
    const end = Number(item.end_time)
    if (start >= 0 && end > start) {
      climax.value = { start: start / 1000, end: end / 1000 }
    }
  }).catch(() => {})
})

onBeforeUnmount(() => {
  clearSkipTimer()
  player.setAudioEl(null)
  if (audioEl.value) audioEl.value.pause()
})
</script>

<style scoped>
.player {
  position: fixed;
  left: var(--sidebar-width);
  right: 0; bottom: 0;
  height: var(--player-height);
  background: var(--color-player-bg);
  backdrop-filter: blur(28px) saturate(1.7);
  -webkit-backdrop-filter: blur(28px) saturate(1.7);
  border-top: 1px solid var(--border);
  display: grid;
  grid-template-columns: 1fr 2fr 1fr;
  align-items: center;
  padding: 0 28px; gap: 20px;
  z-index: 20;
}

.now-playing { display: flex; align-items: center; gap: 14px; min-width: 0; }
.now-cover {
  position: relative; width: 64px; height: 64px; border-radius: 14px; flex-shrink: 0;
  border: 1px solid var(--border);
  display: grid; place-items: center; overflow: hidden;
  background: var(--surface-3);
  box-shadow: var(--shadow-md);
}
.now-cover img { width: 100%; height: 100%; object-fit: cover; }
.now-cover svg { width: 22px; height: 22px; fill: rgba(15,23,42,.3); }
.eq {
  position: absolute; inset: 0; display: none;
  place-items: end center; padding-bottom: 8px; gap: 2.5px;
  background: linear-gradient(180deg, transparent 40%, rgba(44,166,248,.18));
}
.player.playing .eq { display: flex; }
.eq i {
  width: 3px; background: var(--primary); border-radius: 99px;
  animation: eq .9s ease-in-out infinite;
}
.eq i:nth-child(1) { height: 10px; animation-delay: 0s; }
.eq i:nth-child(2) { height: 16px; animation-delay: .25s; }
.eq i:nth-child(3) { height: 12px; animation-delay: .5s; }
@keyframes eq { 0%,100% { transform: scaleY(.4); } 50% { transform: scaleY(1); } }

.now-info { min-width: 0; }
.now-title { font-size: 15px; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.now-artist { font-size: 13px; color: var(--text-3); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.now-actions { display: flex; align-items: center; gap: 8px; margin-left: 12px; flex-shrink: 0; }
.mini-btn {
  width: 34px; height: 34px; border-radius: 50%; border: none; background: transparent;
  cursor: pointer; color: var(--text-2); display: grid; place-items: center; transition: all .15s;
}
.mini-btn:hover { background: var(--surface-3); color: var(--text-1); }
.mini-btn svg { width: 17px; height: 17px; stroke: currentColor; fill: none; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
.mini-btn.liked { color: var(--like); }
.mini-btn.liked svg { fill: currentColor; }
.mini-btn.liked:hover { color: var(--like); background: var(--primary-soft); }

.player-center { display: flex; flex-direction: column; align-items: center; gap: 8px; }
.player-controls-row {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  width: 100%;
  max-width: 540px;
  padding-left: 100px;
}
.player-btns { grid-column: 2; display: flex; align-items: center; gap: 20px; }
.pbtn {
  width: 38px; height: 38px; border-radius: 50%; border: none; background: transparent;
  cursor: pointer; color: var(--text-2); display: grid; place-items: center;
  transition: all .15s var(--ease);
}
.pbtn:hover { color: var(--text-1); background: var(--surface-3); transform: scale(1.06); }
.pbtn svg { width: 19px; height: 19px; stroke: currentColor; fill: none; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
.pbtn-main {
  width: 50px; height: 50px; border-radius: 50%; border: none; cursor: pointer;
  background: linear-gradient(135deg, #2CA6F8, #1E7FD4); color: #fff;
  display: grid; place-items: center;
  transition: all .18s var(--ease);
  box-shadow: 0 5px 16px rgba(44,166,248,.45);
}
.pbtn-main:hover { transform: scale(1.07); box-shadow: 0 7px 22px rgba(44,166,248,.55); }
.pbtn-main svg { width: 20px; height: 20px; fill: currentColor; }

.mini-lyric {
  grid-column: 3;
  justify-self: start;
  padding-left: 20px;
  font-size: 15px;
  color: var(--text-4);
  cursor: pointer;
  transition: color .2s;
  user-select: none;
  line-height: 1;
  width: 280px;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex-shrink: 0;
}
.mini-lyric.active {
  color: var(--text-3);
}
.mini-lyric:hover {
  color: var(--primary);
}
.mini-lyric.active {
  color: var(--text-3);
}
.mini-lyric:hover {
  color: var(--primary);
}

.progress { display: flex; align-items: center; gap: 12px; width: 100%; max-width: 680px; padding-right: 140px; }
.time { font-size: 12px; color: var(--text-3); font-variant-numeric: tabular-nums; min-width: 38px; }
.time.r { text-align: right; }
.bar {
  flex: 1;
  height: 18px;
  display: flex; align-items: center;
  cursor: pointer;
  position: relative;
  touch-action: none;
}
/* 玻璃轨道 */
.bar-track {
  position: relative;
  width: 100%; height: 8px;
  border-radius: 99px;
  background: var(--glass-track, rgba(120,128,140,.22));
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  box-shadow: inset 0 1px 2px rgba(0,0,0,.16), inset 0 0 0 0.5px rgba(255,255,255,.06);
  overflow: hidden;
  transition: height .15s var(--ease);
}
.bar:hover .bar-track { height: 10px; }
/* 缓冲进度（半透明磨砂） */
.bar-buffer {
  position: absolute; left: 0; top: 0; bottom: 0;
  border-radius: 99px;
  background: rgba(140,150,165,.22);
  width: 0%;
}
/* 已播进度：渐变 + 发光 */
.bar-fill {
  position: absolute; left: 0; top: 0; bottom: 0;
  border-radius: 99px;
  background: linear-gradient(90deg, var(--primary-deep), var(--primary) 60%, #5EC5FF);
  box-shadow: 0 0 10px color-mix(in srgb, var(--primary) 55%, transparent);
  width: 0%;
}
.bar-glow {
  position: absolute; right: 0; top: 0; bottom: 0;
  width: 26px;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,.5));
  filter: blur(3px);
}
/* 圆点：毛玻璃白球 */
.bar-dot {
  position: absolute; top: 50%;
  width: 14px; height: 14px;
  border-radius: 50%;
  background: rgba(255,255,255,.95);
  box-shadow: 0 2px 6px rgba(0,0,0,.28), 0 0 0 3.5px color-mix(in srgb, var(--primary) 42%, transparent);
  transform: translate(-50%,-50%) scale(0);
  transition: transform .16s var(--ease), box-shadow .16s var(--ease);
  pointer-events: none;
}
.bar-dot i {
  position: absolute; inset: 3px;
  border-radius: 50%;
  background: var(--primary);
}
.bar:hover .bar-dot, .bar.dragging .bar-dot { transform: translate(-50%,-50%) scale(1); }
.bar.dragging .bar-dot { box-shadow: 0 3px 10px rgba(0,0,0,.32), 0 0 0 4px color-mix(in srgb, var(--primary) 55%, transparent); }
.bar:hover .bar-fill { background: linear-gradient(90deg, var(--primary-deep), var(--primary) 55%, #74CEFF); }
/* 拖动悬停时间提示 */
.bar-hover-time {
  position: absolute; top: -30px;
  transform: translateX(-50%);
  padding: 3px 8px;
  border-radius: 8px;
  background: var(--color-player-bg);
  box-shadow: var(--shadow-md);
  border: 1px solid var(--border);
  font-size: 12px; font-weight: 600;
  color: var(--text-1);
  font-variant-numeric: tabular-nums;
  pointer-events: none;
  white-space: nowrap;
}

.climax-zone {
  position: absolute; top: 0; bottom: 0; border-radius: 99px;
  background: linear-gradient(90deg, #FF6B6B44, #FF9F4344);
  pointer-events: none;
  z-index: 1;
  transition: left .3s, width .3s;
}

.player-right { display: flex; align-items: center; justify-content: flex-end; gap: 12px; }
.vol { display: flex; align-items: center; gap: 10px; }
.vol svg { width: 18px; height: 18px; stroke: var(--text-2); fill: none; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
.vol .bar {
  width: 110px;
  height: 5px;
  border-radius: 99px;
  background: var(--glass-track, rgba(120,128,140,.22));
  box-shadow: inset 0 1px 2px rgba(0,0,0,.16);
  flex: none;
}
.vol .bar:hover { height: 5px; }
.vol .bar-fill { background: var(--text-2); }
.vol .bar-dot {
  width: 11px; height: 11px;
  box-shadow: 0 2px 6px rgba(0,0,0,.28);
}
.vol .bar-dot i { background: var(--text-2); }
.vol .bar:hover .bar-dot, .vol .bar.dragging .bar-dot { transform: translate(-50%,-50%) scale(1); }
.player-right .pbtn svg { width: 19px; height: 19px; }
.quality-btn { width: auto; padding: 0 6px; }
.quality-dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: #F59E0B; /* 切换中：橙色 */
  transition: background .2s;
}
.quality-dot.ok { background: #22C55E; } /* 新源已就绪：绿色 */
.quality-text {
  font-size: 11px; font-weight: 600;
  color: var(--text-2); white-space: nowrap;
}
.quality-btn:hover .quality-text { color: var(--primary-deep); }

/* 音质面板 */
.quality-pop { width: 220px; }
.quality-list { padding: 6px; }
.quality-item {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 12px; border-radius: var(--r-sm);
  cursor: pointer; transition: background .15s;
}
.quality-item:hover { background: var(--surface-3); }
.quality-item.on { background: var(--primary-soft); }
.quality-item.on .q-name { color: var(--primary-deep); font-weight: 700; }
.quality-item .q-tip { font-size: 11px; color: var(--text-3); }
.quality-item .q-check { margin-left: auto; color: var(--primary); font-weight: 700; }
.q-dot { font-size: 9px; margin-left: 4px; }
.q-ok { color: #22C55E; }
.q-no { color: #EF4444; }
.q-dot:hover { cursor: help; }
.quality-msg { padding: 8px 12px 4px; font-size: 12px; color: var(--like); }

/* 轻提示 */
.toast {
  position: absolute;
  bottom: calc(var(--player-height) + 16px);
  left: 50%;
  transform: translateX(-50%);
  padding: 9px 20px;
  background: var(--text-1);
  color: #fff;
  font-size: 13px;
  border-radius: var(--r-full);
  box-shadow: var(--shadow-md);
  z-index: 40;
  white-space: nowrap;
}

/* 队列弹层 */
.queue-pop {
  position: absolute;
  right: 28px; bottom: calc(var(--player-height) + 10px);
  width: 360px; max-height: 360px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-md);
  box-shadow: var(--shadow-lg);
  display: flex; flex-direction: column;
  z-index: 30;
  overflow: hidden;
}
.queue-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 16px; font-size: 13px; font-weight: 600;
  border-bottom: 1px solid var(--border);
}
.queue-close { color: var(--text-3); font-size: 13px; }
.queue-close:hover { color: var(--text-1); }
.queue-list { overflow-y: auto; padding: 6px; }
.queue-item {
  display: flex; align-items: center; gap: 10px;
  padding: 8px 10px; border-radius: var(--r-sm);
  cursor: pointer; transition: background .15s;
}
.queue-item:hover { background: var(--surface-3); }
.queue-item.on .q-idx { color: var(--primary); font-weight: 700; }
.queue-item.on .q-name { color: var(--primary-deep); }
.q-idx { width: 20px; text-align: center; font-size: 12px; color: var(--text-3); flex-shrink: 0; }
.q-name { flex: 1; min-width: 0; font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.q-singer { font-size: 11px; color: var(--text-3); max-width: 120px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.queue-empty { padding: 20px; text-align: center; color: var(--text-3); font-size: 13px; }

@media (max-width: 1100px) {
  .player { grid-template-columns: 1fr 1.4fr 0.8fr; padding: 0 18px; gap: 12px; }
}
@media (max-width: 860px) {
  .player { left: 0; grid-template-columns: 1fr 1.6fr 0.9fr; padding: 0 14px; }
}
</style>
