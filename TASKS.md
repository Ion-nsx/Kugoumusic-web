# VibeMusic 功能对齐清单

对齐目标：`/root/MoeKoeMusic/api/module/*.js`（MakcRe 项目，168 个接口模块）。
`[x]`=后端已实现且前端已接入；`[~]`=后端有/部分、前端未接或未完成；`[ ]`=未实现。

## 基础 & 设备
- [x] 设备注册 `register_dev` → `/api/register/dev`（AES+RSA 设备指纹）
- [x] 图片代理 `images` → `/api/images`

## 搜索
- [x] 单曲搜索 `search` → `/api/search`（mobilecdn /v3/search/song）
- [x] 搜索建议 `search_suggest` → `/api/search/suggest`（searchtip.kugou.com）
- [x] 热搜 `search_hot` → `/api/search/hot`
- [x] 综合搜索 `search_complex` → `/api/search/complex`（后端聚合 song+album+author+special+mv）
- [x] 分类型搜索 `search(album/author/mv/special)` → `/api/search/{type}`（complexsearch.kugou.com）
- [x] 歌词搜索 `search_lyric` → `/api/search/lyric`（`lyrics.kugou.com/v1/search` 仅支持 hash 查询，后端自动降级为歌曲搜索）
- [x] 搜索默认词 `search_default` → `/api/search/default`（searchnofocus/v1/search_no_focus_word）
- [ ] 混合搜索 `search_mixed`（/v3/search/mixed 已失效，可跳过）

## 歌曲 & 播放
- [x] 播放地址 `song_url` / `audio` → `/api/song/url`（gateway v5/url + trackercdn + KeyHash，**登录带 Token 后 VIP 歌可播**；后端按 flac→320→128 降级，v5/url status:2 时降级 tracker priv_url 兜底）
- [x] 音质权限 `privilege_lite` → `/api/privilege/lite`（media.store.kugou.com）
- [x] 歌曲详情 `song_detail` → `/api/song/detail`（已迁移 kmr.service.kugou.com/v1/audio/audio，原 mobilecdn Access Deny）
- [x] 歌曲高潮 `song_climax` → `/api/song/climax`（expendablekmrcdn.kugou.com/v1/audio_climax/audio）
- [ ] 相关歌曲 `audio_related`、相似匹配 `audio_match`、伴奏 `audio_accompany_matching`、KTV `audio_ktv_total`
- [x] 歌词 `lyric` → `/api/lyric`（lyrics.kugou.com + KRC 解码）

## 歌单
- [x] 每日推荐 `everyday_recommend` → `/api/everyday/recommend`（everydayrec.service.kugou.com）
- [x] 个性推荐 `recommend_songs` → `/api/recommend`
- [x] 推荐歌单 `top_playlist` → `/api/top/playlist`
- [x] 歌单详情 `playlist_detail` → `/api/playlist/detail`（global_collection_id）
- [x] 歌单歌曲 `playlist_track_all` → `/api/playlist/song`
- [x] 歌单分类/标签 `playlist_tags` → `/api/playlist/tags`（pubsongs）
- [x] 创建歌单 `playlist_add` → `/api/playlist/create`（social.go，需登录）
- [x] 删除歌单 `playlist_del` → `/api/playlist/delete`（social.go，需登录）
- [x] 歌单加歌 `playlist_tracks_add` → `/api/playlist/tracks/add`（social.go，需登录）
- [x] 歌单删歌 `playlist_tracks_del` → `/api/playlist/tracks/del`（social.go，需登录）
- [ ] 相似歌单 `playlist_similar`、歌单效果 `playlist_effect`
- [ ] 翻唱榜 `sheet_*`（5 个模块）

