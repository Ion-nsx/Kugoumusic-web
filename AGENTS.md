# AGENTS.md

酷狗音乐概念版克隆项目。Go 后端（本目录 `/root/X-music`, module `vibe`）+ Vue3/Vite 前端（`web/`）。

## 参考项目

- **MoeKoeMusic**（后端 + 前端）：`/root/MoeKoeMusic/`
  - API 模块（168 个接口，参数/域名/签名最权威来源）：`/root/MoeKoeMusic/api/module/*.js`
  - 前端页面/组件：`/root/MoeKoeMusic/src/`
  - API 配置常量（appid/clientver 等）：`/root/MoeKoeMusic/api/config.json`
  - 签名/加密工具：`/root/MoeKoeMusic/api/util.js`
- **KuGouMusicApi**（纯后端 API）：`/root/KuGouMusicApi/`，模块在 `module/` 目录
- **MoeKoeMusic API 文档**：https://kugoumusicapi-docs.4everland.app/
- **前端实现务必对照参考项目同功能页面**：先看参考项目的 `.vue` 文件里用的字段名/参数名，再写自己的同名功能
- **单接口调试**：`/tmp/opencode/regtest/*.go`（`cd /tmp/opencode/regtest && go run xxx.go`），验证新接口再迁移

## 版本控制（重要）

- 项目是 git 仓库（曾因无版本控制导致 main.go 被误覆盖丢失，已重建并基线提交）。**任何有意义的改动完成后记得 `git add -A && git commit`**，让每次变更可回滚。
- 需要临时实验/转换/调试时，把临时文件放 `/tmp/opencode`，**绝不直接覆盖 `main.go`、`api/*.go`、`web/src/*` 等源文件**；若确需改动，先 `cp 文件 /tmp/opencode/bak/` 留底。
- 检查未提交改动：`git status` / `git diff`。回滚误改：`git checkout -- <file>`。
- 构建产物（`vibe-server`、`web/dist`、`web/node_modules`）已 gitignore，不入库。

## 构建与运行

- 前端产物用 `//go:embed web/dist/*` 嵌入二进制。**改前端后必须按顺序执行：先 `cd web && npm run build`，再 `cd /root/X-music && go build -o vibe-server .`。顺序反了会导致二进制嵌入旧的 dist，前端改动不生效**（曾因此导致云盘页面编译后不可见）。
- 启动：`./vibe-server`，监听 `:8080`。环境变量 `PORT`。
- **服务重启坑**：`pkill -f vibe-server` 会让 bash 工具卡到超时，必须**单独一条命令**执行；然后用 `cd /root/X-music && setsid -f ./vibe-server >/tmp/vibe-server.log 2>&1` 后台启动，**curl 验证也要单独命令**，不可与启动串在同一条 bash。
- 编译验证：`cd /root/X-music && go build -o vibe-server .`。无 lint/typecheck/test 脚本。

## 平台与签名（api/request.go）

- 仅概念版(lite, appid=3116)：签名盐 `LnT6…`，signKey 盐 `185672…`，RSA 公钥 `publicLiteRSAKey`。
- 签名算法（lib-specific，不要猜）：
  - `signatureAndroidParams`＝`MD5(盐 + 排序key=value + body字节 + 盐)`；盐分平台：lite=`LnT6…`、std=`OIlw…`。
  - `signatureWebParams`＝`MD5(WebSalt + 排序key=value + WebSalt)`，`WebSalt="NVPh5…hwt"`。
  - `signKey`（v5/url 用）＝`MD5(hash+SignKeySalt+appid+mid+userid)`，**userid 为空必须用 "0"**，否则报 illegal key。
- `SendRequest` 默认注入 dfid/mid/uuid/appid/clientver/clienttime 并参与签名。**很多 login-user/gateway 接口依赖这些默认参数参与签名**，`ClearDefault:true` 会导致签名校验失败（扫码 20010）。

## 酷狗接口迁移现状（重点避坑）

大量旧 `mobilecdn.kugou.com` 接口已废弃（返回 `Access Deny`），已迁移到新接口；未迁移的旧接口**不做工作时保持 mobilecdn**：

