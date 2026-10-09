<script setup>
import { RouterView, RouterLink } from 'vue-router'
import Toast from 'primevue/toast'
import {
  BedDouble,
  Heart,
  Award,
  Moon,
  Sun,
  LayoutDashboard,
  CalendarCheck,
  Compass,
  Star,
  Sparkles,
  Menu,
  X,
  LogOut,
  User
} from 'lucide-vue-next'
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale } from './i18n/index.js'
import { useTheme } from './composables/useTheme.js'
import client from './api/client'
import { useAuthStore } from './stores/auth'

const { locale, t } = useI18n()
const { isDark, toggleTheme } = useTheme()
const auth = useAuthStore()
const mobileMenuOpen = ref(false)
const wishlistCount = ref(0)
const loyaltyPoints = ref(null)

const isStaff = computed(() => {
  const role = auth.user?.role
  return ['owner', 'manager', 'receptionist'].includes(role)
})

function toggleLocale() {
  const next = locale.value === 'en' ? 'id' : 'en'
  setLocale(next)
  locale.value = next
}

async function fetchWishlistCount() {
  if (!auth.isAuthenticated) {
    wishlistCount.value = 0
    return
  }
  try {
    const { data } = await client.get('/api/wishlist')
    const list = data.data || data
    wishlistCount.value = Array.isArray(list) ? list.length : 0
  } catch {
    wishlistCount.value = 0
  }
}

async function fetchLoyalty() {
  if (!auth.isAuthenticated) {
    loyaltyPoints.value = null
    return
  }
  try {
    const { data } = await client.get('/api/loyalty/points')
    const payload = data.data || data
    loyaltyPoints.value = payload.points ?? payload.loyalty_points ?? payload.total ?? 0
  } catch {
    try {
      const { data } = await client.get('/api/loyalty')
      const p = data.data || data
      loyaltyPoints.value = p.points ?? 0
    } catch {
      loyaltyPoints.value = null
    }
  }
}

function handleLogout() {
  auth.logout()
  wishlistCount.value = 0
  loyaltyPoints.value = null
}

onMounted(() => {
  fetchWishlistCount()
  fetchLoyalty()
})

watch(() => auth.isAuthenticated, () => {
  fetchWishlistCount()
  fetchLoyalty()
})

if (typeof window !== 'undefined') {
  window.addEventListener('wishlist:updated', fetchWishlistCount)
  window.addEventListener('loyalty:updated', fetchLoyalty)
}
</script>