## 排行榜 & 榜单
- [x] 排行榜列表 `rank_list` → `/api/rank/list`（ocean/v6）
- [x] 榜单歌曲 `rank_audio` → `/api/rank/audio`（openapi/kmr + kg-tid:369）
- [x] 新歌 `top_song` → `/api/top/song`（musicadservice）
- [x] 新专辑 `top_album` → `/api/top/album`（musicadservice）
- [x] 推荐卡片 `top_card` → `/api/top/card`（singlecardrec）
- [x] 排行榜页 /ranking：双列卡片 + 多选(最多6个) + 折叠展开 + 滚动分页 + localStorage 持久化，对齐参考设计

## 专辑 & 歌手
- [x] 专辑详情 `album_detail` → `/api/album/detail`
- [x] 专辑歌曲 `album_songs` → `/api/album/songs`
- [x] 歌手详情 `artist_detail` → `/api/artist/detail`
- [x] 歌手歌曲 `artist_audios` → `/api/artist/audios`
- [x] 歌手专辑 `artist_albums` → `/api/artist/album`
- [x] 歌手热门 `artist_hot` → `/api/artist/hot`
- [x] 歌手列表 `artist_lists` → `/api/artist/list`
- [x] 关注歌手 `artist_follow` → `/api/artist/follow`（需登录）
- [ ] 关注歌手新歌 `artist_follow_newsongs`、歌手荣誉 `artist_honour`

## 评论
- [x] 歌曲评论 `comment_music` → `/api/comment/song`
- [x] 歌单评论 `comment_playlist` → `/api/comment/playlist`
- [x] 专辑评论 `comment_album` → `/api/comment/album`
- [x] 评论数 `comment_count` → `/api/comment/count`
- [x] 评论楼层 `comment_floor` → `/api/comment/floor`（已修复 tid 参数缺失）
- [x] 评论热词 `comment_music_hotword` → `/api/comment/hotword`
- [x] 发送评论 → `/api/comment/send`（需登录）
- [x] CommentModal 分页导航（首页/上页/页码/下页/末页，最多100页）
- [x] 评论输入框（支持 Enter 发送）
- [x] 评论用户头像多字段兼容
- [ ] 评论分类 `comment_music_classify`

## 登录 & 账号
- [x] 登录手机验证码 `login_cellphone` + `captcha_sent` → `/api/login/cellphone`、`/api/captcha/sent`（多账号自动选第一个）
- [x] 扫码登录 `login_qr_create/check/key` → `/api/login/qr/*`（web 签名，**appid=3116 概念版**）
- [~] 账号密码 `login` → `/api/login`（v9/login_by_pwd，AES+RSA；**后端有，前端无入口**）
- [ ] 设备登录 `login_device`、踢线 `login_device_kick`、token `login_token`、openplat、微信 `login_wx_*`

## 用户 & VIP
- [x] 用户信息 `user_detail` → `/api/user/detail`
- [x] 用户歌单 `user_playlist` → `/api/user/playlist`
- [x] 用户收藏 `user_favorite` → `/api/user/favorite`
- [~] VIP 详情 `user_vip_detail` → `/api/user/vip/detail`（kugouvip/v1/get_union_vip，补齐参数，但概念版平台 VIP 检测不稳定）
- [~] 每日领 VIP `youth_vip` → `/api/youth/vip`（youth/v1/ad/play_report，需登录；前端无入口）
- [x] 我听（最近播放） `user_listen` → 本地 localStorage（vibe_history）
- [x] 听歌历史 `user_history` → `/api/user/history`（playhistory/v1/get_songs，需登录）
- [x] 关注列表 `user_follow` → `/api/user/follow`（relationuser.kugou.com/v4/follow_list，响应字段 lists/nickname/pic）
- [ ] 已购歌曲/专辑 `user_purchased_*`
- [ ] 上传历史 `playhistory_upload`

## 猜你喜欢
- [x] 猜你喜欢 `personal_fm` → `/api/fm/songs`（persnfm.service.kugou.com/v2/personal_recommend，2次调用合并去重≈10首）
- [ ] FM 歌曲 `fm_songs`、推荐 `fm_recommend`、分类 `fm_class`、背景 `fm_image`
- [ ] PC 电台 `pc_diantai`、曲库 `yueku`/`yueku_banner`/`yueku_fm`

