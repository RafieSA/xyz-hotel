<script setup>
import { ref, computed, watch } from 'vue'
import DatePicker from 'primevue/datepicker'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import { Calendar, Users, Ticket, Search, CheckCircle2, AlertCircle, X, ChevronDown, ChevronUp } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  checkIn: {
    type: [Date, String, null],
    default: null,
  },
  checkOut: {
    type: [Date, String, null],
    default: null,
  },
  guests: {
    type: Number,
    default: 2,
  },
  loadingAvail: {
    type: Boolean,
    default: false,
  },
  voucherCode: {
    type: String,
    default: '',
  },
  voucherValidating: {
    type: Boolean,
    default: false,
  },
  voucherInfo: {
    type: Object,
    default: null,
  },
  voucherError: {
    type: String,
    default: '',
  },
  availError: {
    type: String,
    default: '',
  },
})

const emit = defineEmits([
  'update:checkIn',
  'update:checkOut',
  'update:guests',
  'update:voucherCode',
  'search',
  'validate-voucher',
  'clear-voucher',
])

const { t } = useI18n()

// Tanggal hari ini sebagai batas awal check-in
const today = new Date()
today.setHours(0, 0, 0, 0)

// Tanggal minimal check-out adalah 1 hari setelah check-in
const minCheckOutDate = computed(() => {
  if (!props.checkIn) {
    const d = new Date()
    d.setDate(d.getDate() + 1)
    return d
  }
  const d = new Date(props.checkIn)
  d.setDate(d.getDate() + 1)
  return d
})

// Jika user memilih check-in yang melewati check-out yang ada, sesuaikan check-out otomatis
watch(() => props.checkIn, (newCheckIn) => {
  if (newCheckIn && props.checkOut) {
    const ci = new Date(newCheckIn).getTime()
    const co = new Date(props.checkOut).getTime()
    if (co <= ci) {
      const nextDay = new Date(newCheckIn)
      nextDay.setDate(nextDay.getDate() + 1)
      emit('update:checkOut', nextDay)
    }
  }
})

// State buka-tutup kolom voucher
const showVoucherInput = ref(false)
const localVoucherCode = ref(props.voucherCode)

watch(() => props.voucherCode, (v) => {
  localVoucherCode.value = v
})

function onApplyVoucher() {
  emit('update:voucherCode', localVoucherCode.value)
  emit('validate-voucher', localVoucherCode.value)
}

function onRemoveVoucher() {
  localVoucherCode.value = ''
  emit('update:voucherCode', '')
  emit('clear-voucher')
}
</script>

