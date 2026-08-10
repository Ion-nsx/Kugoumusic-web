import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// 后端返回 401「请先登录」时，全局派发登录请求事件
// （登录弹窗由 App.vue 监听处理，各页面无需各自判断）
api.interceptors.response.use(
  (res) => res,
  (err) => {
    const status = err.response?.status
    if (status === 401) {
      const msg = err.response?.data?.error || err.response?.data?.message || ''
      if (msg.includes('请先登录') || msg.includes('未登录')) {
        window.dispatchEvent(new CustomEvent('vibe:login-required'))
      }
    }
    return Promise.reject(err)
  }
)

// 设置 Authorization header
export function setAuth(token, userid, dfid, mid, guid) {
  const parts = []
  if (token) parts.push(`token=${token}`)
  if (userid) parts.push(`userid=${userid}`)
  if (dfid) parts.push(`dfid=${dfid}`)
  if (mid) parts.push(`mid=${mid}`)
  if (guid) parts.push(`guid=${guid}`)
  api.defaults.headers.common['Authorization'] = parts.join(';')
}

export function clearAuth() {
  delete api.defaults.headers.common['Authorization']
}

// 设备注册
export const registerDevice = () => api.get('/register/dev')

// 搜索（originalFirst 默认 true，传 false 关闭原版优先排序）
export const searchSongs = (keyword, page = 1, pagesize = 20, originalFirst = true) =>
  api.get('/search', { params: { keyword, page, pagesize, original_first: originalFirst ? undefined : '0' } })

export const searchSuggest = (keyword) =>
  api.get('/search/suggest', { params: { keyword } })

export const searchHot = () => api.get('/search/hot')

export const searchComplex = (keyword, page = 1, pagesize = 20) =>
  api.get('/search/complex', { params: { keyword, page, pagesize } })

// 分类型搜索（song/album/author/mv/special）
export const searchByType = (type, keyword, page = 1, pagesize = 20) =>
  api.get(`/search/${type}`, { params: { keyword, page, pagesize } })

// 歌词搜索
export const searchLyric = (keyword, artist = '', duration = '', album = '', originalFirst = true) =>
  api.get('/search/lyric', { params: { keyword, artist, duration, album, original_first: originalFirst ? undefined : '0' } })

// 歌曲
export const getSongURL = (hash, album_id = '', bitrate = 0, quality = '', vip_token = '', vip_type = '') =>
  api.get('/song/url', { params: { hash, album_id, bitrate, quality, vip_token, vip_type } })

export const getPrivilegeLite = (songIds) =>
  api.get('/privilege/lite', { params: { song_id: songIds.join(',') } })

export const getSongQualities = (hash) =>
  api.get('/song/qualities', { params: { hash } })

export const getSongDetail = (hash) =>
  api.get('/song/detail', { params: { hash } })

// 歌词
export const getLyric = (hash, album_id = '', timelength = '') =>
  api.get('/lyric', { params: { hash, album_id, timelength } })

// 推荐 & 歌单
export const getEverydayRecommend = () => api.get('/everyday/recommend')
export const getRecommend = () => api.get('/recommend')
export const getRecommendPlaylist = () => api.get('/recommend/playlist')
export const getTopPlaylist = (page = 1, pagesize = 20, category_id = '', sort = '') =>
  api.get('/top/playlist', { params: { page, pagesize, category_id, sort } })
export const getPlaylistDetail = (id) => api.get('/playlist/detail', { params: { id } })
export const getPlaylistSongs = (id, page = 1, pagesize = 20) =>
  api.get('/playlist/song', { params: { id, page, pagesize } })
export const getPlaylistCategories = () => api.get('/playlist/category')
export const getPlaylistTags = () => api.get('/playlist/tags')

// 排行榜 & 新歌新专辑
export const getRankList = () => api.get('/rank/list')
export const getRankAudio = (rankid, page = 1, pagesize = 30) =>
  api.get('/rank/audio', { params: { rankid, page, pagesize } })
export const getTopSong = (page = 1, pagesize = 30) =>
  api.get('/top/song', { params: { page, pagesize } })
export const getTopAlbum = (page = 1, pagesize = 30) =>
  api.get('/top/album', { params: { page, pagesize } })
export const getTopCard = (card_id = 1) => api.get('/top/card', { params: { card_id } })

// 专辑
export const getAlbumDetail = (album_id) => api.get('/album/detail', { params: { album_id } })
export const getAlbumSongs = (album_id, page = 1, pagesize = 20) =>
  api.get('/album/songs', { params: { album_id, page, pagesize } })

// 歌手
export const getArtistDetail = (singer_id) => api.get('/artist/detail', { params: { singer_id } })
export const getArtistAudios = (singer_id, page = 1, pagesize = 20) =>
  api.get('/artist/audios', { params: { singer_id, page, pagesize } })
export const getArtistHot = (singer_id) => api.get('/artist/hot', { params: { singer_id } })
export const getArtistAlbums = (singer_id, page = 1, pagesize = 20) =>
  api.get('/artist/album', { params: { singer_id, page, pagesize } })
export const followSinger = (singer_id, follow) =>
  api.post('/artist/follow', null, { params: { singer_id, follow: follow ? 1 : 0 } })
