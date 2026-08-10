import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { setAuth, clearAuth, getUserDetail } from '../utils/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref('')
  const userId = ref('')
  const dfid = ref('')
  const mid = ref('')
  const guid = ref('')
  const user = ref(null)
  const vipToken = ref('') // 概念版 VIP 播放 priv_url 需要
  const vipType = ref('')
  const isLoggedIn = computed(() => !!token.value && !!userId.value)

  // 从 localStorage 恢复登录状态
  function loadFromStorage() {
    const saved = localStorage.getItem('vibe_auth')
    if (saved) {
      try {
        const data = JSON.parse(saved)
        token.value = data.token || ''
        userId.value = data.user_id || ''
        dfid.value = data.dfid || ''
        mid.value = data.mid || ''
        guid.value = data.guid || ''
        user.value = data.user || null
        vipToken.value = data.vip_token || ''
        vipType.value = data.vip_type || ''
        applyAuth()
        return true
      } catch (e) {
        return false
      }
    }
    return false
  }

  function saveToStorage() {
    localStorage.setItem('vibe_auth', JSON.stringify({
      token: token.value,
      user_id: userId.value,
      dfid: dfid.value,
      mid: mid.value,
      guid: guid.value,
      user: user.value,
      vip_token: vipToken.value,
      vip_type: vipType.value
    }))
  }

  function applyAuth() {
    if (token.value) {
      setAuth(token.value, userId.value, dfid.value, mid.value, guid.value)
    } else {
      // 未登录时使用设备信息
      setAuth('', userId.value, dfid.value, mid.value, guid.value)
    }
  }

  function setCredentials(creds) {
    token.value = creds.token || ''
    userId.value = creds.user_id || ''
    // 设备信息：登录响应缺失时保留已有值（扫码登录只返回 token/userid）
    dfid.value = creds.dfid || dfid.value || ''
    mid.value = creds.mid || mid.value || ''
    guid.value = creds.guid || guid.value || ''
    user.value = creds.user || null
    vipToken.value = creds.vip_token || ''
    vipType.value = creds.vip_type || ''
    applyAuth()
    saveToStorage()
  }

  // 设备注册成功后保存设备标识（mid/guid/dfid 参与登录与 VIP 播放签名，必须持久化）
  function setDeviceInfo(device) {
    const d = device || {}
    // 后端 register_dev 返回字段为大写（MID/GUID/DFID），兼容两种写法
    mid.value = d.mid || d.MID || ''
    guid.value = d.guid || d.GUID || ''
    dfid.value = d.dfid || d.DFID || ''
    applyAuth()
    saveToStorage()
  }

  // 登录后拉取用户详情（昵称/头像等），扫码登录只返回凭证，需额外请求
  async function fetchUser() {
    if (!isLoggedIn.value) return null
    try {
      const res = await getUserDetail()
      if (res.data.status === 1 && res.data.data) {
        const d = res.data.data
        user.value = {
          ...(user.value || {}),
          ...d,
          id: d.id || d.user_id || user.value?.id || '',
          name: d.name || d.nickname || user.value?.name || '',
          avatar: d.avatar || d.user_pic || user.value?.avatar || '',
        }
        saveToStorage()
        return user.value
      }
    } catch (e) {
      console.error('获取用户信息失败:', e)
    }
    return null
  }

  function logout() {
    token.value = ''
    userId.value = ''
    user.value = null
    clearAuth()
    localStorage.removeItem('vibe_auth')
  }

  // 初始化时尝试恢复
  const restored = loadFromStorage()
  // 恢复登录态后拉取用户信息（包含 is_vip）
  if (restored && isLoggedIn.value) {
    fetchUser()
  }

  return {
    token, userId, dfid, mid, guid, user, isLoggedIn, vipToken, vipType,
    loadFromStorage, setCredentials, setDeviceInfo, fetchUser, logout
  }
})