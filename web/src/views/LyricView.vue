<template>
  <div class="lyric-view">
    <!-- 背景：封面模糊 + 渐变遮罩 -->
    <div class="lyric-bg">
      <div class="lyric-bg-img" :style="bgStyle" v-if="player.currentSong?.img"></div>
      <div class="lyric-bg-mask"></div>
    </div>

    <!-- 点击空白返回（下层捕鼠器） -->
    <div class="lyric-catcher" @click="goBack"></div>

    <!-- 顶部返回按钮 -->
    <button class="lyric-back" @click="goBack" aria-label="返回">
      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6"/></svg>
    </button>

    <!-- 字体大小控制 -->
    <div class="lyric-font-control" @click.stop>
      <button class="lyric-font-btn" @click="changeFontSize(-1)" title="减小歌词字体">A−</button>
      <button class="lyric-font-btn" @click="changeFontSize(1)" title="增大歌词字体">A+</button>
    </div>

    <!-- 歌词高亮颜色 -->
    <div class="lyric-color-picker" @click.stop>
      <button class="lyric-palette-btn" :style="{ '--hl': hlColor }" @click="showColorPanel = !showColorPanel" title="高亮颜色" aria-label="高亮颜色">
        <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3a9 9 0 0 0 0 18c1.5 0 2-1 2-2 0-.8-.7-1.4-1.5-1.9A2.6 2.6 0 0 1 13 13.8C13 12.6 14 12 15.2 12H17a4 4 0 0 0 4-4 9 9 0 0 0-9-5Z"/><circle cx="7.5" cy="11" r="1.2"/><circle cx="10" cy="7.5" r="1.2"/><circle cx="14.5" cy="7.5" r="1.2"/></svg>
      </button>
      <div class="color-panel" v-if="showColorPanel">
        <div class="color-presets">
          <button
            v-for="c in presetColors"
            :key="c"
            class="color-dot"
            :class="{ on: hlColor === c }"
            :style="{ background: c }"
            @click="hlColor = c"
          ></button>
        </div>
        <label class="color-custom">
          <span class="custom-swatch" :style="{ background: hlColor }"></span>
          <span class="custom-label">自定义</span>
          <input type="color" v-model="hlColor" />
        </label>
      </div>
    </div>

    <!-- 歌曲信息 -->
    <div class="lyric-song" v-if="player.currentSong">
<div class="lyric-cover">
  <img v-if="player.currentSong.img" :src="player.currentSong.img" :alt="player.currentSong.name" />
  <span v-else>♪</span>
</div>
      <div class="lyric-title">{{ player.currentSong.name }}</div>
      <div class="lyric-artist">{{ player.currentSong.singer }}</div>
    </div>

    <!-- 歌词 -->
    <div class="lyric-content" ref="lyricBox" v-if="player.lyrics.length > 0" @click.stop>
      <div
        class="lyric-line"
        v-for="(line, idx) in player.lyrics"
        :key="idx"
        :ref="el => setLineRef(el, idx)"
        :class="{ active: idx === player.currentLyricIndex }"
        :style="lineStyle(idx)"
        @click="seekToLine(line)"
      >
        <template v-if="idx === player.currentLyricIndex && line.characters && line.characters.length">
          <span class="lyric-chars">
            <span
              v-for="(ch, ci) in line.characters"
              :key="ci"
              class="lyric-char"
              :class="{ sung: isCharSung(ch) }"
              :style="charStyle(ch)"
            >{{ ch.char }}</span>
          </span>
        </template>
        <template v-else>{{ line.text }}</template>
      </div>
    </div>
    <div class="lyric-empty" v-else>
      <p>暂无歌词</p>
    </div>

    <!-- 底部控制条 -->
    <div class="lyric-controls" @click.stop>
      <div class="progress">
        <span class="time">{{ formatTime(player.currentTime) }}</span>
        <div class="bar" ref="seekBar" @click="seek" @mousedown="startDrag" @touchstart.prevent="startDrag">
          <div class="bar-fill" :style="{ width: progressPercent + '%' }"></div>
          <div class="bar-dot" :style="{ left: progressPercent + '%' }"></div>
        </div>
        <span class="time r">{{ formatTime(player.duration) }}</span>
      </div>
      <div class="controls-row">
        <button class="ctl small" @click="cyclePlayMode" :title="playModeTitle">
          <svg v-if="player.playMode === 'one'" viewBox="0 0 24 24"><path d="M4 12a8 8 0 0 1 13.7-5.7L21 9.5"/><path d="M21 4.5V9.5H16"/><path d="M20 12a8 8 0 0 1-13.7 5.7L3 14.5"/><path d="M3 19.5V14.5H8"/><text x="13.4" y="17" font-size="6" font-weight="700" fill="currentColor">1</text></svg>
          <svg v-else-if="player.playMode === 'shuffle'" viewBox="0 0 24 24"><path d="M16 3h5v5M4 20l16-16M21 16v5h-5M15 15l6 6M4 4l5 5"/></svg>
          <svg v-else viewBox="0 0 24 24"><path d="M4 12a8 8 0 0 1 13.7-5.7L21 9.5"/><path d="M21 4.5V9.5H16"/><path d="M20 12a8 8 0 0 1-13.7 5.7L3 14.5"/><path d="M3 19.5V14.5H8"/></svg>
        </button>
        <button class="ctl" @click="player.prev()" title="上一首">
          <svg viewBox="0 0 24 24"><path d="M6 5v14M18 6l-8 6 8 6Z"/></svg>
        </button>
        <button class="ctl main" @click="player.togglePlay()" :title="player.isPlaying ? '暂停' : '播放'">
          <svg v-if="player.isPlaying" viewBox="0 0 24 24"><path d="M7 5h4v14H7ZM13 5h4v14h-4Z"/></svg>
          <svg v-else viewBox="0 0 24 24"><path d="M8 5v14l11-7Z"/></svg>
        </button>
        <button class="ctl" @click="player.next()" title="下一首">
          <svg viewBox="0 0 24 24"><path d="M18 5v14M6 6l8 6-8 6Z"/></svg>
        </button>
        <button class="ctl small speed-btn" @click="cycleSpeed" :title="'播放速度 ' + player.playbackRate + 'x'">
          <span class="speed-text">{{ player.playbackRate }}x</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/player'

