package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"vibe/api"
	"vibe/middleware"
)

// 嵌入前端构建产物（web/dist）
//
//go:embed web/dist
var distFS embed.FS

func main() {
	// 平台切换：默认概念版(lite)；VIBE_PLATFORM=std 切标准版
	lite := true
	if os.Getenv("VIBE_PLATFORM") == "std" {
		api.SetUseLite(false)
		lite = false
	} else {
		api.SetUseLite(true)
	}

	mux := http.NewServeMux()

	// ===== 设备 =====
	mux.HandleFunc("/api/register/dev", handleRegisterDevice)

	// ===== 搜索 =====
	mux.HandleFunc("/api/search", handleSearch)
	mux.HandleFunc("/api/search/suggest", handleSearchSuggest)
	mux.HandleFunc("/api/search/hot", handleSearchHot)
	mux.HandleFunc("/api/search/complex", handleSearchComplex)
	mux.HandleFunc("/api/search/album", func(w http.ResponseWriter, r *http.Request) { handleSearchByType(w, r, "album") })
	mux.HandleFunc("/api/search/author", func(w http.ResponseWriter, r *http.Request) { handleSearchByType(w, r, "author") })
	mux.HandleFunc("/api/search/mv", func(w http.ResponseWriter, r *http.Request) { handleSearchByType(w, r, "mv") })
	mux.HandleFunc("/api/search/special", func(w http.ResponseWriter, r *http.Request) { handleSearchByType(w, r, "special") })
	mux.HandleFunc("/api/search/lyric", handleSearchLyric)
	mux.HandleFunc("/api/search/default", handleSearchDefault)

	// ===== 歌曲 & 歌词 =====
	mux.HandleFunc("/api/song/url", handleSongURL)
	mux.HandleFunc("/api/song/qualities", handleSongQualities)
	mux.HandleFunc("/api/privilege/lite", handlePrivilegeLite)
	mux.HandleFunc("/api/song/detail", handleSongDetail)
	mux.HandleFunc("/api/song/climax", handleSongClimax)
	mux.HandleFunc("/api/lyric", handleLyric)

	// ===== 图片代理 =====
	mux.HandleFunc("/api/images", handleImageProxy)

	// ===== 推荐 & 歌单 =====
	mux.HandleFunc("/api/everyday/recommend", handleEverydayRecommend)
	mux.HandleFunc("/api/recommend", handleRecommend)
	mux.HandleFunc("/api/recommend/playlist", handleRecommendPlaylist)
	mux.HandleFunc("/api/top/playlist", handleTopPlaylist)
	mux.HandleFunc("/api/playlist/detail", handlePlaylistDetail)
	mux.HandleFunc("/api/playlist/song", handlePlaylistSongs)
	mux.HandleFunc("/api/playlist/category", handlePlaylistCategories)
	mux.HandleFunc("/api/playlist/tags", handlePlaylistTags)

	// ===== 歌单管理（需登录） =====
	mux.HandleFunc("/api/playlist/create", handlePlaylistCreate)
	mux.HandleFunc("/api/playlist/delete", handlePlaylistDelete)
	mux.HandleFunc("/api/playlist/tracks/add", handlePlaylistTracksAdd)
	mux.HandleFunc("/api/playlist/tracks/del", handlePlaylistTracksDel)

	// ===== 评论 =====
	mux.HandleFunc("/api/comment/song", handleSongComments)
	mux.HandleFunc("/api/comment/playlist", handlePlaylistComments)
	mux.HandleFunc("/api/comment/album", handleAlbumComments)
	mux.HandleFunc("/api/comment/count", handleCommentCount)
	mux.HandleFunc("/api/comment/floor", handleCommentFloor)
	mux.HandleFunc("/api/comment/hotword", handleCommentHotWords)
	mux.HandleFunc("/api/comment/send", handleCommentSend)

	// ===== 排行榜 & 新歌新专辑 =====
	mux.HandleFunc("/api/rank/list", handleRankList)
	mux.HandleFunc("/api/rank/audio", handleRankAudio)
	mux.HandleFunc("/api/top/song", handleTopSong)
	mux.HandleFunc("/api/top/album", handleTopAlbum)
	mux.HandleFunc("/api/top/card", handleTopCard)
	mux.HandleFunc("/api/fm/songs", handleFMSongs)

	// ===== 专辑 & 歌手 =====
	mux.HandleFunc("/api/album/detail", handleAlbumDetail)
	mux.HandleFunc("/api/album/songs", handleAlbumSongs)
	mux.HandleFunc("/api/artist/detail", handleArtistDetail)
	mux.HandleFunc("/api/artist/audios", handleArtistAudios)
	mux.HandleFunc("/api/artist/hot", handleArtistHot)
	mux.HandleFunc("/api/artist/album", handleArtistAlbums)
	mux.HandleFunc("/api/artist/follow", handleArtistFollow)
	mux.HandleFunc("/api/artist/list", handleSingerList)

	// ===== 登录 =====
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/login/qr/create", handleQRCreate)
	mux.HandleFunc("/api/login/qr/get", handleQRGet)
	mux.HandleFunc("/api/login/qr/check", handleQRCheck)
	mux.HandleFunc("/api/login/cellphone", handleLoginCellphone)
	mux.HandleFunc("/api/captcha/sent", handleCaptchaSent)

	// ===== 用户 & VIP =====
	mux.HandleFunc("/api/user/detail", handleUserDetail)
	mux.HandleFunc("/api/user/playlist", handleUserPlaylist)
	mux.HandleFunc("/api/user/favorite", handleUserFavorite)
	mux.HandleFunc("/api/user/vip/detail", handleUserVIPDetail)
	mux.HandleFunc("/api/user/history", handleUserHistory)
	mux.HandleFunc("/api/user/listen", handleUserListen)
	mux.HandleFunc("/api/user/follow", handleUserFollow)
	mux.HandleFunc("/api/youth/vip", handleDailyVIP)

	// ===== 云盘 =====
	mux.HandleFunc("/api/cloud/list", handleCloudList)
	mux.HandleFunc("/api/cloud/delete", handleCloudDelete)
	mux.HandleFunc("/api/cloud/url", handleCloudURL)
	mux.HandleFunc("/api/cloud/match", handleCloudMatch)
	mux.HandleFunc("/api/cloud/upload", handleCloudUpload)

	// ===== 静态文件（前端 SPA） =====
	mux.HandleFunc("/", handleStatic)

	// 中间件链：CORS → 认证解析 → 缓存
	var handler http.Handler = mux
	handler = middleware.CORS(handler)
	handler = middleware.Auth(handler)
	handler = middleware.Cache(handler)

	// 端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	fmt.Printf("VibeMusic server listening on %s (lite=%v)\n", addr, lite)
	if err := http.ListenAndServe(addr, handler); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

// ===== 静态文件服务（SPA 回退到 index.html） =====

func handleStatic(w http.ResponseWriter, r *http.Request) {
	sub, err := fs.Sub(distFS, "web/dist")
	if err != nil {
		writeError(w, 500, "静态资源加载失败")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" || strings.HasSuffix(path, "/") {
		path = "index.html"
	}

	// 若文件不存在（前端路由），回退到 index.html（SPA）
	file, err := sub.Open(path)
	if err != nil {
		path = "index.html"
		file, err = sub.Open(path)
		if err != nil {
			writeError(w, 404, "not found")
			return
		}
	}
	defer file.Close()

	// 设置 Content-Type
	if strings.HasSuffix(path, ".svg") {
		w.Header().Set("Content-Type", "image/svg+xml")
	} else if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	} else if strings.HasSuffix(path, ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else {
		ct := http.DetectContentType(mustReadPrefix(file))
		if ct == "text/plain; charset=utf-8" {
			ct = "application/octet-stream"
		}
		w.Header().Set("Content-Type", ct)
	}

	// embed 文件支持 Seek，可直接用于 ServeContent
	if seeker, ok := file.(io.ReadSeeker); ok {
		http.ServeContent(w, r, path, embedModTime(), seeker)
		return
	}

	// 兜底：不支持 Seek 时直接读全量输出
	data, _ := io.ReadAll(file)
	w.Write(data)
}

// 读取文件前缀用于检测 Content-Type（http.DetectContentType 需要 []byte）
func mustReadPrefix(file fs.File) []byte {
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	return buf[:n]
}

// embed 文件无真实修改时间，返回零值时间（避免误导）
func embedModTime() time.Time {
	return time.Time{}
}

// ===== 通用辅助 =====

func getQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// 数字 id 转字符串，避免 float64 科学计数法（如 7.25779799e+08）
func userIDString(v interface{}) string {
	switch t := v.(type) {
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case string:
		return t
	default:
		return fmt.Sprintf("%v", v)
	}
}

// 从 context 或 header 获取登录凭证
func getCredentials(r *http.Request) *api.Credentials {
	if creds, ok := r.Context().Value(middleware.ContextCredentials).(*api.Credentials); ok && creds != nil {
		return creds
	}
	return api.ParseCredentials(r.Header.Get("Authorization"))
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": 0,
		"error":  msg,
	})
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, map[string]interface{}{
		"status": 1,
		"data":   data,
	})
}

