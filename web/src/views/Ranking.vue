<template>
  <div class="ranking-page">
    <!-- 榜单选择 -->
    <div class="rank-selector-card" v-if="!loading">
      <div class="rank-selector-shell" :class="{ collapsed: showToggle && !expanded }">
        <div class="rank-selector-wrap" :style="{ height: expanded ? 'auto' : (collapsedHeight || 44) + 'px' }">
          <div ref="chipWrapRef" class="rank-selector">
            <button
              v-for="rank in sortedRanks"
              :key="rank.rankid"
              class="rank-chip"
              :class="{ active: selectedIds.includes(rank.rankid) }"
              @click="toggleRank(rank)"
            >{{ rank.rankname }}</button>
          </div>
          <div v-if="showToggle && !expanded" class="selector-fade"></div>
        </div>
      </div>
      <button v-if="showToggle" class="rank-toggle" :class="{ expanded }" @click="expanded = !expanded">
        <span>{{ expanded ? '收起' : '展开更多' }}</span>
        <svg viewBox="0 0 24 24" width="14" height="14"><path d="M6 9l6 6 6-6" fill="none" stroke="currentColor" stroke-width="2"/></svg>
      </button>
    </div>

    <div v-if="loading" class="loading">
      <div class="skeleton" style="width: 100%; height: 400px; border-radius: 12px;"></div>
    </div>

    <!-- 空态 -->
    <div v-else-if="displayedRanks.length === 0" class="empty" style="margin-top: 80px; color: var(--text-3);">
      点击上方标签选择想看的榜单（最多 6 个）
    </div>

    <!-- 榜单面板网格 -->
    <div v-else class="rank-grid">
      <div class="rank-card" v-for="rank in displayedRanks" :key="rank.rankid">
        <div class="rank-card-head">
          <div class="rank-cover">
            <img v-if="rank.imgurl" :src="imgUrl(rank.imgurl, 320)" :alt="rank.rankname" />
          </div>
          <div class="rank-card-info">
            <h2 class="rank-card-title" :style="{ color: rank.album_cover_color }">{{ rank.rankname }}</h2>
            <span class="rank-card-desc">{{ formatIntro(rank.intro) }}</span>
          </div>
          <button class="rank-play" @click.stop="playRankAll(rank)" title="播放全部">
            <svg viewBox="0 0 24 24" width="20" height="20"><polygon points="6,3 20,12 6,21" fill="var(--primary)"/></svg>
          </button>
        </div>

        <div class="rank-song-list" @scroll="onScroll($event, rank.rankid)">
          <div class="rank-song-item" v-for="(song, sIdx) in rank.songs" :key="song.id"
            @click="playSong(rank, song, sIdx)">
            <span class="rsi-idx" :class="{ top: sIdx < 3 }">{{ sIdx + 1 }}</span>
            <div class="rsi-cover">
              <img v-if="song.img" :src="song.img" loading="lazy" />
              <div class="rsi-play-hover">
                <svg viewBox="0 0 24 24" width="14" height="14"><polygon points="6,3 18,12 6,21" fill="#fff"/></svg>
              </div>
            </div>
            <div class="rsi-body">
              <div class="rsi-name">{{ song.name }}</div>
              <div class="rsi-artist">{{ song.singer }}</div>
            </div>
            <div class="rsi-dur">{{ formatDuration(song.duration) }}</div>
          </div>

          <div v-if="pagination[rank.rankid]?.loading" class="rank-loading">加载中...</div>
          <div v-else-if="!pagination[rank.rankid]?.hasMore && rank.songs.length > 0" class="rank-loading">已加载全部</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { usePlayerStore } from '../stores/player'
import { getRankList, getRankAudio, imgUrl } from '../utils/api'

const player = usePlayerStore()
const MAX_RANKS = 6
const PAGE_SIZE = 30

const allRanks = ref([])
const selectedIds = ref([])
const displayedRanks = ref([])
const pagination = ref({})
const loading = ref(true)
const chipWrapRef = ref(null)
const showToggle = ref(false)
const expanded = ref(false)
const collapsedHeight = ref(0)
let resizeObserver = null

const sortedRanks = computed(() => {
  const sel = new Map(selectedIds.value.map((id, i) => [id, i]))
  return [...allRanks.value].sort((a, b) => {
    const ai = sel.get(a.rankid), bi = sel.get(b.rankid)
    if (ai !== undefined && bi !== undefined) return ai - bi
    if (ai !== undefined) return -1
    if (bi !== undefined) return 1
    return 0
  })
})

function initPagination(rankId) {
  if (!pagination.value[rankId]) {
    pagination.value[rankId] = { page: 1, loading: false, hasMore: true }
  }
}

