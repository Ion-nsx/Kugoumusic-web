<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="comment-modal">
      <div class="cm-header">
        <div class="cm-title">
          <span>{{ title }}</span>
        </div>
        <button class="cm-close" @click="$emit('close')">✕</button>
      </div>

      <div class="cm-body">
        <div v-if="loading" class="cm-loading">
          <div class="skeleton" style="width: 100%; height: 120px; border-radius: 8px;"></div>
          <div class="skeleton" style="width: 100%; height: 120px; border-radius: 8px; margin-top: 10px;"></div>
        </div>

        <div v-else-if="comments.length === 0" class="cm-empty">
          {{ emptyText }}
        </div>

        <template v-else>
          <div class="cm-item" v-for="c in comments" :key="c.id">
            <div class="cm-avatar">
              <img v-if="c.user_img" :src="c.user_img" :alt="c.nickname" @error="onImgError($event)" />
              <span v-else>{{ avatarText(c) }}</span>
            </div>
            <div class="cm-main">
              <div class="cm-meta">
                <span class="cm-nick">{{ c.nickname || c.author_name || '音乐用户' }}</span>
                <span class="cm-time">{{ c.addtime }}</span>
              </div>
              <div class="cm-content">{{ c.content }}</div>
              <div class="cm-actions">
                <button class="cm-action" @click="toggleFloor(c)">
                  <svg viewBox="0 0 24 24"><path d="M21 12a8 8 0 0 1-8 8H4l2.2-2.6A8 8 0 1 1 21 12Z"/></svg>
                  {{ c.comments_num || 0 }}
                </button>
              </div>

              <!-- 楼层 -->
              <div class="cm-floor" v-if="openFloors[c.id]">
                <div v-if="floorLoading[c.id]" class="cm-floor-loading">
                  <div class="skeleton" style="height: 40px; border-radius: 6px;"></div>
                </div>
                <div v-else-if="(floors[c.id] || []).length === 0" class="cm-floor-empty">暂无回复</div>
                <div class="cm-floor-item" v-for="f in floors[c.id] || []" :key="f.id">
                  <span class="cf-nick">{{ f.nickname || f.author_name || '音乐用户' }}</span>
                  <span class="cf-text">{{ f.content }}</span>
                </div>
              </div>
            </div>
          </div>
        </template>
      </div>

      <div class="cm-footer" v-if="count > 0 || comments.length > 0">
        <div class="cm-pagination">
          <button class="pg-btn" :disabled="page <= 1" @click="goPage(1)">首页</button>
          <button class="pg-btn" :disabled="page <= 1" @click="goPage(page - 1)">上一页</button>
          <template v-for="p in pageNumbers" :key="p">
            <span v-if="p === '...'" class="pg-ellipsis">...</span>
            <button v-else class="pg-btn" :class="{ on: p === page }" @click="goPage(p)">{{ p }}</button>
          </template>
          <button class="pg-btn" :disabled="page >= totalPages" @click="goPage(page + 1)">下一页</button>
          <button class="pg-btn" :disabled="page >= totalPages" @click="goPage(totalPages)">末页</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import {
  getSongComments, getPlaylistComments, getAlbumComments,
  getCommentFloor, imgUrl
} from '../utils/api'

const props = defineProps({
  resourceType: { type: String, default: 'song' },
  title: { type: String, default: '' },
  mixsongid: { type: [String, Number], default: '' },
  hash: { type: String, default: '' },
  id: { type: String, default: '' }
})

const comments = ref([])
const count = ref(0)
const page = ref(1)
const pagesize = 30
const loading = ref(true)
const floors = ref({})
const floorLoading = ref({})
const openFloors = ref({})

const emptyText = computed(() => '还没有评论，快来抢沙发吧！')

// 最多展示 100 页，避免 API count 虚高造成大量空页
const maxPages = 100
const totalPages = computed(() => Math.max(1, Math.min(Math.ceil(count.value / pagesize), maxPages)))

