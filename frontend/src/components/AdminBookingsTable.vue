<script setup>
import { ref, computed } from 'vue'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import {
  FileText,
  Check,
  X,
  LogIn,
  LogOut,
  Download,
  Eye,
  Calendar,
  Image as ImageIcon,
  AlertCircle,
  ExternalLink,
  Search,
  Filter
} from 'lucide-vue-next'

const props = defineProps({
  bookings: {
    type: Array,
    required: true,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  actionLoading: {
    type: String,
    default: '',
  },
  invoicingId: {
    type: [Number, String, null],
    default: null,
  },
  verifySuccessId: {
    type: [Number, String, null],
    default: null,
  },
})

const emit = defineEmits([
  'verify',
  'confirm-reject',
  'check-in',
  'check-out',
  'download-invoice',
  'refresh',
])

const fmt = (n) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n || 0)
const fmtDate = (d) => {
  if (!d) return '-'
  try {
    return new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  } catch {
    return d
  }
}

// Filter pencarian teks & status
const searchQuery = ref('')
const selectedStatusFilter = ref('all')

const filteredBookings = computed(() => {
  let list = props.bookings || []
  if (selectedStatusFilter.value !== 'all') {
    list = list.filter(b => b.status === selectedStatusFilter.value)
  }
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter(b => {
      const idMatch = String(b.id).includes(q)
      const userMatch = String(b.user_id || b.userId || '').includes(q)
      const guestNameMatch = String(b.guest_name || b.user_name || '').toLowerCase().includes(q)
      const statusMatch = String(b.status || '').toLowerCase().includes(q)
      return idMatch || userMatch || guestNameMatch || statusMatch
    })
  }
  return list
})

// Status styling & label
function getStatusSeverity(status) {
  switch (status) {
    case 'verified': return 'success'
    case 'checked_in': return 'info'
    case 'waiting_verification': return 'warn'
    case 'pending_payment': return 'warn'
    case 'checked_out': return 'secondary'
    case 'cancelled':
    case 'rejected':
    case 'expired': return 'danger'
    default: return 'secondary'
  }
}

function getStatusLabel(status) {
  switch (status) {
    case 'waiting_verification': return 'Menunggu Verifikasi'
    case 'pending_payment': return 'Menunggu Pembayaran'
    case 'verified': return 'Lunas (Siap Check-In)'
    case 'checked_in': return 'Sedang Menginap'
    case 'checked_out': return 'Selesai (Checked-Out)'
    case 'rejected': return 'Ditolak'
    case 'cancelled': return 'Dibatalkan'
    case 'expired': return 'Kadaluarsa'
    default: return status || '-'
  }
}

function canInvoice(status) {
  return ['verified', 'checked_in', 'checked_out'].includes(status)
}

// State Modal Detail & Modal Penolakan
const selectedBooking = ref(null)
const showDetailModal = ref(false)
const showRejectModal = ref(false)
const rejectReason = ref('')
const rejectTargetId = ref(null)

function openDetail(row) {
  selectedBooking.value = row
  showDetailModal.value = true
}

function openRejectDialog(id) {
  rejectTargetId.value = id
  rejectReason.value = ''
  showRejectModal.value = true
}

function submitReject() {
  if (!rejectReason.value.trim()) return
  emit('confirm-reject', { id: rejectTargetId.value, reason: rejectReason.value.trim() })
  showRejectModal.value = false
}
</script>