// 透传酷狗原始响应（resp.Data 已含 status/data/error）
func writeUpstream(w http.ResponseWriter, resp *api.APIResponse) {
	if resp == nil || resp.Error != nil {
		writeError(w, 502, "上游请求失败")
		return
	}
	if resp.Data == nil {
		// 非 JSON 响应（如 mobilecdn 404 页面）或空响应，提示接口不可用
		writeError(w, 502, "接口暂不可用，请稍后重试")
		return
	}
	writeJSON(w, resp.Data)
}

// ===== 设备 =====

func handleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	device, err := api.RegisterDevice()
	if err != nil {
		writeError(w, 502, err.Error())
		return
	}
	writeSuccess(w, device)
}

// ===== 搜索 =====

func handleSearch(w http.ResponseWriter, r *http.Request) {
	keyword := getQueryParam(r, "keyword")
	if keyword == "" {
		writeError(w, 400, "keyword 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 20
	}

	resp := api.SearchSongs(keyword, page, pagesize, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseSearchSongs(resp))
}

func handleSearchSuggest(w http.ResponseWriter, r *http.Request) {
	keyword := getQueryParam(r, "keyword")
	if keyword == "" {
		writeError(w, 400, "keyword 参数必填")
		return
	}
	resp := api.SearchSuggest(keyword, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseSearchSuggest(resp))
}

func handleSearchHot(w http.ResponseWriter, r *http.Request) {
	resp := api.SearchHot(getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseSearchHot(resp))
}

func handleSearchComplex(w http.ResponseWriter, r *http.Request) {
	keyword := getQueryParam(r, "keyword")
	if keyword == "" {
		writeError(w, 400, "keyword 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 10
	}
	writeUpstream(w, api.SearchComplex(keyword, page, pagesize, getCredentials(r)))
}

func handleSearchByType(w http.ResponseWriter, r *http.Request, typ string) {
	keyword := getQueryParam(r, "keyword")
	if keyword == "" {
		writeError(w, 400, "keyword 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 20
	}
	writeUpstream(w, api.SearchByType(typ, keyword, page, pagesize, getCredentials(r)))
}

func handleSearchLyric(w http.ResponseWriter, r *http.Request) {
	keyword := getQueryParam(r, "keyword")
	if keyword == "" {
		writeError(w, 400, "keyword 参数必填")
		return
	}
	artist := getQueryParam(r, "artist")
	duration := getQueryParam(r, "duration")
	album := getQueryParam(r, "album")
	writeUpstream(w, api.SearchLyric(keyword, artist, duration, album, getCredentials(r)))
}

func handleSearchDefault(w http.ResponseWriter, r *http.Request) {
	writeUpstream(w, api.GetSearchDefault(getCredentials(r)))
}

// ===== 歌曲 & 歌词 =====

func handleSongURL(w http.ResponseWriter, r *http.Request) {
	hash := getQueryParam(r, "hash")
	albumID := getQueryParam(r, "album_id")
	bitrate, _ := strconv.Atoi(getQueryParam(r, "bitrate"))
	quality := getQueryParam(r, "quality")

	if hash == "" {
		writeError(w, 400, "hash 参数必填")
		return
	}

	creds := getCredentials(r)

	// 逐级试完取最优：部分歌曲 v5/url 请求高音质时可能返回低音质源，
	// 因此不能拿到第一个结果就返回，必须试完所有等级取 QualityRank 最高的。
	vipRestricted := false
	var bestResult *api.SongURL
	for _, q := range qualityFallbackChain(quality) {
		resp := api.GetSongURL(hash, albumID, bitrate, q, creds)
		if resp.Error != nil || resp.Data == nil {
			continue
		}
		if isVipRestricted(resp.Data) {
			vipRestricted = true
			break
		}
		result := api.ParseSongURL(resp)
		if result != nil {
			if result.BitRate == 0 {
				result.Quality = q
			}
			if bestResult == nil || api.QualityRank(result.Quality) > api.QualityRank(bestResult.Quality) {
				bestResult = result
			}
		}
	}

	if bestResult != nil {
		writeSuccess(w, bestResult)
		return
	}

	// VIP/版权歌曲：用 priv_url 解锁
	if vipRestricted {
		vipToken := getQueryParam(r, "vip_token")
		vipType := getQueryParam(r, "vip_type")
		privResp := api.GetPrivURL(hash, creds, vipToken, vipType)
		if privResult := api.ParsePrivURL(privResp, quality); privResult != nil {
			writeSuccess(w, privResult)
			return
		}
		writeJSON(w, map[string]interface{}{
			"status": 2,
			"error":  "该歌曲为 VIP/版权歌曲，暂无法播放",
		})
		return
	}
	writeError(w, 404, "无法获取播放地址")
}

// 获取歌曲可用音质列表
func handleSongQualities(w http.ResponseWriter, r *http.Request) {
	hash := getQueryParam(r, "hash")
	if hash == "" {
		writeError(w, 400, "hash 参数必填")
		return
	}
	resp := api.GetPrivilegeLite([]string{hash}, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 500, resp.Error.Error())
		return
	}
	qualities := api.ParsePrivilegeQualities(resp)
	if qualities == nil {
		writeError(w, 502, "获取音质信息失败")
		return
	}
	writeSuccess(w, qualities)
}

// circularQualityChain 返回以请求音质为起点、先标准降级链向下的所有等级，
// 最后补充请求音质之上的更高等级（v5/url 可能交叉映射高低音质标签）。
func circularQualityChain(quality string) []string {
	order := []string{"high", "flac", "320", "128"}
	start := len(order)
	for i, q := range order {
		if q == quality {
			start = i
			break
		}
	}
	if start == len(order) {
		chain := make([]string, 0, 2)
		if quality != "" {
			chain = append(chain, quality)
		}
		return append(chain, "128")
	}
	chain := make([]string, 0, len(order))
	// 先从请求位置向下降级（标准链）
	chain = append(chain, order[start:]...)
	// 再补充请求位置以上的更高等级（向上探测，处理 v5/url 标签交叉映射）
	for i := start - 1; i >= 0; i-- {
		chain = append(chain, order[i])
	}
	return chain
}

// 音质降级链：请求音质在链中越靠前越好，缺失则取下一级。
// 请求音质不在链中时返回 [请求音质, 128]，确保至少尝试请求本身。
func qualityFallbackChain(quality string) []string {
	order := []string{"high", "flac", "320", "128"}
	start := len(order)
	for i, q := range order {
		if q == quality {
			start = i
			break
		}
	}
	if start == len(order) {
		// 请求音质不在链中（如 viper_* 或空）：先试请求音质，再兜底 128
		chain := make([]string, 0, 2)
		if quality != "" {
			chain = append(chain, quality)
		}
		return append(chain, "128")
	}
	return order[start:]
}

// 判断 v5/url 响应是否因 VIP/版权限制（status:2）导致无播放地址
func isVipRestricted(data map[string]interface{}) bool {
	if data == nil {
		return false
	}
	st, ok := data["status"]
	if !ok {
		return false
	}
	switch v := st.(type) {
	case float64:
		return int(v) == 2
	case int:
		return v == 2
	case string:
		return v == "2"
	}
	return false
}

func handlePrivilegeLite(w http.ResponseWriter, r *http.Request) {
	songIDs := getQueryParam(r, "song_id")
	if songIDs == "" {
		writeError(w, 400, "song_id 参数必填")
		return
	}
	ids := strings.Split(songIDs, ",")
	resp := api.GetPrivilegeLite(ids, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 500, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParsePrivilegeLite(resp))
}

func handleSongDetail(w http.ResponseWriter, r *http.Request) {
	hash := getQueryParam(r, "hash")
	if hash == "" {
		writeError(w, 400, "hash 参数必填")
		return
	}

	resp := api.GetSongDetail(hash, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	info := api.ParseSongDetail(resp)
	if info == nil {
		writeError(w, 502, "歌曲详情获取失败")
		return
	}
	writeSuccess(w, info)
}

func handleSongClimax(w http.ResponseWriter, r *http.Request) {
	hashStr := getQueryParam(r, "hash")
	if hashStr == "" {
		writeError(w, 400, "hash 参数必填")
		return
	}
	hashes := strings.Split(hashStr, ",")
	writeUpstream(w, api.GetSongClimax(hashes))
}

func handleLyric(w http.ResponseWriter, r *http.Request) {
	hash := getQueryParam(r, "hash")
	if hash == "" {
		writeError(w, 400, "hash 参数必填")
		return
	}
	albumID := getQueryParam(r, "album_id")
	timelength := getQueryParam(r, "timelength")

	resp := api.GetLyric(hash, albumID, timelength, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseLyric(resp))
}

// ===== 图片代理 =====

func handleImageProxy(w http.ResponseWriter, r *http.Request) {
	urlStr := getQueryParam(r, "url")
	size := getQueryParam(r, "size")

	var (
		data []byte
		ct   string
		err  error
	)
	switch {
	case urlStr != "":
		data, ct, err = api.ProxyImage(urlStr, size)
	case getQueryParam(r, "album_id") != "":
		data, ct, err = api.ProxyAlbumImage(getQueryParam(r, "album_id"), size)
	case getQueryParam(r, "artist_id") != "":
		data, ct, err = api.ProxyArtistImage(getQueryParam(r, "artist_id"), size)
	case getQueryParam(r, "playlist_id") != "":
		data, ct, err = api.ProxyPlaylistImage(getQueryParam(r, "playlist_id"), size)
	default:
		writeError(w, 400, "缺少图片参数")
		return
	}

	if err != nil {
		writeError(w, 502, err.Error())
		return
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

// ===== 推荐 & 歌单 =====

func handleEverydayRecommend(w http.ResponseWriter, r *http.Request) {
	writeUpstream(w, api.GetEverydayRecommend(getCredentials(r)))
}

func handleRecommend(w http.ResponseWriter, r *http.Request) {
	writeUpstream(w, api.GetRecommend(getCredentials(r)))
}

func handleRecommendPlaylist(w http.ResponseWriter, r *http.Request) {
	resp := api.GetRecommendPlaylist(getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseRecommendPlaylist(resp))
}

func handleTopPlaylist(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 20
	}
	categoryID := getQueryParam(r, "category_id")
	sort := getQueryParam(r, "sort")

	resp := api.GetTopPlaylist(page, pagesize, categoryID, sort, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	// special_recommend 接口返回 data.special_list，解析为歌单数组
	writeSuccess(w, api.ParseRecommendPlaylist(resp))
}

func handlePlaylistDetail(w http.ResponseWriter, r *http.Request) {
	id := getQueryParam(r, "id")
	if id == "" {
		writeError(w, 400, "id 参数必填")
		return
	}
	resp := api.GetPlaylistDetail(id, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParsePlaylistDetail(resp))
}

func handlePlaylistSongs(w http.ResponseWriter, r *http.Request) {
	id := getQueryParam(r, "id")
	if id == "" {
		writeError(w, 400, "id 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 20
	}
	resp := api.GetPlaylistSongs(id, page, pagesize, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParsePlaylistSongs(resp))
}

func handlePlaylistCategories(w http.ResponseWriter, r *http.Request) {
	writeUpstream(w, api.GetPlaylistCategories(getCredentials(r)))
}

func handlePlaylistTags(w http.ResponseWriter, r *http.Request) {
	writeUpstream(w, api.GetPlaylistTags(getCredentials(r)))
}

// ===== 歌单管理（需登录） =====

func handlePlaylistCreate(w http.ResponseWriter, r *http.Request) {
	name := getQueryParam(r, "name")
	if name == "" {
		writeError(w, 400, "name 参数必填")
		return
	}
	isPri := getQueryParam(r, "is_pri") == "1"
	creds := getCredentials(r)
	if creds.Token == "" {
		writeError(w, 401, "请先登录")
		return
	}
	resp := api.CreatePlaylist(name, isPri, creds)
	if resp.Error != nil {
		writeError(w, 500, resp.Error.Error())
		return
	}
	if resp.Data == nil {
		writeError(w, 502, "创建失败，请稍后重试")
		return
	}
	writeJSON(w, resp.Data)
}

func handlePlaylistDelete(w http.ResponseWriter, r *http.Request) {
	listid := getQueryParam(r, "listid")
	if listid == "" {
		writeError(w, 400, "listid 参数必填")
		return
	}
	creds := getCredentials(r)
	if creds.Token == "" {
		writeError(w, 401, "请先登录")
		return
	}
	resp, aesKey, err := api.DeletePlaylist(listid, creds)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	// delete_list 响应为 AES 加密，解密后返回
	result, err := api.ParseDeletePlaylistResp(resp, aesKey)
	if err == nil {
		writeJSON(w, result)
		return
	}
	// 解密失败则透传原始数据
	if resp.Error != nil {
		writeError(w, 500, resp.Error.Error())
		return
	}
	writeJSON(w, resp.Data)
}

func handlePlaylistTracksAdd(w http.ResponseWriter, r *http.Request) {
	listid := getQueryParam(r, "listid")
	if listid == "" {
		writeError(w, 400, "listid 参数必填")
		return
	}
	creds := getCredentials(r)
	if creds.Token == "" {
		writeError(w, 401, "请先登录")
		return
	}

	// 解析歌曲列表 [{hash,name,album_id,mixsongid}]
	var tracks []api.Song
	body, _ := io.ReadAll(r.Body)
	if len(body) > 0 {
		_ = json.Unmarshal(body, &tracks)
	}
	if len(tracks) == 0 {
		writeError(w, 400, "缺少歌曲数据")
		return
	}

	resp := api.AddPlaylistTracks(listid, tracks, creds)
	if resp.Error != nil {
		writeError(w, 500, resp.Error.Error())
		return
	}
	writeJSON(w, resp.Data)
}

func handlePlaylistTracksDel(w http.ResponseWriter, r *http.Request) {
	listid := getQueryParam(r, "listid")
	if listid == "" {
		writeError(w, 400, "listid 参数必填")
		return
	}
	creds := getCredentials(r)
	if creds.Token == "" {
		writeError(w, 401, "请先登录")
		return
	}

	body, _ := io.ReadAll(r.Body)
	var req struct {
		FileIDs []string `json:"file_ids"`
	}
	_ = json.Unmarshal(body, &req)
	if len(req.FileIDs) == 0 {
		writeError(w, 400, "缺少 file_ids")
		return
	}

	resp := api.DeletePlaylistTracks(listid, req.FileIDs, creds)
	if resp.Error != nil {
		writeError(w, 500, resp.Error.Error())
		return
	}
	writeJSON(w, resp.Data)
}

// ===== 评论 =====

func handleSongComments(w http.ResponseWriter, r *http.Request) {
	mixsongid := getQueryParam(r, "mixsongid")
	if mixsongid == "" {
		writeError(w, 400, "mixsongid 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 30
	}
	writeUpstream(w, api.GetSongComments(mixsongid, page, pagesize, getCredentials(r)))
}

func handlePlaylistComments(w http.ResponseWriter, r *http.Request) {
	id := getQueryParam(r, "id")
	if id == "" {
		writeError(w, 400, "id 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 30
	}
	writeUpstream(w, api.GetPlaylistComments(id, page, pagesize, getCredentials(r)))
}

func handleAlbumComments(w http.ResponseWriter, r *http.Request) {
	id := getQueryParam(r, "id")
	if id == "" {
		writeError(w, 400, "id 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 30
	}
	writeUpstream(w, api.GetAlbumComments(id, page, pagesize, getCredentials(r)))
}

func handleCommentCount(w http.ResponseWriter, r *http.Request) {
	hash := getQueryParam(r, "hash")
	specialID := getQueryParam(r, "special_id")
	writeUpstream(w, api.GetCommentCount(hash, specialID, getCredentials(r)))
}

func handleCommentFloor(w http.ResponseWriter, r *http.Request) {
	resourceType := getQueryParam(r, "resource_type")
	code := getQueryParam(r, "code")
	mixsongid := getQueryParam(r, "mixsongid")
	specialID := getQueryParam(r, "special_id")
	tid := getQueryParam(r, "tid")
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 30
	}
	writeUpstream(w, api.GetCommentFloor(resourceType, code, mixsongid, specialID, tid, page, pagesize, getCredentials(r)))
}

func handleCommentHotWords(w http.ResponseWriter, r *http.Request) {
	mixsongid := getQueryParam(r, "mixsongid")
	if mixsongid == "" {
		writeError(w, 400, "mixsongid 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 30
	}
	writeUpstream(w, api.GetCommentHotWords(mixsongid, page, pagesize, getCredentials(r)))
}

func handleCommentSend(w http.ResponseWriter, r *http.Request) {
	creds := getCredentials(r)
	if creds.Token == "" {
		writeError(w, 401, "请先登录")
		return
	}
	resourceType := getQueryParam(r, "resource_type")
	mixsongid := getQueryParam(r, "mixsongid")
	childrenid := getQueryParam(r, "childrenid")
	content := getQueryParam(r, "content")
	code := getQueryParam(r, "code")

	if content == "" {
		writeError(w, 400, "评论内容不能为空")
		return
	}
	resp := api.SendComment(resourceType, code, mixsongid, childrenid, content, creds)
	if resp.Error != nil {
		writeError(w, 500, resp.Error.Error())
		return
	}
	writeJSON(w, resp.Data)
}

// ===== 排行榜 & 新歌新专辑 =====

func handleRankList(w http.ResponseWriter, r *http.Request) {
	writeUpstream(w, api.GetRankList(getCredentials(r)))
}

func handleRankAudio(w http.ResponseWriter, r *http.Request) {
	rankid := getQueryParam(r, "rankid")
	if rankid == "" {
		writeError(w, 400, "rankid 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 30
	}
	writeUpstream(w, api.GetRankAudio(rankid, page, pagesize, getCredentials(r)))
}

func handleTopSong(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 30
	}
	resp := api.GetTopSong(page, pagesize, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseTopSongs(resp))
}

func handleTopAlbum(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 30
	}
	writeUpstream(w, api.GetTopAlbum(page, pagesize, getCredentials(r)))
}

func handleTopCard(w http.ResponseWriter, r *http.Request) {
	cardID := getQueryParam(r, "card_id")
	if cardID == "" {
		cardID = "1"
	}
	writeUpstream(w, api.GetTopCard(cardID, getCredentials(r)))
}

func handleFMSongs(w http.ResponseWriter, r *http.Request) {
	action := getQueryParam(r, "action")
	if action == "" {
		action = "play"
	}
	songPoolID := getQueryParam(r, "song_pool_id")
	remainSongCnt, _ := strconv.Atoi(getQueryParam(r, "remain_songcnt"))

	writeUpstream(w, api.GetPersonalFM(action, songPoolID, remainSongCnt, getCredentials(r)))
}

// ===== 专辑 & 歌手 =====

func handleAlbumDetail(w http.ResponseWriter, r *http.Request) {
	albumID := getQueryParam(r, "album_id")
	if albumID == "" {
		writeError(w, 400, "album_id 参数必填")
		return
	}
	resp := api.GetAlbumDetail(albumID, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseAlbumDetail(resp))
}

func handleAlbumSongs(w http.ResponseWriter, r *http.Request) {
	albumID := getQueryParam(r, "album_id")
	if albumID == "" {
		writeError(w, 400, "album_id 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 20
	}
	resp := api.GetAlbumSongs(albumID, page, pagesize, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseAlbumSongs(resp))
}

func handleArtistDetail(w http.ResponseWriter, r *http.Request) {
	singerID := getQueryParam(r, "singer_id")
	if singerID == "" {
		writeError(w, 400, "singer_id 参数必填")
		return
	}
	resp := api.GetArtistDetail(singerID, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseArtistDetail(resp))
}

func handleArtistAudios(w http.ResponseWriter, r *http.Request) {
	singerID := getQueryParam(r, "singer_id")
	if singerID == "" {
		writeError(w, 400, "singer_id 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 20
	}
	resp := api.GetArtistAudios(singerID, page, pagesize, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseArtistAudios(resp))
}

func handleArtistHot(w http.ResponseWriter, r *http.Request) {
	singerID := getQueryParam(r, "singer_id")
	if singerID == "" {
		writeError(w, 400, "singer_id 参数必填")
		return
	}
	resp := api.GetArtistHot(singerID, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseArtistAudios(resp))
}

func handleArtistAlbums(w http.ResponseWriter, r *http.Request) {
	singerID := getQueryParam(r, "singer_id")
	if singerID == "" {
		writeError(w, 400, "singer_id 参数必填")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 20
	}
	resp := api.GetArtistAlbums(singerID, page, pagesize, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseArtistAlbums(resp))
}

func handleArtistFollow(w http.ResponseWriter, r *http.Request) {
	singerID := getQueryParam(r, "singer_id")
	if singerID == "" {
		writeError(w, 400, "singer_id 参数必填")
		return
	}
	follow := getQueryParam(r, "follow") != "0"
	creds := getCredentials(r)
	if creds.Token == "" {
		writeError(w, 401, "请先登录")
		return
	}
	resp := api.FollowSinger(singerID, follow, creds)
	if resp.Error != nil {
		writeError(w, 500, resp.Error.Error())
		return
	}
	if resp.Data == nil {
		writeError(w, 502, "操作失败，请稍后重试")
		return
	}
	writeJSON(w, resp.Data)
}

func handleSingerList(w http.ResponseWriter, r *http.Request) {
	typ, _ := strconv.Atoi(getQueryParam(r, "type"))
	sextype, _ := strconv.Atoi(getQueryParam(r, "sextype"))
	musician, _ := strconv.Atoi(getQueryParam(r, "musician"))
	hotsize, _ := strconv.Atoi(getQueryParam(r, "hotsize"))
	writeUpstream(w, api.GetSingerList(typ, sextype, musician, hotsize, getCredentials(r)))
}

// ===== 登录 =====

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &req)
	if req.Username == "" || req.Password == "" {
		writeError(w, 400, "缺少账号或密码")
		return
	}

	resp := api.Login(req.Username, req.Password, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	result := api.ParseLogin(resp)
	if result == nil {
		msg := "登录失败，请检查账号密码"
		if resp.Data != nil {
			if e, ok := resp.Data["error"]; ok {
				msg = fmt.Sprintf("%v", e)
			} else if e, ok := resp.Data["errmsg"]; ok {
				msg = fmt.Sprintf("%v", e)
			}
		}
		writeError(w, 401, msg)
		return
	}
	writeSuccess(w, result)
}

func handleQRCreate(w http.ResponseWriter, r *http.Request) {
	writeUpstream(w, api.GetQRKey(getCredentials(r)))
}

func handleQRGet(w http.ResponseWriter, r *http.Request) {
	key := getQueryParam(r, "key")
	if key == "" {
		writeError(w, 400, "key 参数必填")
		return
	}
	writeUpstream(w, api.GetQRImage(key, getCredentials(r)))
}

func handleQRCheck(w http.ResponseWriter, r *http.Request) {
	key := getQueryParam(r, "key")
	if key == "" {
		writeError(w, 400, "key 参数必填")
		return
	}
	resp := api.CheckQRStatus(key, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseQRCheck(resp))
}

func handleLoginCellphone(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mobile   string `json:"mobile"`
		Code     string `json:"code"`
		AreaCode string `json:"area_code"`
	}
	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &req)
	if req.Mobile == "" || req.Code == "" {
		writeError(w, 400, "缺少手机号或验证码")
		return
	}
	if req.AreaCode == "" {
		req.AreaCode = "86"
	}

	resp := api.LoginCellphone(req.Mobile, req.Code, req.AreaCode, "", getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	result := api.ParseLogin(resp)

	// 手机号绑定多账号：上游返回 status=0 + info_list，自动选第一个账号重新登录
	if result == nil && resp.Data != nil && resp.Data["data"] != nil {
		if dataMap, ok := resp.Data["data"].(map[string]interface{}); ok {
			if infoList, ok := dataMap["info_list"].([]interface{}); ok && len(infoList) > 0 {
				if first, ok := infoList[0].(map[string]interface{}); ok {
					if uid, ok := first["userid"]; ok {
						selUserID := userIDString(uid)
						resp2 := api.LoginCellphone(req.Mobile, req.Code, req.AreaCode, selUserID, getCredentials(r))
						result = api.ParseLogin(resp2)
						if result != nil {
							writeSuccess(w, result)
							return
		}
					}
				}
			}
		}
	}
	if result == nil {
		msg := "登录失败，请检查验证码"
		if resp.Data != nil {
			if e, ok := resp.Data["error"]; ok {
				msg = fmt.Sprintf("%v", e)
			} else if e, ok := resp.Data["errmsg"]; ok {
				msg = fmt.Sprintf("%v", e)
			}
		}
		writeError(w, 401, msg)
		return
	}
	writeSuccess(w, result)
}

func handleCaptchaSent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Mobile   string `json:"mobile"`
		AreaCode string `json:"area_code"`
	}
	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &req)
	if req.Mobile == "" {
		writeError(w, 400, "缺少手机号")
		return
	}
	if req.AreaCode == "" {
		req.AreaCode = "86"
	}
	writeUpstream(w, api.SendCaptcha(req.Mobile, req.AreaCode, getCredentials(r)))
}

// ===== 用户 & VIP =====

func handleUserDetail(w http.ResponseWriter, r *http.Request) {
	resp := api.GetUserDetail(getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseUserDetail(resp))
}

func handleUserPlaylist(w http.ResponseWriter, r *http.Request) {
	userid := getQueryParam(r, "userid")
	token := getQueryParam(r, "token")
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 20
	}
	resp := api.GetUserPlaylist(userid, token, page, pagesize, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseUserPlaylist(resp))
}

func handleUserFavorite(w http.ResponseWriter, r *http.Request) {
	userid := getQueryParam(r, "userid")
	token := getQueryParam(r, "token")
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 20
	}
	resp := api.GetUserFavorite(userid, token, page, pagesize, getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseUserPlaylist(resp))
}

func handleUserVIPDetail(w http.ResponseWriter, r *http.Request) {
	resp := api.GetUserVIPDetail(getCredentials(r))
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeSuccess(w, api.ParseVIPDetail(resp))
}

func handleDailyVIP(w http.ResponseWriter, r *http.Request) {
	creds := getCredentials(r)
	if creds.Token == "" {
		writeError(w, 401, "请先登录")
		return
	}
	resp := api.GetDailyVIP(creds)
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	ok, msg := api.ParseDailyVIP(resp)
	if ok {
		writeSuccess(w, map[string]interface{}{"ok": true})
		return
	}
	if msg == "" {
		msg = "今日已领取或领取失败"
	}
	writeJSON(w, map[string]interface{}{"status": 0, "error": msg})
}

func handleUserHistory(w http.ResponseWriter, r *http.Request) {
	creds := getCredentials(r)
	if creds.Token == "" || creds.UserID == "" {
		writeError(w, 401, "请先登录")
		return
	}
	bp := getQueryParam(r, "bp")
	writeUpstream(w, api.GetUserHistory(creds.UserID, creds.Token, bp))
}

func handleUserListen(w http.ResponseWriter, r *http.Request) {
	creds := getCredentials(r)
	if creds.Token == "" || creds.UserID == "" {
		writeError(w, 401, "请先登录")
		return
	}
	listType, _ := strconv.Atoi(getQueryParam(r, "type"))
	writeUpstream(w, api.GetUserListen(creds.UserID, creds.Token, listType))
}

func handleUserFollow(w http.ResponseWriter, r *http.Request) {
	creds := getCredentials(r)
	if creds.Token == "" || creds.UserID == "" {
		writeError(w, 401, "请先登录")
		return
	}
	resp := api.GetUserFollowList(creds)
	if resp.Error != nil {
		writeError(w, 502, resp.Error.Error())
		return
	}
	writeJSON(w, resp.Data)
}

// ===== 云盘 =====

func handleCloudList(w http.ResponseWriter, r *http.Request) {
	creds := getCredentials(r)
	if creds.Token == "" || creds.UserID == "" {
		writeError(w, 401, "请先登录")
		return
	}
	page, _ := strconv.Atoi(getQueryParam(r, "page"))
	if page < 1 {
		page = 1
	}
	pagesize, _ := strconv.Atoi(getQueryParam(r, "pagesize"))
	if pagesize < 1 {
		pagesize = 30
	}
	writeUpstream(w, api.GetCloudList(creds.UserID, creds.Token, page, pagesize))
}

func handleCloudDelete(w http.ResponseWriter, r *http.Request) {
	creds := getCredentials(r)
	if creds.Token == "" || creds.UserID == "" {
		writeError(w, 401, "请先登录")
		return
	}
	idsStr := getQueryParam(r, "ids")
	if idsStr == "" {
		writeError(w, 400, "ids 参数必填（逗号分隔的 kv_id 列表）")
		return
	}
	parts := strings.Split(idsStr, ",")
	kvIDs := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			writeError(w, 400, "无效的 id: "+p)
			return
		}
		kvIDs = append(kvIDs, id)
	}

	var albumAudioIDs []int64
	if aaStr := getQueryParam(r, "album_audio_ids"); aaStr != "" {
		for _, p := range strings.Split(aaStr, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			id, _ := strconv.ParseInt(p, 10, 64)
			albumAudioIDs = append(albumAudioIDs, id)
		}
	}

	writeUpstream(w, api.DeleteCloudFile(creds.UserID, creds.Token, kvIDs, albumAudioIDs))
}

func handleCloudURL(w http.ResponseWriter, r *http.Request) {
	hash := getQueryParam(r, "hash")
	if hash == "" {
		writeError(w, 400, "hash 参数必填")
		return
	}
	albumAudioID, _ := strconv.ParseInt(getQueryParam(r, "album_audio_id"), 10, 64)
	audioID, _ := strconv.ParseInt(getQueryParam(r, "audio_id"), 10, 64)
	name := getQueryParam(r, "name")
	writeUpstream(w, api.GetCloudURL(hash, albumAudioID, audioID, name))
}

func handleCloudMatch(w http.ResponseWriter, r *http.Request) {
	hash := getQueryParam(r, "hash")
	if hash == "" {
		writeError(w, 400, "hash 参数必填")
		return
	}
	albumAudioID, _ := strconv.ParseInt(getQueryParam(r, "album_audio_id"), 10, 64)
	writeUpstream(w, api.GetCloudMatch(hash, albumAudioID))
}

func handleCloudUpload(w http.ResponseWriter, r *http.Request) {
	creds := getCredentials(r)
	if creds.Token == "" || creds.UserID == "" {
		writeError(w, 401, "请先登录")
		return
	}

	// 限制上传文件大小（50MB）
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeError(w, 400, "文件过大或格式错误: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "请上传文件: "+err.Error())
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		writeError(w, 500, "读取文件失败: "+err.Error())
		return
	}
	if len(fileData) == 0 {
		writeError(w, 400, "文件为空")
		return
	}

	// 从文件名提取扩展名
	filename := ""
	extendname := "mp3"
	if header != nil {
		filename = header.Filename
		// 去掉扩展名用于匹配
		if dot := strings.LastIndex(filename, "."); dot > 0 {
			extendname = filename[dot+1:]
			filename = filename[:dot] // 不含扩展名部分
		}
	}

	autoMatch := getQueryParam(r, "auto_match") != "0"
	resp := api.UploadCloudFile(fileData, filename, extendname, autoMatch, creds.UserID, creds.Token)
	if resp.Error != nil {
		log.Printf("[cloud_upload] 上传失败: %v (userid=%s)", resp.Error, creds.UserID)
	}
	writeUpstream(w, resp)
}
