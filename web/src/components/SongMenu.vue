<template>
  <!-- 行内「更多操作」按钮：绝对定位，不占网格列（避免改动 7 列表格布局） -->
  <button
    ref="btnEl"
    class="row-more"
    :class="{ open }"
    title="更多操作"
    @click.stop="toggle"
  >
    <svg viewBox="0 0 24 24"><circle cx="5" cy="12" r="1.7"/><circle cx="12" cy="12" r="1.7"/><circle cx="19" cy="12" r="1.7"/></svg>
  </button>

  <!-- 菜单挂到 body：避免被 .song-table { overflow: hidden } 裁掉 -->
  <Teleport to="body">
    <div v-if="open" class="song-menu-mask" @click="close"></div>
    <div v-if="open" class="song-menu lg-glass-strong" :style="menuStyle" @click.stop>
      <template v-if="flash">
        <div class="sm-flash">✓ {{ flash }}</div>
      </template>

      <template v-else-if="picking">
        <div class="sm-head">
          <button class="sm-back" @click="picking = false">‹ 返回</button>
          <span>选择歌单</span>
        </div>
        <div class="sm-pl-list">
          <div v-if="loadingPl" class="sm-empty">加载中…</div>
          <div v-else-if="!playlists.length" class="sm-empty">暂无歌单</div>
          <button
            v-else
            v-for="pl in playlists"
            :key="pl.id"
            class="sm-pl"
            @click="addTo(pl)"
          >
            <span class="sm-pl-name">{{ pl.name }}</span>
            <span class="sm-pl-count">{{ pl.count || 0 }} 首</span>
          </button>
        </div>
      </template>

      <template v-else>
        <button class="sm-item" @click="doNext">
          <svg viewBox="0 0 24 24"><path d="M4 6l9 6-9 6Z"/><path d="M18 5v14"/></svg>
          <span>下一首播放</span>
        </button>
        <button class="sm-item" @click="doLike">
          <svg viewBox="0 0 24 24"><path d="M12 21s-7.5-4.6-9.5-8.7C1 9 3 5.5 6.5 5.5c2 0 3.6 1.2 4.5 3 .9-1.8 2.5-3 4.5-3 3.5 0 5.5 3.5 4 6.8C18.5 16.4 12 21 12 21Z"/></svg>
          <span>{{ liked ? '取消收藏' : '收藏' }}</span>
        </button>
        <button class="sm-item" @click="doQueue">
          <svg viewBox="0 0 24 24"><path d="M4 6h11M4 12h11M4 18h7"/><path d="M18 12v7M15 16h6"/></svg>
          <span>添加到播放队列</span>
        </button>
        <button class="sm-item" @click="openPlaylists">
          <svg viewBox="0 0 24 24"><path d="M12 5v14M5 12h14"/></svg>
          <span>添加到歌单</span>
        </button>
      </template>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed } from 'vue'
import { usePlayerStore } from '../stores/player'
import { useAuthStore } from '../stores/auth'
import { getUserPlaylist, addPlaylistTracks } from '../utils/api'

const props = defineProps({
  song: { type: Object, required: true }
})

const player = usePlayerStore()
const auth = useAuthStore()

const btnEl = ref(null)
const open = ref(false)
const pos = ref({ top: 0, left: 0 })
const flash = ref('')
const picking = ref(false)
const loadingPl = ref(false)
const playlists = ref([])

const liked = computed(() => player.isLiked(props.song))
const menuStyle = computed(() => ({ top: pos.value.top + 'px', left: pos.value.left + 'px' }))

const MENU_W = 190

// 根据按钮位置定位弹层，空间不足时自动翻转
function toggle() {
  if (open.value) { close(); return }
  const r = btnEl.value.getBoundingClientRect()
  let left = r.right - MENU_W
  if (left < 8) left = 8
  let top = r.bottom + 6
  const estH = 190
  if (top + estH > window.innerHeight - 8) top = Math.max(8, r.top - 6 - estH)
  pos.value = { top, left }
  flash.value = ''
  picking.value = false
  open.value = true
}

function close() {
  open.value = false
  picking.value = false
  flash.value = ''
}

// 短暂显示结果提示后自动关闭
function done(text) {
  flash.value = text
  setTimeout(close, 700)
}

function doNext() {
  player.insertNext(props.song)
  done('已添加到下一首播放')
}

function doLike() {
  const wasLiked = liked.value
  player.toggleLike(props.song)
  done(wasLiked ? '已取消收藏' : '已收藏')
}

function doQueue() {
  player.addToQueue(props.song)
  done('已添加到播放队列')
}

async function openPlaylists() {
  if (!auth.isLoggedIn) {
    window.dispatchEvent(new CustomEvent('vibe:login-required'))
    close()
    return
  }
  picking.value = true
  if (playlists.value.length) return
  loadingPl.value = true
  try {
    const res = await getUserPlaylist()
    if (res.data.status === 1 && Array.isArray(res.data.data)) {
      playlists.value = res.data.data
    }
  } catch (e) {
    console.error('获取歌单失败:', e)
  } finally {
    loadingPl.value = false
  }
}

async function addTo(pl) {
  // 加歌用数字 listid，老版本接口返回可能没有 list_id，退回 id
  const listid = pl.list_id || pl.id
  try {
    const res = await addPlaylistTracks(listid, [props.song])
    const errCode = res.data?.error_code != null ? Number(res.data.error_code) : (res.data?.status === 0 ? -1 : 0)
    done(errCode === 0 ? `已添加到「${pl.name}」` : '添加失败')
  } catch (e) {
    console.error('添加歌曲到歌单失败:', e)
    done('添加失败')
  }
}
</script>
