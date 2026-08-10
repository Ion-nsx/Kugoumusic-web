<template>
  <div class="page search-page">
    <div class="search-header">
      <div class="search-box">
        <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/></svg>
        <input
          v-model="keyword"
          class="search-input"
          placeholder="搜索歌曲、歌手、歌单..."
          @keyup.enter="doSearch"
        />
        <button v-if="keyword" class="clear-btn" @click="clearSearch">✕</button>
        <button class="search-go" @click="doSearch">搜索</button>
      </div>

      <!-- 类型 Tab -->
      <div class="search-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.type"
          class="search-tab"
          :class="{ active: searchType === tab.type }"
          @click="switchType(tab.type)"
        >{{ tab.name }}</button>
      </div>
    </div>

    <!-- 热搜榜 -->
    <div v-if="!keyword && searchType === 'complex'" class="hot-section">
      <h2 class="section-title">热搜榜</h2>
      <div class="hot-card">
        <div
          class="hot-item"
          v-for="(word, idx) in hotWords"
          :key="idx"
          @click="keyword = word.word; doSearch()"
        >
          <span class="hot-rank" :class="{ 'top': idx < 3 }">{{ idx + 1 }}</span>
          <span class="hot-word">{{ word.word }}</span>
          <span class="hot-heat" v-if="word.hot > 0">{{ formatHeat(word.hot) }}</span>
        </div>
      </div>

      <!-- 搜索推荐词 -->
      <h2 v-if="defaultWords.length" class="section-title" style="margin-top: 24px">推荐搜索</h2>
      <div class="default-tags" v-if="defaultWords.length">
        <span
          v-for="(w, idx) in defaultWords"
          :key="idx"
          class="default-tag"
          @click="keyword = w; doSearch()"
        >{{ w }}</span>
      </div>
    </div>

    <!-- 搜索结果 -->
    <div v-else class="search-results">
      <div v-if="loading" class="loading">
        <div class="skeleton" style="width: 100%; height: 200px; border-radius: 12px;"></div>
      </div>

      <div v-else-if="isEmpty" class="empty">未找到相关结果</div>

      <!-- 综合搜索：分段展示 -->
      <template v-else-if="searchType === 'complex'">
        <section v-for="seg in complexSegments" :key="seg.type" v-show="seg.lists.length" class="seg-section">
          <h3 class="seg-title">{{ segName(seg.type) }} <span class="seg-count">{{ seg.total }}</span></h3>
          <div class="seg-body">
            <template v-if="seg.type === 'song'">
              <div class="song-table" v-if="seg.lists.length">
                <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span></span><span>时长</span></div>
                <div class="song-row" v-for="(song, i) in seg.lists" :key="'s' + i" @click="playSong(seg.lists, i)">
                  <div class="idx"><span class="num">{{ pad(i + 1) }}</span></div>
                  <div class="s-cover">
                    <img v-if="song.img" :src="imgUrl(song.img)" :alt="song.name" loading="lazy" />
                  </div>
                  <div class="s-name">{{ song.name }}</div>
                  <div class="s-album link" @click.stop="goAlbum(song.album_id)">{{ song.album }}</div>
                  <div class="s-artist link" @click.stop="goArtist(song.singer_id)">{{ song.singer }}</div>
                  <button class="like" :class="{ liked: player.isLiked(song) }" @click.stop="player.toggleLike(song)">
                    <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
                  </button>
                  <div class="dur">{{ formatDuration(song.duration) }}</div>
                </div>
              </div>
            </template>
            <div v-else-if="seg.type === 'album'" class="card-grid">
              <div class="playlist-card" v-for="a in seg.lists" :key="a.albumid" @click="goAlbum(a.albumid)">
                <div class="cover">
                  <img v-if="a.img" :src="a.img.replace('{size}', '240')" :alt="a.albumname" loading="lazy" style="width:100%;height:100%;object-fit:cover" />
                  <div v-else class="cover-tile">♪</div>
                </div>
                <div class="card-name">{{ a.albumname }}</div>
                <div class="card-sub">{{ a.singer }}</div>
              </div>
            </div>
            <div v-else-if="seg.type === 'author'" class="author-list">
              <div class="author-item" v-for="a in seg.lists" :key="a.AuthorId" @click="goArtist(a.AuthorId)">
                <img v-if="a.Avatar" class="author-avatar" :src="a.Avatar.replace('{size}', '120')" loading="lazy" />
                <div v-else class="author-avatar placeholder"></div>
                <div class="author-meta">
                  <div class="author-name">{{ a.AuthorName }}</div>
                  <div class="author-sub">{{ a.AudioCount || 0 }} 首作品</div>
                </div>
                <span class="author-arrow">›</span>
              </div>
            </div>
            <div v-else-if="seg.type === 'collect' || seg.type === 'special'" class="card-grid">
              <div class="playlist-card" v-for="pl in seg.lists" :key="pl.specialid || pl.gid" @click="goPlaylist(pl.gid || pl.specialid)">
                <div class="cover">
                  <img v-if="pl.img" :src="imgUrl(pl.img)" :alt="pl.specialname" loading="lazy" style="width:100%;height:100%;object-fit:cover" />
                  <div v-else class="cover-tile">♪</div>
                </div>
                <div class="card-name">{{ pl.specialname || pl.name }}</div>
                <div class="card-sub">{{ pl.nickname }}</div>
              </div>
            </div>
          </div>
        </section>
        <div v-if="!hasAnySegment" class="empty">未找到相关结果</div>
      </template>

      <!-- 单曲 -->
      <template v-else-if="searchType === 'song'">
        <div class="song-table" v-if="songs.length">
          <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span></span><span>时长</span></div>
          <div class="song-row" v-for="(song, idx) in songs" :key="song.id" @click="playSong(songs, idx, true)">
                    <div class="idx"><span class="num">{{ pad((page - 1) * PAGE_SIZE + idx + 1) }}</span></div>
            <div class="s-cover"><img v-if="song.img" :src="imgUrl(song.img)" :alt="song.name" loading="lazy" /></div>
            <div class="s-name">{{ song.name }}</div>
            <div class="s-album link" @click.stop="goAlbum(song.album_id)">{{ song.album }}</div>
            <div class="s-artist link" @click.stop="goArtist(song.singer_id)">{{ song.singer }}</div>
            <button class="like" :class="{ liked: player.isLiked(song) }" @click.stop="player.toggleLike(song)">
              <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
            </button>
            <div class="dur">{{ formatDuration(song.duration) }}</div>
          </div>
        </div>
      </template>

      <!-- 歌单 -->
      <template v-else-if="searchType === 'special'">
        <h3 class="seg-title" v-if="!keyword">热门歌单</h3>
        <div class="card-grid">
          <div class="playlist-card" v-for="pl in typedLists" :key="pl.id" @click="goPlaylist(pl.id)">
            <div class="cover">
              <img v-if="pl.cover" :src="imgUrl(pl.cover)" :alt="pl.name" loading="lazy" style="width:100%;height:100%;object-fit:cover" />
              <div v-else class="cover-tile">♪</div>
            </div>
            <div class="card-name">{{ pl.name }}</div>
            <div class="card-sub">{{ pl.author || '' }}{{ pl.play_count ? ' · ' + fmtCnt(pl.play_count) : '' }}</div>
          </div>
        </div>
      </template>

      <!-- 专辑 -->
      <template v-else-if="searchType === 'album'">
        <div class="card-grid">
          <div class="playlist-card" v-for="a in typedLists" :key="a.albumid" @click="goAlbum(a.albumid)">
            <div class="cover">
              <img v-if="a.img" :src="a.img.replace('{size}', '240')" :alt="a.albumname" loading="lazy" style="width:100%;height:100%;object-fit:cover" />
              <div v-else class="cover-tile">♪</div>
            </div>
            <div class="card-name">{{ a.albumname }}</div>
            <div class="card-sub">{{ a.singer }} · {{ a.publish_time }}</div>
          </div>
        </div>
      </template>

      <!-- 歌手 -->
      <template v-else-if="searchType === 'author'">
        <div class="author-list">
          <div class="author-item" v-for="a in typedLists" :key="a.AuthorId" @click="goArtist(a.AuthorId)">
            <img v-if="a.Avatar" class="author-avatar" :src="a.Avatar.replace('{size}', '120')" loading="lazy" />
            <div v-else class="author-avatar placeholder"></div>
            <div class="author-meta">
              <div class="author-name">{{ a.AuthorName }}</div>
              <div class="author-sub">{{ a.AudioCount || 0 }} 首作品 · {{ formatHeat(a.FansNum) }} 粉丝</div>
            </div>
            <span class="author-arrow">›</span>
          </div>
        </div>
      </template>

      <!-- 歌词 -->
      <template v-else-if="searchType === 'lyric'">
        <div class="song-table" v-if="lyricList.length">
          <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span></span><span>时长</span></div>
          <div class="song-row" v-for="(item, idx) in lyricList" :key="item.id || idx" @click="playLyricSong(idx)">
            <div class="idx"><span class="num">{{ pad(idx + 1) }}</span></div>
            <div class="s-cover">
              <img v-if="item.img" :src="imgUrl(item.img)" :alt="item.song" loading="lazy" />
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/></svg>
            </div>
            <div class="s-name">{{ item.song }}</div>
            <div class="s-album link" @click.stop="goAlbum(item.album_id)">{{ item.album }}</div>
            <div class="s-artist link" @click.stop="goArtist(item.singer_id)">{{ item.singer }}</div>
            <div class="dur">{{ formatDuration(item.durationSec) }}</div>
          </div>
        </div>
      </template>
    </div>

    <!-- 分页 -->
    <div class="pagination" v-if="totalPages > 1 && keyword && searchType !== 'complex' && searchType !== 'lyric'">
      <button class="pg-btn" :disabled="page === 1" @click="goPage(1)">首页</button>
      <button class="pg-btn" :disabled="page === 1" @click="goPage(page - 1)">‹</button>
      <button
        class="pg-num"
        v-for="p in pages"
        :key="p"
        :class="{ active: p === page, dots: p === '...' }"
        :disabled="p === '...'"
        @click="p !== '...' && goPage(p)"
      >{{ p }}</button>
      <button class="pg-btn" :disabled="page >= totalPages" @click="goPage(page + 1)">›</button>
      <button class="pg-btn" :disabled="page >= totalPages" @click="goPage(totalPages)">末页</button>
      <span class="pg-jump">跳至 <input class="pg-input" v-model="jumpPage" @keyup.enter="goJump" /> 页 <button class="pg-go" @click="goJump">GO</button></span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { usePlayerStore } from '../stores/player'
