import axios from 'axios'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

// Attach JWT from localStorage
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Extract a helpful message from API error responses
export function getApiErrorMessage(err, fallback = 'Something went wrong. Please try again') {
  if (err?.response?.data?.message) return err.response.data.message
  if (err?.response?.data?.error) return err.response.data.error
  if (err?.message) return err.message
  return fallback
}

export function getApiErrorDetails(err) {
  return err?.response?.data?.details || null
}

// Handle auth errors with clear messaging + rate limit 429
client.interceptors.response.use(
  (res) => res,
  (err) => {
    const status = err.response?.status
    if (status === 429) {
      err.userMessage = err.response?.data?.message || 'Too many requests, try again in a minute'
    } else if (status === 401) {
      err.userMessage = 'Your session expired. Please sign in again'
    } else if (status === 403) {
      err.userMessage = err.response?.data?.message || 'You do not have permission for this'
    } else if (err.response?.data?.message) {
      err.userMessage = err.response.data.message
    } else if (!err.response) {
      err.userMessage = 'Cannot reach the server. Check your connection and try again'
    }
    return Promise.reject(err)
  }
)

export default client