const router = useRouter()
const player = usePlayerStore()
const lyricBox = ref(null)
const lineRefs = ref([])
const seekBar = ref(null)

// 歌词高亮颜色（默认蓝色），持久化
const hlColor = ref(localStorage.getItem('vibe_lyric_hl_color') || '#2CA6F8')
watch(hlColor, (c) => localStorage.setItem('vibe_lyric_hl_color', c))
const showColorPanel = ref(false)
// 预设高亮颜色
const presetColors = ['#2CA6F8', '#6BC3FF', '#3AD07A', '#FFD166', '#FF7A45', '#FF5D8F', '#BF7AFF', '#FFFFFF']

// 歌词字体大小（相对基准的 px 增量），持久化
const fontSize = ref(parseInt(localStorage.getItem('vibe_lyric_font') || '0'))
watch(fontSize, (v) => localStorage.setItem('vibe_lyric_font', String(v)))
function changeFontSize(delta) {
  fontSize.value = Math.max(-4, Math.min(6, fontSize.value + delta))
}

// 播放模式循环：loop → one → shuffle
const playModeTitle = computed(() => {
  const map = { loop: '列表循环', one: '单曲循环', shuffle: '随机播放' }
  return map[player.playMode] || '列表循环'
})
function cyclePlayMode() {
  const order = ['loop', 'one', 'shuffle']
  const i = order.indexOf(player.playMode)
  player.playMode = order[(i + 1) % order.length]
}

// 倍速循环：1 → 1.25 → 1.5 → 2 → 1
const speedOptions = [1, 1.25, 1.5, 2]
function cycleSpeed() {
  const i = speedOptions.indexOf(player.playbackRate)
  player.setPlaybackRate(speedOptions[(i + 1) % speedOptions.length])
}

// 点击歌词行定位到该句时间点播放
function seekToLine(line) {
  if (!line || typeof line.time !== 'number') return
  player.seekTo(line.time)
  if (!player.isPlaying) player.play()
}

function setLineRef(el, idx) {
  if (el) lineRefs.value[idx] = el
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}

// 歌词行渐变小：离当前行越远越小越淡；fontSize 为整体字号偏移
function lineStyle(idx) {
  const dist = Math.abs(idx - player.currentLyricIndex)
  const scale = Math.max(0.82, 1 - dist * 0.045)
  const opacity = Math.max(0.3, 1 - dist * 0.16)
  const isActive = idx === player.currentLyricIndex
  const base = isActive ? 'var(--font-size-xl)' : 'var(--font-size-lg)'
  const off = fontSize.value
  const fs = off > 0 ? `calc(${base} + ${off}px)` : (off < 0 ? `calc(${base} - ${-off}px)` : base)
  return {
    transform: `scale(${scale})`,
    opacity,
    fontSize: fs
  }
}