import { searchSongs, searchHot, searchComplex, searchByType, searchLyric, imgUrl, getTopPlaylist, getSearchDefault } from '../utils/api'

const router = useRouter()
const route = useRoute()
const player = usePlayerStore()

const keyword = ref('')
const searchType = ref('complex')
const songs = ref([])
const hotWords = ref([])
const defaultWords = ref([])
const loading = ref(false)
const complexData = ref(null)
const typedLists = ref([])
const lyricList = ref([])
const page = ref(1)
const total = ref(0)
const switchingPage = ref(false)
const jumpPage = ref('')

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))

const PAGE_SIZE = 30

const pages = computed(() => {
  const t = totalPages.value
  const p = page.value
  const result = []
  if (t <= 7) { for (let i = 1; i <= t; i++) result.push(i); return result }
  result.push(1)
  if (p > 3) result.push('...')
  for (let i = Math.max(2, p - 1); i <= Math.min(t - 1, p + 1); i++) result.push(i)
  if (p < t - 2) result.push('...')
  result.push(t)
  return result
})

const tabs = [
  { type: 'complex', name: '综合' },
  { type: 'song', name: '单曲' },
  { type: 'special', name: '歌单' },
  { type: 'album', name: '专辑' },
  { type: 'author', name: '歌手' },
  { type: 'lyric', name: '歌词' }
]