<template>
  <div class="w-full max-w-4xl mx-auto">
    <!-- Main Search Card (Floating Bar) -->
    <div class="bg-white/95 dark:bg-[#2D3748]/95 backdrop-blur-md rounded-2xl md:rounded-3xl shadow-xl border border-white/60 dark:border-gray-700 p-4 md:p-6 text-left">
      <!-- Grid Field: CheckIn, CheckOut, Guests, Button -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-12 gap-3.5 items-end">
        <!-- Field Check-in -->
        <div class="lg:col-span-4">
          <label class="block text-[11px] font-bold text-[#6B7280] dark:text-[#C9A86A] uppercase tracking-wider mb-1.5 flex items-center gap-1.5">
            <Calendar class="w-3.5 h-3.5 text-[#8B5A2B]" />
            {{ t('search.checkin') }}
          </label>
          <DatePicker
            :modelValue="checkIn"
            @update:modelValue="emit('update:checkIn', $event)"
            :minDate="today"
            :placeholder="t('search.selectDate')"
            dateFormat="dd/mm/yy"
            showIcon
            class="w-full !rounded-xl"
            inputClass="!py-2.5 !text-sm !font-medium !text-[#1F2937] dark:!text-white"
          />
        </div>

        <!-- Field Check-out -->
        <div class="lg:col-span-4">
          <label class="block text-[11px] font-bold text-[#6B7280] dark:text-[#C9A86A] uppercase tracking-wider mb-1.5 flex items-center gap-1.5">
            <Calendar class="w-3.5 h-3.5 text-[#8B5A2B]" />
            {{ t('search.checkout') }}
          </label>
          <DatePicker
            :modelValue="checkOut"
            @update:modelValue="emit('update:checkOut', $event)"
            :minDate="minCheckOutDate"
            :placeholder="t('search.selectDate')"
            dateFormat="dd/mm/yy"
            showIcon
            class="w-full !rounded-xl"
            inputClass="!py-2.5 !text-sm !font-medium !text-[#1F2937] dark:!text-white"
          />
        </div>

        <!-- Field Guests -->
        <div class="sm:col-span-1 lg:col-span-2">
          <label class="block text-[11px] font-bold text-[#6B7280] dark:text-[#C9A86A] uppercase tracking-wider mb-1.5 flex items-center gap-1.5">
            <Users class="w-3.5 h-3.5 text-[#8B5A2B]" />
            {{ t('search.guests') }}
          </label>
          <div class="relative">
            <select
              :value="guests"
              @change="emit('update:guests', Number($event.target.value))"
              class="w-full border border-[#E5E7EB] dark:border-gray-600 rounded-xl px-3 py-2.5 text-sm font-medium text-[#1F2937] dark:text-white bg-white dark:bg-[#1A3A4A] focus:outline-none focus:ring-2 focus:ring-[#8B5A2B]/30 focus:border-[#8B5A2B] transition"
            >
              <option :value="1">{{ t('search.guest1') }}</option>
              <option :value="2">{{ t('search.guest2') }}</option>
              <option :value="3">{{ t('search.guest3') }}</option>
              <option :value="4">{{ t('search.guest4') }}</option>
            </select>
          </div>
        </div>

        <!-- Tombol Cari -->
        <div class="sm:col-span-1 lg:col-span-2">
          <Button
            :label="loadingAvail ? 'Mencari...' : 'Cari Kamar'"
            :loading="loadingAvail"
            icon="pi pi-search"
            class="w-full !min-h-[44px] !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] text-white !rounded-xl font-semibold !text-sm shadow-md hover:shadow-lg transition flex items-center justify-center whitespace-nowrap"
            @click="emit('search')"
          />
        </div>
      </div>

      <!-- Area Pesan Error Tanggal (Jika ada) -->
      <div v-if="availError" class="mt-3 flex items-center gap-2 p-3 bg-red-50 dark:bg-red-950/40 text-red-700 dark:text-red-300 rounded-xl text-xs font-medium border border-red-200 dark:border-red-800">
        <AlertCircle class="w-4 h-4 shrink-0 text-red-600" />
        <span>{{ availError }}</span>
      </div>

      <!-- Bagian Voucher Terpadu (Accordion Ringkas) -->
      <div class="mt-4 pt-3.5 border-t border-[#F3F4F6] dark:border-gray-700/80 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 text-xs">
        <!-- Status Jika Voucher Sudah Terpasang -->
        <div v-if="voucherInfo" class="flex items-center gap-2 bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 px-3 py-1.5 rounded-xl text-emerald-800 dark:text-emerald-300">
          <CheckCircle2 class="w-4 h-4 text-emerald-600" />
          <span class="font-medium">
            Kode <b>{{ voucherCode }}</b> aktif: Diskon <b>{{ voucherInfo.discount_percent ?? voucherInfo.discount }}%</b>
          </span>
          <button
            type="button"
            @click="onRemoveVoucher"
            class="ml-2 hover:bg-emerald-200/50 p-1 rounded-full text-emerald-700 dark:text-emerald-300 transition"
            title="Hapus voucher"
          >
            <X class="w-3.5 h-3.5" />
          </button>
        </div>

        <!-- Tombol Buka/Tutup Input Voucher Jika Belum Ada Voucher -->
        <button
          v-else
          type="button"
          @click="showVoucherInput = !showVoucherInput"
          class="inline-flex items-center gap-1.5 text-[#8B5A2B] dark:text-[#C9A86A] font-semibold hover:underline cursor-pointer"
        >
          <Ticket class="w-3.5 h-3.5" />
          <span>{{ showVoucherInput ? 'Tutup kolom kode promo' : 'Punya kode promo / voucher?' }}</span>
          <component :is="showVoucherInput ? ChevronUp : ChevronDown" class="w-3 h-3" />
        </button>

        <!-- Informasi Ringkas Jaminan Hotel -->
        <div class="text-[#6B7280] dark:text-gray-400 text-[11px] sm:text-right">
          <span>✨ Pembatalan gratis · Tanpa biaya tersembunyi</span>
        </div>
      </div>

      <!-- Form Input Voucher (Hanya Muncul Jika Diklik dan Belum Ada Voucher) -->
      <div v-if="showVoucherInput && !voucherInfo" class="mt-3 bg-[#FDF6EC] dark:bg-[#1A3A4A]/60 rounded-xl p-3 border border-[#E5E7EB] dark:border-gray-700 flex flex-col sm:flex-row gap-2.5 items-stretch sm:items-center">
        <div class="relative flex-1">
          <InputText
            v-model="localVoucherCode"
            placeholder="Contoh: HEMAT20"
            class="w-full !py-2 !px-3 !text-xs md:!text-sm !rounded-lg"
            @keyup.enter="onApplyVoucher"
          />
        </div>
        <Button
          label="Terapkan"
          :loading="voucherValidating"
          class="!min-h-[38px] !px-5 !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] text-white !rounded-lg !text-xs font-semibold whitespace-nowrap shadow-sm"
          @click="onApplyVoucher"
        />
      </div>

      <!-- Pesan Error Voucher -->
      <div v-if="voucherError" class="mt-2 text-xs text-red-600 dark:text-red-400 flex items-center gap-1">
        <AlertCircle class="w-3.5 h-3.5 shrink-0" />
        <span>{{ voucherError }}</span>
      </div>
    </div>
  </div>
</template>