const pageNumbers = computed(() => {
  const total = totalPages.value
  const cur = page.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const nums = [1]
  if (cur > 3) nums.push('...')
  const start = Math.max(2, cur - 1)
  const end = Math.min(total - 1, cur + 1)
  for (let i = start; i <= end; i++) nums.push(i)
  if (cur < total - 2) nums.push('...')
  nums.push(total)
  return nums
})

function avatarText(c) {
  const n = c.nickname || c.author_name || '音'
  return n.slice(0, 1)
}

function onImgError(e) {
  e.target.style.display = 'none'
  e.target.nextElementSibling && (e.target.nextElementSibling.style.display = '')
}

async function load(pageNo) {
  loading.value = true
  try {
    const res = await fetchByType(pageNo)
    if (res.data.status === 1) {
      const list = res.data.list || res.data.comments || res.data.reply_list || []
      if (list.length > 0) {
        if (list.length < pagesize) {
          // 最后一页：总数以实际为准（API 的 combine_count 可能包含回复数导致虚高）
          count.value = (pageNo - 1) * pagesize + list.length
        } else {
          count.value = Number(res.data.count) || Number(res.data.combine_count) || count.value || list.length
        }
        comments.value = normalize(list)
        page.value = pageNo
      } else if (pageNo === 1) {
        // 第一页就没数据
        count.value = 0
        comments.value = []
        page.value = 1
      } else {
        // 翻页返回空，API 统计数虚高，修正总数
        count.value = (pageNo - 1) * pagesize
        loading.value = false
        return
      }
    }
  } catch (e) {
    console.error('加载评论失败:', e)
  } finally {
    loading.value = false
  }
}

function goPage(p) {
  p = Number(p)
  if (p < 1 || p > totalPages.value || loading.value || p === page.value) return
  load(p)
  const body = document.querySelector('.cm-body')
  if (body) body.scrollTop = 0
}

function fetchByType(pageNo) {
  if (props.resourceType === 'playlist') {
    return getPlaylistComments(props.id, pageNo, pagesize)
  }
  if (props.resourceType === 'album') {
    return getAlbumComments(props.id, pageNo, pagesize)
  }
  return getSongComments(props.mixsongid, pageNo, pagesize)
}

// 提取用户头像 URL（兼容多种字段名）
function pickAvatar(c) {
  return c.user_img || c.sizable_avatar || c.img || c.user_pic || c.cover || c.icon || ''
}

function normalize(list) {
  return list.map(c => ({
    id: c.id,
    content: c.content,
    addtime: c.addtime,
    like: c.like,
    comments_num: c.comments_num || 0,
    nickname: c.nickname || c.author_name || c.user_name || '',
    author_name: c.author_name || '',
    user_img: imgUrl(pickAvatar(c), 80)
  }))
}

async function toggleFloor(c) {
  if (openFloors.value[c.id]) {
    openFloors.value[c.id] = false
    return
  }
  openFloors.value[c.id] = true
  if (floors.value[c.id]) return
  floorLoading.value[c.id] = true
  try {
    const res = await getCommentFloor(props.resourceType, {
      code: '',
      mixsongid: String(props.mixsongid || ''),
      special_id: props.resourceType === 'song' ? '' : props.id,
      tid: String(c.id || '')
    })
    if (res.data.status === 1) {
      floors.value[c.id] = normalize(res.data.list || res.data.replys || res.data.comments || [])
    } else {
      floors.value[c.id] = []
    }
  } catch (e) {
    floors.value[c.id] = []
  } finally {
    floorLoading.value[c.id] = false
  }
}