async function loadSongs(rankId, page = 1, append = false) {
  const p = pagination.value[rankId]
  if (!p || p.loading) return
  p.loading = true
  try {
    const res = await getRankAudio(rankId, page, PAGE_SIZE)
    if (res.data.status === 1 && res.data.data) {
      const list = (res.data.data.songlist || []).map(normalizeSong).filter(Boolean)
      const rank = displayedRanks.value.find(r => r.rankid === rankId)
      if (rank) {
        rank.songs = append ? [...(rank.songs || []), ...list] : list
      }
      p.hasMore = list.length >= PAGE_SIZE
      p.page = page
    }
  } catch (e) {
    console.error('加载榜单歌曲失败:', e)
  } finally {
    p.loading = false
  }
}

function onScroll(e, rankId) {
  const el = e.target
  const p = pagination.value[rankId]
  if (!p || p.loading || !p.hasMore) return
  if (el.scrollTop + el.clientHeight >= el.scrollHeight - 60) {
    loadSongs(rankId, p.page + 1, true)
  }
}

function normalizeSong(item) {
  const hash = item?.deprecated?.hash || item?.hash
  if (!hash) return null
  const cover = (item?.trans_param?.union_cover || item?.album_info?.sizable_cover || '').replace('{size}', '240')
  return {
    id: String(hash),
    name: item.songname || item.name || '',
    singer: item.author_name || item.authors?.[0]?.author_name || '',
    album: item.album_info?.album_name || item.album_name || '',
    album_id: item.album_id != null ? String(item.album_id) : '',
    duration: Math.round((item.deprecated?.duration || item.audio_info?.duration_128 || 0) / 1000),
    img: cover
  }
}

async function toggleRank(rank) {
  const idx = selectedIds.value.indexOf(rank.rankid)
  if (idx === -1) {
    if (selectedIds.value.length >= MAX_RANKS) return
    selectedIds.value.push(rank.rankid)
    initPagination(rank.rankid)
    displayedRanks.value.push({ ...rank, songs: [] })
    await loadSongs(rank.rankid, 1, false)
  } else {
    selectedIds.value.splice(idx, 1)
    displayedRanks.value = displayedRanks.value.filter(r => r.rankid !== rank.rankid)
    delete pagination.value[rank.rankid]
  }
  saveSelected()
  calcCollapsed()
}

function saveSelected() {
  localStorage.setItem('vibe_selected_ranks', JSON.stringify(selectedIds.value))
}

function calcCollapsed() {
  nextTick(() => {
    const el = chipWrapRef.value
    if (!el) return
    const chips = Array.from(el.querySelectorAll('.rank-chip'))
    if (!chips.length) return
    const rowTops = [...new Set(chips.map(c => c.offsetTop))].sort((a, b) => a - b)
    if (rowTops.length > 2) {
      const chipH = chips[0].offsetHeight
      collapsedHeight.value = rowTops[1] + chipH + 16
      showToggle.value = true
    } else {
      showToggle.value = false
      expanded.value = false
    }
  })
}

async function loadRankList() {
  try {
    const res = await getRankList()
    if (res.data.status === 1 && res.data.data?.info) {
      allRanks.value = res.data.data.info
      const saved = localStorage.getItem('vibe_selected_ranks')
      let ids = []
      if (saved) {
        try { ids = JSON.parse(saved).filter(id => allRanks.value.find(r => r.rankid === id)) } catch (_) {}
      }
      // 无保存则随机选 4 个
      if (!ids.length) {
        const shuffled = [...allRanks.value].sort(() => Math.random() - 0.5)
        ids = shuffled.slice(0, 4).map(r => r.rankid)
      }
      for (const id of ids) {
        const rank = allRanks.value.find(r => r.rankid === id)
        if (rank && selectedIds.value.length < MAX_RANKS) {
          selectedIds.value.push(id)
          initPagination(id)
          displayedRanks.value.push({ ...rank, songs: [] })
          loadSongs(id, 1, false)
        }
      }
      saveSelected()
      await nextTick()
      calcCollapsed()
    }
  } catch (e) {
    console.error('获取排行榜失败:', e)
  } finally {
    loading.value = false
  }
}

function formatIntro(intro) {
  if (!intro) return ''
  const sort = intro.match(/排序[方方式]*[：:]\s*(.+?)(?:\n|$)/)
  const freq = intro.match(/更新频率[：:]\s*(.+?)(?:\n|$)/)
  return sort && freq ? `${sort[1]} · ${freq[1]}` : intro
}

function playRankAll(rank) {
  player.playList(rank.songs || [], 0)
}

function playSong(rank, song, _) {
  const all = displayedRanks.value.reduce((arr, r) => arr.concat(r.songs || []), [])
  const idx = all.findIndex(s => s.id === song.id)
  player.playList(all, idx >= 0 ? idx : 0)
}