const complexSegments = computed(() => {
  if (!complexData.value || !complexData.value.lists) return []
  // 过滤 MV 段（项目明确不做 MV，前端也无对应渲染）
  return complexData.value.lists.filter(s => s.type !== 'mv')
})

const hasAnySegment = computed(() => complexSegments.value.some(s => (s.lists || []).length > 0))

const isEmpty = computed(() => {
  if (searchType.value === 'complex') return !hasAnySegment.value
  if (searchType.value === 'lyric') return lyricList.value.length === 0
  return typedLists.value.length === 0 && songs.value.length === 0
})

// 支持 /search?type=xxx&keyword=xxx 直达
watch(
  () => [route.query.type, route.query.keyword],
  ([type, kw]) => {
    if (type && tabs.some(t => t.type === type)) {
      searchType.value = type
    }
    if (kw && kw !== keyword.value) {
      keyword.value = kw
      doSearch()
    } else if (!kw && type === 'special') {
      loadPopularPlaylists()
    }
  },
  { immediate: true }
)

onMounted(() => {
  searchHot().then(res => {
    if (res.data.status === 1 && res.data.data) {
      if (Array.isArray(res.data.data)) {
        hotWords.value = res.data.data
      } else if (res.data.data.info) {
        hotWords.value = res.data.data.info
      }
    }
  }).catch(() => {})

  // 搜索推荐词
  getSearchDefault().then(res => {
    if (res.data.status === 1 && res.data.data && Array.isArray(res.data.data.ads)) {
      defaultWords.value = res.data.data.ads
        .filter(a => a.main_title)
        .map(a => a.main_title)
        .slice(0, 8)
    }
  }).catch(() => {})
})