onMounted(() => {
  load(1)
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: var(--color-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.comment-modal {
  width: 680px;
  max-width: 94vw;
  height: 85vh;
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  animation: modal-in 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}
@keyframes modal-in {
  from { opacity: 0; transform: scale(0.96) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.cm-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 22px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.cm-title { display: flex; align-items: center; gap: 10px; font-size: var(--font-size-lg); font-weight: var(--font-weight-bold); min-width: 0; }
.cm-title span:first-child { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cm-count { font-size: var(--font-size-xs); font-weight: var(--font-weight-semibold); color: var(--text-3); background: var(--surface-3); padding: 2px 10px; border-radius: var(--radius-full); flex-shrink: 0; }
.cm-close { width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; border-radius: 50%; color: var(--text-2); transition: background var(--transition-fast); }
.cm-close:hover { background: var(--surface-3); }

.cm-body { flex: 1; overflow-y: auto; padding: 10px 0; }
.cm-loading { padding: 20px; }
.cm-empty { padding: 48px; text-align: center; color: var(--text-3); font-size: var(--font-size-sm); }

.cm-item { display: flex; gap: 12px; padding: 14px 22px; }
.cm-item:hover { background: var(--surface-2); }
.cm-avatar {
  width: 36px; height: 36px; flex-shrink: 0;
  border-radius: 50%; overflow: hidden;
  background: linear-gradient(135deg, #2CA6F8, #7C6BFF);
  display: grid; place-items: center;
  color: #fff; font-size: 13px; font-weight: 700;
}
.cm-avatar img { width: 100%; height: 100%; object-fit: cover; }
.cm-main { flex: 1; min-width: 0; }
.cm-meta { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.cm-nick { font-size: var(--font-size-sm); font-weight: var(--font-weight-semibold); color: var(--text-2); }
.cm-time { font-size: var(--font-size-xs); color: var(--text-3); }
.cm-content { font-size: var(--font-size-base); line-height: 1.6; word-break: break-word; }
.cm-actions { display: flex; align-items: center; gap: 18px; margin-top: 8px; }
.cm-action {
  display: flex; align-items: center; gap: 4px;
  font-size: var(--font-size-xs); color: var(--text-3);
  transition: color var(--transition-fast);
}
.cm-action:hover { color: var(--text-1); }
.cm-action svg { width: 14px; height: 14px; stroke: currentColor; fill: none; stroke-width: 1.8; stroke-linecap: round; stroke-linejoin: round; }
.cm-action.like.on { color: var(--like); }
.cm-action.like.on svg { fill: currentColor; }

.cm-floor {
  margin-top: 8px; padding: 10px 12px;
  background: var(--surface-3); border-radius: var(--r-sm);
}
.cm-floor-loading { padding: 4px; }
.cm-floor-empty { font-size: var(--font-size-xs); color: var(--text-3); padding: 4px; }
.cm-floor-item { font-size: var(--font-size-sm); line-height: 1.5; padding: 3px 0; word-break: break-word; }
.cf-nick { color: var(--primary-deep); font-weight: 600; margin-right: 6px; }
.cf-text { color: var(--text-1); }

/* 分页 */
.cm-footer { flex-shrink: 0; padding: 10px 22px; border-top: 1px solid var(--border); }
.cm-pagination { display: flex; align-items: center; justify-content: center; gap: 4px; }
.pg-btn {
  min-width: 32px; height: 30px;
  display: inline-flex; align-items: center; justify-content: center;
  border-radius: var(--r-sm);
  font-size: var(--font-size-xs); color: var(--text-2);
  padding: 0 8px;
  transition: all var(--transition-fast);
}
.pg-btn:hover:not(:disabled) { background: var(--surface-3); color: var(--text-1); }
.pg-btn:disabled { opacity: 0.35; cursor: default; }
.pg-btn.on { background: var(--primary); color: #fff; font-weight: 600; }
.pg-ellipsis { font-size: var(--font-size-xs); color: var(--text-3); padding: 0 2px; }
.pg-info { font-size: var(--font-size-xs); color: var(--text-3); margin-left: 8px; }
</style>
