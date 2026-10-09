<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { ChevronLeft, ChevronRight, X, Maximize2, Sparkles, Image as ImageIcon } from 'lucide-vue-next'

const props = defineProps({
  rooms: {
    type: Array,
    required: true,
  },
  galleryImages: {
    type: Object,
    default: () => ({}),
  },
  fallbackGallery: {
    type: Object,
    required: true,
  },
})

// Tab kamar yang sedang aktif dilihat (default kamar pertama atau kamar Deluxe)
const activeRoomId = ref(props.rooms[1]?.id || props.rooms[0]?.id || 1)

// Foto utama yang sedang disorot di galeri
const activePhotoIndex = ref(0)

// Data kamar yang aktif saat ini
const activeRoom = computed(() => {
  return props.rooms.find(r => r.id === activeRoomId.value) || props.rooms[0]
})

// Daftar foto kamar yang aktif
const currentPhotos = computed(() => {
  const custom = props.galleryImages[activeRoomId.value]
  if (custom && custom.length) return custom
  return props.fallbackGallery[activeRoomId.value] || []
})

// State Lightbox (Layar Penuh)
const lightboxOpen = ref(false)
const lightboxIndex = ref(0)

function selectRoom(id) {
  activeRoomId.value = id
  activePhotoIndex.value = 0
}

function openLightbox(index) {
  lightboxIndex.value = index
  lightboxOpen.value = true
}

function closeLightbox() {
  lightboxOpen.value = false
}

function nextPhoto() {
  if (!currentPhotos.value.length) return
  lightboxIndex.value = (lightboxIndex.value + 1) % currentPhotos.value.length
}

function prevPhoto() {
  if (!currentPhotos.value.length) return
  lightboxIndex.value = (lightboxIndex.value - 1 + currentPhotos.value.length) % currentPhotos.value.length
}

// Navigasi Keyboard ramah aksesibilitas
function onKeyDown(e) {
  if (!lightboxOpen.value) return
  if (e.key === 'Escape') closeLightbox()
  if (e.key === 'ArrowRight') nextPhoto()
  if (e.key === 'ArrowLeft') prevPhoto()
}

onMounted(() => {
  window.addEventListener('keydown', onKeyDown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeyDown)
})
</script>