<template>
  <div class="bg-white dark:bg-[#2D3748] rounded-3xl border border-[#E5E7EB] dark:border-gray-700 shadow-sm p-4 md:p-6">
    <!-- Header Tabel: Filter & Search Bar -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-5 border-b border-[#F3F4F6] dark:border-gray-700">
      <div>
        <div class="flex items-center gap-2">
          <FileText class="w-5 h-5 text-[#8B5A2B]" />
          <h2 class="font-display font-bold text-xl text-[#1A3A4A] dark:text-white" style="font-family:'Playfair Display',serif">
            Daftar Reservasi & Tindakan Resepsionis
          </h2>
        </div>
        <p class="text-xs text-[#6B7280] dark:text-gray-400 mt-1">
          Verifikasi bukti transfer, kelola status kedatangan tamu (*Check-In/Out*), dan terbitkan invoice.
        </p>
      </div>

      <!-- Quick Search Bar -->
      <div class="flex items-center gap-2.5">
        <div class="relative w-full sm:w-64">
          <Search class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari ID, nama tamu..."
            class="w-full pl-9 pr-3 py-2 text-xs rounded-xl border border-[#E5E7EB] dark:border-gray-600 bg-[#F9FAFB] dark:bg-[#1A3A4A] text-[#1F2937] dark:text-white focus:outline-none focus:ring-2 focus:ring-[#8B5A2B]/30 focus:border-[#8B5A2B]"
          />
        </div>
        <Button
          label="Segarkan"
          icon="pi pi-refresh"
          outlined
          size="small"
          class="!rounded-xl !text-xs !border-[#8B5A2B] !text-[#8B5A2B] whitespace-nowrap"
          :loading="loading"
          @click="emit('refresh')"
        />
      </div>
    </div>

    <!-- Quick Filter Tabs Berdasarkan Status -->
    <div class="flex items-center gap-1.5 overflow-x-auto py-3 no-scrollbar border-b border-gray-100 dark:border-gray-800">
      <button
        v-for="f in [
          { key: 'all', label: 'Semua' },
          { key: 'waiting_verification', label: '⏳ Perlu Verifikasi' },
          { key: 'verified', label: '✓ Siap Check-In' },
          { key: 'checked_in', label: '🏨 Sedang Menginap' },
          { key: 'checked_out', label: 'Selesai' },
        ]"
        :key="f.key"
        type="button"
        @click="selectedStatusFilter = f.key"
        class="px-3 py-1.5 rounded-full text-xs font-semibold whitespace-nowrap transition cursor-pointer"
        :class="selectedStatusFilter === f.key
          ? 'bg-[#1A3A4A] text-white shadow-sm'
          : 'bg-[#F9FAFB] dark:bg-[#1A3A4A]/50 text-[#6B7280] dark:text-gray-300 hover:bg-gray-100'"
      >
        {{ f.label }}
      </button>
    </div>

    <!-- DataTable Reservasi -->
    <div class="mt-4">
      <DataTable
        :value="filteredBookings"
        paginator
        :rows="8"
        stripedRows
        responsiveLayout="scroll"
        :loading="loading"
        dataKey="id"
        class="text-xs"
      >
        <!-- Kolom ID -->
        <Column field="id" header="ID" sortable style="width: 75px">
          <template #body="{ data }">
            <span class="font-bold text-[#1A3A4A] dark:text-white">#{{ data.id }}</span>
          </template>
        </Column>

        <!-- Kolom Tamu & Kamar -->
        <Column header="Tamu & Kamar" style="min-width: 170px">
          <template #body="{ data }">
            <div>
              <p class="font-semibold text-[#1A3A4A] dark:text-gray-100">
                {{ data.guest_name || data.user_name || `Tamu #${data.user_id}` }}
              </p>
              <p class="text-[11px] text-[#6B7280] dark:text-gray-400">
                Tipe Kamar #{{ data.room_type_id }} · {{ data.guests || 2 }} Tamu
              </p>
            </div>
          </template>
        </Column>

        <!-- Kolom Periode Menginap -->
        <Column header="Jadwal Menginap" style="min-width: 170px">
          <template #body="{ data }">
            <div>
              <p class="text-[#1A3A4A] dark:text-gray-200">
                <span class="font-medium">Masuk:</span> {{ fmtDate(data.check_in) }}
              </p>
              <p class="text-[#6B7280] dark:text-gray-400">
                <span class="font-medium">Keluar:</span> {{ fmtDate(data.check_out) }}
              </p>
            </div>
          </template>
        </Column>

        <!-- Kolom Status Reservasi -->
        <Column field="status" header="Status" sortable style="min-width: 150px">
          <template #body="{ data }">
            <Tag
              :value="getStatusLabel(data.status)"
              :severity="getStatusSeverity(data.status)"
              rounded
              class="!text-[11px] font-semibold"
            />
          </template>
        </Column>

        <!-- Kolom Total Biaya -->
        <Column field="total_price" header="Total Biaya" sortable style="min-width: 120px">
          <template #body="{ data }">
            <span class="font-bold text-[#8B5A2B] dark:text-[#C9A86A]">
              {{ fmt(data.total_price) }}
            </span>
          </template>
        </Column>

        <!-- Kolom Bukti Transfer (Thumbnail Jelas) -->
        <Column header="Bukti Pembayaran" style="min-width: 120px">
          <template #body="{ data }">
            <!-- Jika Ada Bukti Transfer -->
            <div v-if="data.proof_url" class="flex items-center gap-1.5">
              <!-- Jika File PDF -->
              <a
                v-if="data.proof_url.endsWith('.pdf')"
                :href="data.proof_url"
                target="_blank"
                class="inline-flex items-center gap-1 text-[11px] font-semibold text-red-600 bg-red-50 dark:bg-red-950/40 px-2 py-1 rounded-lg border border-red-200"
              >
                <FileText class="w-3.5 h-3.5" /> PDF
              </a>

              <!-- Jika Gambar (Bisa di-zoom) -->
              <button
                v-else
                type="button"
                @click="openDetail(data)"
                class="group relative w-12 h-10 rounded-lg overflow-hidden border border-[#E5E7EB] dark:border-gray-600 shadow-xs cursor-pointer hover:border-[#8B5A2B]"
                title="Klik untuk memperbesar bukti transfer"
              >
                <img :src="data.proof_url" alt="Bukti Transfer" class="w-full h-full object-cover group-hover:scale-110 transition" />
                <div class="absolute inset-0 bg-black/20 group-hover:bg-transparent transition"></div>
              </button>
            </div>

            <!-- Jika Belum Ada Bukti -->
            <span v-else class="text-[#9CA3AF] text-xs italic">Belum diunggah</span>
          </template>
        </Column>

        <!-- Kolom Tindakan Cepat Resepsionis -->
        <Column header="Tindakan Resepsionis" style="min-width: 290px">
          <template #body="{ data }">
            <div class="flex items-center flex-wrap gap-1.5">
              <!-- Aksi 1: Verifikasi Pembayaran (Jika waiting_verification) -->
              <template v-if="data.status === 'waiting_verification'">
                <Button
                  label="Setujui"
                  icon="pi pi-check"
                  size="small"
                  class="!py-1 !px-2.5 !text-xs !bg-emerald-600 !border-emerald-600 hover:!bg-emerald-700 text-white !rounded-lg font-semibold shadow-xs"
                  :loading="actionLoading === `verify-${data.id}`"
                  @click="emit('verify', data.id, 'verified')"
                />
                <Button
                  label="Tolak"
                  icon="pi pi-times"
                  size="small"
                  severity="danger"
                  outlined
                  class="!py-1 !px-2.5 !text-xs !rounded-lg font-semibold"
                  :loading="actionLoading === `verify-${data.id}`"
                  @click="openRejectDialog(data.id)"
                />
              </template>

              <!-- Aksi 2: Check-In (Jika verified) -->
              <Button
                v-if="data.status === 'verified'"
                label="Check-In Tamu"
                icon="pi pi-sign-in"
                size="small"
                class="!py-1 !px-2.5 !text-xs !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] text-white !rounded-lg font-semibold shadow-xs"
                :loading="actionLoading === `checkin-${data.id}`"
                @click="emit('check-in', data.id)"
              />

              <!-- Aksi 3: Check-Out (Jika checked_in) -->
              <Button
                v-if="data.status === 'checked_in'"
                label="Check-Out"
                icon="pi pi-sign-out"
                size="small"
                class="!py-1 !px-2.5 !text-xs !bg-[#1A3A4A] !border-[#1A3A4A] hover:!bg-[#2D3748] text-white !rounded-lg font-semibold shadow-xs"
                :loading="actionLoading === `checkout-${data.id}`"
                @click="emit('check-out', data.id)"
              />

              <!-- Aksi 4: Unduh Invoice -->
              <Button
                v-if="canInvoice(data.status)"
                label="Invoice"
                size="small"
                outlined
                class="!py-1 !px-2.5 !text-xs !rounded-lg !border-gray-300 dark:!border-gray-600 !text-gray-700 dark:!text-gray-200"
                :loading="invoicingId === data.id"
                @click="emit('download-invoice', data)"
              >
                <template #icon>
                  <Download class="w-3 h-3 mr-1" />
                </template>
              </Button>

              <!-- Aksi 5: Detail -->
              <Button
                icon="pi pi-eye"
                size="small"
                text
                class="!w-7 !h-7 !p-0 !text-gray-500 hover:!text-[#8B5A2B]"
                title="Lihat detail lengkap"
                @click="openDetail(data)"
              />
            </div>
          </template>
        </Column>
      </DataTable>
    </div>

    <!-- Modal Pratinjau Bukti Transfer & Rincian Reservasi -->
    <Dialog
      v-model:visible="showDetailModal"
      modal
      header="Rincian Reservasi & Bukti Pembayaran"
      :style="{ width: '560px' }"
      class="!rounded-2xl"
    >
      <div v-if="selectedBooking" class="space-y-4 text-xs md:text-sm">
        <div class="p-3.5 bg-[#FDF6EC] dark:bg-[#1A3A4A] rounded-xl border border-[#E5E7EB] dark:border-gray-700 space-y-1.5">
          <div class="flex justify-between">
            <span class="text-gray-500">Nomor Reservasi:</span>
            <span class="font-bold text-[#1A3A4A] dark:text-white">#{{ selectedBooking.id }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">Status Saat Ini:</span>
            <Tag :value="getStatusLabel(selectedBooking.status)" :severity="getStatusSeverity(selectedBooking.status)" rounded class="text-xs" />
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500">Total Pembayaran:</span>
            <span class="font-bold text-[#8B5A2B] dark:text-[#C9A86A]">{{ fmt(selectedBooking.total_price) }}</span>
          </div>
        </div>

        <!-- Area Gambar Bukti Transfer -->
        <div>
          <label class="block text-xs font-bold text-gray-500 uppercase tracking-wider mb-2">
            Bukti Transfer Tamu
          </label>
          <div v-if="selectedBooking.proof_url" class="rounded-xl overflow-hidden border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-center p-2">
            <img
              v-if="!selectedBooking.proof_url.endsWith('.pdf')"
              :src="selectedBooking.proof_url"
              alt="Bukti Transfer Penuh"
              class="max-h-80 mx-auto rounded-lg object-contain"
            />
            <div v-else class="py-8">
              <FileText class="w-12 h-12 text-red-500 mx-auto mb-2" />
              <p class="text-xs text-gray-600 dark:text-gray-300">File bukti berformat dokumen PDF.</p>
              <a :href="selectedBooking.proof_url" target="_blank" class="mt-2 inline-flex items-center gap-1 text-xs font-semibold text-[#8B5A2B] underline">
                Buka Dokumen PDF <ExternalLink class="w-3.5 h-3.5" />
              </a>
            </div>
          </div>
          <p v-else class="text-xs text-gray-400 italic">Tamu belum mengunggah bukti transfer.</p>
        </div>
      </div>
      <template #footer>
        <Button label="Tutup" class="!bg-[#8B5A2B] !border-[#8B5A2B] text-white !rounded-xl !text-xs" @click="showDetailModal = false" />
      </template>
    </Dialog>

    <!-- Modal Alasan Penolakan Reservasi -->
    <Dialog
      v-model:visible="showRejectModal"
      modal
      header="Tolak Bukti Pembayaran"
      :style="{ width: '420px' }"
      class="!rounded-2xl"
    >
      <div class="space-y-3">
        <p class="text-xs text-[#6B7280]">
          Alasan penolakan akan disampaikan secara transparan kepada tamu melalui notifikasi pemesanan.
        </p>
        <div>
          <label class="block text-xs font-bold text-gray-700 dark:text-gray-300 mb-1">Alasan Penolakan:</label>
          <InputText
            v-model="rejectReason"
            placeholder="Contoh: Bukti buram, nominal transfer tidak sesuai"
            class="w-full !rounded-xl !text-xs md:!text-sm"
          />
        </div>
      </div>
      <template #footer>
        <Button label="Batal" text class="!text-xs" @click="showRejectModal = false" />
        <Button
          label="Konfirmasi Tolak"
          severity="danger"
          class="!rounded-xl !text-xs"
          :loading="actionLoading === `verify-${rejectTargetId}`"
          @click="submitReject"
        />
      </template>
    </Dialog>
  </div>
</template>
