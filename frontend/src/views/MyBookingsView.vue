<script setup>
import { ref, onMounted } from 'vue'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import FileUpload from 'primevue/fileupload'
import { useToast } from 'primevue/usetoast'
import client from '../api/client'
import { Upload, Eye, Calendar, Receipt } from 'lucide-vue-next'

const toast = useToast()
const bookings = ref([])
const loading = ref(false)
const uploadingId = ref(null)
const selectedFile = ref({})

const statusSeverity = (s) => ({ pending_payment: 'warn', waiting_verification: 'warn', verified: 'success', checked_in: 'info', checked_out: 'secondary', rejected: 'danger', expired: 'danger', cancelled: 'danger' }[s] || 'secondary')
const fmt = (n) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n)
const fmtDate = (d) => { try { return new Date(d).toLocaleDateString('id-ID') } catch { return d } }

async function fetchBookings() {
  loading.value = true
  try {
    const { data } = await client.get('/api/bookings')
    const list = data.data || data
    bookings.value = Array.isArray(list) ? list : []
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Gagal memuat booking', detail: e?.response?.data?.message || e.message, life: 3000 })
  } finally { loading.value = false }
}

function onFileSelect(event, bookingId) {
  const file = event.files?.[0] || event.target?.files?.[0]
  if (file) selectedFile.value[bookingId] = file
}

async function uploadProof(booking) {
  const file = selectedFile.value[booking.id]
  if (!file) { toast.add({ severity: 'warn', summary: 'Pilih file dulu', life: 2000 }); return }
  if (file.size > 5 * 1024 * 1024) { toast.add({ severity: 'error', summary: 'File terlalu besar', detail: 'Max 5MB', life: 2500 }); return }
  const ext = '.' + file.name.split('.').pop().toLowerCase()
  if (!['.jpg','.jpeg','.png','.pdf'].includes(ext)) { toast.add({ severity: 'error', summary: 'Ekstensi tidak valid', detail: 'Hanya jpg/png/pdf', life: 2500 }); return }
  uploadingId.value = booking.id
  try {
    const form = new FormData()
    form.append('proof', file)
    const { data } = await client.post(`/api/bookings/${booking.id}/proof`, form, { headers: { 'Content-Type': 'multipart/form-data' } })
    toast.add({ severity: 'success', summary: 'Proof terupload', detail: data.message || 'Waiting verification', life: 3000 })
    await fetchBookings()
    selectedFile.value[booking.id] = null
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Upload gagal', detail: e?.response?.data?.message || e.message, life: 3500 })
  } finally { uploadingId.value = null }
}

