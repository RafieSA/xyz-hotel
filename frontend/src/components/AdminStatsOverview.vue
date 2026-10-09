<script setup>
import { ref } from 'vue'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Chart from 'primevue/chart'
import {
  CalendarDays,
  Wallet,
  TrendingUp,
  Bed,
  BarChart3,
  ChevronDown,
  ChevronUp,
  PieChart,
  LineChart
} from 'lucide-vue-next'

const props = defineProps({
  totalBookings: {
    type: Number,
    default: 0,
  },
  totalRevenue: {
    type: Number,
    default: 0,
  },
  occupancyRate: {
    type: Number,
    default: 0,
  },
  availableUnits: {
    type: Number,
    default: 0,
  },
  totalUnits: {
    type: [Number, String],
    default: '-',
  },
  bookingsByStatus: {
    type: Object,
    default: () => ({}),
  },
  revenueChartData: {
    type: Object,
    required: true,
  },
  revenueChartOptions: {
    type: Object,
    required: true,
  },
  doughnutData: {
    type: Object,
    required: true,
  },
  doughnutOptions: {
    type: Object,
    required: true,
  },
  barChartData: {
    type: Object,
    required: true,
  },
  barChartOptions: {
    type: Object,
    required: true,
  },
  showStatPop: {
    type: Boolean,
    default: false,
  },
})

// Toggle buka-tutup grafik agar resepsionis tidak terbebani scroll panjang
const showCharts = ref(false)

const fmt = (n) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n || 0)

function formatStatusName(status) {
  const map = {
    waiting_verification: 'Perlu Verifikasi',
    pending_payment: 'Belum Bayar',
    verified: 'Lunas',
    checked_in: 'Menginap',
    checked_out: 'Selesai',
    rejected: 'Ditolak',
    cancelled: 'Dibatalkan',
    expired: 'Kadaluarsa',
  }
  return map[status] || status.replace('_', ' ')
}

function getStatusPillColor(status) {
  if (status === 'waiting_verification') return 'bg-amber-50 text-amber-700 border-amber-200'
  if (status === 'verified') return 'bg-emerald-50 text-emerald-700 border-emerald-200'
  if (status === 'checked_in') return 'bg-teal-50 text-teal-700 border-teal-200'
  if (status === 'rejected' || status === 'cancelled' || status === 'expired') return 'bg-red-50 text-red-700 border-red-200'
  return 'bg-gray-50 text-gray-700 border-gray-200'
}
</script>

