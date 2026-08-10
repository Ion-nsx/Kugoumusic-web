<template>
  <div class="home">
    <!-- Banner 轮播 -->
    <div class="banner fade-in" @mouseenter="stopAuto" @mouseleave="startAuto" v-if="banners.length">
      <div class="banner-inner" :style="{ transform: `translateX(-${bannerIdx * 100}%)` }">
        <div class="banner-slide" v-for="(b, i) in banners" :key="i" :style="{ '--banner-glow': b.color }" @click="goBannerTarget(b)">
          <div class="banner-art" :style="{ background: b.color }">
            <img v-if="b.img" :src="b.img" :alt="b.title" />
            <svg v-else viewBox="0 0 24 24"><path d="M8.5 17.5a3.5 3.5 0 1 1-1-2.47l8-2.2V7.4L6.8 10.4a4.5 4.5 0 1 0 1.2 2.9l8-2.2V4.4a1 1 0 0 1 1.3-.95l3.2 1A1 1 0 0 1 21 5.4v8.1a4.5 4.5 0 1 0-1.3-3.2v3.9l-11.2 3.3Z"/></svg>
          </div>
          <div class="banner-info">
            <div class="banner-tag">{{ b.tag }}</div>
            <h2>{{ b.title }}</h2>
            <p>{{ b.desc }}</p>
            <button class="banner-play" @click.stop="goBannerTarget(b)">
              <svg viewBox="0 0 24 24"><path d="M8 5v14l11-7Z"/></svg>
              立即播放
            </button>
          </div>
        </div>
      </div>
      <div class="banner-dots">
        <i
          v-for="(_, i) in banners"
          :key="i"
          :class="{ on: i === bannerIdx }"
          @click="bannerIdx = i"
        ></i>
      </div>
    </div>

    <!-- 快捷功能 -->
    <div class="quick-grid fade-in">
      <div class="quick-card" @click="onQuick('daily')">
        <div class="quick-ic" style="background:linear-gradient(135deg,#2CA6F8,#1E7FD4)"><svg viewBox="0 0 24 24"><path d="M12 3v4M12 17v4M3 12h4M17 12h4"/><circle cx="12" cy="12" r="5"/></svg></div>
        <div><div class="quick-name">每日推荐</div><div class="quick-sub">根据口味每日更新</div></div>
        <svg class="quick-arrow" viewBox="0 0 24 24"><path d="m9 6 6 6-6 6"/></svg>
      </div>
      <div class="quick-card" @click="onQuick('playlist')">
        <div class="quick-ic" style="background:linear-gradient(135deg,#FF8A3D,#F5642A)"><svg viewBox="0 0 24 24"><rect x="3" y="5" width="18" height="14" rx="2"/><path d="M3 9h18"/><path d="M8 5v14"/></svg></div>
        <div><div class="quick-name">歌单广场</div><div class="quick-sub">千万歌单等你发现</div></div>
        <svg class="quick-arrow" viewBox="0 0 24 24"><path d="m9 6 6 6-6 6"/></svg>
      </div>
      <div class="quick-card" @click="onQuick('rank')">
        <div class="quick-ic" style="background:linear-gradient(135deg,#FF4D6A,#E02F4A)"><svg viewBox="0 0 24 24"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2Z"/></svg></div>
        <div><div class="quick-name">排行榜</div><div class="quick-sub">热歌风向标</div></div>
        <svg class="quick-arrow" viewBox="0 0 24 24"><path d="m9 6 6 6-6 6"/></svg>
      </div>
    </div>

    <!-- 热门歌单 -->
    <section class="section fade-in">
      <div class="section-head">
        <div class="section-title">热门歌单</div>
        <div class="section-more" @click="onQuick('playlist')">查看全部 <svg viewBox="0 0 24 24"><path d="m9 6 6 6-6 6"/></svg></div>
      </div>
      <div v-if="loading" class="loading">
        <div class="skeleton" style="width: 100%; height: 240px; border-radius: 12px;"></div>
      </div>
      <div v-else-if="playlists.length" class="card-grid">
        <div class="playlist-card" v-for="(p, i) in playlists" :key="p.id" @click="goPlaylist(p.id)">
          <div class="cover">
            <img v-if="p.cover" :src="imgUrl(p.cover)" :alt="p.name" loading="lazy" style="width:100%;height:100%;object-fit:cover" />
            <div v-else class="cover-tile" :style="{ background: tileColor(i) }">{{ noteIcon }}</div>
            <div class="cover-count"><svg viewBox="0 0 24 24"><path d="M8 5v14l11-7Z"/></svg>{{ fmtCount(p.play_count) }}</div>
            <button class="cover-play" @click.stop="goPlaylist(p.id)"><svg viewBox="0 0 24 24"><path d="M8 5v14l11-7Z"/></svg></button>
          </div>
          <div class="card-name">{{ p.name }}</div>
          <div class="card-sub">{{ p.author || p.count + ' 首' }}</div>
        </div>
      </div>
      <div v-else class="empty">暂无歌单</div>
    </section>

    <!-- 排行榜 -->
    <section class="section fade-in">
      <div class="section-head">
        <div class="section-title">排行榜</div>
        <div class="section-more" @click="router.push('/ranking')">更多榜单 <svg viewBox="0 0 24 24"><path d="m9 6 6 6-6 6"/></svg></div>
      </div>
      <div class="rank-wrap" v-if="ranks.length">
        <div class="rank-card" v-for="r in ranks" :key="r.id">
          <div class="rank-title-row">
            <div class="t">{{ r.name }}</div>
            <div class="m" @click="router.push('/ranking')">更多 <svg viewBox="0 0 24 24"><path d="m9 6 6 6-6 6"/></svg></div>
          </div>
          <div
            class="rank-row"
            v-for="(s, i) in r.songs"
            :key="i"
            @click="player.playList(r.songs, i)"
          >
            <span class="rank-no">{{ i + 1 }}</span>
            <div class="rank-info">
              <div class="n">{{ s.name }}</div>
              <div class="a">{{ s.singer }}</div>
            </div>
            <button class="play-sm" @click.stop="player.playList(r.songs, i)"><svg viewBox="0 0 24 24"><path d="M8 5v14l11-7Z"/></svg></button>
          </div>
        </div>
      </div>
      <div v-else class="empty">榜单加载中...</div>
    </section>

    <!-- 人气歌手 -->
    <section class="section fade-in">
      <div class="section-head">
        <div class="section-title">人气歌手</div>
        <div class="section-more" @click="router.push('/search?type=author')">全部歌手 <svg viewBox="0 0 24 24"><path d="m9 6 6 6-6 6"/></svg></div>
      </div>
      <div class="h-scroll" v-if="artists.length">
        <div class="artist" v-for="(a, i) in artists" :key="i" @click="goArtist(a)">
          <div class="artist-photo" :style="{ background: tileColor(i * 2 + 1) }">
            <img v-if="a.pic" :src="a.pic" :alt="a.name" loading="lazy" />
            <svg v-else viewBox="0 0 24 24"><path d="M8.5 17.5a3.5 3.5 0 1 1-1-2.47l8-2.2V7.4L6.8 10.4a4.5 4.5 0 1 0 1.2 2.9l8-2.2V4.4a1 1 0 0 1 1.3-.95l3.2 1A1 1 0 0 1 21 5.4v8.1a4.5 4.5 0 1 0-1.3-3.2v3.9l-11.2 3.3Z"/></svg>
          </div>
          <div class="artist-name">{{ a.name }}</div>
          <div class="artist-tag">{{ a.tag }}</div>
        </div>
      </div>
    </section>

    <!-- 新歌首发 -->
    <section class="section fade-in">
      <div class="section-head">
        <div class="section-title">新歌首发</div>
        <div class="section-more" v-if="newSongs.length">更多新歌 <svg viewBox="0 0 24 24"><path d="m9 6 6 6-6 6"/></svg></div>
      </div>
      <div v-if="newSongs.length" class="song-table">
        <div class="song-head"><span></span><span></span><span class="h-name">歌曲</span><span>专辑</span><span>歌手</span><span></span><span>时长</span></div>
        <div
          class="song-row"
          v-for="(song, idx) in newSongs"
          :key="song.id + idx"
          :class="{ playing: isCurrent(song) }"
          @click="player.playList(newSongs, idx)"
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
          <div class="s-album link" @click.stop="goAlbum(song.album_id)">{{ song.album }}</div>
          <div class="s-artist link" @click.stop="goArtist(song.singer_id)">{{ song.singer }}</div>
          <button class="like" :class="{ liked: player.isLiked(song) }" @click.stop="player.toggleLike(song)">
            <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
          </button>
          <div class="dur">{{ fmtDuration(song.duration) }}</div>
        </div>
      </div>
      <div v-else class="empty">新歌加载中...</div>
    </section>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '../stores/player'
