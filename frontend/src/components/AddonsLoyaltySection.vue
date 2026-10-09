<script setup>
import { computed } from 'vue'
import { Utensils, Car, BedDouble, Award, Check, Sparkles, CheckCircle2, Gift } from 'lucide-vue-next'

const props = defineProps({
  addons: {
    type: Array,
    required: true,
  },
  selectedAddons: {
    type: Object, // Set of ids
    required: true,
  },
  isAuthenticated: {
    type: Boolean,
    default: false,
  },
  loyaltyPoints: {
    type: Number,
    default: 0,
  },
  useLoyalty: {
    type: Boolean,
    default: false,
  },
  loyaltyCanRedeem: {
    type: Boolean,
    default: false,
  },
  addonsTotal: {
    type: Number,
    default: 0,
  },
})

const emit = defineEmits(['toggle-addon', 'update:useLoyalty'])

const fmt = (n) => new Intl.NumberFormat('id-ID').format(n || 0)

// Helper icon sesuai nama layanan
function getAddonIcon(name) {
  const n = (name || '').toLowerCase()
  if (n.includes('breakfast') || n.includes('sarapan')) return Utensils
  if (n.includes('airport') || n.includes('transfer') || n.includes('antar')) return Car
  return BedDouble
}

// Progress poin loyalty menuju 100
const progressPercent = computed(() => {
  return Math.min(100, Math.round((props.loyaltyPoints / 100) * 100))
})

const pointsNeeded = computed(() => {
  return Math.max(0, 100 - props.loyaltyPoints)
})
</script>