function switchType(type) {
  if (searchType.value === type) return
  searchType.value = type
  page.value = 1
  total.value = 0
  if (!keyword.value && type === 'special') { loadPopularPlaylists(); return }
  doSearch()
}

function clearSearch() {
  keyword.value = ''
  songs.value = []
  typedLists.value = []
  lyricList.value = []
  complexData.value = null
  page.value = 1
  total.value = 0
  router.replace({ path: '/search' })
}

async function doSearch() {
  if (!keyword.value) {
    if (searchType.value === 'special') loadPopularPlaylists()
    return
  }
  loading.value = true
  try {
    if (searchType.value === 'complex') {
      const res = await searchComplex(keyword.value, page.value, 15)
      if (res.data.status === 1 && res.data.data) {
        complexData.value = res.data.data
        // 综合搜索所有分段共用一个页码，取各分段 total 最大值
        let maxT = 0
        if (res.data.data.lists) {
          res.data.data.lists.forEach(s => { if (s.total > maxT) maxT = s.total })
        }
        total.value = maxT
      }
      songs.value = []
      typedLists.value = []
    } else if (searchType.value === 'song') {
      const res = await searchSongs(keyword.value, page.value, PAGE_SIZE)
      if (res.data.status === 1 && res.data.data) {
        songs.value = res.data.data.songs || res.data.data.lists || []
        total.value = res.data.data.total || 0
      }
      typedLists.value = []
    } else if (searchType.value === 'lyric') {
      const res = await searchLyric(keyword.value)
      if (res.data.status === 1 && res.data.data) {
        const raw = res.data.data.songs || res.data.data.info || []
        lyricList.value = raw.map((s, i) => ({
          id: s.hash || s.id || '',
          song: s.name || s.songname || '',
          singer: s.singer || s.singername || '',
          album: s.album || s.albumname || '',
          album_id: s.album_id || '',
          singer_id: s.singer_id || '',
          durationSec: Math.round((Number(s.duration) || Number(s.time) || 0)),
          img: s.img || s.album_pic || ''
        }))
        total.value = res.data.data.total || 0
      }
      songs.value = []
      typedLists.value = []
    } else {
      const res = await searchByType(searchType.value, keyword.value, page.value, PAGE_SIZE)
      if (res.data.status === 1 && res.data.data) {
        typedLists.value = (res.data.data.lists || []).map(normalizeCard)
        total.value = res.data.data.total || 0
      }
      songs.value = []
    }
  } catch (e) {
    console.error('搜索失败:', e)
  } finally {
    loading.value = false
    switchingPage.value = false
  }
}

async function goPage(pg) {
  if (switchingPage.value || pg === page.value || pg < 1 || pg > totalPages.value) return
  switchingPage.value = true
  page.value = pg
  jumpPage.value = ''
  await doSearch()
}

function goJump() {
  const n = parseInt(jumpPage.value)
  if (n >= 1 && n <= totalPages.value) goPage(n)
}

// 归一词条卡片字段（search结果和top_playlist字段名不同）
function normalizeCard(item) {
  return {
    ...item,
    id: item.id || item.gid || item.specialid || item.albumid || '',
    cover: item.cover || item.img || '',
    name: item.name || item.specialname || item.albumname || '',
    author: item.author || item.nickname || item.singer || '',
    play_count: item.play_count || 0
  }
}

