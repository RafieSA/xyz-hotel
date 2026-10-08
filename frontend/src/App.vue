<script setup>
import { RouterView, RouterLink } from 'vue-router'
import Toast from 'primevue/toast'
import { BedDouble } from 'lucide-vue-next'
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
          <RouterLink to="/" class="hover:text-[#C9A86A] transition-colors">Kamar</RouterLink>
          <a href="#" class="hover:text-[#C9A86A] transition-colors">Fasilitas</a>
          <a href="#" class="hover:text-[#C9A86A] transition-colors">Kontak</a>
        </div>
        <div class="flex items-center gap-3">
          <RouterLink to="/login" class="hidden sm:inline text-sm hover:text-[#C9A86A] transition-colors">Masuk</RouterLink>
          <RouterLink to="/admin" class="text-xs border border-white/30 rounded-full px-3 py-1 hover:bg-white hover:text-[#1A3A4A] transition-colors">Backoffice</RouterLink>
          <RouterLink to="/" class="bg-[#8B5A2B] hover:bg-[#6F4620] text-white text-sm font-semibold px-5 py-2.5 rounded-xl transition-colors">Booking Sekarang</RouterLink>
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