## 云盘
- [x] 云盘列表 `user_cloud` → `/api/cloud/list`（mcloudservice + AES+RSA 加密，**自动批量匹配曲库补充封面**）
- [x] 云盘删除 `user_cloud_del` → `/api/cloud/delete`
- [x] 云盘播放 `user_cloud_url` → `/api/cloud/url`（bsstrackercdngz + signCloudKey）
- [x] 云盘上传 `user_cloud_upload` → `/api/cloud/upload`（5步流程：授权→初始化→分片→完成→add_files，上限50MB）
- [x] 曲库匹配 `user_cloud_match` → `/api/cloud/match`（上传前/列表查询时自动匹配酷狗曲库补齐元数据）
- [x] 前端 Cloud.vue：列表/播放/删除/分页/上传/歌名解析/封面匹配
- [ ] 云盘匹配增强 `user_cloud_match`（上传后更新已有文件的匹配数据，目前仅新增时匹配）

## 明确不做（用户要求）
- [ ] MV/视频：`video_*`、`kmr_audio_mv`、`artist_videos`、`user_video_*`
- [ ] 青少年频道 `youth_channel_*`、长音频 `longaudio_*`、场景 `scene_*`、主题 `theme_*`
- [ ] 音乐人 `brush`/`sidedt`/`get_model`/`get_mode_info`/`get_verify_info`、IP `ip*`

## 前端页面现状
- [x] Home / Search / Ranking / Me / Liked / Recent / Local / LyricView
- [x] Playlist / Album / Artist 详情页（沉浸式，BackBar 回退）
- [x] 侧栏 Sidebar + 顶栏 TopBar + 底部 PlayerBar 布局（BottomTab 已废弃）
- [x] LoginModal 登录（扫码/手机验证码；账号密码前端已删）
- [x] CommentModal 评论弹层（播放器、歌单、专辑入口）
- [x] 未登录引导：`requireLogin(reason)` + 后端 401 全局拦截弹登录框
- [x] 右上角设置：默认音质（128/320/flac/高解析Hi-Res）+ 深色模式切换，持久化 localStorage
- [x] 歌曲列表统一表头 + 7 列对齐（索引/时长右对齐，歌手左对齐）
- [x] 歌词展示：PlayerBar 迷你歌词行 + LyricView 自动滚动居中
- [x] LyricView 歌词页 UI 优化：封面模糊背景氛围 + 圆角方形封面 + 播放控制条（进度条拖拽/上下一首/播放暂停）+ 顶部返回按钮
- [x] 逐字歌词高亮：后端 `/api/lyric` 优先返回 KRC（带字符级时间戳），前端 `parseLyrics` 解析 `<start,dur>字` 标签，当前行逐字渲染（唱到的字上高亮色，未唱保持白色）；无字符标签时按字数均分行级时间兜底
- [x] 歌词高亮颜色自定义：LyricView 右下角调色板按钮弹出面板（8 预设色 + 自定义取色），持久化 localStorage `vibe_lyric_hl_color`
- [x] LyricView 交互增强：点击歌词行定位播放、A−/A+ 歌词字号调节（localStorage）、播放模式切换（列表/单曲/随机，图标与 PlayerBar 一致）、倍速循环 1x→1.25x→1.5x→2x（store 导出 playbackRate/setPlaybackRate，换源后 onLoadedMetadata 重新应用）
- [x] PlayerBar 进度条 UI 优化：磨砂玻璃轨道（backdrop-filter + 内阴影）+ 缓冲进度层（audio.buffered）+ 已播蓝渐变发光填充 + 毛玻璃圆点（hover/拖动放大）+ 拖动时暂停并显示目标时间 tooltip；音量条同步改为细长轨道（5px 高 110px 宽）
- [x] 猜你喜欢页面（列表展示 + 换一批 + 播放全部，更名为猜你喜欢）
- [x] 歌手列表页（`/api/artist/list` 后端已通，前端 SongSearch 用）
- [x] 听歌排行页已砍（`/api/user/listen` 数据陈旧不可用，参考项目也仅作推荐卡片用）
- [x] 音乐云盘页 /cloud（列表浏览/点击播放/删除确认/分页解析歌名，KMR 匹配失败时封面回退默认图标）
- [x] 搜索分页：单曲/专辑/歌手/歌单支持首页/上页/页码/下页/末页 + 跳页输入框
- [x] 搜索默认词 → 热搜榜下方推荐搜索标签
- [x] 专辑名/歌手名可点击跳转详情页
- [x] 歌曲高潮片段 → PlayerBar 进度条橙色区间标记
- [x] 鼠标指针美化 + 点击缩放/高亮反馈动效
- [x] 红心收藏/取消收藏（数字 listid+mixsongid=0+add_song 去掉 x-router）
- [x] 歌单歌曲名称去歌手前缀
- [x] 专辑页修复（template 包裹 v-if 链 + pagesize=50 + 歌曲封面回退专辑封面）
- [x] 播放栏进度条拖拽（拖拽时暂停，松手恢复）
- [x] 播放栏歌词溢出修复（text-overflow: ellipsis + max-width 160→280px）
- [x] player.js 新增 seekTo / setAudioEl 支持
- [x] setQuality() Bug 修复：先赋值再判断导致永远不换源，改为先存旧值
- [x] PlayerBar 音质选择补上 Hi-Res(high) 档（与 TopBar 设置对齐 4 档）
- [x] migrateLocalLiked 用 likeListNum（数字 listid）替代 likePlaylistId
- [x] 深色模式修复：App.vue watch(theme) immediate 覆盖 dark → 删除 + index.html 前置脚本；TopBar 改实色 var(--bg) 防 backdrop-filter 透白；Home banner/圆点/骨架屏暗色适配；补 --text-4