- **已迁移的新接口**（原 mobilecdn 全部 Access Deny，已切换）：
  - 搜索建议 `/v2/getSearchTip` + `x-router: searchtip.kugou.com`（响应 `data[].RecordDatas[].HintInfo`）。
  - 分类型搜索 `/v1/search/{album|author|mv|special}` + `x-router: complexsearch.kugou.com`；单曲 `/v3/search/song`（mobilecdn 仍可用，保留）。
  - **综合搜索 `/search/complex` 已重构为后端聚合**：并行 song(mobilecdn)+album+author+special+mv，因为 `/v6/search/complex` 与 `/v3/search/mixed` 均返回 error 152（对所有人失效）。
  - 每日推荐 `POST /everyday_song_recommend` + `x-router: everydayrec.service.kugou.com`（data.song_list=30 首，字段 hash/ori_audio_name/sizable_cover/author_name/time_length）。
  - 音质权限 `POST /v2/get_res_privilege/lite` + `x-router: media.store.kugou.com`（data 数组，解析见 ParsePrivilegeLite）。
  - 用户歌单 `POST /v7/get_all_list` + `x-router: cloudlist.service.kugou.com`；用户信息 `POST /v3/get_my_info` + `x-router: usercenter.kugou.com`（p=RSA 加密）。
  - 发送验证码 `POST login.user.kugou.com/v7/send_mobile_code`。
  - 账号登录 `POST /v9/login_by_pwd`（x-router: login.user.kugou.com）：params=AES(JSON)、pk=裸RSA(JSON)、secu_params 需 AES 解密合并进 data。
  - 手机登录 `POST loginserviceretry.kugou.com/v7/login_by_verifycode`：概念版(lite) 需 t1/t2（GUID/MAC/DEV 固定 AES key）+ key=signParamsKey + pk=裸RSA。
  - 领 VIP `POST /youth/v1/ad/play_report`（需登录）。
- **新增接口（对齐 MoeKoeMusic 功能）**：`/api/rank/list`(`/ocean/v6/rank/list`)、`/api/rank/audio`(`/openapi/kmr/v2/rank/audio`+kg-tid:369)、`/api/top/song`(`/musicadservice/container/v1/newsong_publish`)、`/api/top/album`(`/musicadservice/v1/mobile_newalbum_sp`)、`/api/top/card`(`/singlecardrec.service/v1/single_card_recommend`)、`/api/playlist/tags`(`/pubsongs/v1/get_tags_by_type`)。
- **裸 RSA 加密**：登录/用户信息用裸 RSA（raw modPow 无 PKCS1 填充，明文右对齐补0，输出 keyLen*2 位大写 hex），与 `RSAEncrypt`（PKCS1v15，设备注册用）不同，勿混用。`RawRSAEncryptJSON` 用 Go map 序列化（**按字母序排 key**），需保序的接口（如 get_my_info 的 `{"token":..,"clienttime":..}`）必须手工构造 JSON 字节再 `RawRSAEncrypt`。
- **登录后用户类接口的坑（get_my_info / get_all_list，踩过重雷）**：
  - **平台匹配（关键）**：这两接口必须用**与登录 token 相同的平台**签名+RSA 公钥。本项目全局是概念版(lite)，账号/登录也走概念版：
    - `GetUserDetail`（get_my_info）：**必须用概念版 RSA 公钥**（`RawRSAEncrypt(json, true)`）+ lite 签名。用标准版公钥/签名 → 20018/20006。
    - `GetUserPlaylist`（get_all_list）：用 lite 签名，概念版 token 返回歌单正常；用标准版签名 → 20017。
  - **RSA 明文 JSON 字段顺序**：`get_my_info` 的 `p` 参数用 `fmt.Sprintf` 手工构造 `{"token":..,"clienttime":..}` 保持与 JS 一致；`RawRSAEncryptJSON(map)` 会按字母序排成 `{"clienttime":..,"token":..}`，RSA 明文不同 → 上游 20018。**登录接口不受影响**（clienttime_ms<key 字母序恰好等于 JS 插入序）。
  - **mid 必须传字符串 `"undefined"`**（`MID:"undefined"` + `MIDSet:true` 禁止 fallback）：空字符串返回 20006、真实 MID 返回 20018。dfid 固定 `"-"`（真实 dfid 会 20006）。
  - **userid 科学计数法**：登录/QR 解析 userid 用 `playlistIDString()`，`fmt.Sprintf("%v", float64)` 会输出 `7.25779799e+08` 导致上游 20017/20018。
  - `RequestOptions` 含 `MIDSet`/`DFIDSet`（空值不自动 fallback 到全局 GetMID/GetDFID）。