import {
  getTopPlaylist, getTopSong, getRankList, getRankAudio,
  imgUrl
} from '../utils/api'

const router = useRouter()
const player = usePlayerStore()

const TILES = ['#EAF3FD','#F0EAFD','#FDEEEA','#EAFBF1','#FDF5E6','#EAF9FA','#F6EAFD','#EDF2F7','#FDF0F3','#EAF7E6','#F5F0EA','#E8EEFD']
const tileColor = i => TILES[((i % TILES.length) + TILES.length) % TILES.length]
const noteIcon = '♪'

const playlists = ref([])
const ranks = ref([])
const artists = ref([])
const newSongs = ref([])
const loading = ref(true)

const banners = computed(() => {
  return playlists.value.slice(0, 5).map((p, i) => ({
    title: p.name,
    desc: (p.author || '精选歌单') + ' · ' + fmtCount(p.play_count),
    tag: i === 0 ? '热门推荐' : i === 1 ? '编辑精选' : '为你发现',
    img: imgUrl(p.cover, 400),
    color: tileColor(i * 3),
    id: p.id
  }))
})

const bannerIdx = ref(0)
let bannerTimer = null

onMounted(async () => {
  try {
    const [plRes, topRes, rankRes] = await Promise.all([
      getTopPlaylist(1, 10),
      getTopSong(1, 20),
      getRankList()
    ])

    if (plRes.data.status === 1 && Array.isArray(plRes.data.data)) {
      playlists.value = plRes.data.data.slice(0, 10)
    }

    if (topRes.data.status === 1 && Array.isArray(topRes.data.data)) {
      newSongs.value = topRes.data.data.map(normalizeSong).filter(Boolean).slice(0, 12)
      buildArtists(newSongs.value)
    }

    if (rankRes.data.status === 1 && rankRes.data.data && rankRes.data.data.info) {
      const info = rankRes.data.data.info.slice(0, 3)
      ranks.value = await Promise.all(info.map(async (r) => {
        const id = r.id != null ? String(r.id) : ''
        let songs = []
        try {
          const audioRes = await getRankAudio(id, 1, 5)
          if (audioRes.data.status === 1 && audioRes.data.data && audioRes.data.data.songlist) {
            songs = audioRes.data.data.songlist.map(normalizeRankSong).filter(Boolean).slice(0, 5)
          }
        } catch (e) { /* ignore */ }
        return { id, name: r.rankname || '', cover: r.img_9 || '', songs }
      }))
    }
  } catch (e) {
    console.error('首页数据加载失败:', e)
  } finally {
    loading.value = false
  }

  startAuto()
})