## 遗留技术债
- [x] `parseOpenAPISong`/`parseSongFromTrack` 封面/时长字段缺失 — 已补 `album_sizable_cover`/`timelength`/`deprecated.duration`
- [x] 歌手歌曲无封面 — 已批量查专辑 `sizable_cover` 补充
- [x] 歌单广场/搜索歌单字段映射混乱 — 已加 `normalizeCard` 统一 `id/cover/name/author`
- [x] 歌词搜索前端已完成接入（搜索页新增"歌词"Tab，后端自动将关键词转为歌曲搜索）
- [x] singer_id 解析缺失 — 已补 parseSongFromTrack/parseOpenAPISong/Daily
- [x] isLiked 模糊匹配 → hash/mixsongid/名称+歌手+专辑四重验证
- [x] add_song/delete_songs 使用数字 listid 而非 global_collection_id
- [x] setQuality() 切音质 Bug：质量值覆盖早于比较，导致 always true 永远不换源 — 已修复为先存旧值再比较
- [x] 深色模式全局失效：App.vue `watch(theme, immediate)` 在子组件之后执行覆盖 dark → 已删除 + index.html 前置脚本
- [x] 深色下顶栏泛白：backdrop-filter 透 html 默认白 → html/topbar 加 background var(--bg)
- [x] Home banner 白底/暗色圆点不可见 → 改 var(--surface) + 暗色 override
- [x] --text-4 浅色模式缺失 → :root 补全
- [x] 骨架屏硬编码灰 #E7E9ED → 改 var(--surface-2)
- [x] 我的歌单评论无数据：CommentModal 用 route.params.id 而非 playlist.id + 用户歌单需要创建者 global_collection_id（`comment_id` 字段）— 已修复
- [x] 评论分页虚高：API combine_count 包含回复数导致翻页白屏，改为以实际返回条数修正 — 已修复
