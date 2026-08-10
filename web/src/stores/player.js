import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth'
import { getSongURL, getPrivilegeLite, getLyric, getSongQualities, setAuth, clearAuth, imgUrl,
  getUserPlaylist, getPlaylistSongs, addPlaylistTracks, deletePlaylistTracks, getCloudURL } from '../utils/api'

export const usePlayerStore = defineStore('player', () => {
  const currentSong = ref(null)
  const audioUrl = ref('')
  const isPlaying = ref(false)
  const currentTime = ref(0)
  const duration = ref(0)
  const volume = ref(0.8)
  const isMuted = ref(false)
  // 当前/默认音质：128/320/flac/high(Hi-Res)。持久化到 localStorage
  const quality = ref(localStorage.getItem('vibe_quality') || '128')
  const actualQuality = ref('')         // 当前歌曲实际播放的音质（后端降级后的）
  const songQualities = ref([])          // 当前歌曲所有可用音质列表
  const lyrics = ref([])
  const currentLyricIndex = ref(0)
  const playMode = ref('loop') // loop, one, shuffle
  const queue = ref([])
  const queueIndex = ref(-1)
  const playError = ref('') // 播放地址获取失败原因（如 VIP 限制）

  // 云端「我喜欢」歌单收藏（登录后从酷狗云端加载，不再用本地 localStorage）
  const likedSongs = ref([])
  const likePlaylistId = ref('') // 歌单 global_collection_id，getPlaylistSongs 用
  const likeListNum = ref('')    // 数字 listid，add_song 用
  const likeLoading = ref(false)
  // 最近播放（本地历史）
  const history = ref(JSON.parse(localStorage.getItem('vibe_history') || '[]'))
  // 本地音乐文件（session 内，URL 刷新后失效需重新选择）
  const localFiles = ref(JSON.parse(sessionStorage.getItem('vibe_local') || '[]'))

  const hasNext = computed(() => queueIndex.value < queue.value.length - 1)
  const hasPrev = computed(() => queueIndex.value > 0)

  // 加载歌曲
  async function loadSong(song) {
    // 清空播放地址让上一首立即停止（即使新歌加载失败也不会继续出声），
    // 但不改 isPlaying：playNow 会重新 play()，next/prev 依赖它保持播放意图
    audioUrl.value = ''
    currentSong.value = { ...song, img: imgUrl(song.img, 300) }
    currentTime.value = 0
    duration.value = 0
    currentLyricIndex.value = 0
    playError.value = ''
    recordHistory(song)

    if (song.localUrl) {
      // 本地文件直接播放，跳过在线接口
      audioUrl.value = song.localUrl
      return
    }
    if (String(song.id || '').startsWith('local_')) {
      // 本地文件 URL 已失效（刷新后），停止加载
      audioUrl.value = ''
      return
    }

    // 获取播放地址（VIP 歌由后端用概念版 priv_url 解锁，需传 vip_token）
    try {
      const auth = useAuthStore()

      // 云盘文件：使用云盘专用 URL 接口
      if (song.cloudFile) {
        const urlRes = await getCloudURL(song.id, song.cloudAlbumAudioID || 0, song.cloudAudioID || 0, song.name || '')
        if (urlRes.data.status === 1 && urlRes.data.data && urlRes.data.data.url) {
          audioUrl.value = urlRes.data.data.url
          actualQuality.value = 'cloud'
        } else {
          audioUrl.value = ''
          playError.value = urlRes.data.error || '无法获取云盘文件播放地址'
        }
        return
      }

      const res = await getSongURL(song.id, song.album_id || '', 0, quality.value, auth.vipToken, auth.vipType)
      if (res.data.status === 2) {
        audioUrl.value = ''
        playError.value = res.data.error || '该歌曲暂无法播放'
      } else if (res.data.status === 1 && res.data.data) {
        audioUrl.value = res.data.data.url
        actualQuality.value = res.data.data.quality || quality.value
      } else {
        audioUrl.value = ''
        playError.value = res.data.error || '无法获取播放地址'
      }
    } catch (e) {
      audioUrl.value = ''
      playError.value = e.response?.data?.error || '无法获取播放地址'
      console.error('获取播放地址失败:', e)
    }

    // 获取可用音质列表
    getSongQualities(song.id).then(res => {
      if (res.data.status === 1 && Array.isArray(res.data.data)) {
        songQualities.value = res.data.data
      }
    }).catch(() => {})

    // 获取歌词
    try {
      const res = await getLyric(song.id, song.album_id || '')
      if (res.data.status === 1 && res.data.data) {
        parseLyrics(res.data.data.content || '')
      }
    } catch (e) {
      console.error('获取歌词失败:', e)
    }
  }

  // 切换音质：设置默认音质并持久化；有正在播放的歌曲时重新请求播放地址，成功才生效
  async function setQuality(q) {
    const old = quality.value
    quality.value = q
    localStorage.setItem('vibe_quality', q)
    if (!currentSong.value) return true
    if (q === old && audioUrl.value) return true
    try {
      const auth = useAuthStore()
      const res = await getSongURL(currentSong.value.id, currentSong.value.album_id || '', 0, q, auth.vipToken, auth.vipType)
      if (res.data.status === 1 && res.data.data && res.data.data.url) {
        audioUrl.value = res.data.data.url
        actualQuality.value = res.data.data.quality || q
        playError.value = ''
        return true
      }
      return false
    } catch (e) {
      quality.value = old
      localStorage.setItem('vibe_quality', old)
      return false
    }
  }

  // 仅切换当前歌曲音质（不持久化，切歌后仍按设置里的默认音质）
  async function switchQuality(q) {
    if (!currentSong.value) return false
    try {
      const auth = useAuthStore()
      const res = await getSongURL(currentSong.value.id, currentSong.value.album_id || '', 0, q, auth.vipToken, auth.vipType)
      if (res.data.status === 1 && res.data.data && res.data.data.url) {
        audioUrl.value = res.data.data.url
        actualQuality.value = res.data.data.quality || q
        playError.value = ''
        return true
      }
      return false
    } catch (e) {
      return false
    }
  }

  function parseLyrics(lrcText) {
    if (!lrcText) {
      lyrics.value = []
      return
    }
    const lines = lrcText.split('\n')
    const parsed = []
    for (const line of lines) {
      const match = line.match(/\[(\d{2}):(\d{2})\.(\d{2,3})\](.*)/)
      if (match) {
        const min = parseInt(match[1])
        const sec = parseInt(match[2])
        const ms = parseInt(match[3].padEnd(3, '0'))
        const time = min * 60 + sec + ms / 1000
        const text = match[4].trim()
        if (text) {
          parsed.push({ time, text })
        }
      }
    }
    parsed.sort((a, b) => a.time - b.time)
    lyrics.value = parsed
  }

  function updateCurrentTime(time) {
    currentTime.value = time
    // 更新当前歌词索引
    const idx = lyrics.value.findIndex((l, i) => {
      const next = lyrics.value[i + 1]
      return time >= l.time && (!next || time < next.time)
    })
    if (idx >= 0) currentLyricIndex.value = idx
  }

  // 跳转到指定时间（由外部 audio 元素调用）
  let _audioEl = null
  function setAudioEl(el) { _audioEl = el }
  function seekTo(time) {
    currentTime.value = time
    if (_audioEl) _audioEl.currentTime = time
  }

  function play() { isPlaying.value = true }
  function pause() { isPlaying.value = false }
  function togglePlay() { isPlaying.value = !isPlaying.value }

  function setVolume(v) {
    volume.value = Math.max(0, Math.min(1, v))
  }

  function toggleMute() { isMuted.value = !isMuted.value }

  function next() {
    if (queue.value.length === 0) return
    if (playMode.value === 'shuffle') {
      queueIndex.value = Math.floor(Math.random() * queue.value.length)
    } else if (queueIndex.value < queue.value.length - 1) {
      queueIndex.value++
    } else {
      queueIndex.value = 0
    }
    loadSong(queue.value[queueIndex.value])
  }

  function prev() {
    if (queue.value.length === 0) return
    if (queueIndex.value > 0) {
      queueIndex.value--
    } else {
      queueIndex.value = queue.value.length - 1
    }
    loadSong(queue.value[queueIndex.value])
  }

  function addToQueue(song) {
    // 检查是否已存在
    const exists = queue.value.find(s => s.id === song.id)
    if (!exists) {
      queue.value.push(song)
    }
    if (queueIndex.value === -1) {
      queueIndex.value = 0
      loadSong(queue.value[0])
    }
  }

  function playNow(song) {
    const idx = queue.value.findIndex(s => s.id === song.id)
    if (idx >= 0) {
      queueIndex.value = idx
    } else {
      queue.value.push(song)
      queueIndex.value = queue.value.length - 1
    }
    loadSong(song)
    play()
  }

  // 播放整个列表：列表点击/播放全部时使用，保证「下一首」能切到列表下一首
  function playList(list, index = 0) {
    if (!Array.isArray(list) || list.length === 0) return
    const songs = list.slice()
    queue.value = songs
    queueIndex.value = Math.min(Math.max(0, index), songs.length - 1)
    loadSong(songs[queueIndex.value])
    play()
  }

  // 在现有队列中按索引播放
  function playAt(i) {
    if (i < 0 || i >= queue.value.length) return
    queueIndex.value = i
    loadSong(queue.value[i])
    play()
  }

  function clearQueue() {
    queue.value = []
    queueIndex.value = -1
    currentSong.value = null
    audioUrl.value = ''
    lyrics.value = []
    pause()
  }

  // ===== 我喜欢（云端歌单） =====

  // 登录后加载云端「我喜欢」歌单：先找歌单 id，再拉取全部歌曲（分页循环），
  // 同时把旧版 localStorage 收藏一次性迁移到云端（仅迁移云端没有的）
  async function loadLiked() {
    const auth = useAuthStore()
    if (!auth.isLoggedIn) {
      likedSongs.value = []
      likePlaylistId.value = ''
      return
    }
    likeLoading.value = true
    try {
      // 从用户歌单列表中找到默认「我喜欢」歌单
      const plRes = await getUserPlaylist(undefined, undefined, 1, 100)
      let pl = null
      if (plRes.data.status === 1 && Array.isArray(plRes.data.data)) {
        pl = plRes.data.data.find(p => p.name === '我喜欢' || p.name === '我喜欢的音乐') || null
      }
      if (!pl) {
        likedSongs.value = []
        likePlaylistId.value = ''
        return
      }
      likePlaylistId.value = pl.id
      likeListNum.value = pl.list_id || ''

      // 分页拉取全部歌单歌曲（每页 50 首，最多 20 页 = 1000 首）
      const MAX_PAGES = 20
      const PAGE_SIZE = 50
      const allSongs = []
      for (let page = 1; page <= MAX_PAGES; page++) {
        const sRes = await getPlaylistSongs(pl.id, page, PAGE_SIZE)
        if (sRes.data.status === 1 && Array.isArray(sRes.data.data) && sRes.data.data.length > 0) {
          allSongs.push(...sRes.data.data)
          if (sRes.data.data.length < PAGE_SIZE) break
        } else {
          break
        }
      }
      likedSongs.value = allSongs
      await migrateLocalLiked()
    } catch (e) {
      console.error('加载云端喜欢列表失败:', e)
    } finally {
      likeLoading.value = false
    }
  }

  // 旧版本地收藏（localStorage vibe_liked）一次性迁入云端「我喜欢」歌单
  async function migrateLocalLiked() {
    const raw = localStorage.getItem('vibe_liked')
    if (!raw || !likePlaylistId.value) return
    let old = []
    try { old = JSON.parse(raw) } catch (e) { old = [] }
    if (!Array.isArray(old) || !old.length) {
      localStorage.removeItem('vibe_liked')
      return
    }
    const existing = new Set(likedSongs.value.map(s => String(s.id)))
    const toAdd = old.filter(s => s && s.id && !existing.has(String(s.id)))
    if (toAdd.length) {
      try {
        await addPlaylistTracks(likeListNum.value, toAdd)
        likedSongs.value = [...toAdd, ...likedSongs.value]
      } catch (e) {
        console.error('迁移旧收藏到云端失败:', e)
      }
    }
    localStorage.removeItem('vibe_liked')
  }

  // 红心收藏/取消：登录后操作云端「我喜欢」歌单，未登录弹登录框
  async function toggleLike(song) {
    if (!song || !song.id) return
    const auth = useAuthStore()
    if (!auth.isLoggedIn) {
      window.dispatchEvent(new CustomEvent('vibe:login-required'))
      return
    }
    if (!likePlaylistId.value) {
      await loadLiked()
    }
    if (!likePlaylistId.value) return
    const liked = isLiked(song)
    try {
      if (liked) {
        let item = likedSongs.value.find(s => {
          const sid = String(song.id)
          const smix = String(song.mixsongid || '')
          const sname = normalizeKey(song.name)
          const ssinger = normalizeKey(song.singer)
          if (String(s.id) === sid) return true
          if (smix && String(s.mixsongid || '') === smix) return true
          if (smix && String(s.id) === smix) return true
          if (sid && String(s.mixsongid || '') === sid) return true
          if (sname && ssinger && s.name && s.singer) {
            const sn = normalizeKey(s.name)
            if (sn !== sname) return false
            const ss = normalizeKey(s.singer)
            if (!(ss === ssinger || ss.includes(ssinger) || ssinger.includes(ss))) return false
            const sa = normalizeKey(s.album || '')
            const sla = normalizeKey(song.album || '')
            if (sa && sla && sa !== sla) return false
            return true
          }
          return false
        })
        // 如果本地记录缺少 file_id（刚添加的搜索歌曲），重新从云端加载获取正确数据
        if (item && !item.file_id) {
          await loadLiked()
          item = likedSongs.value.find(s => {
            if (String(s.id) === String(song.id)) return true
            const smix2 = String(song.mixsongid || '')
            if (smix2 && String(s.mixsongid || '') === smix2) return true
            return false
          })
        }
        if (item && item.file_id) {
          await deletePlaylistTracks(likeListNum.value, [item.file_id])
          likedSongs.value = likedSongs.value.filter(s => String(s.file_id) !== String(item.file_id))
        }
      } else {
        const res = await addPlaylistTracks(likeListNum.value, [song])
        // 检查云端返回是否成功
        const errCode = res.data?.error_code != null ? Number(res.data.error_code) : (res.data?.status === 0 ? -1 : 0)
        if (errCode !== 0) {
          console.warn('收藏未成功, error_code:', errCode)
          return
        }
        // 云端成功，更新本地状态
        likedSongs.value.unshift({ ...song })
        // 异步刷新云端列表获取正确 file_id（后续取消收藏需要）
        loadLiked().catch(() => {})
      }
    } catch (e) {
      console.error('收藏操作失败:', e.response?.status, e.response?.data?.error || e.response?.data?.error_msg || '')
    }
  }

  // 标准化后比较（去除非字母数字字符、转小写）
  function normalizeKey(s) {
    return String(s || '').replace(/[^a-zA-Z0-9\u4e00-\u9fff]/g, '').toLowerCase()
  }
  function isLiked(songOrId, mixsongid, name, singer) {
    const id   = typeof songOrId === 'object' ? songOrId.id : songOrId
    const mid  = typeof songOrId === 'object' ? songOrId.mixsongid : mixsongid
    const sname = typeof songOrId === 'object' ? songOrId.name : name
    const ssinger = typeof songOrId === 'object' ? songOrId.singer : singer
    const salbum = typeof songOrId === 'object' ? songOrId.album : ''

    const result = likedSongs.value.some(s => {
      const sid = String(s.id || '')
      const smix = String(s.mixsongid || '')
      const tid = String(id || '')
      const tmix = String(mid || '')
      if (sid === tid || smix === tmix || (tid && smix === tid) || (tmix && sid === tmix)) {
        return true
      }
      if (sname && ssinger && s.name && s.singer) {
        const sn = normalizeKey(s.name)
        const tn = normalizeKey(sname)
        if (sn && tn && sn === tn) {
          const ss = normalizeKey(s.singer)
          const ts = normalizeKey(ssinger)
          if (ss && ts && (ss === ts || ss.includes(ts) || ts.includes(ss))) {
            const sa = normalizeKey(s.album || '')
            const ta = normalizeKey(salbum || '')
            if (!sa || !ta || sa === ta) return true
          }
        }
      }
      return false
    })

    return result
  }

  // 登出时清空云端收藏缓存
  function clearLiked() {
    likedSongs.value = []
    likePlaylistId.value = ''
    likeLoading.value = false
  }

  // ===== 最近播放（本地历史） =====
  function persistHistory() {
    localStorage.setItem('vibe_history', JSON.stringify(history.value.slice(0, 100)))
  }
  function recordHistory(song) {
    if (!song || !song.id) return
    history.value = [{ ...song, played_at: Date.now() }]
      .concat(history.value.filter(h => h.id !== song.id))
      .slice(0, 100)
    persistHistory()
  }

  // ===== 本地音乐 =====
  function addLocalFiles(files) {
    const added = []
    for (const f of files) {
      if (!f.type.startsWith('audio/')) continue
      const song = {
        id: 'local_' + f.name + '_' + f.size,
        name: f.name.replace(/\.[^.]+$/, ''),
        singer: '本地音乐',
        album: '',
        album_id: '',
        duration: 0,
        img: '',
        localUrl: URL.createObjectURL(f),
        size: f.size
      }
      localFiles.value = [song, ...localFiles.value.filter(s => s.name !== f.name)]
      added.push(song)
    }
    sessionStorage.setItem('vibe_local', JSON.stringify(
      localFiles.value.map(s => ({ id: s.id, name: s.name, singer: s.singer, size: s.size }))
    ))
    return added
  }
  function removeLocalFile(song) {
    if (song.localUrl) URL.revokeObjectURL(song.localUrl)
    localFiles.value = localFiles.value.filter(s => s.id !== song.id)
    sessionStorage.setItem('vibe_local', JSON.stringify(
      localFiles.value.map(s => ({ id: s.id, name: s.name, singer: s.singer, size: s.size }))
    ))
  }

  return {
    currentSong, audioUrl, isPlaying, currentTime, duration,
    volume, isMuted, quality, actualQuality, songQualities, lyrics, currentLyricIndex, playMode,
    queue, queueIndex, hasNext, hasPrev, playError,
    likedSongs, likeLoading, history, localFiles,
    loadSong, play, pause, togglePlay, setVolume, toggleMute,
    next, prev, addToQueue, playNow, playList, playAt, clearQueue, updateCurrentTime, setQuality, switchQuality, seekTo, setAudioEl,
    toggleLike, isLiked, loadLiked, clearLiked, addLocalFiles, removeLocalFile
  }
})