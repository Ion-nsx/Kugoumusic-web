package api

// 统一的响应结构
// 前端约定：{ status: 1, data: {...} } 或 { status: 0, error: "..." }

// 歌曲信息（播放核心）
type Song struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Album     string   `json:"album"`
	AlbumID   string   `json:"album_id"`
	Singer    string   `json:"singer"`
	SingerID  string   `json:"singer_id"`
	MixSongID string   `json:"mixsongid,omitempty"` // 歌曲唯一ID（评论/添加歌单用）
	FileID    string   `json:"file_id,omitempty"`   // 歌单内歌曲ID（删除歌单歌曲用）
	Time      int64    `json:"time"`
	Duration  int64    `json:"duration"`
	Img       string   `json:"img"`
	HQImg     string   `json:"hq_img"`
	SQImg     string   `json:"sq_img"`
	PayType   string   `json:"pay_type"`
	Pay       int      `json:"pay"`
	Owned     bool     `json:"owned"`
	Extname   string   `json:"extname"`
	Type      string   `json:"type"`
	BitRate   int      `json:"bitrate"`
	Size      int64    `json:"size"`
	Playable  bool     `json:"playable"`
	Url       string   `json:"url"`
	Language  string   `json:"language"`
	Lyric     string   `json:"lyric"`
	Subtitle  string   `json:"subtitle"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
	Em        string   `json:"em"`
	RQContent string   `json:"rq_content"`
}

// 歌单信息
type Playlist struct {
	ID        string  `json:"id"`
	ListID    string  `json:"list_id,omitempty"`     // 数字 listid，add_song 需要
	CommentID string  `json:"comment_id,omitempty"`  // 评论用 childrenid（创建者视角的 global_collection_id）
	Name      string  `json:"name"`
	Cover     string  `json:"cover"`
	Count     int     `json:"count"`
	PlayCount int64   `json:"play_count"`
	Author    string  `json:"author"`
	Desc      string  `json:"desc"`
	Tag       []string `json:"tag,omitempty"`
	IsVIP     bool    `json:"is_vip"`
	Songs     []Song  `json:"songs,omitempty"`
}

// 专辑信息
type Album struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Singer   string `json:"singer"`
	Cover    string `json:"cover"`
	Count    int    `json:"count"`
	Time     string `json:"time"`
	Company  string `json:"company"`
	Language string `json:"language"`
	Songs    []Song `json:"songs,omitempty"`
}

// 歌手信息
type Artist struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Pic     string `json:"pic"`
	Count   int    `json:"count"`
	Gender  string `json:"gender"`
	Country string `json:"country"`
	Desc    string `json:"desc"`
	Songs   []Song `json:"songs,omitempty"`
}

// 搜索结果
type SearchResult struct {
	Total int         `json:"total"`
	Songs []Song      `json:"songs"`
	Albums []Album    `json:"albums,omitempty"`
	Artists []Artist  `json:"artists,omitempty"`
	Lists  []Playlist `json:"lists,omitempty"`
	Users  []interface{} `json:"users,omitempty"`
}

// 歌词信息
type LyricResult struct {
	Content string `json:"content"`
	Author  string `json:"author"`
	Title   string `json:"title"`
	Album   string `json:"album"`
	Time    string `json:"time"`
}

// 歌曲播放信息
type SongURL struct {
	URL      string   `json:"url"`
	Backup   []string `json:"backup,omitempty"`
	BitRate  int      `json:"bitrate"`
	Ext      string   `json:"ext"`
	Size     int64    `json:"size"`
	Duration int64    `json:"duration"`
	Hash     string   `json:"hash"`
	FileName string   `json:"file_name"`
	Quality  string   `json:"quality"`
}

// 音质权限信息
type Privilege struct {
	BitRate      int    `json:"bitrate"`
	Format       string `json:"format"`
	Length       string `json:"length"`
	CanPlay      bool   `json:"can_play"`
	CanDownload  bool   `json:"can_download"`
	CanShare     bool   `json:"can_share"`
	PayType      string `json:"pay_type"`
	Extname      string `json:"extname"`
	Duration     int64  `json:"duration"`
	Image        string `json:"image"`
}

// 用户信息
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Sex      string `json:"sex"`
	Level    int    `json:"level"`
	IsVIP    bool   `json:"is_vip"`
	VIPType  string `json:"vip_type"`
	Email    string `json:"email"`
}

// 登录结果
type LoginResult struct {
	Token   string `json:"token"`
	UserID  string `json:"user_id"`
	DFID    string `json:"dfid"`
	MID     string `json:"mid"`
	GUID    string `json:"guid"`
	Expire  int64  `json:"expire"`
	User    User   `json:"user"`
	VIPType string `json:"vip_type,omitempty"` // 概念版 VIP 播放 priv_url 需要
	VIPToken string `json:"vip_token,omitempty"`
}

// 排行榜
type Rank struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Cover string `json:"cover"`
	Songs []Song `json:"songs,omitempty"`
}

// 热搜榜
type HotWord struct {
	Word string `json:"word"`
	Rank int    `json:"rank"`
	Hot  int64  `json:"hot"`
}