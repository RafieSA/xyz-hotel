import { ref } from 'vue'

const isDark = ref(false)

export function initTheme() {
  if (typeof window === 'undefined' || typeof document === 'undefined') return
  const stored = localStorage.getItem('theme')
  let dark = false
  if (stored === 'dark') dark = true
  else if (stored === 'light') dark = false
  else dark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
}

export function toggleTheme() {
  isDark.value = !isDark.value
  if (typeof document !== 'undefined') {
    document.documentElement.classList.toggle('dark', isDark.value)
  }
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  }
}

export function useTheme() {
  return { isDark, initTheme, toggleTheme }
}

export default useTheme