<template>
  <section class="max-w-7xl mx-auto px-4 sm:px-6 py-12 md:py-16">
    <!-- Header Section -->
    <div class="text-center max-w-2xl mx-auto mb-10">
      <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#FDF6EC] text-[#8B5A2B] text-xs font-semibold uppercase tracking-wider mb-2">
        <Sparkles class="w-3.5 h-3.5 text-[#C9A86A]" />
        Resort Ambience
      </div>
      <h2 class="font-display font-bold text-3xl md:text-4xl text-[#1A3A4A] dark:text-[#FDF6EC]" style="font-family:'Playfair Display',serif">
        Sudut Kehangatan Resor
      </h2>
      <div class="h-0.5 w-16 bg-[#C9A86A] mx-auto mt-3"></div>
      <p class="text-sm md:text-base text-[#6B7280] dark:text-gray-300 mt-3">
        Dirancang dengan kayu jati alami dan sentuhan khas Ubud yang menenangkan. Pilih tipe kamar untuk melihat setiap sudutnya.
      </p>
    </div>

    <!-- Tab Pemilih Tipe Kamar (Scrollable di Layar HP) -->
    <div class="flex items-center justify-start sm:justify-center gap-2 overflow-x-auto pb-4 mb-8 no-scrollbar">
      <button
        v-for="r in rooms"
        :key="'tab-' + r.id"
        type="button"
        @click="selectRoom(r.id)"
        class="flex items-center gap-2 px-5 py-2.5 rounded-full text-xs md:text-sm font-semibold transition-all whitespace-nowrap cursor-pointer shadow-sm"
        :class="activeRoomId === r.id
          ? 'bg-[#8B5A2B] text-white shadow-md scale-105'
          : 'bg-white dark:bg-[#2D3748] text-[#1A3A4A] dark:text-gray-200 border border-[#E5E7EB] dark:border-gray-700 hover:border-[#8B5A2B]/50'"
      >
        <component :is="r.icon" class="w-4 h-4" :class="activeRoomId === r.id ? 'text-[#C9A86A]' : 'text-[#8B5A2B]'" />
        <span>{{ r.type }} Room</span>
      </button>
    </div>

    <!-- Showcase Grid: 1 Foto Besar + 4 Thumbnail Pendamping -->
    <div class="bg-white dark:bg-[#2D3748] rounded-3xl p-4 md:p-6 border border-[#E5E7EB] dark:border-gray-700 shadow-md">
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-4 items-stretch">
        <!-- Foto Utama Besar (8 Kolom di Layar Lebar) -->
        <div class="lg:col-span-8 relative rounded-2xl overflow-hidden group bg-[#FDF6EC] aspect-[16/10] shadow-sm">
          <img
            :src="currentPhotos[activePhotoIndex] || fallbackGallery[activeRoomId][0]"
            :alt="`${activeRoom?.type} main view`"
            class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105 cursor-pointer"
            @click="openLightbox(activePhotoIndex)"
          />
          <div class="absolute inset-0 bg-gradient-to-t from-black/50 via-transparent to-transparent pointer-events-none"></div>

          <!-- Keterangan & Tombol Layar Penuh di Bawah Foto Utama -->
          <div class="absolute bottom-4 left-4 right-4 flex items-center justify-between text-white">
            <div>
              <span class="text-xs font-semibold px-2.5 py-1 rounded-full bg-white/20 backdrop-blur-md">
                {{ activeRoom?.type }} Room · Sudut {{ activePhotoIndex + 1 }} dari {{ currentPhotos.length }}
              </span>
            </div>
            <button
              type="button"
              @click="openLightbox(activePhotoIndex)"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-white/90 hover:bg-white text-[#1A3A4A] text-xs font-semibold shadow-md transition"
              aria-label="Lihat foto layar penuh"
            >
              <Maximize2 class="w-3.5 h-3.5 text-[#8B5A2B]" />
              <span class="hidden sm:inline">Layar Penuh</span>
            </button>
          </div>
        </div>

        <!-- 4 Thumbnail di Samping Foto Utama (4 Kolom di Layar Lebar) -->
        <div class="lg:col-span-4 grid grid-cols-2 lg:grid-cols-2 gap-3">
          <button
            v-for="(photo, idx) in currentPhotos.slice(0, 4)"
            :key="'thumb-' + idx"
            type="button"
            @click="activePhotoIndex = idx"
            class="relative rounded-xl overflow-hidden aspect-[4/3] border-2 transition-all cursor-pointer group"
            :class="activePhotoIndex === idx
              ? 'border-[#8B5A2B] shadow-md ring-2 ring-[#8B5A2B]/20'
              : 'border-transparent hover:border-[#8B5A2B]/40 opacity-80 hover:opacity-100'"
          >
            <img
              :src="photo"
              :alt="`${activeRoom?.type} sudut ${idx + 1}`"
              class="w-full h-full object-cover group-hover:scale-105 transition duration-300"
            />
            <div
              v-if="activePhotoIndex === idx"
              class="absolute inset-0 bg-[#8B5A2B]/10 pointer-events-none"
            ></div>
            <span class="absolute bottom-1.5 right-1.5 text-[10px] font-bold bg-black/60 text-white px-1.5 py-0.5 rounded backdrop-blur-sm">
              0{{ idx + 1 }}
            </span>
          </button>
        </div>
      </div>

      <!-- Keterangan Ramah Pengguna -->
      <div class="mt-4 pt-4 border-t border-[#F3F4F6] dark:border-gray-700 flex flex-col sm:flex-row items-center justify-between text-xs text-[#6B7280] dark:text-gray-400 gap-2">
        <span class="flex items-center gap-1.5">
          <ImageIcon class="w-4 h-4 text-[#8B5A2B]" />
          Klik foto mana saja untuk menjelajah dalam resolusi penuh
        </span>
        <span>Semua foto asli diambil di lokasi XYZ Hotel Ubud</span>
      </div>
    </div>

    <!-- Modal Lightbox (Layar Penuh) yang Mewah -->
    <Teleport to="body">
      <div
        v-if="lightboxOpen"
        class="fixed inset-0 z-50 bg-black/90 backdrop-blur-md flex flex-col items-center justify-between p-4 md:p-6"
        @click.self="closeLightbox"
      >
        <!-- Header Lightbox: Judul Kamar & Tombol Tutup -->
        <div class="w-full max-w-5xl flex items-center justify-between text-white py-2">
          <div>
            <h3 class="text-base md:text-lg font-bold text-white flex items-center gap-2">
              <span class="text-[#C9A86A]">{{ activeRoom?.type }} Room</span>
              <span class="text-xs text-gray-400 font-normal">({{ lightboxIndex + 1 }} / {{ currentPhotos.length }})</span>
            </h3>
          </div>
          <button
            type="button"
            @click="closeLightbox"
            class="w-10 h-10 rounded-full bg-white/10 hover:bg-white/20 text-white flex items-center justify-center transition cursor-pointer"
            aria-label="Tutup foto layar penuh"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Area Gambar Utama Lightbox + Tombol Panah -->
        <div class="relative w-full max-w-5xl flex-1 flex items-center justify-center my-4 overflow-hidden">
          <!-- Tombol Kiri -->
          <button
            type="button"
            @click="prevPhoto"
            class="absolute left-2 md:left-4 z-10 w-11 h-11 rounded-full bg-black/60 hover:bg-black/90 text-white flex items-center justify-center shadow-lg transition border border-white/20"
            aria-label="Foto sebelumnya"
          >
            <ChevronLeft class="w-6 h-6" />
          </button>

          <!-- Foto Resolusi Layar Penuh -->
          <img
            :src="currentPhotos[lightboxIndex]"
            :alt="`${activeRoom?.type} full view`"
            class="max-w-full max-h-[75vh] object-contain rounded-2xl shadow-2xl select-none"
          />

          <!-- Tombol Kanan -->
          <button
            type="button"
            @click="nextPhoto"
            class="absolute right-2 md:right-4 z-10 w-11 h-11 rounded-full bg-black/60 hover:bg-black/90 text-white flex items-center justify-center shadow-lg transition border border-white/20"
            aria-label="Foto berikutnya"
          >
            <ChevronRight class="w-6 h-6" />
          </button>
        </div>

        <!-- Footer Lightbox: Dot Navigator & Tips Keyboard -->
        <div class="w-full max-w-5xl flex flex-col sm:flex-row items-center justify-between gap-3 text-white/70 text-xs py-2">
          <span class="hidden sm:inline">Gunakan tombol <b>←</b> dan <b>→</b> di keyboard, atau tekan <b>ESC</b> untuk keluar</span>
          <div class="flex items-center gap-2">
            <button
              v-for="(_, i) in currentPhotos"
              :key="'dot-' + i"
              type="button"
              @click="lightboxIndex = i"
              class="w-2.5 h-2.5 rounded-full transition-all cursor-pointer"
              :class="i === lightboxIndex ? 'bg-[#C9A86A] scale-125' : 'bg-white/40 hover:bg-white/70'"
              :aria-label="`Pindah ke foto ${i + 1}`"
            ></button>
          </div>
        </div>
      </div>
    </Teleport>
  </section>
</template>
