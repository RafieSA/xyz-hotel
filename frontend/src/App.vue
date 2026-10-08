<script setup>
import { RouterView, RouterLink } from 'vue-router'
import Toast from 'primevue/toast'
import { BedDouble, Heart } from 'lucide-vue-next'
import { ref, onMounted, watch } from 'vue'
import client from './api/client'
import { useAuthStore } from './stores/auth'

const auth = useAuthStore()
const wishlistCount = ref(0)
async function fetchWishlistCount(){
  if(!auth.isAuthenticated){ wishlistCount.value=0; return }
  try{
    const { data } = await client.get('/api/wishlist')
    const list = data.data || data
    wishlistCount.value = Array.isArray(list) ? list.length : 0
  }catch{ wishlistCount.value=0 }
}
onMounted(fetchWishlistCount)
watch(()=>auth.isAuthenticated, fetchWishlistCount)
if(typeof window !== 'undefined'){
  window.addEventListener('wishlist:updated', fetchWishlistCount)
}
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <nav class="sticky top-0 z-40 bg-[#1A3A4A] text-white shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 flex items-center justify-between h-16">
        <RouterLink to="/" class="flex items-center gap-2 font-display font-bold text-xl tracking-tight">
          <span class="bg-white text-[#1A3A4A] rounded-lg p-1.5"><BedDouble class="w-5 h-5" /></span>
          xyz<span class="text-[#C9A86A]">hotel</span>
        </RouterLink>
        <div class="hidden md:flex items-center gap-6 text-sm font-medium">
          <RouterLink to="/" class="hover:text-[#C9A86A] transition-colors">Rooms</RouterLink>
          <RouterLink to="/bookings" class="hover:text-[#C9A86A] transition-colors">My Bookings</RouterLink>
          <RouterLink to="/wishlist" class="hover:text-[#C9A86A] transition-colors flex items-center gap-1">Wishlist</RouterLink>
          <a href="#" class="hover:text-[#C9A86A] transition-colors">Amenities</a>
          <a href="#" class="hover:text-[#C9A86A] transition-colors">Contact</a>
        </div>
        <div class="flex items-center gap-2 sm:gap-3">
          <RouterLink to="/wishlist" class="relative p-2 rounded-full hover:bg-white/10 transition-colors" aria-label="Wishlist">
            <Heart class="w-5 h-5" :class="wishlistCount>0 ? 'fill-[#C9A86A] text-[#C9A86A]' : 'text-white'" />
            <span v-if="wishlistCount>0" class="absolute -top-1 -right-1 bg-[#8B5A2B] text-white text-[10px] font-bold rounded-full min-w-[18px] h-[18px] flex items-center justify-center px-1 border border-white">{{ wishlistCount > 99 ? '99+' : wishlistCount }}</span>
          </RouterLink>
          <RouterLink to="/bookings" class="hidden sm:inline text-xs border border-white/30 rounded-full px-3 py-1.5 hover:bg-white hover:text-[#1A3A4A] transition-colors">My Bookings</RouterLink>
          <RouterLink to="/login" class="hidden sm:inline text-sm hover:text-[#C9A86A] transition-colors">Sign In</RouterLink>
          <RouterLink to="/admin" class="text-xs border border-white/30 rounded-full px-3 py-1 hover:bg-white hover:text-[#1A3A4A] transition-colors">Dashboard</RouterLink>
          <RouterLink to="/" class="bg-[#8B5A2B] hover:bg-[#6F4620] text-white text-sm font-semibold px-5 py-2.5 rounded-xl transition-colors">Book Your Stay</RouterLink>
        </div>
      </div>
    </nav>
    <main class="flex-1"><RouterView /></main>
    <footer class="bg-[#1A3A4A] text-white/80 text-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 flex flex-col md:flex-row justify-between gap-4">
        <p>© 2026 xyz-hotel · Jl. Hangat No. 8B, Ubud, Bali · WA 0812-3456-7890</p>
        <p class="text-[#C9A86A]">Warm brown #8B5A2B · Aura WarmAura</p>
      </div>
    </footer>
    <Toast />
  </div>
</template>

