<script setup>
import { computed } from 'vue'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Rating from 'primevue/rating'
import Textarea from 'primevue/textarea'
import { Star, MessageSquareQuote, CheckCircle2, User, Send, Calendar } from 'lucide-vue-next'

const props = defineProps({
  rooms: {
    type: Array,
    required: true,
  },
  reviewsByType: {
    type: Object,
    default: () => ({}),
  },
  avgRating: {
    type: Function,
    required: true,
  },
  ratingCount: {
    type: Function,
    required: true,
  },
  isAuthenticated: {
    type: Boolean,
    default: false,
  },
  eligibleBookings: {
    type: Array,
    default: () => [],
  },
  reviewForm: {
    type: Object,
    required: true,
  },
  submittingReview: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['submit-review'])

// Helper mendapatkan inisial nama tamu
function getInitial(name) {
  if (!name) return 'T'
  return name.trim().charAt(0).toUpperCase()
}

// Format tanggal ulasan
function formatDate(dateStr) {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  } catch {
    return dateStr
  }
}
</script>

<template>
  <section class="max-w-7xl mx-auto px-4 sm:px-6 py-12 md:py-16">
    <!-- Header Section -->
    <div class="text-center max-w-2xl mx-auto mb-10">
      <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#FDF6EC] text-[#8B5A2B] text-xs font-semibold uppercase tracking-wider mb-2">
        <Star class="w-3.5 h-3.5 text-[#C9A86A] fill-[#C9A86A]" />
        Kesan & Pengalaman Tamu
      </div>
      <h2 class="font-display font-bold text-3xl md:text-4xl text-[#1A3A4A] dark:text-[#FDF6EC]" style="font-family:'Playfair Display',serif">
        Ulasan Tamu Nyata
      </h2>
      <div class="h-0.5 w-16 bg-[#C9A86A] mx-auto mt-3"></div>
      <p class="text-sm md:text-base text-[#6B7280] dark:text-gray-300 mt-3">
        Cerita otentik dari para tamu yang telah menikmati ketenangan dan keramahan layanan kami di Ubud.
      </p>
    </div>

    <!-- Grid Kartu Ulasan Per Tipe Kamar -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
      <div
        v-for="r in rooms"
        :key="'rev-card-' + r.id"
        class="bg-white dark:bg-[#2D3748] rounded-3xl p-5 border border-[#E5E7EB] dark:border-gray-700 shadow-sm flex flex-col justify-between"
      >
        <div>
          <!-- Header Kartu: Tipe Kamar & Rating Rata-rata -->
          <div class="flex items-center justify-between pb-3.5 border-b border-[#F3F4F6] dark:border-gray-700">
            <div>
              <h3 class="font-bold text-sm text-[#1A3A4A] dark:text-white">
                {{ r.type }} Room
              </h3>
              <p class="text-[11px] text-[#6B7280] dark:text-gray-400">
                {{ ratingCount(r.id) }} ulasan terverifikasi
              </p>
            </div>
            <div class="flex items-center gap-1 bg-[#FDF6EC] dark:bg-[#1A3A4A] px-2.5 py-1 rounded-full border border-[#8B5A2B]/20">
              <Star class="w-3.5 h-3.5 text-[#C9A86A] fill-[#C9A86A]" />
              <span class="text-xs font-bold text-[#8B5A2B] dark:text-[#C9A86A]">
                {{ avgRating(r.id).toFixed(1) }}
              </span>
            </div>
          </div>

          <!-- Daftar Ulasan (Maksimal 2-3 Ulasan Terbaru) -->
          <div v-if="(reviewsByType[r.id] || []).length" class="mt-4 space-y-3.5">
            <div
              v-for="rev in (reviewsByType[r.id] || []).slice(0, 2)"
              :key="'item-' + rev.id"
              class="bg-[#F9FAFB] dark:bg-[#1A3A4A]/50 p-3.5 rounded-2xl border border-gray-100 dark:border-gray-700/60"
            >
              <div class="flex items-center justify-between mb-1.5">
                <div class="flex items-center gap-1.5">
                  <div class="w-6 h-6 rounded-full bg-[#8B5A2B] text-white text-[10px] font-bold flex items-center justify-center">
                    {{ getInitial(rev.user_name || rev.userName || 'Tamu') }}
                  </div>
                  <span class="text-xs font-semibold text-[#1A3A4A] dark:text-gray-200">
                    {{ rev.user_name || rev.userName || 'Tamu Terverifikasi' }}
                  </span>
                </div>
                <div class="flex items-center">
                  <Rating :modelValue="rev.rating" readonly :stars="5" class="!gap-0.5 scale-90 origin-right" />
                </div>
              </div>
              <p class="text-xs text-[#4B5563] dark:text-gray-300 leading-relaxed italic">
                "{{ rev.comment || 'Pelayanan sangat memuaskan dan kamar sangat nyaman.' }}"
              </p>
              <span class="text-[10px] text-[#9CA3AF] block mt-1.5 text-right">
                {{ formatDate(rev.created_at) }}
              </span>
            </div>
          </div>

          <!-- Empty State Jika Belum Ada Ulasan -->
          <div v-else class="py-8 text-center">
            <MessageSquareQuote class="w-8 h-8 text-gray-300 mx-auto mb-2" />
            <p class="text-xs text-[#6B7280] dark:text-gray-400">
              Belum ada ulasan untuk kamar ini.
            </p>
            <span class="text-[11px] text-[#8B5A2B] dark:text-[#C9A86A] mt-1 block">
              Jadilah yang pertama menginap!
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Formulir Tulis Ulasan (Hanya Muncul Jika Tamu Login) -->
    <div v-if="isAuthenticated" class="mt-10 max-w-3xl mx-auto bg-white dark:bg-[#2D3748] rounded-3xl p-6 md:p-8 border border-[#E5E7EB] dark:border-gray-700 shadow-md">
      <div class="flex items-center gap-2.5 mb-2">
        <MessageSquareQuote class="w-5 h-5 text-[#8B5A2B]" />
        <h3 class="font-bold text-lg text-[#1A3A4A] dark:text-white">
          Tulis Ulasan Pengalaman Anda
        </h3>
      </div>
      <p class="text-xs text-[#6B7280] dark:text-gray-400 mb-5">
        Khusus untuk reservasi yang telah selesai (*checked-out*). Ulasan Anda membantu kami menjaga standar pelayanan terbaik.
      </p>

      <div class="space-y-4">
        <!-- Pilihan Reservasi yang Selesai -->
        <div>
          <label class="block text-xs font-bold text-[#6B7280] dark:text-gray-300 uppercase tracking-wider mb-1.5">
            Pilih Reservasi Anda
          </label>
          <Select
            v-model="reviewForm.booking_id"
            :options="eligibleBookings"
            optionLabel="id"
            optionValue="id"
            placeholder="Pilih kamar yang telah Anda selesaikan"
            class="w-full !rounded-xl"
            :emptyMessage="'Tidak ada reservasi yang selesai untuk diulas'"
          >
            <template #option="{ option }">
              <span class="text-xs">
                #{{ option.id }} · Tipe Kamar {{ option.room_type_id }} ({{ option.check_in }} s/d {{ option.check_out }})
              </span>
            </template>
            <template #value="{ value }">
              <span v-if="value" class="text-xs font-semibold">Reservasi #{{ value }} Terpilih</span>
              <span v-else class="text-xs text-gray-400">Pilih reservasi yang ingin diulas</span>
            </template>
          </Select>
        </div>

        <!-- Rating Bintang -->
        <div>
          <label class="block text-xs font-bold text-[#6B7280] dark:text-gray-300 uppercase tracking-wider mb-1.5">
            Beri Penilaian Bintang
          </label>
          <Rating v-model="reviewForm.rating" :stars="5" class="!gap-1" />
        </div>

        <!-- Kolom Komentar -->
        <div>
          <label class="block text-xs font-bold text-[#6B7280] dark:text-gray-300 uppercase tracking-wider mb-1.5">
            Cerita & Masukan Anda
          </label>
          <Textarea
            v-model="reviewForm.comment"
            rows="3"
            placeholder="Ceritakan momen terbaik Anda selama menginap di XYZ Hotel Ubud..."
            class="w-full !rounded-xl !p-3 !text-sm"
            autoResize
          />
        </div>

        <!-- Tombol Kirim -->
        <Button
          label="Kirimkan Ulasan"
          icon="pi pi-send"
          class="!min-h-[44px] !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] text-white !rounded-xl !px-6 !text-sm font-semibold shadow-sm transition"
          :loading="submittingReview"
          @click="emit('submit-review')"
        />
      </div>
    </div>

    <!-- Ajakan Login Jika Belum Masuk -->
    <div v-else class="mt-8 text-center text-xs text-[#6B7280] dark:text-gray-400">
      <span>Ingin membagikan ulasan menginap Anda? </span>
      <router-link to="/login" class="text-[#8B5A2B] dark:text-[#C9A86A] font-semibold underline">
        Masuk ke akun Anda
      </router-link>
    </div>
  </section>
</template>
