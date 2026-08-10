# <img src="https://raw.githubusercontent.com/Ion-nsx/Kugoumusic-web/main/web/public/favicon.svg" width="32" height="32" align="top"> Kugoumusic-web

<div align="center">

**酷狗音乐概念版 · 第三方全栈客户端**

Go 后端 + Vue3 前端，单二进制部署，开箱即用

**参考项目**：[MoeKoeMusic](https://github.com/MoeKoeMusic/MoeKoeMusic) · [KuGouMusicApi](https://github.com/MakcRe/KuGouMusicApi)

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.5-4FC08D?logo=vue.js)](https://vuejs.org/)
[![Vite](https://img.shields.io/badge/Vite-8-646CFF?logo=vite)](https://vitejs.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

</div>

---

## ✨ 特性

- 🎧 **高品质播放** — 支持 128k / 320k / FLAC / Hi-Res 多音质，自动降级
- 🔐 **VIP 解锁** — 登录后概念版 VIP 歌曲直接播放
- 🔍 **全能搜索** — 单曲 / 歌手 / 专辑 / MV / 歌单 / 歌词，综合搜索后端聚合
- 📋 **歌单管理** — 创建 / 删除 / 收藏 / 加歌删歌 / 每日推荐
- 💬 **评论系统** — 歌曲 / 歌单 / 专辑评论，支持楼层回复和发表评论
- 🏆 **排行榜** — 多榜同列（最多 6 个），滚动分页，偏好持久化
- 👤 **用户中心** — 扫码 / 手机验证码登录，个人信息 / 关注歌手 / 我的歌单
- 🌙 **深色模式** — CSS 变量驱动的深色主题，零闪烁切换
- 📱 **响应式设计** — iOS19 毛玻璃设计语言，侧栏 + 顶栏 + 底部播放器布局
- 🚀 **单二进制** — 前端 SPA 嵌入 Go binary，`./vibe-server` 一键启动

## 🏗️ 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.26, net/http, crypto (AES/RSA) |
| 前端 | Vue 3 `<script setup>` SFC, Pinia, Vue Router, Axios |
| 构建 | Vite, `//go:embed` 嵌入前端产物 |
| 协议 | 酷狗 Android/Web 签名算法, KRC 歌词解码, RSA 裸加密 |
| 设计 | CSS 变量, iOS19 毛玻璃, 深色/浅色主题 |

## 📂 项目结构

```
├── main.go              # 入口 + 路由 + HTTP 服务
├── api/                 # 后端 API 模块（19 个文件）
│   ├── request.go       # 请求签名、平台切换、HTTP 客户端
│   ├── crypto.go        # AES/RSA 加密工具
│   ├── search.go        # 搜索（建议/热搜/综合/分类型）
│   ├── song.go          # 播放地址、歌词（KRC 解码）
│   ├── playlist.go      # 歌单（详情/歌曲/标签/推荐/每日推荐）
│   ├── social.go        # 歌单管理 + 评论（增删歌/创建删除/评论增删查）
│   ├── album.go         # 专辑（详情/歌曲）
│   ├── artist.go        # 歌手（详情/歌曲/专辑/热门/列表/关注）
│   ├── rank.go          # 排行榜（榜单列表/歌曲）
│   ├── user.go          # 用户信息/VIP/关注
│   ├── auth.go          # 登录（扫码/手机/密码）
│   ├── discover.go      # 发现页（新歌/新专/推荐卡片）
│   └── types.go         # 公共类型定义
├── middleware/           # 中间件（CORS/日志）
├── web/                 # 前端项目
│   ├── src/
│   │   ├── views/       # 页面（Home/Me/Search/Ranking/Playlist/Album/Artist）
│   │   ├── components/  # 通用组件（Sidebar/TopBar/PlayerBar/CommentModal 等）
│   │   ├── stores/      # Pinia stores（player/auth）
│   │   ├── utils/       # API 封装（api.js/auth.js）
│   │   └── assets/      # 静态资源 + 全局样式
│   └── dist/            # 构建产物（go embed）
├── TASKS.md             # 功能对齐清单
└── AGENTS.md            # 开发备忘（签名/接口迁移/坑位记录）
```

## 🚀 快速开始

### 前置要求

- Go 1.26+
- Node.js 18+（仅开发/构建前端时需要）

### 构建

```bash
# 1. 构建前端
cd web && npm install && npm run build

# 2. 构建后端（前端产物嵌入二进制）
cd .. && go build -o vibe-server .
```

### 运行

```bash
# 概念版（默认）
./vibe-server

# 自定义端口（默认 :8080）
PORT=3000 ./vibe-server
```

浏览器打开 `http://localhost:8080`

## 🎯 功能状态

- ✅ **基础 & 设备** — 设备注册、图片代理
- ✅ **搜索** — 单曲/建议/热搜/综合/分类型/歌词/默认词
- ✅ **播放** — 多音质播放 + 降级链、KRC 歌词解码、VIP 解锁
- ✅ **歌单** — 每日推荐/个性推荐/标签/详情/歌曲/增删改查
- ✅ **排行榜** — 榜单列表/歌曲、新歌/新专/推荐卡片
- ✅ **歌手 & 专辑** — 详情/歌曲/专辑/热门/列表/关注
- ✅ **评论** — 歌曲/歌单/专辑评论 + 分页 + 楼层 + 发表
- ✅ **登录** — 扫码/手机验证码（概念版 appid）
- ✅ **用户** — 信息/VIP/歌单/关注歌手
- ✅ **iOS19 毛玻璃设计** — 深色模式 + 响应式布局
- ☐ **待做** — 云盘、MV/视频、电台、场景/主题、音乐人

## ⚠️ 免责声明

本项目仅供学习和研究使用，请勿用于商业用途。API 版权归酷狗音乐所有。

## 📄 License

Apache-2.0
