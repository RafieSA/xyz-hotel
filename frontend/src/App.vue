<script setup>
import { RouterView, RouterLink } from 'vue-router'
import Toast from 'primevue/toast'
import { BedDouble, Heart, Award, Moon, Sun } from 'lucide-vue-next'
import { ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale } from './i18n/index.js'
import { useTheme } from './composables/useTheme.js'
import client from './api/client'
import { useAuthStore } from './stores/auth'

const { locale, t } = useI18n()
const { isDark, toggleTheme } = useTheme()
const auth = useAuthStore()
const wishlistCount = ref(0)
const loyaltyPoints = ref(null)

function toggleLocale() {
  const next = locale.value === 'en' ? 'id' : 'en'
  setLocale(next)
  locale.value = next
}

async function fetchWishlistCount(){
  if(!auth.isAuthenticated){ wishlistCount.value=0; return }
  try{
    const { data } = await client.get('/api/wishlist')
    const list = data.data || data
    wishlistCount.value = Array.isArray(list) ? list.length : 0
  }catch{ wishlistCount.value=0 }
}
async function fetchLoyalty(){
  if(!auth.isAuthenticated){ loyaltyPoints.value=null; return }
  try{
    const { data } = await client.get('/api/loyalty/points')
    const payload = data.data || data
    loyaltyPoints.value = payload.points ?? payload.loyalty_points ?? payload.total ?? 0
  }catch{
    try{
      const { data } = await client.get('/api/loyalty')
      const p = data.data || data
      loyaltyPoints.value = p.points ?? 0
    }catch{ loyaltyPoints.value = null }
  }
}
onMounted(()=>{ fetchWishlistCount(); fetchLoyalty() })
watch(()=>auth.isAuthenticated, ()=>{ fetchWishlistCount(); fetchLoyalty() })
if(typeof window !== 'undefined'){
  window.addEventListener('wishlist:updated', fetchWishlistCount)
  window.addEventListener('loyalty:updated', fetchLoyalty)
}
</script>

<template>
  <div class="min-h-screen flex flex-col bg-[#FDF6EC] dark:bg-[#1A3A4A] text-[#1F2937] dark:text-[#FDF6EC] transition-colors duration-200">
    <nav class="sticky top-0 z-40 bg-[#1A3A4A] dark:bg-[#0F2A36] text-white shadow-sm border-b border-transparent dark:border-[#4A5568]">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 flex items-center justify-between h-16">
        <RouterLink to="/" class="flex items-center gap-2 font-display font-bold text-xl tracking-tight">
          <span class="bg-white text-[#1A3A4A] rounded-lg p-1.5"><BedDouble class="w-5 h-5" /></span>
          xyz<span class="text-[#C9A86A]">hotel</span>
        </RouterLink>
        <div class="hidden md:flex items-center gap-6 text-sm font-medium">
          <RouterLink to="/" class="hover:text-[#C9A86A] transition-colors">{{ t('nav.rooms') }}</RouterLink>
          <RouterLink to="/bookings" class="hover:text-[#C9A86A] transition-colors">{{ t('nav.bookings') }}</RouterLink>
          <RouterLink to="/wishlist" class="hover:text-[#C9A86A] transition-colors flex items-center gap-1">{{ t('nav.wishlist') }}</RouterLink>
          <a href="#" class="hover:text-[#C9A86A] transition-colors">{{ t('nav.amenities') }}</a>
          <a href="#" class="hover:text-[#C9A86A] transition-colors">{{ t('nav.contact') }}</a>
        </div>
        <div class="flex items-center gap-2 sm:gap-2">
          <button @click="toggleLocale" class="w-9 h-9 rounded-full bg-white/10 hover:bg-white/20 border border-white/20 flex items-center justify-center text-sm transition-colors" :title="locale==='en' ? 'Switch to ID' : 'Switch to EN'" :aria-label="locale==='en' ? 'Switch to Indonesian' : 'Switch to English'">
            {{ locale==='en' ? '🇬🇧' : '🇮🇩' }}
          </button>
          <button @click="toggleTheme" class="w-9 h-9 rounded-full bg-[#8B5A2B] hover:bg-[#6F4620] text-white flex items-center justify-center shadow-sm border border-white/20 transition-colors" :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'" :title="isDark ? 'Light mode' : 'Dark mode'">
            <Moon v-if="!isDark" class="w-4 h-4" />
            <Sun v-else class="w-4 h-4" />
          </button>
          <RouterLink to="/wishlist" class="relative p-2 rounded-full hover:bg-white/10 transition-colors" aria-label="Wishlist">
            <Heart class="w-5 h-5" :class="wishlistCount>0 ? 'fill-[#C9A86A] text-[#C9A86A]' : 'text-white'" />
            <span v-if="wishlistCount>0" class="absolute -top-1 -right-1 bg-[#8B5A2B] text-white text-[10px] font-bold rounded-full min-w-[18px] h-[18px] flex items-center justify-center px-1 border border-white">{{ wishlistCount > 99 ? '99+' : wishlistCount }}</span>
          </RouterLink>
          <div v-if="auth.isAuthenticated && loyaltyPoints!==null" class="hidden sm:flex items-center gap-1.5 bg-[#C9A86A] text-[#1A3A4A] rounded-full px-3 py-1.5 text-xs font-bold shadow-sm" title="Loyalty points: 10 per night, 100 points = IDR 100k">
            <Award class="w-3.5 h-3.5" /> {{ loyaltyPoints }} pts
          </div>
          <div v-else-if="auth.isAuthenticated" class="hidden sm:flex items-center gap-1.5 bg-white/15 text-white border border-white/20 rounded-full px-3 py-1.5 text-xs font-semibold">
            <Award class="w-3.5 h-3.5 text-[#C9A86A]" /> {{ t('nav.loyalty') }}
          </div>
          <RouterLink to="/bookings" class="hidden sm:inline text-xs border border-white/30 rounded-full px-3 py-1.5 hover:bg-white hover:text-[#1A3A4A] transition-colors">{{ t('nav.bookings') }}</RouterLink>
          <RouterLink to="/login" class="hidden sm:inline text-sm hover:text-[#C9A86A] transition-colors">{{ t('nav.signin') }}</RouterLink>
          <RouterLink to="/admin" class="text-xs border border-white/30 rounded-full px-3 py-1 hover:bg-white hover:text-[#1A3A4A] transition-colors">{{ t('nav.dashboard') }}</RouterLink>
          <RouterLink to="/" class="bg-[#8B5A2B] hover:bg-[#6F4620] text-white text-sm font-semibold px-5 py-2.5 rounded-xl transition-colors">{{ t('cta.book') }}</RouterLink>
        </div>
      </div>
    </nav>
    <main class="flex-1 bg-[#FDF6EC] dark:bg-[#1A3A4A] transition-colors"><RouterView /></main>
    <footer class="bg-[#1A3A4A] dark:bg-[#0F2A36] text-white/80 text-sm border-t border-transparent dark:border-[#4A5568]">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 flex flex-col md:flex-row justify-between gap-4">
        <p>© 2026 xyz-hotel · Jl. Hangat No. 8B, Ubud, Bali · WA 0812-3456-7890</p>
        <p class="text-[#C9A86A]">Warm brown #8B5A2B · Aura WarmAura</p>
      </div>
    </footer>
    <Toast />
  </div>
</template>
