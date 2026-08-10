<template>
  <div class="lyric-view" @click="onClose">
    <div class="lyric-song" v-if="player.currentSong">
      <div class="lyric-cover">
        <img v-if="player.currentSong.img" :src="player.currentSong.img" :alt="player.currentSong.name" />
        <span v-else>♪</span>
      </div>
      <div class="lyric-title">{{ player.currentSong.name }}</div>
      <div class="lyric-artist">{{ player.currentSong.singer }}</div>
    </div>

    <div class="lyric-content" ref="lyricBox" v-if="player.lyrics.length > 0">
      <div
        class="lyric-line"
        v-for="(line, idx) in player.lyrics"
        :key="idx"
        :ref="el => setLineRef(el, idx)"
        :class="{ active: idx === player.currentLyricIndex }"
      >
        {{ line.text }}
      </div>
    </div>
    <div class="lyric-empty" v-else>
      <p>暂无歌词</p>
    </div>

    <div class="lyric-hint">点击任意位置返回</div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/player'

const router = useRouter()
const player = usePlayerStore()
const lyricBox = ref(null)
const lineRefs = ref([])

function setLineRef(el, idx) {
  if (el) lineRefs.value[idx] = el
}

function onClose(e) {
  if (e.target.closest('.lyric-content')) return // 允许选中歌词文字
  router.back()
}

watch(() => player.currentLyricIndex, async (idx) => {
  await nextTick()
  const el = lineRefs.value[idx]
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
})
</script>

<style scoped>
.lyric-view {
  position: fixed;
  inset: 0;
  background: var(--bg);
  z-index: 500;
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  padding: var(--spacing-2xl) var(--spacing-4xl) var(--spacing-4xl);
}

.lyric-song { text-align: center; flex-shrink: 0; padding-top: 24px; }
.lyric-cover {
  width: 96px; height: 96px;
  margin: 0 auto 14px;
  border-radius: 18px;
  overflow: hidden;
  background: var(--surface);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-lg);
  display: grid; place-items: center;
  font-size: 40px; color: var(--primary);
}
.lyric-cover img { width: 100%; height: 100%; object-fit: cover; }
.lyric-title { font-size: var(--font-size-xl); font-weight: var(--font-weight-bold); }
.lyric-artist { font-size: var(--font-size-sm); color: var(--text-3); margin-top: 4px; }

.lyric-content {
  flex: 1;
  width: 100%;
  max-width: 640px;
  text-align: center;
  overflow-y: auto;
  padding: 40px 0 80px;
  scrollbar-width: none;
}
.lyric-content::-webkit-scrollbar { display: none; }

.lyric-line {
  padding: var(--spacing-sm) 0;
  font-size: var(--font-size-lg);
  color: var(--text-3);
  transition: all var(--transition-slow);
  line-height: 1.8;
}
.lyric-line.active {
  color: var(--text-1);
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-semibold);
}

.lyric-empty {
  flex: 1;
  display: flex;
  align-items: center;
  color: var(--text-3);
  font-size: var(--font-size-lg);
}

.lyric-hint {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--text-3);
  padding-bottom: 16px;
}
</style>