async function loadPopularPlaylists() {
  loading.value = true
  try {
    const res = await getTopPlaylist(1, 30)
    if (res.data.status === 1 && res.data.data) {
      typedLists.value = (res.data.data || []).map(normalizeCard)
    }
  } catch (e) {
    console.error('加载热门歌单失败:', e)
  } finally {
    loading.value = false
  }
}

function segName(type) {
  const map = { song: '单曲', album: '专辑', author: '歌手', collect: '歌单', mv: 'MV', special: '歌单' }
  return map[type] || type
}

function playSong(list, index, alreadyNormalized = false) {
  const arr = Array.isArray(list) ? list : []
  const normalized = arr.map(s => alreadyNormalized ? s : normalizeSong(s)).filter(Boolean)
  player.playList(normalized, index)
}

function normalizeSong(song) {
  return {
    id: String(song.id || song.hash || ''),
    name: song.name || song.songname || '',
    singer: song.singer || song.singername || '',
    album: song.album || song.albumname || '',
    album_id: song.album_id != null ? String(song.album_id) : '',
    singer_id: song.singer_id != null ? String(song.singer_id) : '',
    duration: Number(song.duration) || Number(song.time) || 0,
    img: imgUrl(song.img || song.album_pic, 240)
  }
}

function goAlbum(id) { if (id) router.push(`/album/${id}`) }
function goArtist(id) { if (id) router.push(`/artist/${id}`) }
function goArtistSearch(name) {
  if (name) {
    keyword.value = name
    searchType.value = 'song'
    page.value = 1
    doSearch()
  }
}

function playLyricSong(idx) {
  player.playList(lyricList.value, idx)
}
function goPlaylist(id) { if (id) router.push(`/playlist/${id}`) }