<template>
  <div class="min-h-screen flex flex-col bg-[#FDF6EC] dark:bg-[#1A3A4A] text-[#1F2937] dark:text-[#FDF6EC] transition-colors duration-200">
    <!-- Navbar Premium Ubud (z-[1050] to always float above Leaflet maps) -->
    <nav class="sticky top-0 z-[1050] bg-[#1A3A4A] dark:bg-[#0F2A36] text-white shadow-md border-b border-white/10 dark:border-[#4A5568] backdrop-blur-md">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 flex items-center justify-between h-18 sm:h-20">
        <!-- Logo Resort -->
        <RouterLink to="/" class="flex items-center gap-2.5 font-display font-bold text-xl sm:text-2xl tracking-tight group">
          <span class="bg-white text-[#1A3A4A] rounded-xl p-2 shadow-sm group-hover:scale-105 transition-transform flex items-center justify-center">
            <BedDouble class="w-5 h-5 text-[#8B5A2B]" />
          </span>
          <span style="font-family:'Playfair Display',serif">
            xyz<span class="text-[#C9A86A]">hotel</span>
          </span>
        </RouterLink>

        <!-- Menu Utama Tengah (Desktop) -->
        <div class="hidden md:flex items-center gap-7 text-sm font-medium">
          <a href="/#rooms" class="hover:text-[#C9A86A] transition-colors py-1 flex items-center gap-1.5">
            <span>Kamar</span>
          </a>
          <a href="/#amenities" class="hover:text-[#C9A86A] transition-colors py-1 flex items-center gap-1.5">
            <span>Fasilitas</span>
          </a>
          <a href="/#location" class="hover:text-[#C9A86A] transition-colors py-1 flex items-center gap-1.5">
            <span>Lokasi</span>
          </a>
          <a href="/#reviews" class="hover:text-[#C9A86A] transition-colors py-1 flex items-center gap-1.5">
            <span>Ulasan</span>
          </a>
          <RouterLink
            v-if="auth.isAuthenticated"
            to="/bookings"
            class="hover:text-[#C9A86A] transition-colors py-1 flex items-center gap-1"
          >
            <span>Pesanan Saya</span>
          </RouterLink>
        </div>

        <!-- Tombol Aksi Kanan -->
        <div class="flex items-center gap-2 sm:gap-3">
          <!-- Ganti Bahasa -->
          <button
            @click="toggleLocale"
            class="w-9 h-9 rounded-full bg-white/10 hover:bg-white/20 border border-white/20 flex items-center justify-center text-sm transition-all active:scale-95 shrink-0"
            :title="locale === 'en' ? 'Ganti ke Bahasa Indonesia' : 'Switch to English'"
            :aria-label="locale === 'en' ? 'Ganti ke Bahasa Indonesia' : 'Switch to English'"
          >
            {{ locale === 'en' ? '🇬🇧' : '🇮🇩' }}
          </button>

          <!-- Toggle Mode Gelap/Terang -->
          <button
            @click="toggleTheme"
            class="w-9 h-9 rounded-full bg-[#8B5A2B] hover:bg-[#704620] text-white flex items-center justify-center shadow-sm border border-white/20 transition-all active:scale-95 shrink-0"
            :title="isDark ? 'Mode Terang' : 'Mode Gelap'"
            :aria-label="isDark ? 'Mode Terang' : 'Mode Gelap'"
          >
            <Sun v-if="isDark" class="w-4 h-4 text-amber-300" />
            <Moon v-else class="w-4 h-4 text-white" />
          </button>

          <!-- Wishlist Favorit -->
          <RouterLink
            to="/wishlist"
            class="relative w-9 h-9 rounded-full bg-white/10 hover:bg-white/20 border border-white/20 flex items-center justify-center transition-all active:scale-95 shrink-0"
            title="Kamar Favorit"
            aria-label="Wishlist"
          >
            <Heart class="w-4 h-4" :class="wishlistCount > 0 ? 'fill-[#C9A86A] text-[#C9A86A]' : 'text-white'" />
            <span
              v-if="wishlistCount > 0"
              class="absolute -top-1 -right-1 bg-[#8B5A2B] text-white text-[10px] font-bold rounded-full min-w-[18px] h-[18px] flex items-center justify-center px-1 border border-[#1A3A4A]"
            >
              {{ wishlistCount > 99 ? '99+' : wishlistCount }}
            </span>
          </RouterLink>

          <!-- Poin Loyalitas (Jika Login) -->
          <div
            v-if="auth.isAuthenticated && loyaltyPoints !== null"
            class="hidden lg:flex items-center gap-1.5 bg-[#C9A86A] text-[#1A3A4A] rounded-full px-3 py-1.5 text-xs font-bold shadow-sm"
            title="Poin loyalitas aktif (100 poin = diskon Rp 100.000)"
          >
            <Award class="w-3.5 h-3.5" />
            <span>{{ loyaltyPoints }} pts</span>
          </div>

          <!-- Tombol Khusus Staf / Dashboard Admin -->
          <RouterLink
            v-if="auth.isAuthenticated && isStaff"
            to="/admin"
            class="hidden sm:inline-flex items-center gap-1.5 text-xs bg-white/15 hover:bg-white/25 border border-white/30 rounded-full px-3.5 py-1.5 text-white font-medium transition-colors"
            title="Buka Dashboard Operasional Resort"
          >
            <LayoutDashboard class="w-3.5 h-3.5 text-[#C9A86A]" />
            <span>Dashboard</span>
          </RouterLink>

          <!-- Status Akun (Login / Logout) -->
          <template v-if="!auth.isAuthenticated">
            <RouterLink
              to="/login"
              class="hidden sm:inline-block text-xs sm:text-sm font-medium hover:text-[#C9A86A] transition-colors px-2 py-1"
            >
              Masuk
            </RouterLink>
          </template>
          <template v-else>
            <button
              @click="handleLogout"
              class="hidden sm:inline-flex items-center gap-1 text-xs text-white/70 hover:text-white hover:bg-white/10 rounded-full px-2.5 py-1.5 transition-colors"
              title="Keluar dari akun"
            >
              <LogOut class="w-3.5 h-3.5" />
            </button>
          </template>

          <!-- CTA Pesan Sekarang -->
          <a
            href="/#rooms"
            class="bg-[#8B5A2B] hover:bg-[#704620] text-white text-xs sm:text-sm font-semibold px-4 sm:px-5 py-2 sm:py-2.5 rounded-xl transition-all shadow-sm hover:shadow-md active:scale-95 flex items-center gap-1.5"
          >
            <span>Pesan Kamar</span>
          </a>

          <!-- Hamburger Button (Mobile) -->
          <button
            @click="mobileMenuOpen = !mobileMenuOpen"
            class="md:hidden p-2 rounded-xl bg-white/10 hover:bg-white/20 transition-colors ml-1"
            aria-label="Buka Menu"
          >
            <Menu v-if="!mobileMenuOpen" class="w-5 h-5 text-white" />
            <X v-else class="w-5 h-5 text-white" />
          </button>
        </div>
      </div>

      <!-- Mobile Dropdown Menu -->
      <div
        v-if="mobileMenuOpen"
        class="md:hidden bg-[#1A3A4A] dark:bg-[#0F2A36] border-t border-white/10 px-4 py-4 space-y-3 shadow-xl"
        @click="mobileMenuOpen = false"
      >
        <a href="/#rooms" class="block text-sm text-white/90 hover:text-[#C9A86A] py-1.5">Kamar & Tipe Kamar</a>
        <a href="/#amenities" class="block text-sm text-white/90 hover:text-[#C9A86A] py-1.5">Fasilitas & Layanan Tambahan</a>
        <a href="/#location" class="block text-sm text-white/90 hover:text-[#C9A86A] py-1.5">Lokasi & Wisata Sekitar</a>
        <a href="/#reviews" class="block text-sm text-white/90 hover:text-[#C9A86A] py-1.5">Ulasan Tamu</a>
        <div class="h-px bg-white/10 my-2"></div>
        <RouterLink v-if="auth.isAuthenticated" to="/bookings" class="block text-sm text-[#C9A86A] py-1.5">
          Pesanan Saya
        </RouterLink>
        <RouterLink v-if="auth.isAuthenticated && isStaff" to="/admin" class="block text-sm text-[#C9A86A] py-1.5">
          Dashboard Staf / Admin
        </RouterLink>
        <RouterLink v-if="!auth.isAuthenticated" to="/login" class="block text-sm text-white font-semibold py-1.5">
          Masuk ke Akun
        </RouterLink>
        <button v-else @click="handleLogout" class="block text-sm text-red-300 py-1.5 text-left w-full">
          Keluar (Logout)
        </button>
      </div>
    </nav>

    <!-- Main Content Area -->
    <main class="flex-1 bg-[#FDF6EC] dark:bg-[#1A3A4A] transition-colors">
      <RouterView />
    </main>

    <!-- Footer Resort Ubud -->
    <footer class="bg-[#1A3A4A] dark:bg-[#0F2A36] text-white/80 text-sm border-t border-white/10 dark:border-[#4A5568]">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 flex flex-col md:flex-row items-center justify-between gap-4">
        <div class="flex items-center gap-2">
          <span class="bg-white/10 text-white rounded-lg p-1.5">
            <BedDouble class="w-4 h-4 text-[#C9A86A]" />
          </span>
          <p class="font-display font-medium text-white" style="font-family:'Playfair Display',serif">
            xyz<span class="text-[#C9A86A]">hotel</span> · Ubud, Bali
          </p>
        </div>
        <p class="text-xs text-white/60 text-center md:text-left">
          Jl. Suweta No. 8B, Ubud, Gianyar, Bali 80571 · WhatsApp: +62 812-3456-7890
        </p>
        <p class="text-xs text-[#C9A86A] font-medium">
          Kenyamanan & Kehangatan Otentik di Jantung Ubud
        </p>
      </div>
    </footer>

    <Toast />
  </div>
</template>