- **设备标识持久化**：`GetGUID` 从 `/root/X-music/.device-guid` 读取/持久化（`loadOrCreateGUID`），保证服务重启后 GUID/MID 不变。**mid 与登录 token 绑定，重启后 mid 变了会导致 VIP 播放/用户接口 20018**。前端 `auth.setDeviceInfo` 保存 register_dev 返回的 `MID/GUID/DFID`（**注意是大写字段**）到 localStorage，设备注册后刷新不丢。
- **VIP 歌播放（踩过重雷）**：
  - `GetSongURL`（v5/url）**必须带 `Token: creds.Token`**，否则 VIP 歌返回 status:2（免费歌不受影响）。带 token 后概念版 VIP 歌直接返回完整 mp3 url。
  - v5/url 仍 status:2 时降级 `GetPrivURL`（`tracker.kugou.com/v6/priv_url`，概念版参数，`tracker_param.key=MD5(hash+lite盐+appid+mid+userid)`）兜底。priv_url 完整音频是 **mgg 加密**（`en_tracker_url`，浏览器不可播），仅 `climax_info.url`（高潮片段 mp3）可直接播，故只作兜底。
  - **登录 token 必须与播放接口平台一致**：概念版 VIP 播放需要概念版(lite) token。扫码登录 appid 已改为概念版（`QRAppID="3116"`，见 `CheckQRStatus`/`GetQRKey` 的 qrcode_txt），不能用标准版（1005）否则 v5/url 不解锁。