function pad(n) { return String(n).padStart(2, '0') }
function formatDuration(seconds) {
  if (!seconds) return '--:--'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${pad(s)}`
}

function fmtCnt(n) {
  if (n >= 100000000) return (n / 100000000).toFixed(1) + '亿'
  if (n >= 10000) return (n / 10000).toFixed(1) + '万'
  return String(n)
}
function formatHeat(heat) {
  if (!heat) return ''
  if (heat >= 10000) return (heat / 10000).toFixed(1) + '万'
  return heat.toString()
}
</script>

<style scoped>
.search-page { max-width: 960px; }

.search-header { margin-bottom: var(--spacing-xl); }

.search-box {
  display: flex; align-items: center; gap: var(--spacing-sm);
  padding: 4px 6px 4px 16px;
  background: var(--surface);
  border: 1px solid var(--border-strong);
  border-radius: var(--r-full);
  transition: border-color .2s, box-shadow .2s;
}
.search-box:focus-within { border-color: var(--primary); box-shadow: 0 0 0 3px var(--primary-soft); }
.search-icon { width: 16px; height: 16px; color: var(--text-3); flex-shrink: 0; }
.search-input {
  flex: 1; border: none; background: transparent;
  font-size: var(--font-size-lg); outline: none;
  padding: 8px 0; color: var(--text-1);
}
.search-input::placeholder { color: var(--text-3); }
.clear-btn {
  width: 24px; height: 24px; display: flex; align-items: center; justify-content: center;
  border-radius: 50%; color: var(--text-3); font-size: 12px; flex-shrink: 0;
}
.clear-btn:hover { background: var(--surface-3); }
.search-go {
  flex-shrink: 0; height: 34px; padding: 0 18px;
  background: var(--primary); color: #fff;
  border-radius: var(--r-full); font-size: 13px; font-weight: 600;
  transition: background .15s;
}
.search-go:hover { background: var(--primary-deep); }

.search-tabs { display: flex; gap: var(--spacing-sm); margin-top: var(--spacing-md); }
.search-tab {
  flex-shrink: 0; padding: 6px 16px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-sm); font-weight: var(--font-weight-medium);
  color: var(--text-2); transition: all var(--transition-fast);
  background: var(--surface);
  border: 1px solid var(--border);
}
.search-tab:hover { color: var(--text-1); }
.search-tab.active { background: var(--primary); color: #FFF; border-color: var(--primary); }

.hot-section { max-width: 600px; }
.section-title { font-size: var(--font-size-xl); font-weight: var(--font-weight-semibold); margin-bottom: var(--spacing-lg); }
.hot-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-md);
  padding: 8px;
  box-shadow: var(--shadow-sm);
}
.hot-item {
  display: flex; align-items: center; gap: var(--spacing-md);
  padding: 10px 14px; border-radius: var(--r-sm);
  cursor: pointer; transition: background .15s;
}
.hot-item:hover { background: var(--surface-3); }
.hot-rank { width: 24px; font-size: var(--font-size-lg); font-weight: var(--font-weight-bold); color: var(--text-3); text-align: center; }
.hot-rank.top { color: var(--primary); }
.hot-word { flex: 1; font-size: var(--font-size-base); }
.hot-heat { font-size: var(--font-size-xs); color: var(--text-3); }

.seg-section { margin-bottom: var(--spacing-3xl); }
.seg-title {
  font-size: var(--font-size-lg); font-weight: var(--font-weight-semibold);
  margin-bottom: var(--spacing-md); display: flex; align-items: center; gap: var(--spacing-sm);
}
.seg-count {
  font-size: var(--font-size-xs); font-weight: var(--font-weight-semibold);
  color: var(--text-3); background: var(--surface-3);
  padding: 2px 10px; border-radius: var(--radius-full);
}

.author-list { display: flex; flex-direction: column; gap: var(--spacing-sm); }
.author-item {
  display: flex; align-items: center; gap: var(--spacing-lg);
  padding: var(--spacing-md); border-radius: var(--r-md);
  cursor: pointer; transition: background .15s;
  background: var(--surface);
  border: 1px solid var(--border);
}
.author-item:hover { background: var(--surface-3); }
.author-avatar {
  width: 52px; height: 52px; border-radius: 50%;
  object-fit: cover; background: var(--surface-3);
  flex-shrink: 0;
}
.author-avatar.placeholder { display: block; }
.author-meta { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.author-name { font-size: var(--font-size-base); font-weight: var(--font-weight-semibold); }
.author-sub { font-size: var(--font-size-xs); color: var(--text-3); }
.author-arrow { color: var(--text-3); font-size: 20px; }

.default-tags { display: flex; flex-wrap: wrap; gap: 10px; }
.default-tag {
  padding: 6px 16px; border-radius: 20px; font-size: 13px;
  background: var(--glass-2); color: var(--text-secondary);
  cursor: pointer; transition: all .2s; border: 1px solid transparent;
}
.default-tag:hover { background: var(--primary-soft); color: var(--primary); border-color: var(--primary); }

.pagination { display: flex; align-items: center; justify-content: center; gap: 6px; padding: 24px 0 8px; }
.pg-btn {
  height: 32px; padding: 0 12px; border: 1px solid var(--border); border-radius: 8px;
  background: var(--surface); color: var(--text-2); font-size: 13px; cursor: pointer; transition: all .15s;
}
.pg-btn:hover:not(:disabled) { border-color: var(--primary); color: var(--primary); }
.pg-btn:disabled { opacity: .4; cursor: default; }
.pg-num {
  min-width: 32px; height: 32px; border: 1px solid var(--border); border-radius: 8px;
  background: var(--surface); color: var(--text-2); font-size: 13px; cursor: pointer; transition: all .15s;
  display: flex; align-items: center; justify-content: center;
}
.pg-num:hover:not(:disabled) { border-color: var(--primary); color: var(--primary); }
.pg-num.active { background: var(--primary); color: #fff; border-color: var(--primary); }
.pg-num.dots { border: none; cursor: default; color: var(--text-3); }
.pg-jump { font-size: 13px; color: var(--text-2); margin-left: 12px; display: flex; align-items: center; gap: 4px; }
.pg-input { width: 44px; height: 28px; border: 1px solid var(--border); border-radius: 6px; text-align: center; font-size: 13px; background: var(--surface); }
.pg-go { height: 28px; padding: 0 10px; border: 1px solid var(--border); border-radius: 6px; background: var(--surface); font-size: 12px; cursor: pointer; color: var(--text-2); }
.pg-go:hover { border-color: var(--primary); color: var(--primary); }
</style>
