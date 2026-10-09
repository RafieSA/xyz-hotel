<script setup>
import { ref, computed } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import {
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  RotateCw,
  KeyRound,
  CheckCircle2,
  Sparkles,
  Wrench,
  Bed,
  Info
} from 'lucide-vue-next'

const props = defineProps({
  roomUnits: {
    type: Array,
    required: true,
  },
  roomTypes: {
    type: Array,
    default: () => [],
  },
  calendarData: {
    type: [Object, Array],
    default: null,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  startDate: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['refresh', 'change-range'])

// Offset hari (0 = hari ini sampai +6 hari)
const dayOffset = ref(0)

// Helper format YYYY-MM-DD
function formatDateISO(d) {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 7 hari terhitung dari hari ini + dayOffset
const days = computed(() => {
  const result = []
  const base = new Date()
  base.setHours(0, 0, 0, 0)
  base.setDate(base.getDate() + dayOffset.value)

  for (let i = 0; i < 7; i++) {
    const d = new Date(base)
    d.setDate(base.getDate() + i)
    result.push(formatDateISO(d))
  }
  return result
})

// Cek apakah tanggal adalah hari ini
function isToday(dateStr) {
  return dateStr === formatDateISO(new Date())
}

// Navigasi tanggal
function prevWeek() {
  dayOffset.value -= 7
  notifyRangeChange()
}

function nextWeek() {
  dayOffset.value += 7
  notifyRangeChange()
}

function resetToToday() {
  dayOffset.value = 0
  notifyRangeChange()
}

function notifyRangeChange() {
  emit('change-range', {
    from: days.value[0],
    to: days.value[6],
    days: days.value,
  })
}

// Helper nama tipe kamar
function getRoomTypeName(typeId) {
  const found = props.roomTypes.find(t => t.id === typeId)
  return found?.name || `Type #${typeId}`
}

// Helper resolusi status per unit per tanggal
function getUnitCellInfo(unit, dateStr) {
  // 1. Jika calendarData memiliki struktur backend { days, bookings, ... }
  if (props.calendarData?.bookings) {
    const bkg = props.calendarData.bookings.find(b => {
      if (b.room_unit_id && b.room_unit_id === unit.id) {
        return b.check_in <= dateStr && b.check_out > dateStr && b.status !== 'cancelled'
      }
      return false
    })
    if (bkg) {
      return {
        status: 'occupied',
        bookingId: bkg.id,
        label: `Booking #${bkg.id}`,
      }
    }
  }

  // 2. Jika calendarData memiliki struktur array per tanggal (mock/fallback)
  if (props.calendarData && props.calendarData[dateStr] && Array.isArray(props.calendarData[dateStr])) {
    const found = props.calendarData[dateStr].find(x => x.unit_id === unit.id || x.code === unit.code)
    if (found) {
      return {
        status: found.status || 'available',
        bookingId: found.booking_id || null,
        label: found.booking_id ? `BKG #${found.booking_id}` : found.status,
      }
    }
  }

  // 3. Fallback ke status asli unit jika tanggal adalah hari ini
  if (isToday(dateStr)) {
    return {
      status: unit.status || 'available',
      bookingId: null,
      label: unit.status || 'available',
    }
  }

  // Default hari lain: available
  return {
    status: 'available',
    bookingId: null,
    label: 'Tersedia',
  }
}

// Konfigurasi visual tema Ubud per status
const statusMeta = {
  available: {
    label: 'Tersedia',
    color: '#2E7D32',
    bg: 'bg-emerald-50/80 dark:bg-emerald-950/30',
    border: 'border-emerald-200/70 dark:border-emerald-800/40',
    text: 'text-emerald-800 dark:text-emerald-300',
    icon: CheckCircle2,
  },
  occupied: {
    label: 'Terisi',
    color: '#8B5A2B',
    bg: 'bg-[#8B5A2B] text-white shadow-sm',
    border: 'border-[#704620]',
    text: 'text-white',
    icon: KeyRound,
  },
  dirty: {
    label: 'Perlu Dibersihkan',
    color: '#EF6C00',
    bg: 'bg-amber-50 dark:bg-amber-950/40',
    border: 'border-amber-200 dark:border-amber-800/50',
    text: 'text-amber-800 dark:text-amber-300',
    icon: Sparkles,
  },
  maintenance: {
    label: 'Perawatan',
    color: '#6B7280',
    bg: 'bg-gray-100 dark:bg-gray-800/60',
    border: 'border-gray-200 dark:border-gray-700',
    text: 'text-gray-700 dark:text-gray-300',
    icon: Wrench,
  },
}

// Fallback jika unit kosong
const displayUnits = computed(() => {
  if (props.roomUnits && props.roomUnits.length > 0) {
    return props.roomUnits
  }
  return [
    { id: 101, code: 'STD-101', room_type_id: 1, status: 'available' },
    { id: 102, code: 'DLX-101', room_type_id: 2, status: 'occupied' },
    { id: 103, code: 'FAM-101', room_type_id: 3, status: 'dirty' },
    { id: 104, code: 'STE-101', room_type_id: 4, status: 'maintenance' },
  ]
})

// Format label rentang tanggal (header)
const rangeLabel = computed(() => {
  if (!days.value.length) return ''
  const first = new Date(days.value[0])
  const last = new Date(days.value[6])
  const opt = { day: 'numeric', month: 'short' }
  const optWithYear = { day: 'numeric', month: 'short', year: 'numeric' }
  return `${first.toLocaleDateString('id-ID', opt)} — ${last.toLocaleDateString('id-ID', optWithYear)}`
})
</script>

<template>
  <div class="bg-white dark:bg-[#2D3748] rounded-2xl border border-[#E5E7EB] dark:border-[#4A5568] shadow-sm overflow-hidden">
    <!-- Header Kalender & Navigasi -->
    <div class="p-4 sm:p-6 border-b border-[#E5E7EB] dark:border-[#4A5568]">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-[#FDF6EC] dark:bg-[#1A3A4A]/60 flex items-center justify-center text-[#8B5A2B]">
              <CalendarDays class="w-4 h-4" />
            </div>
            <h3 class="text-[#1A3A4A] dark:text-white font-display font-bold text-lg" style="font-family:'Playfair Display',serif">
              Matriks Okupansi Kamar
            </h3>
          </div>
          <p class="text-xs text-[#6B7280] dark:text-gray-300 mt-1 flex items-center gap-2">
            <span>Rentang 7 Hari: <strong class="text-[#1A3A4A] dark:text-white">{{ rangeLabel }}</strong></span>
            <span class="text-gray-300">·</span>
            <span>Total {{ displayUnits.length }} unit kamar</span>
          </p>
        </div>

        <!-- Tombol Aksi & Kontrol Minggu -->
        <div class="flex items-center flex-wrap gap-2">
          <div class="inline-flex items-center rounded-xl border border-[#E5E7EB] dark:border-[#4A5568] bg-[#F9FAFB] dark:bg-[#1F2937] p-1">
            <button
              type="button"
              class="px-2.5 py-1 text-xs rounded-lg font-medium text-[#1A3A4A] dark:text-gray-200 hover:bg-white dark:hover:bg-gray-700 transition flex items-center gap-1"
              title="7 Hari Sebelumnya"
              @click="prevWeek"
            >
              <ChevronLeft class="w-3.5 h-3.5" />
              <span class="hidden md:inline">Lalu</span>
            </button>
            <button
              type="button"
              class="px-3 py-1 text-xs rounded-lg font-semibold transition"
              :class="dayOffset === 0 ? 'bg-[#8B5A2B] text-white shadow-xs' : 'text-[#8B5A2B] hover:bg-white dark:hover:bg-gray-700'"
              @click="resetToToday"
            >
              Hari Ini
            </button>
            <button
              type="button"
              class="px-2.5 py-1 text-xs rounded-lg font-medium text-[#1A3A4A] dark:text-gray-200 hover:bg-white dark:hover:bg-gray-700 transition flex items-center gap-1"
              title="7 Hari Berikutnya"
              @click="nextWeek"
            >
              <span class="hidden md:inline">Depan</span>
              <ChevronRight class="w-3.5 h-3.5" />
            </button>
          </div>

          <Button
            outlined
            size="small"
            class="!rounded-xl !border-[#8B5A2B] !text-[#8B5A2B] !text-xs !py-1.5"
            :loading="loading"
            @click="emit('refresh')"
          >
            <template #icon>
              <RotateCw class="w-3.5 h-3.5 mr-1" />
            </template>
            Segarkan
          </Button>
        </div>
      </div>

      <!-- Legend Warna Status -->
      <div class="mt-4 pt-3 border-t border-gray-100 dark:border-gray-700/60 flex flex-wrap items-center gap-4 text-xs text-[#6B7280] dark:text-gray-300">
        <span class="text-[11px] font-semibold uppercase tracking-wider text-gray-400">Petunjuk Status:</span>
        <div class="inline-flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
          <span>Tersedia (Available)</span>
        </div>
        <div class="inline-flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-[#8B5A2B]"></span>
          <span class="font-medium text-[#1A3A4A] dark:text-white">Terisi (Occupied)</span>
        </div>
        <div class="inline-flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-amber-500"></span>
          <span>Perlu Dibersihkan (Dirty)</span>
        </div>
        <div class="inline-flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-gray-500"></span>
          <span>Perawatan (Maintenance)</span>
        </div>
      </div>
    </div>

    <!-- Area Tabel Matriks Scrollable -->
    <div class="relative overflow-x-auto max-w-full">
      <div v-if="loading" class="py-16 text-center">
        <RotateCw class="w-8 h-8 text-[#8B5A2B] animate-spin mx-auto mb-2" />
        <p class="text-xs text-[#6B7280] dark:text-gray-400">Memuat data ketersediaan kamar...</p>
      </div>

      <table v-else class="w-full text-xs border-collapse min-w-[760px]">
        <thead>
          <tr class="bg-[#F9FAFB] dark:bg-[#1F2937]/80 border-b border-[#E5E7EB] dark:border-[#4A5568]">
            <!-- Kolom Unit (Sticky Left) -->
            <th
              scope="col"
              class="sticky left-0 z-20 bg-[#F9FAFB] dark:bg-[#1F2937] px-4 py-3 text-left font-bold text-[#1A3A4A] dark:text-white border-r border-[#E5E7EB] dark:border-[#4A5568] w-[140px] shadow-[2px_0_5px_-2px_rgba(0,0,0,0.06)]"
            >
              <div class="flex items-center gap-1.5 text-xs">
                <Bed class="w-3.5 h-3.5 text-[#8B5A2B]" />
                <span>Unit Kamar</span>
              </div>
            </th>

            <!-- Kolom Tiap Hari (7 Hari) -->
            <th
              v-for="d in days"
              :key="d"
              scope="col"
              class="px-2 py-2.5 text-center border-r border-[#E5E7EB] dark:border-[#4A5568] last:border-r-0 min-w-[95px]"
              :class="isToday(d) ? 'bg-[#FDF6EC]/70 dark:bg-[#1A3A4A]/40' : ''"
            >
              <div class="flex flex-col items-center justify-center">
                <span
                  v-if="isToday(d)"
                  class="inline-block bg-[#8B5A2B] text-white text-[9px] font-bold px-1.5 py-0.2 rounded-full mb-0.5 tracking-wider uppercase"
                >
                  Hari Ini
                </span>
                <span class="font-bold text-[#1A3A4A] dark:text-white text-xs">
                  {{ new Date(d).toLocaleDateString('id-ID', { weekday: 'short' }) }}
                </span>
                <span class="text-[10px] text-[#6B7280] dark:text-gray-400">
                  {{ new Date(d).toLocaleDateString('id-ID', { day: '2-digit', month: 'short' }) }}
                </span>
              </div>
            </th>
          </tr>
        </thead>

        <tbody class="divide-y divide-[#E5E7EB] dark:divide-[#4A5568]">
          <tr
            v-for="unit in displayUnits"
            :key="unit.id"
            class="hover:bg-gray-50/50 dark:hover:bg-gray-800/40 transition-colors"
          >
            <!-- Header Unit Kamar (Sticky Left) -->
            <td
              class="sticky left-0 z-10 bg-white dark:bg-[#2D3748] px-4 py-3 border-r border-[#E5E7EB] dark:border-[#4A5568] shadow-[2px_0_5px_-2px_rgba(0,0,0,0.06)]"
            >
              <div class="flex flex-col">
                <span class="font-bold text-[#1A3A4A] dark:text-white text-xs tracking-wide">
                  {{ unit.code || 'UNIT-' + unit.id }}
                </span>
                <span class="text-[10px] text-[#6B7280] dark:text-gray-400">
                  {{ getRoomTypeName(unit.room_type_id) }}
                </span>
              </div>
            </td>

            <!-- Sel Status untuk Masing-masing Tanggal -->
            <td
              v-for="d in days"
              :key="d + '-' + unit.id"
              class="p-1.5 text-center border-r border-[#E5E7EB] dark:border-[#4A5568] last:border-r-0 align-middle"
              :class="isToday(d) ? 'bg-[#FDF6EC]/30 dark:bg-[#1A3A4A]/20' : ''"
            >
              <template v-for="cell in [getUnitCellInfo(unit, d)]" :key="cell.status">
                <div
                  class="rounded-xl px-2 py-2 border text-[11px] font-medium transition flex items-center justify-center gap-1 select-none"
                  :class="[
                    statusMeta[cell.status]?.bg || 'bg-gray-50 text-gray-700',
                    statusMeta[cell.status]?.border || 'border-gray-200',
                    statusMeta[cell.status]?.text || 'text-gray-700'
                  ]"
                  :title="`${unit.code} pada ${d}: ${statusMeta[cell.status]?.label || cell.status}`"
                >
                  <component
                    :is="statusMeta[cell.status]?.icon"
                    class="w-3 h-3 shrink-0"
                    :class="cell.status === 'occupied' ? 'text-white' : ''"
                  />
                  <span class="truncate max-w-[65px] font-semibold text-[10px]">
                    {{ cell.label }}
                  </span>
                </div>
              </template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Footer Catatan Operasional -->
    <div class="px-4 py-3 bg-[#F9FAFB] dark:bg-[#1F2937]/60 border-t border-[#E5E7EB] dark:border-[#4A5568] flex items-center justify-between text-xs text-[#6B7280] dark:text-gray-400">
      <div class="flex items-center gap-1.5">
        <Info class="w-3.5 h-3.5 text-[#8B5A2B]" />
        <span>Matriks otomatis mencerminkan data reservasi aktif dan status kamar fisik.</span>
      </div>
      <span class="hidden sm:inline text-[11px]">Check-in 14:00 · Check-out 12:00</span>
    </div>
  </div>
</template>