export const getSingerList = (type = 0, sextype = 0, musician = 0, hotsize = 60) =>
  api.get('/artist/list', { params: { type, sextype, musician, hotsize } })

// 登录
export const login = (username, password) =>
  api.post('/login', { username, password })
export const qrCreate = () => api.get('/login/qr/create')
export const qrGet = (key) => api.get('/login/qr/get', { params: { key } })
export const qrCheck = (key) => api.get('/login/qr/check', { params: { key } })
export const loginCellphone = (mobile, code, area_code = '86') =>
  api.post('/login/cellphone', { mobile, code, area_code })
export const sendCaptcha = (mobile, area_code = '86') =>
  api.post('/captcha/sent', { mobile, area_code })

// 用户
export const getUserDetail = () => api.get('/user/detail')
export const getUserPlaylist = (userid = '', token = '', page = 1, pagesize = 20) =>
  api.get('/user/playlist', { params: { userid, token, page, pagesize } })
export const getUserFavorite = (userid = '', token = '', page = 1, pagesize = 20) =>
  api.get('/user/favorite', { params: { userid, token, page, pagesize } })
export const getUserVIPDetail = () => api.get('/user/vip/detail')
export const getDailyVIP = () => api.get('/youth/vip')
export const getUserFollow = () => api.get('/user/follow')

// 歌单管理（需登录）
export const createPlaylist = (name, is_pri = 0) =>
  api.post('/playlist/create', null, { params: { name, is_pri } })
export const deletePlaylist = (listid) =>
  api.post('/playlist/delete', null, { params: { listid } })
export const addPlaylistTracks = (listid, songs) =>
  api.post('/playlist/tracks/add', songs, { params: { listid } })
export const deletePlaylistTracks = (listid, file_ids) =>
  api.post('/playlist/tracks/del', { file_ids }, { params: { listid } })

// 猜你喜欢
export const getFMSongs = (action = 'play', songPoolId = '0', remainSongCnt = 0) =>
  api.get('/fm/songs', { params: { action, song_pool_id: songPoolId, remain_songcnt: remainSongCnt } })

// 搜索默认词
export const getSearchDefault = () => api.get('/search/default')

// 歌曲高潮
export const getSongClimax = (hash) =>
  api.get('/song/climax', { params: { hash } })

// 听歌历史
export const getUserHistory = (bp = '') =>
  api.get('/user/history', { params: { bp } })

// 听歌排行
export const getUserListen = (type = 1) => api.get('/user/listen', { params: { type } })
export const getSongComments = (mixsongid, page = 1, pagesize = 30) =>
  api.get('/comment/song', { params: { mixsongid, page, pagesize } })
export const getPlaylistComments = (id, page = 1, pagesize = 30) =>
  api.get('/comment/playlist', { params: { id, page, pagesize } })
export const getAlbumComments = (id, page = 1, pagesize = 30) =>
  api.get('/comment/album', { params: { id, page, pagesize } })
export const getCommentCount = (hash = '', special_id = '') =>
  api.get('/comment/count', { params: { hash, special_id } })
export const getCommentFloor = (resource_type, { code = '', mixsongid = '', special_id = '', tid = '' }, page = 1, pagesize = 30) =>
  api.get('/comment/floor', { params: { resource_type, code, mixsongid, special_id, tid, page, pagesize } })
export const getCommentHotWords = (mixsongid, page = 1, pagesize = 30) =>
  api.get('/comment/hotword', { params: { mixsongid, page, pagesize } })
export const sendComment = (resource_type, { mixsongid = '', childrenid = '', content = '', code = '' }) =>
  api.post('/comment/send', null, { params: { resource_type, mixsongid, childrenid, content, code } })

// 处理酷狗图片 URL 中的 {size} 占位符（统一替换为实际尺寸）
export function imgUrl(url, size = 480) {
  if (!url) return ''
  return String(url).replace('{size}', size)
}

// ===== 云盘 =====
// 获取云盘文件列表
export const getCloudList = (page = 1, pagesize = 30) =>
  api.get('/cloud/list', { params: { page, pagesize } })
// 删除云盘文件（ids 为 kv_id 逗号分隔列表）
export const deleteCloudFile = (ids, albumAudioIds = '') =>
  api.get('/cloud/delete', { params: { ids, album_audio_ids: albumAudioIds } })
// 获取云盘文件播放地址
export const getCloudURL = (hash, album_audio_id = 0, audio_id = 0, name = '') =>
  api.get('/cloud/url', { params: { hash, album_audio_id, audio_id, name } })
// 曲库匹配（上传前匹配文件 hash）
export const cloudMatch = (hash, album_audio_id = 0) =>
  api.get('/cloud/match', { params: { hash, album_audio_id } })
// 上传文件到云盘（multipart form，auto_match=1 自动匹配曲库）
// onProgress 可选回调：接收 0-100 的百分比，反映文件上传到后端服务器的进度
export const cloudUpload = (formData, onProgress) =>
  api.post('/cloud/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 600000,
    onUploadProgress: (e) => {
      if (onProgress && e.total) {
        onProgress(Math.round((e.loaded / e.total) * 100))
      }
    },
  })

export default api