function pad(n) { return String(n).padStart(2, '0') }
function formatDuration(seconds) {
  if (!seconds) return '--:--'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${pad(s)}`
}

onMounted(() => {
  loadRankList()
  resizeObserver = new ResizeObserver(calcCollapsed)
  if (chipWrapRef.value) resizeObserver.observe(chipWrapRef.value)
  window.addEventListener('resize', calcCollapsed)
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  window.removeEventListener('resize', calcCollapsed)
})
</script>

<style scoped>
.ranking-page {
  max-width: 1100px;
}

/* ---- 榜单选择器卡片 ---- */
.rank-selector-card {
  background: var(--surface);
  border-radius: var(--radius-lg);
  padding: var(--spacing-md);
  margin-bottom: var(--spacing-2xl);
  box-shadow: 0 4px 20px rgba(0,0,0,.06);
}

.rank-selector-shell.collapsed .rank-selector-wrap {
  position: relative;
}
.rank-selector-wrap {
  overflow: hidden;
  transition: height .35s cubic-bezier(.4,0,.2,1);
}
.rank-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding-bottom: 8px;
}

.selector-fade {
  position: absolute;
  inset: auto 0 0;
  height: 36px;
  background: linear-gradient(180deg, transparent, var(--surface));
  pointer-events: none;
}

.rank-chip {
  height: 34px;
  line-height: 34px;
  padding: 0 14px;
  border-radius: 8px;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-2);
  background: var(--bg);
  border: none;
  cursor: pointer;
  transition: all .25s ease;
}
.rank-chip:hover {
  background: var(--border);
  color: var(--text-1);
  transform: translateY(-1px);
}
.rank-chip.active {
  background: var(--primary);
  color: #fff;
}

.rank-toggle {
  width: 100%;
  margin-top: 8px;
  height: 36px;
  padding: 0 14px;
  border: none;
  border-radius: 8px;
  background: var(--bg);
  color: var(--text-3);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  cursor: pointer;
  transition: all .25s ease;
}
.rank-toggle:hover { color: var(--primary); }
.rank-toggle svg { transition: transform .3s; }
.rank-toggle.expanded svg { transform: rotate(180deg); }

/* ---- 榜单面板网格 ---- */
.rank-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-2xl);
}
@media (max-width: 900px) { .rank-grid { grid-template-columns: 1fr; } }

.rank-card {
  background: var(--surface);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0,0,0,.06);
  height: 600px;
  display: flex;
  flex-direction: column;
  transition: transform .25s ease, box-shadow .25s ease;
}
.rank-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 8px 30px rgba(0,0,0,.1);
}

.rank-card-head {
  display: flex;
  align-items: center;
  padding: var(--spacing-md);
  gap: var(--spacing-md);
  background: linear-gradient(135deg, color-mix(in srgb, var(--primary) 8%, transparent), var(--surface));
  position: relative;
  flex-shrink: 0;
}

.rank-cover {
  width: 88px;
  height: 88px;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0,0,0,.12);
  flex-shrink: 0;
}
.rank-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform .3s ease;
}
.rank-cover:hover img { transform: scale(1.06); }

.rank-card-info { flex: 1; min-width: 0; }

.rank-card-title {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  margin: 0 0 4px;
  line-height: 1.3;
}

.rank-card-desc {
  font-size: var(--font-size-xs);
  color: var(--text-3);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.rank-play {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 42px;
  height: 42px;
  border-radius: 50%;
  background: var(--surface);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0,0,0,.08);
  transition: all .25s ease;
  flex-shrink: 0;
}
.rank-play:hover {
  background: var(--primary);
  transform: scale(1.08);
}
.rank-play:hover svg polygon { fill: #fff; }

/* ---- 歌曲列表（卡片内滚动） ---- */
.rank-song-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 10px;
}

.rank-song-list::-webkit-scrollbar { width: 5px; }
.rank-song-list::-webkit-scrollbar-thumb {
  background: var(--border);
  border-radius: 3px;
}

.rank-song-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 6px;
  border-radius: 8px;
  cursor: pointer;
  transition: background .15s ease;
}
.rank-song-item:hover { background: var(--bg); }

.rsi-idx {
  width: 24px;
  text-align: center;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--text-3);
  flex-shrink: 0;
}
.rsi-idx.top { color: var(--primary); }

.rsi-cover {
  width: 42px;
  height: 42px;
  border-radius: 6px;
  overflow: hidden;
  flex-shrink: 0;
  position: relative;
}
.rsi-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.rsi-play-hover {
  position: absolute;
  inset: 0;
  background: rgba(0,0,0,.35);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity .15s ease;
}
.rank-song-item:hover .rsi-play-hover { opacity: 1; }

.rsi-body {
  flex: 1;
  min-width: 0;
}
.rsi-name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--text-1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rsi-artist {
  font-size: var(--font-size-xs);
  color: var(--text-3);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-top: 2px;
}

.rsi-dur {
  font-size: var(--font-size-xs);
  color: var(--text-3);
  flex-shrink: 0;
}

.rank-loading {
  text-align: center;
  padding: 16px;
  font-size: var(--font-size-xs);
  color: var(--text-3);
}
</style>