// 滚动到当前歌词行（页面进入 + 索引变化时复用）
function scrollToCurrentLine() {
  const el = lineRefs.value[player.currentLyricIndex]
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
}

watch(() => player.currentLyricIndex, async () => {
  await nextTick()
  scrollToCurrentLine()
})

// ===== 逐字高亮：按当前播放时间计算该行高亮进度 =====
const playMs = ref(0)
let rafId = 0
function syncPlayTime() {
  const el = player.audioEl
  const t = el && isFinite(el.currentTime) ? el.currentTime * 1000 : player.currentTime * 1000
  playMs.value = t
  rafId = requestAnimationFrame(syncPlayTime)
}
onMounted(async () => {
  rafId = requestAnimationFrame(syncPlayTime)
  // 进入页面时立刻滚动到当前播放的歌词行
  await nextTick()
  setTimeout(scrollToCurrentLine, 80)
})
onBeforeUnmount(() => { cancelAnimationFrame(rafId) })

// 该字符是否已唱到（当前播放时间 >= 字符开始时间）
function isCharSung(ch) {
  return playMs.value >= ch.startTime
}

// 单字符样式：已唱的字用高亮色 + 柔光，未唱的保持白色
function charStyle(ch) {
  if (!isCharSung(ch)) return {}
  return {
    color: hlColor.value,
    textShadow: `0 0 18px ${hlColor.value}88, 0 1px 8px rgba(0,0,0,.4)`
  }
}

// 背景封面
const bgStyle = computed(() => ({
  backgroundImage: `url(${player.currentSong?.img || ''})`
}))

// 进度条
const progressPercent = computed(() => {
  if (player.duration === 0) return 0
  return (player.currentTime / player.duration) * 100
})

