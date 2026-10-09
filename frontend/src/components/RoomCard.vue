<script setup>
import { computed } from 'vue'
import Button from 'primevue/button'
import { Star, Bed, Users, Heart, Sparkles } from 'lucide-vue-next'

const props = defineProps({
  room: {
    type: Object,
    required: true,
  },
  avail: {
    type: Object,
    default: null,
  },
  loadingAvail: {
    type: Boolean,
    default: false,
  },
  isWished: {
    type: Boolean,
    default: false,
  },
  wishLoading: {
    type: Boolean,
    default: false,
  },
  bookingLoading: {
    type: Boolean,
    default: false,
  },
  nights: {
    type: Number,
    default: 1,
  },
  canSearch: {
    type: Boolean,
    default: false,
  },
  discountPercent: {
    type: Number,
    default: 0,
  },
  weekendMultiplier: {
    type: Number,
    default: 1,
  },
  avgRating: {
    type: Number,
    default: 5.0,
  },
  ratingCount: {
    type: Number,
    default: 0,
  },
  isSelected: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['toggle-wishlist', 'view-details', 'book'])

const fmt = (n) => new Intl.NumberFormat('id-ID').format(Math.round(n || 0))

// Fallback gambar jika gambar eksternal gagal dimuat
const fallbackImage = 'https://images.unsplash.com/photo-1566665797739-1674de7a421a?w=800&q=80&auto=format&fit=crop'
function onImageError(e) {
  if (e?.target) {
    e.target.src = fallbackImage
  }
}

// Perhitungan harga bersih
const effectivePricePerNight = computed(() => {
  let p = props.room.price
  if (props.discountPercent > 0) {
    p = p * (1 - props.discountPercent / 100)
  }
  if (props.canSearch && props.weekendMultiplier > 1) {
    p = p * props.weekendMultiplier
  }
  return Math.round(p)
})

const estimatedTotalPrice = computed(() => {
  return effectivePricePerNight.value * (props.nights || 1)
})

// Status ketersediaan kamar
const isFullyBooked = computed(() => {
  return props.avail && props.avail.available <= 0
})

const isLowAvailability = computed(() => {
  return props.avail && props.avail.available > 0 && props.avail.available <= 2
})

const availabilityLabel = computed(() => {
  if (!props.canSearch) return ''
  if (props.loadingAvail) return 'Mengecek ketersediaan...'
  if (!props.avail) return 'Ketersediaan belum dicek'
  if (isFullyBooked.value) return 'Kamar Penuh di Tanggal Ini'
  if (isLowAvailability.value) return `Sisa ${props.avail.available} kamar saja!`
  return `${props.avail.available} kamar tersedia`
})

const availabilityBadgeClass = computed(() => {
  if (isFullyBooked.value) return 'bg-red-50 text-[#C62828] border-red-200'
  if (isLowAvailability.value) return 'bg-amber-50 text-[#EF6C00] border-amber-200'
  return 'bg-emerald-50 text-[#2E7D32] border-emerald-200'
})
</script>

<template>
  <article
    :id="'room-card-' + room.id"
    class="group relative flex flex-col h-full bg-white dark:bg-[#2D3748] rounded-2xl border border-[#E5E7EB] dark:border-[#4A5568] shadow-sm hover:shadow-lg transition-all duration-300 overflow-hidden"
    :class="isSelected ? 'ring-2 ring-[#8B5A2B] dark:ring-[#C9A86A] ring-offset-2' : ''"
  >
    <!-- Header Gambar & Badge -->
    <div class="relative w-full aspect-[16/10] overflow-hidden bg-[#FDF6EC]">
      <img
        :src="room.img"
        :alt="`${room.type} Room - XYZ Hotel`"
        loading="lazy"
        @error="onImageError"
        class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
        :class="isFullyBooked ? 'grayscale-[30%] opacity-90' : ''"
      />

      <!-- Overlay Gradien Lembut Bawah Gambar -->
      <div class="absolute inset-0 bg-gradient-to-t from-black/40 via-transparent to-black/10 pointer-events-none"></div>

      <!-- Badge Rating di Kiri Atas -->
      <div class="absolute top-3 left-3 flex items-center gap-1.5 px-2.5 py-1 bg-white/95 dark:bg-[#1A3A4A]/90 backdrop-blur-sm rounded-full shadow-sm">
        <Star class="w-3.5 h-3.5 text-[#C9A86A] fill-[#C9A86A]" />
        <span class="text-xs font-bold text-[#1A3A4A] dark:text-[#FDF6EC]">{{ avgRating.toFixed(1) }}</span>
        <span v-if="ratingCount > 0" class="text-[11px] text-[#6B7280] dark:text-[#C9A86A]/80">({{ ratingCount }})</span>
      </div>

      <!-- Badge Tipe Kamar di Kanan Atas -->
      <div class="absolute top-3 right-3 px-3 py-1 bg-[#8B5A2B]/95 text-white text-xs font-semibold uppercase tracking-wider rounded-full shadow-sm backdrop-blur-sm">
        {{ room.type }}
      </div>

      <!-- Tombol Wishlist di Kanan Bawah Gambar (Touch target 44x44px) -->
      <button
        type="button"
        @click.stop="emit('toggle-wishlist', room)"
        :disabled="wishLoading"
        :aria-label="isWished ? `Hapus ${room.type} dari wishlist` : `Simpan ${room.type} ke wishlist`"
        class="absolute bottom-3 right-3 w-10 h-10 rounded-full bg-white/95 dark:bg-[#1A3A4A]/90 backdrop-blur-sm shadow-md flex items-center justify-center hover:scale-110 active:scale-95 transition border border-white/60 dark:border-white/10 disabled:opacity-50"
      >
        <Heart
          class="w-5 h-5 transition-colors"
          :class="isWished ? 'text-red-500 fill-red-500' : 'text-[#6B7280] dark:text-gray-300 hover:text-red-500'"
        />
      </button>

      <!-- Overlay Khusus Kamar Penuh -->
      <div
        v-if="canSearch && isFullyBooked"
        class="absolute inset-0 bg-black/55 backdrop-blur-[1px] flex flex-col items-center justify-center text-white text-center p-4 pointer-events-none"
      >
        <span class="px-3 py-1 rounded-full bg-red-600/90 text-xs font-bold tracking-wider uppercase mb-1">Penuh</span>
        <p class="text-xs text-white/90">Kamar tidak tersedia di tanggal yang dipilih</p>
      </div>
    </div>

    <!-- Konten Informasi Kamar -->
    <div class="flex-1 p-5 flex flex-col justify-between">
      <div>
        <!-- Nama Kamar & Kapasitas -->
        <div class="flex items-start justify-between gap-2">
          <h3 class="font-display text-lg font-bold text-[#1A3A4A] dark:text-[#FDF6EC] line-clamp-1">
            {{ room.type }} Room
          </h3>
          <span class="inline-flex items-center gap-1 text-xs text-[#6B7280] dark:text-gray-300 font-medium whitespace-nowrap bg-[#F9FAFB] dark:bg-[#1A3A4A] px-2 py-0.5 rounded-md border border-[#E5E7EB] dark:border-gray-700">
            <Users class="w-3.5 h-3.5 text-[#8B5A2B]" />
            {{ room.cap }} Tamu
          </span>
        </div>

        <!-- Fasilitas Utama -->
        <p class="mt-2 text-xs text-[#6B7280] dark:text-gray-400 flex items-center gap-1.5 line-clamp-1">
          <Bed class="w-3.5 h-3.5 text-[#8B5A2B] shrink-0" />
          <span>{{ room.facility }}</span>
        </p>

        <!-- Status Ketersediaan (Jika tanggal sudah dipilih) -->
        <div v-if="canSearch" class="mt-3">
          <div
            class="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-lg border"
            :class="availabilityBadgeClass"
          >
            <span
              class="w-1.5 h-1.5 rounded-full"
              :class="isFullyBooked ? 'bg-[#C62828]' : isLowAvailability ? 'bg-[#EF6C00]' : 'bg-[#2E7D32]'"
            ></span>
            <span>{{ availabilityLabel }}</span>
          </div>
        </div>
      </div>

      <!-- Informasi Harga & Tombol Aksi -->
      <div class="mt-5 pt-4 border-t border-[#F3F4F6] dark:border-[#4A5568]">
        <!-- Area Harga Bersih -->
        <div class="mb-4">
          <div class="flex items-baseline gap-1.5">
            <span class="text-xs text-[#6B7280] dark:text-gray-400">Mulai dari</span>
            <span class="text-xl font-bold text-[#8B5A2B] dark:text-[#C9A86A]">
              Rp {{ fmt(effectivePricePerNight) }}
            </span>
            <span class="text-xs text-[#6B7280] dark:text-gray-400">/ malam</span>
          </div>

          <!-- Indikator Diskon atau Weekend (Jika Ada) -->
          <div v-if="discountPercent > 0 || weekendMultiplier > 1" class="flex flex-wrap items-center gap-2 mt-1">
            <span v-if="discountPercent > 0" class="inline-flex items-center gap-0.5 text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-1.5 py-0.5 rounded">
              <Sparkles class="w-3 h-3" /> Hemat {{ discountPercent }}%
            </span>
            <span v-if="discountPercent > 0" class="text-xs text-gray-400 line-through">
              Rp {{ fmt(room.price) }}
            </span>
            <span v-if="canSearch && weekendMultiplier > 1" class="text-[11px] text-amber-700 bg-amber-50 px-1.5 py-0.5 rounded">
              Tarif akhir pekan (+20%)
            </span>
          </div>

          <!-- Total untuk periode tanggal terpilih -->
          <div v-if="canSearch && nights > 1" class="mt-1 text-xs text-[#1A3A4A] dark:text-[#FDF6EC] font-medium">
            Total {{ nights }} malam: <span class="font-bold text-[#8B5A2B]">Rp {{ fmt(estimatedTotalPrice) }}</span>
          </div>
        </div>

        <!-- Tombol Aksi (Tinggi minimal 44px untuk kenyamanan sentuhan) -->
        <div class="grid grid-cols-2 gap-2.5">
          <Button
            label="Lihat Foto"
            outlined
            class="!min-h-[44px] !rounded-xl !text-xs md:!text-sm !font-semibold !text-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#FDF6EC] dark:hover:!bg-[#1A3A4A] transition"
            @click="emit('view-details', room)"
          />

          <Button
            :label="isFullyBooked ? 'Kamar Penuh' : bookingLoading ? 'Memproses...' : 'Pesan Kamar'"
            :loading="bookingLoading"
            :disabled="isFullyBooked"
            class="!min-h-[44px] !rounded-xl !text-xs md:!text-sm !font-semibold !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] text-white disabled:!bg-gray-300 disabled:!border-gray-300 disabled:!text-gray-500 shadow-sm transition"
            @click="emit('book', room)"
          />
        </div>
      </div>
    </div>
  </article>
</template>