onBeforeUnmount(() => stopAuto())

function startAuto() {
  stopAuto()
  bannerTimer = setInterval(() => {
    bannerIdx.value = (bannerIdx.value + 1) % Math.max(1, banners.value.length)
  }, 5000)
}
function stopAuto() {
  if (bannerTimer) { clearInterval(bannerTimer); bannerTimer = null }
}

function buildArtists(songs) {
  const seen = new Set()
  const list = []
  for (const s of songs) {
    if (!s.singer) continue
    const key = s.singer
    if (seen.has(key)) continue
    seen.add(key)
    list.push({
      name: s.singer,
      id: s.singer_id,
      pic: s.img || '',
      tag: list.length % 3 === 0 ? '热门歌手' : list.length % 3 === 1 ? '新晋实力' : '经典之声'
    })
    if (list.length >= 10) break
  }
  artists.value = list
}

function normalizeSong(item) {
  const id = item.id != null ? String(item.id) : ''
  if (!id) return null
  return {
    id,
    name: item.name || '',
    singer: item.singer || '',
    album: item.album || '',
    album_id: item.album_id != null ? String(item.album_id) : '',
    singer_id: item.singer_id != null ? String(item.singer_id) : '',
    duration: Number(item.duration) || 0,
    img: imgUrl(item.img, 240)
  }
}

