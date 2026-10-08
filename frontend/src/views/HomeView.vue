<script setup>
import Card from 'primevue/card'
import Button from 'primevue/button'
import DatePicker from 'primevue/datepicker'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import { Bed, Users, Calendar, Star, MapPin, Wifi, Coffee, Waves } from 'lucide-vue-next'
import { ref, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import client from '../api/client'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const toast = useToast()
const auth = useAuthStore()

const checkIn = ref(null)
const checkOut = ref(null)
const guests = ref(2)
const availMap = ref({}) // { roomTypeId: {available, occupied, total_units} }
const loadingAvail = ref(false)
const bookingLoading = ref('')
const availError = ref('')

const rooms = [
  { id: 1, type: 'Standard', price: 350000, cap: 2, facility: 'Smart TV · Breakfast', img: 'https://images.unsplash.com/photo-1566665797739-1674de7a421a?w=600&q=80&auto=format&fit=crop', rating: 4.6, icon: Bed },
  { id: 2, type: 'Deluxe', price: 550000, cap: 2, facility: 'Balkon · Mini fridge', img: 'https://images.unsplash.com/photo-1611892440504-42a792e24d32?w=600&q=80&auto=format&fit=crop', rating: 4.8, icon: Coffee },
  { id: 3, type: 'Family', price: 850000, cap: 4, facility: '2 Bedroom · Kitchen', img: 'https://images.unsplash.com/photo-1590490360182-c33d57733427?w=600&q=80&auto=format&fit=crop', rating: 4.9, icon: Users },
  { id: 4, type: 'Suite', price: 1250000, cap: 3, facility: 'Living room · Bathtub', img: 'https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=600&q=80&auto=format&fit=crop', rating: 5.0, icon: Waves },
]

const fmt = (n) => new Intl.NumberFormat('id-ID').format(n)

function toISO(d) {
  if (!d) return ''
  const dt = d instanceof Date ? d : new Date(d)
  if (isNaN(dt)) return ''
  const y = dt.getFullYear()
  const m = String(dt.getMonth() + 1).padStart(2, '0')
  const day = String(dt.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const canSearch = computed(() => checkIn.value && checkOut.value && toISO(checkIn.value) && toISO(checkOut.value))

async function fetchAvailability() {
  if (!canSearch.value) {
    availMap.value = {}
    availError.value = ''
    return
  }
  const ci = toISO(checkIn.value)
  const co = toISO(checkOut.value)
  if (ci >= co) {
    availError.value = 'Check-out harus setelah check-in'
    return
  }
  availError.value = ''
  loadingAvail.value = true
  try {
    const results = await Promise.all(rooms.map(async (r) => {
      try {
        const { data } = await client.get('/api/availability', { params: { room_type_id: r.id, check_in: ci, check_out: co } })
        const payload = data.data || data
        return [r.id, payload]
      } catch (e) {
        return [r.id, null]
      }
    }))
    const m = {}
    results.forEach(([id, payload]) => { if (payload) m[id] = payload })
    availMap.value = m
  } finally {
    loadingAvail.value = false
  }
}

watch([checkIn, checkOut], fetchAvailability)

async function onSearch() {
  await fetchAvailability()
  if (Object.keys(availMap.value).length) {
    toast.add({ severity: 'info', summary: 'Ketersediaan diperbarui', detail: `${Object.keys(availMap.value).length} tipe diperiksa`, life: 2000 })
  }
}

async function onBooking(room) {
  if (!canSearch.value) {
    toast.add({ severity: 'warn', summary: 'Pilih tanggal', detail: 'Pilih check-in & check-out dulu', life: 2500 })
    return
  }
  if (!auth.isAuthenticated) {
    toast.add({ severity: 'warn', summary: 'Login diperlukan', detail: 'Silakan login dulu untuk booking', life: 2500 })
    router.push('/login')
    return
  }
  const ci = toISO(checkIn.value)
  const co = toISO(checkOut.value)
  const avail = availMap.value[room.id]
  if (avail && avail.available <= 0) {
    toast.add({ severity: 'error', summary: 'Penuh', detail: `${room.type} penuh di tanggal tersebut`, life: 2500 })
    return
  }
  bookingLoading.value = room.type
  try {
    const { data } = await client.post('/api/bookings', {
      room_type_id: room.id,
      check_in: ci,
      check_out: co,
      guests: guests.value,
    })
    const booking = data.data || data
    toast.add({ severity: 'success', summary: 'Booking berhasil', detail: `Booking #${booking.id || ''} pending payment — Rp ${fmt(booking.total_price || room.price)}`, life: 4000 })
    // refresh availability after booking
    await fetchAvailability()
  } catch (e) {
    const msg = e?.response?.data?.message || e.message || 'Booking gagal'
    toast.add({ severity: 'error', summary: 'Gagal booking', detail: msg, life: 3500 })
  } finally {
    bookingLoading.value = ''
  }
}

function availText(room) {
  const a = availMap.value[room.id]
  if (!a) return ''
  return `${a.available} tersedia / ${a.total_units} unit`
}
function availSeverity(room) {
  const a = availMap.value[room.id]
  if (!a) return 'secondary'
  if (a.available <= 0) return 'danger'
  if (a.available <= 2) return 'warn'
  return 'success'
}
</script>

<template>
  <div>
    <!-- Hero -->
    <section class="relative overflow-hidden bg-[#1A3A4A]">
      <img src="https://images.unsplash.com/photo-1566073771259-6a8506099945?w=1400&q=80&auto=format&fit=crop" alt="Hotel lobby warm" class="absolute inset-0 w-full h-full object-cover opacity-50" />
      <div class="absolute inset-0 bg-gradient-to-t from-[#1A3A4A]/80 via-[#1A3A4A]/30 to-transparent"></div>
      <div class="relative max-w-7xl mx-auto px-4 sm:px-6 py-16 md:py-24 text-center text-white">
        <p class="inline-flex items-center gap-2 bg-white/15 backdrop-blur rounded-full px-4 py-1.5 text-xs tracking-widest uppercase"><MapPin class="w-3.5 h-3.5 text-[#C9A86A]" /> Ubud · Bali · Since 2024</p>
        <h1 class="font-display font-bold text-4xl md:text-5xl leading-tight mt-4">Menginap hangat,<br /><span class="text-[#C9A86A]">seperti di rumah.</span></h1>
        <p class="mt-4 text-white/80 max-w-2xl mx-auto">4 tipe kamar curated — Standard hingga Suite. Harga transparan, booking 3 langkah, konfirmasi real-time.</p>

        <!-- Search box -->
        <div class="mt-8 bg-white rounded-2xl shadow-lg p-4 md:p-5 flex flex-col md:flex-row gap-3 items-stretch md:items-end text-left max-w-4xl mx-auto">
          <div class="flex-1">
            <label class="text-xs font-semibold text-[#6B7280] uppercase tracking-wide flex items-center gap-1.5"><Calendar class="w-3.5 h-3.5" /> Check-in</label>
            <DatePicker v-model="checkIn" placeholder="Pilih tanggal" dateFormat="dd/mm/yy" showIcon class="w-full mt-1" />
          </div>
          <div class="flex-1">
            <label class="text-xs font-semibold text-[#6B7280] uppercase tracking-wide flex items-center gap-1.5"><Calendar class="w-3.5 h-3.5" /> Check-out</label>
            <DatePicker v-model="checkOut" placeholder="Pilih tanggal" dateFormat="dd/mm/yy" showIcon class="w-full mt-1" />
          </div>
          <div class="w-full md:w-36">
            <label class="text-xs font-semibold text-[#6B7280] uppercase tracking-wide flex items-center gap-1.5"><Users class="w-3.5 h-3.5" /> Tamu</label>
            <select v-model="guests" class="mt-1 w-full border border-[#E5E7EB] rounded-xl px-3 py-2.5 text-sm text-[#1F2937] bg-white focus:outline-none focus:ring-2 focus:ring-[#8B5A2B]/30 focus:border-[#8B5A2B]">
              <option :value="1">1 Tamu</option>
              <option :value="2">2 Tamu</option>
              <option :value="3">3 Tamu</option>
              <option :value="4">4 Tamu</option>
            </select>
          </div>
          <Button :label="loadingAvail ? 'Mencari...' : 'Cari Kamar'" :loading="loadingAvail" icon="pi pi-search" class="md:w-auto w-full !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl !px-8 !py-3 font-semibold whitespace-nowrap" @click="onSearch" />
        </div>
        <Message v-if="availError" severity="error" class="max-w-4xl mx-auto mt-3 text-left">{{ availError }}</Message>
        <p class="mt-3 text-xs text-white/60">Free cancellation · Bayar di hotel · No hidden fee</p>
      </div>
    </section>

    <!-- Rooms -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 py-12 md:py-16">
      <div class="flex items-end justify-between gap-4">
        <div>
          <h2 class="font-display font-bold text-2xl md:text-3xl text-[#1A3A4A]">Pilih tipe kamar</h2>
          <p class="text-[#6B7280] text-sm mt-1">4 tipe locked — harga per malam, sudah termasuk pajak.</p>
        </div>
        <span class="hidden md:inline-flex items-center gap-1.5 text-xs bg-[#FDF6EC] text-[#8B5A2B] border border-[#8B5A2B]/20 rounded-full px-3 py-1.5 font-semibold"><Wifi class="w-3.5 h-3.5" /> Free Wi-Fi · Breakfast</span>
      </div>

      <div v-if="canSearch && loadingAvail" class="mt-6 text-center text-sm text-[#6B7280]">Memeriksa ketersediaan...</div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mt-8">
        <Card v-for="r in rooms" :key="r.type" class="overflow-hidden !rounded-xl !shadow-sm hover:!shadow-md hover:-translate-y-0.5 transition-all duration-200 !border !border-[#E5E7EB]">
          <template #header>
            <div class="relative">
              <img :src="r.img" :alt="'Kamar ' + r.type" class="w-full aspect-[16/10] object-cover" />
              <span class="absolute top-3 left-3 bg-white/95 backdrop-blur text-[#1A3A4A] text-xs font-bold rounded-full px-2.5 py-1 flex items-center gap-1 shadow-sm"><Star class="w-3.5 h-3.5 text-[#C9A86A] fill-[#C9A86A]" /> {{ r.rating }}</span>
              <span class="absolute top-3 right-3 bg-[#8B5A2B] text-white text-xs font-semibold rounded-full px-2.5 py-1">{{ r.type }}</span>
            </div>
          </template>
          <template #title><span class="text-[#1A3A4A] font-display font-semibold">{{ r.type }}</span></template>
          <template #subtitle><span class="text-xs text-[#6B7280] flex items-center gap-1.5"><Bed class="w-3.5 h-3.5" /> Kapasitas {{ r.cap }} orang · {{ r.facility }}</span></template>
          <template #content>
            <p class="font-bold text-[#8B5A2B] text-lg leading-none">Rp {{ fmt(r.price) }} <span class="font-normal text-sm text-[#6B7280]">/ malam</span></p>
            <div v-if="availMap[r.id]" class="mt-2">
              <Tag :value="availText(r)" :severity="availSeverity(r)" rounded class="text-xs" />
              <p class="text-xs text-[#6B7280] mt-1">Occupied: {{ availMap[r.id].occupied }} · Available: {{ availMap[r.id].available }}</p>
            </div>
            <p v-else-if="canSearch" class="text-xs text-[#9CA3AF] mt-2">Pilih tanggal untuk cek ketersediaan</p>
          </template>
          <template #footer>
            <div class="flex gap-2 pt-1">
              <Button label="Lihat Detail" outlined class="!rounded-xl !text-[#8B5A2B] !border-[#8B5A2B] flex-1 !py-2 text-sm" />
              <Button :label="bookingLoading === r.type ? 'Booking...' : 'Booking'" :loading="bookingLoading === r.type" :disabled="availMap[r.id] && availMap[r.id].available <= 0" class="!bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl flex-1 !py-2 text-sm font-semibold disabled:!bg-gray-300 disabled:!border-gray-300" @click="onBooking(r)" />
            </div>
          </template>
        </Card>
      </div>
    </section>

    <section class="bg-[#FDF6EC] border-y border-[#E5E7EB]/60">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 flex flex-col md:flex-row gap-6 justify-between text-sm">
        <p class="text-[#1A3A4A] font-semibold">Butuh bantuan? <span class="font-normal text-[#6B7280]">WA 0812-3456-7890 · check-in 14:00, check-out 12:00 · 1 booking = 1 tipe kamar</span></p>
        <p class="text-[#6B7280]">Harga snapshot saat booking · Voucher opsional</p>
      </div>
    </section>
  </div>
</template>