<template>
  <section class="max-w-7xl mx-auto px-4 sm:px-6 py-12 md:py-16">
    <div class="bg-white dark:bg-[#2D3748] rounded-3xl p-6 md:p-8 border border-[#E5E7EB] dark:border-gray-700 shadow-md">
      <!-- Header Section -->
      <div class="flex flex-col md:flex-row md:items-end justify-between gap-4 pb-6 border-b border-[#F3F4F6] dark:border-gray-700">
        <div>
          <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#FDF6EC] text-[#8B5A2B] text-xs font-semibold uppercase tracking-wider mb-2">
            <Sparkles class="w-3.5 h-3.5 text-[#C9A86A]" />
            Kenyamanan Tambahan
          </div>
          <h2 class="font-display font-bold text-2xl md:text-3xl text-[#1A3A4A] dark:text-[#FDF6EC]" style="font-family:'Playfair Display',serif">
            Layanan Pelengkap Menginap
          </h2>
          <p class="text-xs md:text-sm text-[#6B7280] dark:text-gray-300 mt-1">
            Pilih fasilitas ekstra untuk menyempurnakan liburan Anda. Biaya dihitung transparan tanpa biaya tersembunyi.
          </p>
        </div>

        <!-- Ringkasan Biaya Tambahan -->
        <div class="flex items-center gap-3 bg-[#FDF6EC] dark:bg-[#1A3A4A] px-4 py-3 rounded-2xl border border-[#E5E7EB] dark:border-gray-600 self-start md:self-auto">
          <div>
            <p class="text-[11px] font-medium text-[#6B7280] dark:text-gray-300">Total Tambahan</p>
            <p class="text-base font-bold text-[#8B5A2B] dark:text-[#C9A86A]">
              Rp {{ fmt(addonsTotal) }}
            </p>
          </div>
          <span v-if="selectedAddons.size > 0" class="px-2 py-0.5 rounded-full bg-[#8B5A2B] text-white text-[11px] font-bold">
            {{ selectedAddons.size }} dipilih
          </span>
        </div>
      </div>

      <!-- Kartu-Kartu Pilihan Add-ons -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4.5 mt-6">
        <div
          v-for="a in addons"
          :key="'addon-' + a.id"
          @click="emit('toggle-addon', a.id)"
          class="relative p-5 rounded-2xl border-2 transition-all cursor-pointer flex flex-col justify-between select-none group"
          :class="selectedAddons.has(a.id)
            ? 'border-[#8B5A2B] bg-[#FDF6EC]/70 dark:bg-[#1A3A4A]/80 shadow-sm'
            : 'border-[#E5E7EB] dark:border-gray-700 bg-white dark:bg-[#2D3748] hover:border-[#8B5A2B]/40 hover:shadow-sm'"
        >
          <div>
            <!-- Ikon & Check Indicator -->
            <div class="flex items-center justify-between mb-3">
              <div
                class="w-10 h-10 rounded-xl flex items-center justify-center transition"
                :class="selectedAddons.has(a.id)
                  ? 'bg-[#8B5A2B] text-white'
                  : 'bg-[#FDF6EC] dark:bg-[#1A3A4A] text-[#8B5A2B] group-hover:scale-105'"
              >
                <component :is="getAddonIcon(a.name)" class="w-5 h-5" />
              </div>

              <!-- Checkbox Kustom yang Ramah Sentuhan -->
              <div
                class="w-6 h-6 rounded-full flex items-center justify-center border-2 transition"
                :class="selectedAddons.has(a.id)
                  ? 'bg-[#8B5A2B] border-[#8B5A2B] text-white'
                  : 'border-gray-300 dark:border-gray-600 bg-white dark:bg-[#1A3A4A]'"
              >
                <Check v-if="selectedAddons.has(a.id)" class="w-3.5 h-3.5 stroke-[3]" />
              </div>
            </div>

            <h3 class="font-bold text-sm md:text-base text-[#1A3A4A] dark:text-white">
              {{ a.name }}
            </h3>
            <p class="text-xs text-[#6B7280] dark:text-gray-300 mt-1 leading-relaxed">
              {{ a.description }}
            </p>
          </div>

          <div class="mt-4 pt-3 border-t border-[#E5E7EB]/70 dark:border-gray-700 flex items-center justify-between">
            <span class="text-xs text-[#6B7280] dark:text-gray-400">Tarif per malam</span>
            <span class="text-sm font-bold text-[#8B5A2B] dark:text-[#C9A86A]">
              + Rp {{ fmt(a.price) }}
            </span>
          </div>
        </div>
      </div>

      <!-- Area Program Loyalitas Poin Terpadu -->
      <div class="mt-8 pt-6 border-t border-[#F3F4F6] dark:border-gray-700">
        <div class="bg-gradient-to-br from-[#FDF6EC] to-white dark:from-[#1A3A4A] dark:to-[#2D3748] rounded-2xl p-5 md:p-6 border border-[#C9A86A]/40 flex flex-col md:flex-row items-stretch md:items-center justify-between gap-6">
          <div class="flex-1">
            <div class="flex items-center gap-2">
              <Award class="w-5 h-5 text-[#C9A86A]" />
              <h3 class="font-bold text-base text-[#1A3A4A] dark:text-white">
                Program Reward Loyalitas Tamu
              </h3>
            </div>
            <p class="text-xs md:text-sm text-[#6B7280] dark:text-gray-300 mt-1">
              Dapatkan 10 poin untuk setiap malam menginap. Setiap <b>100 poin</b> dapat ditukar dengan potongan langsung <b>Rp 100.000</b>.
            </p>

            <!-- Bar Status Poin Tamu (Jika Sudah Login) -->
            <div v-if="isAuthenticated" class="mt-3.5 max-w-md">
              <div class="flex items-center justify-between text-xs mb-1.5">
                <span class="font-semibold text-[#1A3A4A] dark:text-white">Saldo Anda: {{ loyaltyPoints }} Poin</span>
                <span class="text-[#6B7280] dark:text-gray-400">
                  {{ loyaltyCanRedeem ? 'Siap digunakan!' : `${pointsNeeded} poin lagi untuk Rp 100k` }}
                </span>
              </div>
              <div class="h-2.5 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-[#C9A86A] to-[#8B5A2B] rounded-full transition-all duration-500"
                  :style="{ width: progressPercent + '%' }"
                ></div>
              </div>
            </div>

            <!-- Jika Tamu Belum Login -->
            <div v-else class="mt-2 text-xs text-[#8B5A2B] dark:text-[#C9A86A] flex items-center gap-1.5">
              <Gift class="w-4 h-4" />
              <span>Masuk akun untuk mengumpulkan dan menukarkan poin hadiah Anda.</span>
            </div>
          </div>

          <!-- Aksi Penukaran Poin -->
          <div class="flex flex-col items-start md:items-end justify-center pt-2 md:pt-0 border-t md:border-t-0 border-[#E5E7EB] dark:border-gray-700">
            <label
              class="inline-flex items-center gap-3 p-3 rounded-xl border transition cursor-pointer select-none"
              :class="!loyaltyCanRedeem
                ? 'opacity-60 bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700 cursor-not-allowed'
                : useLoyalty
                  ? 'bg-emerald-50 dark:bg-emerald-950/40 border-emerald-300 dark:border-emerald-700 text-emerald-800 dark:text-emerald-300'
                  : 'bg-white dark:bg-[#1A3A4A] border-[#8B5A2B]/40 hover:border-[#8B5A2B]'"
            >
              <input
                type="checkbox"
                :checked="useLoyalty"
                :disabled="!loyaltyCanRedeem"
                @change="emit('update:useLoyalty', $event.target.checked)"
                class="w-4 h-4 accent-[#8B5A2B] cursor-pointer disabled:cursor-not-allowed"
              />
              <div class="text-left">
                <p class="text-xs font-bold">Gunakan 100 Poin</p>
                <p class="text-[11px] text-[#6B7280] dark:text-gray-300">Potongan langsung Rp 100.000</p>
              </div>
            </label>

            <span v-if="useLoyalty && loyaltyCanRedeem" class="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-600 dark:text-emerald-400 mt-1.5">
              <CheckCircle2 class="w-3.5 h-3.5" /> Diskon Rp 100.000 diterapkan
            </span>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