onMounted(fetchBookings)
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 space-y-6">
    <div class="flex items-center gap-3">
      <span class="bg-[#8B5A2B] text-white rounded-xl p-2.5"><Receipt class="w-5 h-5" /></span>
      <div>
        <h1 class="font-display font-bold text-2xl text-[#1A3A4A]">Booking Saya</h1>
        <p class="text-sm text-[#6B7280]">Kelola bukti pembayaran & status booking Anda.</p>
      </div>
      <Button label="Muat ulang" icon="pi pi-refresh" outlined class="!rounded-xl !ml-auto !border-[#8B5A2B] !text-[#8B5A2B]" :loading="loading" @click="fetchBookings" />
    </div>

    <Message v-if="!bookings.length && !loading" severity="info">Belum ada booking. Silakan booking dari halaman utama.</Message>

    <div v-for="b in bookings" :key="b.id" class="border border-[#E5E7EB] rounded-xl bg-white shadow-sm p-5 space-y-4">
      <div class="flex flex-wrap gap-3 items-start justify-between">
        <div>
          <p class="font-bold text-[#1A3A4A]">Booking #{{ b.id }} <Tag :value="b.status.replace('_',' ')" :severity="statusSeverity(b.status)" rounded class="ml-2 capitalize" /></p>
          <p class="text-sm text-[#6B7280] flex items-center gap-1.5 mt-1"><Calendar class="w-4 h-4" /> {{ fmtDate(b.check_in) }} → {{ fmtDate(b.check_out) }} · {{ b.guests }} tamu · Room Type #{{ b.room_type_id }}</p>
          <p class="text-sm font-semibold text-[#8B5A2B] mt-1">{{ fmt(b.total_price) }}</p>
          <p v-if="b.voucher_id" class="text-xs text-[#6B7280] mt-1">Voucher applied: #{{ b.voucher_id }}</p>
        </div>
        <div v-if="b.proof_url" class="text-right">
          <p class="text-xs text-[#6B7280] mb-1">Bukti pembayaran</p>
          <img v-if="b.proof_url.endsWith('.pdf')" src="https://cdn-icons-png.flaticon.com/512/337/337946.png" alt="pdf" class="w-20 h-20 object-contain border rounded" />
          <img v-else :src="`http://localhost:8080/storage/uploads/${b.proof_url}`" :alt="'proof-'+b.id" class="w-28 h-20 object-cover rounded border" @error="(e)=>e.target.style.display='none'" />
          <p class="text-xs text-[#6B7280] mt-1 break-all">{{ b.proof_url }}</p>
        </div>
      </div>

      <!-- Upload proof only when pending_payment -->
      <div v-if="b.status==='pending_payment'" class="bg-[#FDF6EC] border border-[#E5E7EB] rounded-xl p-4 space-y-3">
        <p class="text-sm font-semibold text-[#1A3A4A] flex items-center gap-2"><Upload class="w-4 h-4 text-[#8B5A2B]" /> Upload Bukti Bayar (JPG/PNG/PDF, max 5MB)</p>
        <div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
          <input type="file" accept=".jpg,.jpeg,.png,.pdf" class="text-sm file:mr-3 file:py-2 file:px-4 file:rounded-full file:border-0 file:bg-[#8B5A2B] file:text-white file:text-sm hover:file:bg-[#6F4620] file:cursor-pointer" @change="e=>onFileSelect({target:e}, b.id)" />
          <span v-if="selectedFile[b.id]" class="text-xs text-[#6B7280]">{{ selectedFile[b.id].name }} ({{ (selectedFile[b.id].size/1024).toFixed(1) }} KB)</span>
          <Button :label="uploadingId===b.id ? 'Mengupload...' : 'Upload Proof'" icon="pi pi-upload" :loading="uploadingId===b.id" class="!bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl !ml-auto" @click="uploadProof(b)" />
        </div>
        <p class="text-xs text-[#6B7280]">File akan disimpan sebagai <code>bookings/{{'{'}}id{{'}'}}_{{ '{' }}uuid{{ '}' }}.ext</code> di backend storage (private).</p>
      </div>

      <div v-else-if="b.status==='waiting_verification'" class="text-sm text-amber-700 bg-amber-50 border border-amber-200 rounded-lg px-3 py-2">Menunggu verifikasi admin. Anda akan mendapat konfirmasi setelah proof diverifikasi.</div>
      <div v-else-if="b.status==='verified'" class="text-sm text-green-700 bg-green-50 border border-green-200 rounded-lg px-3 py-2">Terverifikasi! Silakan check-in sesuai jadwal.</div>
      <div v-else-if="b.status==='rejected'" class="text-sm text-red-700 bg-red-50 border border-red-200 rounded-lg px-3 py-2">Ditolak<span v-if="b.reject_reason">: {{ b.reject_reason }}</span></div>
      <div v-else-if="b.status==='checked_in'" class="text-sm text-blue-700 bg-blue-50 border border-blue-200 rounded-lg px-3 py-2 flex items-center gap-2"><Eye class="w-4 h-4" /> Sudah check-in<span v-if="b.room_unit_id"> · Unit #{{ b.room_unit_id }}</span></div>
      <div v-else-if="b.status==='checked_out'" class="text-sm text-gray-700 bg-gray-50 border border-gray-200 rounded-lg px-3 py-2">Check-out selesai. Unit kamar telah ditandai dirty.</div>
    </div>
  </div>
</template>
