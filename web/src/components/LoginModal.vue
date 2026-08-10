<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h2 class="modal-title">{{ reason || '登录' }}</h2>
        <button class="close-btn" @click="$emit('close')">✕</button>
      </div>
      <p v-if="reason" class="modal-reason">该功能需要登录后使用，请先登录</p>

      <!-- 登录方式切换 -->
      <div class="tab-bar">
        <button class="tab" :class="{ active: tab === 'qr' }" @click="tab = 'qr'">扫码登录</button>
        <button class="tab" :class="{ active: tab === 'phone' }" @click="tab = 'phone'">手机验证码</button>
      </div>

      <!-- 扫码登录 -->
      <div v-if="tab === 'qr'" class="tab-content">
        <div class="qr-container">
          <div class="qr-placeholder" v-if="qrLoading">
            <div class="skeleton" style="width: 200px; height: 200px; border-radius: 12px;"></div>
            <p style="margin-top: 12px; font-size: 13px; color: var(--text-3);">正在生成二维码...</p>
          </div>
          <div v-else-if="qrImage" class="qr-image">
            <img :src="qrImage" alt="扫码登录" />
            <p style="margin-top: 12px; font-size: 13px; color: var(--text-3);">请使用酷狗App扫码</p>
          </div>
          <div v-else-if="qrError" class="qr-error">
            <p>{{ qrError }}</p>
            <button class="btn btn-primary" @click="initQR" style="margin-top: 12px;">重试</button>
          </div>
        </div>
      </div>

      <!-- 手机验证码登录 -->
      <div v-if="tab === 'phone'" class="tab-content">
        <div class="form-group">
          <label>手机号</label>
          <input v-model="phone" type="tel" placeholder="请输入手机号" class="input" />
        </div>
        <div class="form-group code-group">
          <input v-model="smsCode" type="text" placeholder="验证码" class="input" />
          <button class="btn btn-ghost code-btn" @click="sendSMS" :disabled="smsCountdown > 0">
            {{ smsCountdown > 0 ? `${smsCountdown}s` : '发送' }}
          </button>
        </div>
        <button class="btn btn-primary login-btn" @click="loginPhone" :disabled="loading">
          {{ loading ? '登录中...' : '登录' }}
        </button>
        <p v-if="phoneError" class="error-msg">{{ phoneError }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { qrCreate, qrCheck, qrGet, loginCellphone, sendCaptcha } from '../utils/api'

const props = defineProps({
  reason: { type: String, default: '' }
})

const emit = defineEmits(['close'])
const auth = useAuthStore()

const tab = ref('qr')

const qrLoading = ref(false)
const qrImage = ref('')
const qrError = ref('')
const qrKey = ref('')
let qrTimer = null

const loading = ref(false)

const phone = ref('')
const smsCode = ref('')
const smsCountdown = ref(0)
const phoneError = ref('')
let smsTimer = null

onMounted(() => {
  initQR()
})

async function initQR() {
  qrLoading.value = true
  qrImage.value = ''
  qrError.value = ''
  qrKey.value = ''

  try {
    const res = await qrCreate()
    if (res.data.status === 1 && res.data.data) {
      const key = res.data.data.key || res.data.data.qr_key || res.data.data.qrcode
      qrKey.value = key

      if (res.data.data.qrcode_img) {
        qrImage.value = res.data.data.qrcode_img
      } else {
        const imgRes = await qrGet(key)
        if (imgRes.data.status === 1) {
          qrImage.value = imgRes.data.data
        }
      }

      startQRPolling(key)
    } else {
      qrError.value = '获取二维码失败'
    }
  } catch (e) {
    qrError.value = '网络错误'
  } finally {
    qrLoading.value = false
  }
}

function startQRPolling(key) {
  clearInterval(qrTimer)
  qrTimer = setInterval(async () => {
    try {
      const res = await qrCheck(key)
      if (res.data.status === 1 && res.data.data) {
        const data = res.data.data
        // 状态码: 0=过期 1=等待扫码 2=待确认 4=登录成功
        if (data.status === 4) {
          clearInterval(qrTimer)
          console.log('[QR Login] raw data:', JSON.stringify(data))
          auth.setCredentials({
            token: data.token,
            user_id: data.user_id,
            dfid: data.dfid,
            mid: data.mid,
            guid: data.guid,
            vip_token: data.vip_token || '',
            vip_type: data.vip_type || ''
          })
          auth.fetchUser()
          emit('close')
        } else if (data.status === 0) {
          clearInterval(qrTimer)
          qrError.value = '二维码已过期，请重试'
        }
      }
    } catch (e) {
      // ignore polling errors
    }
  }, 2000)
}

async function sendSMS() {
  if (!phone.value) {
    phoneError.value = '请输入手机号'
    return
  }
  try {
    const res = await sendCaptcha(phone.value)
    if (res.data.status === 1) {
      smsCountdown.value = 60
      smsTimer = setInterval(() => {
        smsCountdown.value--
        if (smsCountdown.value <= 0) clearInterval(smsTimer)
      }, 1000)
    } else {
      phoneError.value = res.data.error || '发送失败'
    }
  } catch (e) {
    phoneError.value = '网络错误'
  }
}

async function loginPhone() {
  if (!phone.value || !smsCode.value) {
    phoneError.value = '请输入手机号和验证码'
    return
  }
  loading.value = true
  phoneError.value = ''
  try {
    const res = await loginCellphone(phone.value, smsCode.value)
    if (res.data.status === 1 && res.data.data) {
      auth.setCredentials(res.data.data)
      auth.fetchUser()
      emit('close')
    } else {
      phoneError.value = res.data.error || '登录失败'
    }
  } catch (e) {
    phoneError.value = '网络错误'
  } finally {
    loading.value = false
  }
}
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

.modal-content {
  width: 400px;
  max-width: 90vw;
  padding: var(--spacing-2xl);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-lg);
  animation: modal-in 0.35s cubic-bezier(0.34, 1.56, 0.64, 1);
}
@keyframes modal-in {
  from { opacity: 0; transform: scale(0.94) translateY(8px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-xl);
}
.modal-title { font-size: var(--font-size-xl); font-weight: var(--font-weight-bold); }
.modal-reason {
  margin: -8px 0 var(--spacing-lg);
  padding: 10px 14px;
  font-size: var(--font-size-sm);
  color: var(--primary-deep);
  background: var(--surface-2);
  border: 1px solid var(--border);
  border-radius: var(--r-sm);
}
.close-btn {
  width: 32px; height: 32px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 50%; color: var(--text-2); transition: all var(--transition-fast);
}
.close-btn:hover { background: var(--surface-3); }

.tab-bar {
  display: flex; gap: var(--spacing-xs);
  margin-bottom: var(--spacing-xl);
  background: var(--surface-3);
  border-radius: var(--radius-md);
  padding: 3px;
}
.tab {
  flex: 1; padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--font-size-sm); font-weight: var(--font-weight-medium);
  color: var(--text-2); border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}