- **已接入**：歌单管理、用户 VIP 详情、歌手关注/歌手列表、**关注歌手列表**（`/api/user/follow`，relationuser/v4/follow_list，RSA 加密，响应 `data.lists` 含 nickname/pic/singerid，Me 页支持取消关注）。**仍未接入**：MV 播放、已购歌曲、上传历史等（见 TASKS.md）。
- **新接口通用模式**：`POST gateway.kugou.com/{path}` + header `x-router: <服务名>` + android 签名；openapi 接口额外带 `kg-tid`。各接口在 api/*.go 及 `/root/MoeKoeMusic/api/module/*.js`（MakcRe 参考）都有验证过的实现。
  - 播放 v5/url：`gateway.kugou.com/v5/url` + `x-router: trackercdn.kugou.com` + `EncryptKey` + `KeyHash`(signKey)。
  - 歌词：`lyrics.kugou.com/v1/search`（hash+keyword）→ `download`；krc 格式需 KRC 解码（**跳过4字节头** + XOR + zlib）。
  - QR 登录：`login-user.kugou.com`（web 签名，保留默认参数）。
- 数字 ID 用 `playlistIDString()` 格式化，避免 float64 科学计数法（如 `1.845e+08`）。
- 歌单详情/歌曲用 `global_collection_id` 格式 ID（`collection_3_xxx`），不能用普通数字 id。
- **歌单评论需用创建者视角的 `comment_id`**：用户收藏/转存的歌单（`collection_3_`），评论关联在原始创建者那边，后端 `parsePlaylist` 从 `list_create_userid` + `list_create_listid` 构造 `comment_id` 字段。公共歌单（`collection_1_`）直接使用 `global_collection_id`。前端 CommentModal 用 `playlist.comment_id || playlist.id`。
- **评论分页数修正**：酷狗评论 API 的 `combine_count` 包含子回复数，可能虚高。CommentModal 在列表实际条数 < pagesize 时以实际数为准，翻页返回空时自动修正总数。
- 版权/VIP 歌的 v5/url **登录后带 Token 即可播放**（见上「VIP 歌播放」）；仅当未登录或平台不匹配时才返回 `status:2` 无 url。前端对 status:2 显示「无法获取播放地址」而非报错。
- **云盘上传（踩过重雷）**：
  - **BSS upload filename 必须用文件内容 MD5**（不能是用户原文件名如 `羽·泉 - 狂流 (live).flac`），否则初始化分片返回 `error_code: 80024`。参考 `MoeKoeMusic/api/module/user_cloud_upload.js` 始终用 `fileMd5(fileData).toLowerCase()`。
  - **`doBssRequest` 的 `DialContext` 不能传 `nil` context**（Go 1.21+ 会 panic），必须使用外层传入的 `ctx` 参数：`dialer.DialContext(ctx, "tcp4", addr)`。

## 功能对齐进度（详见 TASKS.md）

- **已对齐**：搜索/建议/热搜/综合/分类型、播放 url、音质权限、歌词(+KRC)、每日推荐、歌单(详情/歌曲/标签/增删/加删歌)、排行(rank_list/audio)、新歌新专辑、专辑、歌手(详情/歌曲/专辑/热门/列表/关注)、评论全家桶、登录(密码/手机/扫码)、用户信息/歌单/收藏/VIP 详情、每日领 VIP。
- **未对齐**：我听/听歌历史/关注列表、榜单增强(rank_top/song_ranking 等)、歌词搜索(后端有未接路由)、歌曲详情(mobilecdn 失效)。
- **明确不做**：MV/视频、青少年频道、长音频、场景/主题、音乐人/IP。
- **进行中**：未登录引导（`requireLogin` + 401 全局拦截）已上线；歌单管理、评论、VIP 详情、歌手关注已完成。

## 前端

- Vue3 `<script setup>` SFC + Pinia + axios。API 封装在 `web/src/utils/api.js`（各 api 调后端 `/api/*` 固定路由，参数名与后端 handler 读取的 query 一致）。
- 注意参数命名：歌手接口后端读 `singer_id`（前端 `getArtistDetail(singer_id)`），专辑读 `album_id`。改后端 query 名需同步前端。
- **布局为侧栏 + 顶栏**：`App.vue` 只做容器，`Sidebar.vue`（发现音乐/每日推荐/排行榜/猜你喜欢 + 我的音乐/我的/最近播放/音乐云盘，导航分两组、激活项带蓝色图标块+左侧指示条）+ `TopBar.vue`（搜索框居中 + 设置按钮 + 用户区）+ `PlayerBar.vue`（底部播放器浮层）+ `LoginModal`。`BottomTab.vue` 已废弃无引用，勿使用。改动布局先看这三个组件。
- **默认音质设置**：右上角设置按钮（`TopBar.vue` 齿轮）弹层选择默认音质（128/320/flac/high(Hi-Res)），调用 `player.setQuality` 写入 localStorage `vibe_quality`；弹层保持打开不自动关闭（用户需实时看勾选状态）。**无当前播放歌曲时也能要改默认音质**（player.js `setQuality` 先更新 quality+持久化，有 currentSong 才请求换源）。**注意 `setQuality()` 必须先存旧值再赋新值，否则 `q===quality.value` 永远 true 导致从不会换源**。
- **音质降级链**：后端 `handleSongURL`（main.go）逐级尝试请求音质及以下等级（4 级链：`high → flac → 320 → 128`），所有等级试完后取 `QualityRank` 最高者返回。不再只取第一个结果，因为部分歌曲 v5/url 请求高音质时会返回低音质源。VIP/版权限制（status:2）不降级、改走 priv_url 兜底。
- **音质推断**：`inferQualityFromResponse` 按 bitRate+扩展名区分：>2M 为 Hi-Res，FLAC 为无损，>=320k 为高品，其余为标准。对 VIP 歌曲 v5/url 不返 bitRate 的情况，`handleSongURL` 直接信任请求的音质值。
- **播放失败自动跳过**：`PlayerBar.vue` 监听 `playError`，3 秒后自动跳下一首（需队列>1首）。
- **深色模式**：`index.html` 头部阻塞脚本在 Vue 挂载前读 localStorage `vibe_dark` 设 `html[data-theme]`（防闪烁）；`TopBar.vue` 设置弹层提供切换并持久化；**App.vue 勿写 `watch(theme, immediate)` 会在子组件之后把 dark 覆盖回 light**。CSS 变量在 `global.css` 的 `[data-theme="dark"]` 段定义，所有组件统一用 `var(--*)` 不可硬编码颜色。`html` 需设 `background: var(--bg)` 否则 `backdrop-filter` 透出浏览器默认白。`color-mix(in srgb, var(--bg) 72%, transparent)` 在暗色下不可靠，用实色 `var(--bg)`。
- **歌曲列表**：全局 `.song-table`/`.song-head`/`.song-row`（global.css）固定 7 列（索引/封面/歌曲/专辑/歌手/红心/时长，44+52+1fr+1.1fr+1fr+48+54px，行高 64px）。表头文字对齐必须与数据列一致（歌手/时长右对齐），改列序要同时同步表头与行。播放中整行 `--primary-soft` 高亮。
- **详情页头部统一**：Playlist/Album/Artist 共用全局 `.detail-header`/`.detail-cover`/`.detail-avatar`（208px 方形封面 + 30px 标题 + 操作按钮），详情页 SFC 内不再自带重复样式。
- **Me 页无 VIP 卡片**：VIP 详情/每日领 VIP 卡片已删除。**关注歌手列表**登录后可查看（`/api/user/follow`），支持取消关注和跳转歌手详情。
- **未登录引导**：`App.vue` 提供 `requireLogin(reason)`（provide 注入），未登录的功能入口统一调用它弹登录框并提示原因；后端 401「请先登录」由 `api.js` 响应拦截器捕获后派发 `vibe:login-required` 全局事件，`App.vue` 监听弹出登录框（带 `reason` 提示）。各页 catch 到 401 时不要误报「网络错误」。
- **登录方式**：`LoginModal.vue` 目前只保留**扫码登录**和**手机验证码**两种（账号密码登录已删除，`/api/login` 后端路由仍在但前端无入口）。手机号绑定多账号时后端自动选第一个账号登录（`handleLoginCellphone` 解析 `info_list` 取首个 userid 重发）。手机/扫码登录响应均可能带 `vip_token`/`vip_type`，前端 `auth.js` 存入 localStorage（`setCredentials`），播放接口传参用于 VIP 解锁。- **登录后用户信息**：扫码/密码/手机登录成功后 `LoginModal.vue` 都会调 `auth.fetchUser()`（`auth.js`）拉取 `/api/user/detail` 拿昵称/头像（扫码登录只返回凭证，无 user）。头像/昵称渲染在 `TopBar.vue` 用户区（`auth.user?.avatar`）与 `Me.vue`（`me-avatar-img`）。`auth.js` 从 localStorage 恢复登录态时也会自动 `fetchUser()`，旧缓存缺头像/昵称也能补上。**改动这些文件必须重新 `npm run build` + `go build`，否则二进制 embed 的还是旧 dist**。
- **详情页判断**：路由 `detailPaths = ['/playlist','/album','/artist','/lyric']`，命中时 `App.vue` 自动隐藏 Sidebar/PlayerBar（沉浸式）。这些页面顶部回退用 `BackBar.vue`。新增详情类路由必须同步该数组，否则布局错乱。
- **首页搜索栏**：Home.vue 顶部搜索栏是个假输入框，点击 `router.push('/search')` 跳独立搜索 Tab，不是真正的输入框。
- **全局样式**：`web/src/assets/style/global.css` 用 CSS 变量（`--color-*`/`--glass-*`/`--radius-*`/`--tabbar-height` 等）统一定义 iOS19 毛玻璃设计语言；各视图模板复用全局类（`.glass-card`/`.card`/`.song-table`/`.btn`），不加新组件只需调变量。改视觉优先改这个文件而非逐个 SFC。
- **鼠标指针**：不使用自定义 SVG 指针（不跨组件兼容），统一用原生 OS `cursor: pointer`。非可点区域用原生默认指针。`global.css` 底部集中管理可点元素列表。