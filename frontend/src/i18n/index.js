import { createI18n } from 'vue-i18n'
import en from '../locales/en.json'
import id from '../locales/id.json'

const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('locale') : null
const locale = stored === 'id' || stored === 'en' ? stored : 'en'

const i18n = createI18n({
  legacy: false,
  locale,
  fallbackLocale: 'en',
  messages: { en, id }
})

export function setLocale(newLocale) {
  i18n.global.locale.value = newLocale
  if (typeof localStorage !== 'undefined') localStorage.setItem('locale', newLocale)
  if (typeof document !== 'undefined') document.documentElement.lang = newLocale
}

export default i18n