// 排行榜歌曲归一化（新接口）
function normalizeRankSong(item) {
  const hash = item?.deprecated?.hash || item?.hash
  if (!hash) return null
  const cover = (item?.album_info?.sizable_cover || item?.audio_info?.image || '').replace('{size}', '240')
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

function isCurrent(song) {
  return player.currentSong && player.currentSong.id === song.id
}

function goPlaylist(id) {
  if (id) router.push(`/playlist/${id}`)
}

function goBannerTarget(b) {
  if (b.id) router.push(`/playlist/${b.id}`)
}

function goArtist(a) {
  if (typeof a === 'string' || typeof a === 'number') {
    if (a) router.push(`/artist/${a}`)
    return
  }
  if (a.id) {
    router.push(`/artist/${a.id}`)
  } else if (a.name) {
    router.push(`/search?type=author&keyword=${encodeURIComponent(a.name)}`)
  }
}

function onQuick(act) {
  const map = {
    daily: () => router.push('/daily'),
    playlist: () => router.push('/search?type=special'),
    rank: () => router.push('/ranking')
  }
  if (map[act]) map[act]()
}

function pad(n) { return String(n).padStart(2, '0') }
function fmtDuration(sec) {
  if (!sec) return '--:--'
  const m = Math.floor(sec / 60)
  const s = Math.floor(sec % 60)
  return `${m}:${pad(s)}`
}
function fmtCount(n) {
  n = Number(n) || 0
  return n >= 10000 ? (n / 10000).toFixed(1) + '万' : String(n)
}

function goAlbum(id) { if (id) router.push('/album/' + id) }
</script>

<style scoped>
.home { max-width: 1280px; margin: 0 auto; }

/* ============ Banner ============ */
.banner {
  position: relative; height: 230px; border-radius: var(--r-xl);
  overflow: hidden; margin-bottom: 30px;
  background: var(--surface);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-md);
  cursor: pointer;
}
.banner-inner { position: absolute; inset: 0; display: flex; transition: transform .5s var(--ease); }
.banner-slide {
  flex: 0 0 100%; display: flex; align-items: center;
  padding: 0 48px; gap: 44px;
  position: relative;
}
/* 色块渐变氛围：用每个歌单的 tile 色做柔和背景 */
.banner-slide::before {
  content: '';
  position: absolute; inset: 0;
  background: radial-gradient(560px 280px at 20% 30%, var(--banner-glow, rgba(44,166,248,.14)), transparent 70%);
  pointer-events: none;
}
.banner-art {
  flex-shrink: 0; width: 168px; height: 168px; border-radius: 16px;
  background: var(--surface); border: 1px solid var(--border);
  box-shadow: var(--shadow-lg);
  display: grid; place-items: center; overflow: hidden;
  position: relative;
}
.banner-art img { width: 100%; height: 100%; object-fit: cover; }
.banner-art svg { width: 56px; height: 56px; fill: var(--primary); opacity: .9; }
.banner-info h2 { font-size: 30px; font-weight: 800; letter-spacing: -.03em; margin-bottom: 8px; }
.banner-info p { font-size: 14px; color: var(--text-2); max-width: 420px; margin-bottom: 20px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.banner-tag {
  display: inline-block; font-size: 11px; font-weight: 700;
  color: var(--primary-deep); background: var(--primary-soft);
  padding: 3px 10px; border-radius: 99px; margin-bottom: 12px; letter-spacing: .05em;
}
.banner-play {
  display: inline-flex; align-items: center; gap: 8px;
  background: var(--primary); color: #fff;
  font-size: 13.5px; font-weight: 600;
  border: none; border-radius: var(--r-full);
  padding: 10px 22px; cursor: pointer;
  box-shadow: 0 4px 14px rgba(44,166,248,.35);
  transition: transform .15s var(--ease), box-shadow .15s;
}
.banner-play:hover { transform: translateY(-1px); box-shadow: 0 6px 18px rgba(44,166,248,.45); }
.banner-play svg { width: 14px; height: 14px; fill: #fff; }
.banner-dots {
  position: absolute; bottom: 16px; left: 50%; transform: translateX(-50%);
  display: flex; gap: 6px;
}
.banner-dots i {
  width: 5px; height: 5px; border-radius: 99px; background: rgba(15,23,42,.18);
  transition: all .3s var(--ease); cursor: pointer;
}
[data-theme="dark"] .banner-dots i { background: rgba(255,255,255,.22); }
.banner-dots i.on { width: 18px; background: var(--primary); }

/* ============ 快捷功能宫格 ============ */
.quick-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 30px; }
.quick-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-md);
  padding: 18px 16px;
  display: flex; align-items: center; gap: 12px;
  cursor: pointer;
  transition: all .18s var(--ease);
}
.quick-card:hover { transform: translateY(-2px); box-shadow: var(--shadow-md); border-color: var(--border-strong); }
.quick-ic {
  width: 40px; height: 40px; border-radius: 12px; flex-shrink: 0;
  display: grid; place-items: center;
  box-shadow: 0 4px 12px rgba(15,23,42,.10);
}
.quick-ic svg { width: 19px; height: 19px; stroke: #fff; fill: none; stroke-width: 1.9; stroke-linecap: round; stroke-linejoin: round; }
.quick-name { font-size: 13.5px; font-weight: 600; }
.quick-sub { font-size: 11px; color: var(--text-3); margin-top: 1px; }
.quick-arrow { width: 14px; height: 14px; margin-left: auto; stroke: var(--text-3); fill: none; stroke-width: 2; stroke-linecap: round; stroke-linejoin: round; flex-shrink: 0; }
.quick-card:hover .quick-arrow { stroke: var(--primary); }
@media (max-width: 860px) {
  .banner { height: 160px; }
  .banner-slide { padding: 0 24px; gap: 20px; }
  .banner-art { width: 110px; height: 110px; }
  .banner-info h2 { font-size: 22px; }
  .quick-grid { grid-template-columns: repeat(2, 1fr); }
}
</style>
