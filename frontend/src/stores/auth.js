import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import client from '../api/client'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const refreshToken = ref(localStorage.getItem('refresh_token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAuthenticated = computed(() => !!token.value)

  function setSession(newToken, newRefresh, newUser) {
    token.value = newToken || ''
    refreshToken.value = newRefresh || ''
    user.value = newUser || null
    if (newToken) localStorage.setItem('token', newToken)
    else localStorage.removeItem('token')
    if (newRefresh) localStorage.setItem('refresh_token', newRefresh)
    else localStorage.removeItem('refresh_token')
    if (newUser) localStorage.setItem('user', JSON.stringify(newUser))
    else localStorage.removeItem('user')
  }

  function login(newToken, newUser, newRefresh) {
    setSession(newToken, newRefresh, newUser)
  }

  function logout() {
    setSession('', '', null)
  }

  async function loginRequest(email, password) {
    const { data } = await client.post('/api/auth/login', { email, password })
    const payload = data.data || data
    const at = payload.access_token || payload.token
    const rt = payload.refresh_token || ''
    const u = payload.user || null
    login(at, u, rt)
    return payload
  }

  async function registerRequest(name, email, password) {
    const { data } = await client.post('/api/auth/register', { name, email, password })
    const payload = data.data || data
    const at = payload.access_token || payload.token
    const rt = payload.refresh_token || ''
    const u = payload.user || null
    login(at, u, rt)
    return payload
  }

  async function fetchMe() {
    const { data } = await client.get('/api/auth/me')
    const u = data.data || data
    user.value = u
    localStorage.setItem('user', JSON.stringify(u))
    return u
  }

  async function refresh() {
    const rt = refreshToken.value || localStorage.getItem('refresh_token')
    if (!rt) throw new Error('no refresh token')
    const { data } = await client.post('/api/auth/refresh', { refresh_token: rt })
    const payload = data.data || data
    token.value = payload.access_token
    if (payload.refresh_token) {
      refreshToken.value = payload.refresh_token
      localStorage.setItem('refresh_token', payload.refresh_token)
    }
    localStorage.setItem('token', payload.access_token)
    return payload
  }

  return { token, refreshToken, user, isAuthenticated, login, logout, setSession, loginRequest, registerRequest, fetchMe, refresh }
})