.tab.active {
  background: var(--surface);
  color: var(--text-1);
  box-shadow: var(--shadow-sm);
}

.tab-content { min-height: 200px; }

.qr-container {
  display: flex; flex-direction: column; align-items: center;
  padding: var(--spacing-xl) 0;
}
.qr-image img { width: 200px; height: 200px; border-radius: var(--radius-md); }
.qr-error p { color: var(--text-2); }

.form-group { margin-bottom: var(--spacing-lg); }
.form-group label {
  display: block;
  font-size: var(--font-size-sm); font-weight: var(--font-weight-medium);
  color: var(--text-2); margin-bottom: var(--spacing-xs);
}
.input {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  border: 1px solid var(--border-strong);
  border-radius: var(--radius-md);
  background: var(--surface-2);
  font-size: var(--font-size-base);
  outline: none;
  transition: border var(--transition-fast);
}
.input:focus { border-color: var(--primary); }

.code-group { display: flex; gap: var(--spacing-sm); }
.code-group .input { flex: 1; }
.code-btn { white-space: nowrap; flex-shrink: 0; }

.login-btn { width: 100%; padding: var(--spacing-md); font-size: var(--font-size-base); }
.login-btn:disabled { opacity: .6; cursor: not-allowed; }

.error-msg { color: var(--like); font-size: var(--font-size-sm); margin-top: var(--spacing-sm); text-align: center; }
</style>