<template>
  <div class="space-y-4">
    <!-- 4 Kartu Metrik Utama -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Total Pemesanan -->
      <Card class="!rounded-2xl !shadow-xs !border !border-[#E5E7EB] dark:!border-gray-700 dark:!bg-[#2D3748] hover:!shadow-md transition">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-[11px] uppercase tracking-wider text-[#6B7280] dark:text-gray-400 font-bold">Total Reservasi</p>
              <p class="text-3xl font-bold text-[#1A3A4A] dark:text-white mt-1">
                {{ totalBookings }}
              </p>
              <p class="text-xs text-[#6B7280] dark:text-gray-400 mt-1">Semua status pemesanan</p>
            </div>
            <span class="bg-[#FDF6EC] dark:bg-[#1A3A4A] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-2xl p-2.5">
              <CalendarDays class="w-5 h-5" />
            </span>
          </div>
        </template>
      </Card>

      <!-- Total Pendapatan -->
      <Card class="!rounded-2xl !shadow-xs !border !border-[#E5E7EB] dark:!border-gray-700 dark:!bg-[#2D3748] hover:!shadow-md transition">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-[11px] uppercase tracking-wider text-[#6B7280] dark:text-gray-400 font-bold">Total Pendapatan</p>
              <p class="text-2xl font-bold text-[#8B5A2B] dark:text-[#C9A86A] mt-1">
                {{ fmt(totalRevenue) }}
              </p>
              <p class="text-xs text-[#6B7280] dark:text-gray-400 mt-1">Pemesanan terverifikasi & selesai</p>
            </div>
            <span class="bg-[#FDF6EC] dark:bg-[#1A3A4A] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-2xl p-2.5">
              <Wallet class="w-5 h-5" />
            </span>
          </div>
        </template>
      </Card>

      <!-- Tingkat Okupansi -->
      <Card class="!rounded-2xl !shadow-xs !border !border-[#E5E7EB] dark:!border-gray-700 dark:!bg-[#2D3748] hover:!shadow-md transition">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-[11px] uppercase tracking-wider text-[#6B7280] dark:text-gray-400 font-bold">Tingkat Okupansi</p>
              <p class="text-3xl font-bold text-[#1A3A4A] dark:text-white mt-1">
                {{ occupancyRate }}%
              </p>
              <div class="mt-2 h-1.5 w-24 bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                <div class="h-full bg-[#8B5A2B] rounded-full transition-all duration-500" :style="{ width: occupancyRate + '%' }"></div>
              </div>
            </div>
            <span class="bg-[#FDF6EC] dark:bg-[#1A3A4A] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-2xl p-2.5">
              <TrendingUp class="w-5 h-5" />
            </span>
          </div>
        </template>
      </Card>

      <!-- Kamar Siap Huni -->
      <Card class="!rounded-2xl !shadow-xs !border !border-[#E5E7EB] dark:!border-gray-700 dark:!bg-[#2D3748] hover:!shadow-md transition">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-[11px] uppercase tracking-wider text-[#6B7280] dark:text-gray-400 font-bold">Kamar Siap Huni</p>
              <p class="text-3xl font-bold text-[#1A3A4A] dark:text-white mt-1">
                {{ availableUnits }} <span class="text-sm font-normal text-[#6B7280]">/ {{ totalUnits }}</span>
              </p>
              <p class="text-xs text-emerald-600 dark:text-emerald-400 font-medium mt-1">Tersedia untuk tamu</p>
            </div>
            <span class="bg-[#FDF6EC] dark:bg-[#1A3A4A] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-2xl p-2.5">
              <Bed class="w-5 h-5" />
            </span>
          </div>
        </template>
      </Card>
    </div>

    <!-- Bar Status Ringkas & Tombol Toggle Grafik -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-3 bg-white dark:bg-[#2D3748] rounded-2xl border border-[#E5E7EB] dark:border-gray-700">
      <!-- Badge Ringkasan Jumlah per Status -->
      <div class="flex items-center gap-1.5 overflow-x-auto no-scrollbar py-0.5">
        <span
          v-for="(count, status) in bookingsByStatus"
          :key="status"
          class="inline-flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full border whitespace-nowrap"
          :class="getStatusPillColor(status)"
        >
          <span class="font-medium capitalize">{{ formatStatusName(status) }}:</span>
          <span class="font-bold">{{ count }}</span>
        </span>
      </div>

      <!-- Tombol Buka/Tutup Grafik Analitik -->
      <button
        type="button"
        @click="showCharts = !showCharts"
        class="inline-flex items-center gap-1.5 text-xs font-semibold px-3 py-1.5 rounded-xl border border-[#8B5A2B]/30 hover:border-[#8B5A2B] text-[#8B5A2B] dark:text-[#C9A86A] bg-[#FDF6EC]/50 dark:bg-[#1A3A4A] transition whitespace-nowrap self-start sm:self-auto cursor-pointer"
      >
        <BarChart3 class="w-3.5 h-3.5" />
        <span>{{ showCharts ? 'Sembunyikan Grafik Analitik' : 'Buka Grafik Analitik Bisnis' }}</span>
        <component :is="showCharts ? ChevronUp : ChevronDown" class="w-3.5 h-3.5" />
      </button>
    </div>

    <!-- Area Grafik Analitik (Collapsible / Bisa Dilipat) -->
    <div v-if="showCharts" class="space-y-4 pt-1 transition-all">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <!-- Grafik Pendapatan Harian (Line Chart) -->
        <Card class="!rounded-2xl !shadow-xs !border !border-[#E5E7EB] dark:!border-gray-700 dark:!bg-[#2D3748] lg:col-span-2">
          <template #title>
            <div class="flex items-center gap-2 pb-2 border-b border-gray-100 dark:border-gray-700">
              <LineChart class="w-4 h-4 text-[#8B5A2B]" />
              <span class="text-[#1A3A4A] dark:text-white font-bold text-sm">Tren Pendapatan Harian</span>
            </div>
          </template>
          <template #content>
            <div class="h-[250px] mt-2">
              <Chart type="line" :data="revenueChartData" :options="revenueChartOptions" />
            </div>
          </template>
        </Card>

        <!-- Grafik Rasio Okupansi (Doughnut Chart) -->
        <Card class="!rounded-2xl !shadow-xs !border !border-[#E5E7EB] dark:!border-gray-700 dark:!bg-[#2D3748]">
          <template #title>
            <div class="flex items-center gap-2 pb-2 border-b border-gray-100 dark:border-gray-700">
              <PieChart class="w-4 h-4 text-[#8B5A2B]" />
              <span class="text-[#1A3A4A] dark:text-white font-bold text-sm">Rasio Keterisian Kamar</span>
            </div>
          </template>
          <template #content>
            <div class="h-[220px] flex items-center justify-center mt-2">
              <Chart type="doughnut" :data="doughnutData" :options="doughnutOptions" />
            </div>
            <p class="text-center text-xs text-[#6B7280] dark:text-gray-400 mt-2">
              {{ occupancyRate }}% terisi · {{ 100 - occupancyRate }}% tersedia
            </p>
          </template>
        </Card>
      </div>

      <!-- Grafik Pemesanan per Tipe Kamar (Bar Chart) -->
      <Card class="!rounded-2xl !shadow-xs !border !border-[#E5E7EB] dark:!border-gray-700 dark:!bg-[#2D3748]">
        <template #title>
          <div class="flex items-center gap-2 pb-2 border-b border-gray-100 dark:border-gray-700">
            <BarChart3 class="w-4 h-4 text-[#8B5A2B]" />
            <span class="text-[#1A3A4A] dark:text-white font-bold text-sm">Distribusi Pemesanan per Tipe Kamar</span>
          </div>
        </template>
        <template #content>
          <div class="h-[240px] mt-2">
            <Chart type="bar" :data="barChartData" :options="barChartOptions" />
          </div>
        </template>
      </Card>
    </div>
  </div>
</template>