function formatTime(seconds) {
  if (!seconds || isNaN(seconds)) return '00:00'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${pad(m)}:${pad(s)}`
}
function pad(n) { return String(n).padStart(2, '0') }

function seekByEvent(e) {
  if (!seekBar.value || player.duration === 0) return null
  const rect = seekBar.value.getBoundingClientRect()
  const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width))
  return ratio * player.duration
}

function seek(e) {
  const t = seekByEvent(e)
  if (t !== null) player.seekTo(t)
}

// 拖动进度条
let dragging = false
let wasPlaying = false
let cleanup = null

function startDrag(e) {
  if (player.duration === 0) return
  dragging = true
  wasPlaying = player.isPlaying
  player.pause()

  const t = seekByEvent(e)
  if (t !== null) player.seekTo(t)

  const onMove = (ev) => {
    const tt = seekByEvent(ev)
    if (tt !== null) player.seekTo(tt)
  }
  const onUp = () => {
    dragging = false
    if (wasPlaying) player.play()
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    document.removeEventListener('touchmove', onMove)
    document.removeEventListener('touchend', onUp)
    cleanup = null
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
  document.addEventListener('touchmove', onMove)
  document.addEventListener('touchend', onUp)
  cleanup = onUp
}
</script>

<style scoped>
.lyric-view {
  position: fixed;
  inset: 0;
  z-index: 500;
  display: flex;
  flex-direction: column;
  align-items: center;
  overflow: hidden;
  cursor: pointer;
  color: #fff;
  user-select: none;
}

/* ===== 背景层 ===== */
.lyric-bg { position: fixed; inset: 0; overflow: hidden; }
.lyric-bg-img {
  position: absolute; inset: -80px;
  background-size: cover;
  background-position: center;
  filter: blur(56px) saturate(1.15) brightness(.9);
  transform: scale(1.1);
  animation: bg-breathe 12s ease-in-out infinite alternate;
}
.lyric-bg-mask {
  position: absolute; inset: 0;
  background: linear-gradient(180deg, rgba(8,12,18,.68) 0%, rgba(8,12,18,.45) 38%, rgba(8,12,18,.62) 78%, rgba(8,12,18,.82) 100%);
}
@keyframes bg-breathe {
  from { transform: scale(1.1); }
  to { transform: scale(1.18); }
}

/* ===== 点击捕鼠器（空背景） ===== */
.lyric-catcher { position: fixed; inset: 0; }

/* ===== 顶部返回 ===== */
.lyric-back {
  position: fixed; top: 22px; left: 22px;
  z-index: 4;
  width: 40px; height: 40px;
  border-radius: 50%;
  display: grid; place-items: center;
  background: rgba(255,255,255,.12);
  border: 1px solid rgba(255,255,255,.18);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  color: rgba(255,255,255,.92);
  transition: background .18s, transform .12s;
}
.lyric-back:hover { background: rgba(255,255,255,.22); }
.lyric-back:active { transform: scale(.92); }

/* ===== 歌曲信息 ===== */
.lyric-song {
  position: relative; z-index: 2;
  text-align: center; flex-shrink: 0;
  margin-top: 40px;
  text-shadow: 0 1px 8px rgba(0,0,0,.35);
}
.lyric-cover {
  width: 132px; height: 132px;
  margin: 0 auto 18px;
  border-radius: var(--r-lg);
  overflow: hidden;
  border: 1px solid rgba(255,255,255,.28);
  box-shadow: 0 12px 40px rgba(0,0,0,.45);
  display: grid; place-items: center;
  font-size: 52px;
  background: rgba(255,255,255,.08);
}
.lyric-cover img { width: 100%; height: 100%; object-fit: cover; }
.lyric-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  letter-spacing: -.01em;
}
.lyric-artist {
  font-size: var(--font-size-sm);
  color: rgba(255,255,255,.72);
  margin-top: 6px;
}

/* ===== 歌词 ===== */
.lyric-content {
  position: relative; z-index: 2;
  flex: 1;
  width: 100%;
  max-width: 620px;
  text-align: center;
  overflow-y: auto;
  padding: 34px 0 16px;
  scrollbar-width: none;
  cursor: default;
}
.lyric-content::-webkit-scrollbar { display: none; }

.lyric-line {
  padding: 13px 0;
  font-size: var(--font-size-lg);
  color: rgba(255,255,255,.55);
  transition: transform .5s cubic-bezier(.25,.1,.25,1), opacity .5s cubic-bezier(.25,.1,.25,1), color .5s;
  line-height: 1.6;
  text-shadow: 0 1px 6px rgba(0,0,0,.3);
  will-change: transform, opacity;
  cursor: pointer;
}
.lyric-line:hover { color: rgba(255,255,255,.85); }
.lyric-line.active {
  color: #fff;
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  text-shadow: 0 0 22px rgba(44,166,248,.55), 0 1px 8px rgba(0,0,0,.4);
}

/* 当前行逐字渲染：已唱的字单独上色，未唱的保持白色 */
.lyric-chars {
  white-space: nowrap;
}
.lyric-char {
  display: inline-block;
  color: #fff;
  transition: color .15s linear, text-shadow .15s linear;
}

/* 字体大小控制（右下角调色板左侧） */
.lyric-font-control {
  position: fixed; right: 78px; bottom: 26px;
  z-index: 5;
  display: flex; gap: 8px;
}
.lyric-font-btn {
  width: 40px; height: 40px;
  border-radius: 50%;
  display: grid; place-items: center;
  background: rgba(255,255,255,.12);
  border: 1px solid rgba(255,255,255,.18);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  color: rgba(255,255,255,.92);
  font-size: 15px;
  font-weight: 700;
  cursor: pointer;
  transition: background .18s, transform .12s;
}
.lyric-font-btn:hover { background: rgba(255,255,255,.22); }
.lyric-font-btn:active { transform: scale(.9); }

/* 歌词高亮颜色选择器 */
.lyric-color-picker {
  position: fixed; right: 22px; bottom: 26px;
  z-index: 5;
}
.lyric-palette-btn {
  width: 44px; height: 44px;
  border-radius: 50%;
  display: grid; place-items: center;
  background: rgba(255,255,255,.12);
  border: 1px solid rgba(255,255,255,.18);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  color: rgba(255,255,255,.92);
  cursor: pointer;
  box-shadow: 0 6px 20px rgba(0,0,0,.3);
  transition: background .18s, transform .12s;
  position: relative;
}
.lyric-palette-btn::after {
  content: '';
  position: absolute; right: 5px; bottom: 5px;
  width: 10px; height: 10px;
  border-radius: 50%;
  background: var(--hl);
  border: 2px solid rgba(255,255,255,.85);
  box-shadow: 0 1px 4px rgba(0,0,0,.35);
}
.lyric-palette-btn:hover { background: rgba(255,255,255,.22); }
.lyric-palette-btn:active { transform: scale(.9); }

/* 颜色面板 */
.color-panel {
  position: absolute;
  right: 0; bottom: 56px;
  padding: 14px;
  border-radius: 16px;
  background: rgba(24,28,34,.82);
  border: 1px solid rgba(255,255,255,.14);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  box-shadow: 0 12px 40px rgba(0,0,0,.45);
  display: flex;
  flex-direction: column;
  gap: 12px;
  transform-origin: bottom right;
  animation: color-pop .2s var(--ease);
}
@keyframes color-pop {
  from { opacity: 0; transform: scale(.92) translateY(8px); }
  to { opacity: 1; transform: none; }
}
.color-presets {
  display: flex;
  gap: 9px;
  align-items: center;
}
.color-dot {
  width: 26px; height: 26px;
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid transparent;
  outline: none;
  transition: transform .12s, box-shadow .12s, border-color .12s;
}
.color-dot:hover { transform: scale(1.15); }
.color-dot.on {
  border-color: rgba(255,255,255,.92);
  box-shadow: 0 0 0 2px rgba(255,255,255,.25);
}
.color-custom {
  display: flex;
  align-items: center;
  gap: 9px;
  cursor: pointer;
  padding: 7px 10px;
  border-radius: 10px;
  background: rgba(255,255,255,.07);
  transition: background .15s;
}
.color-custom:hover { background: rgba(255,255,255,.13); }
.custom-swatch {
  width: 20px; height: 20px;
  border-radius: 6px;
  border: 1px solid rgba(255,255,255,.3);
  flex-shrink: 0;
}
.custom-label { font-size: 12.5px; color: rgba(255,255,255,.8); }
.color-custom input {
  width: 0; height: 0;
  opacity: 0;
  position: absolute;
}

.lyric-empty {
  position: relative; z-index: 2;
  flex: 1;
  display: flex;
  align-items: center;
  color: rgba(255,255,255,.6);
  font-size: var(--font-size-lg);
}

/* ===== 底部控制条 ===== */
.lyric-controls {
  position: relative; z-index: 3;
  width: 100%;
  max-width: 560px;
  padding: 6px 24px 40px;
  flex-shrink: 0;
  cursor: default;
}
.progress { display: flex; align-items: center; gap: 14px; margin-bottom: 22px; }
.progress .time { font-size: 14px; color: rgba(255,255,255,.75); font-variant-numeric: tabular-nums; min-width: 46px; }
.progress .time.r { text-align: right; }
.bar {
  flex: 1; height: 12px; border-radius: 99px;
  background: rgba(255,255,255,.18);
  position: relative; cursor: pointer;
  transition: height .15s;
}
.bar:hover { height: 14px; }
.bar-fill {
  position: absolute; left: 0; top: 0; bottom: 0;
  border-radius: 99px;
  background: linear-gradient(90deg, #2CA6F8, #6BC3FF);
}
.bar-dot {
  position: absolute; top: 50%;
  width: 22px; height: 22px;
  border-radius: 50%;
  background: #fff;
  transform: translate(-50%, -50%);
  box-shadow: 0 2px 8px rgba(0,0,0,.45);
  opacity: 0;
  transition: opacity .15s;
}
.bar:hover .bar-dot { opacity: 1; }

.controls-row { display: flex; align-items: center; justify-content: center; gap: 34px; }
.ctl {
  width: 46px; height: 46px;
  border-radius: 50%;
  display: grid; place-items: center;
  background: rgba(255,255,255,.1);
  border: 1px solid rgba(255,255,255,.16);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  color: rgba(255,255,255,.92);
  transition: background .18s, transform .12s;
}
.ctl:hover { background: rgba(255,255,255,.2); }
.ctl:active { transform: scale(.9); }
.ctl svg { width: 20px; height: 20px; fill: currentColor; }
.ctl.small {
  width: 40px; height: 40px;
  background: transparent;
  border: none;
  color: rgba(255,255,255,.75);
}
.ctl.small:hover { background: rgba(255,255,255,.14); }
.ctl.small svg {
  width: 19px; height: 19px;
  stroke: currentColor; fill: none;
  stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round;
}
.speed-btn .speed-text {
  font-size: 13px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  letter-spacing: .5px;
}
.ctl.main {
  width: 62px; height: 62px;
  background: #fff;
  border: none;
  color: #0E1116;
  box-shadow: 0 8px 28px rgba(0,0,0,.35);
}
.ctl.main:hover { background: #f2f3f5; }
.ctl.main svg { width: 26px; height: 26px; }
</style>
