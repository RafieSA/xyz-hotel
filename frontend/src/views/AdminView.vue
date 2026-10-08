<script setup>
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Message from 'primevue/message'
import { LayoutDashboard, TrendingUp, CalendarDays, Eye } from 'lucide-vue-next'
import { ref, onMounted, computed } from 'vue'
import { useToast } from 'primevue/usetoast'
import client from '../api/client'

const toast = useToast()
const bookings = ref([])
const roomUnits = ref([])
const loading = ref(false)
const actionLoading = ref('')
const selected = ref(null)
const showDetail = ref(false)
const showReject = ref(false)
const rejectReason = ref('')
const rejectId = ref(null)
const unitStatusMap = ref({}) // unitId -> selected status

const statusSeverity = (s) => ({ verified: 'success', checked_in: 'info', pending_payment: 'warn', waiting_verification: 'warn', checked_out: 'secondary', cancelled: 'danger', rejected: 'danger', expired: 'danger' }[s] || 'secondary')
const statusLabel = (s) => s?.replace('_', ' ') || s
const fmt = (n) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n)
const fmtDate = (d) => { try { return new Date(d).toISOString().slice(0,10) } catch { return d } }

async function fetchBookings() {
  loading.value = true
  try {
    const { data } = await client.get('/api/admin/bookings')
    const list = data.data || data
    bookings.value = Array.isArray(list) ? list : []
  } catch (e) {
    // fallback to /api/bookings if admin route not allowed
    try {
      const { data } = await client.get('/api/bookings')
      const list = data.data || data
      bookings.value = Array.isArray(list) ? list : []
    } catch (e2) {
      toast.add({ severity: 'error', summary: 'Gagal memuat bookings', detail: e?.response?.data?.message || e.message, life: 3000 })
    }
  } finally { loading.value = false }
}

async function fetchUnits() {
  try {
    const { data } = await client.get('/api/admin/room-units')
    const list = data.data || data
    roomUnits.value = Array.isArray(list) ? list : []
  } catch {}
}

onMounted(()=>{ fetchBookings(); fetchUnits() })

async function doVerify(id, action) {
  if (action==='rejected') { rejectId.value=id; showReject.value=true; return }
  actionLoading.value = `verify-${id}`
  try {
    const { data } = await client.patch(`/api/admin/bookings/${id}/verify`, { action })
    toast.add({ severity: 'success', summary: 'Verified', detail: `Booking #${id} verified`, life: 2500 })
    await fetchBookings()
  } catch (e) { toast.add({ severity: 'error', summary: 'Verify gagal', detail: e?.response?.data?.message || e.message, life: 3500 }) }
  finally { actionLoading.value='' }
}
async function confirmReject() {
  if (!rejectReason.value.trim()) { toast.add({ severity:'warn', summary:'Isi alasan reject', life:2000}); return }
  const id = rejectId.value
  actionLoading.value=`verify-${id}`
  try {
    await client.patch(`/api/admin/bookings/${id}/verify`, { action:'rejected', reject_reason: rejectReason.value })
    toast.add({ severity:'success', summary:'Rejected', detail:`Booking #${id} rejected`, life:2500 })
    showReject.value=false; rejectReason.value=''; await fetchBookings()
  } catch(e){ toast.add({ severity:'error', summary:'Reject gagal', detail:e?.response?.data?.message||e.message, life:3500}) }
  finally{ actionLoading.value='' }
}
async function doCheckIn(id){
  actionLoading.value=`checkin-${id}`
  try{ await client.patch(`/api/admin/bookings/${id}/checkin`); toast.add({ severity:'success', summary:'Check-In berhasil', detail:`Booking #${id} checked_in (unit assigned)`, life:2500 }); await fetchBookings(); await fetchUnits() } catch(e){ toast.add({ severity:'error', summary:'Check-In gagal', detail:e?.response?.data?.message||e.message, life:3500}) } finally{ actionLoading.value='' }
}
async function doCheckOut(id){
  actionLoading.value=`checkout-${id}`
  try{ await client.patch(`/api/admin/bookings/${id}/checkout`); toast.add({ severity:'success', summary:'Check-Out berhasil', detail:`Booking #${id} checked_out (unit dirty)`, life:2500 }); await fetchBookings(); await fetchUnits() } catch(e){ toast.add({ severity:'error', summary:'Check-Out gagal', detail:e?.response?.data?.message||e.message, life:3500}) } finally{ actionLoading.value='' }
}
function openDetail(row){ selected.value=row; showDetail.value=true }

const unitStatusOptions = [
  { label:'available', value:'available' },
  { label:'occupied', value:'occupied' },
  { label:'dirty', value:'dirty' },
  { label:'maintenance', value:'maintenance' },
]
async function updateUnitStatus(unit){
  const newStatus = unitStatusMap.value[unit.id]
  if(!newStatus) return
  try{ await client.patch(`/api/admin/room-units/${unit.id}/status`, { status:newStatus }); toast.add({ severity:'success', summary:'Unit updated', detail:`${unit.code} → ${newStatus}`, life:2500 }); await fetchUnits() } catch(e){ toast.add({ severity:'error', summary:'Update gagal', detail:e?.response?.data?.message||e.message, life:3500}) }
}

const unitSeverity = (s)=>({ available:'success', occupied:'info', dirty:'warn', maintenance:'danger' }[s]||'secondary')
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 space-y-6">
    <div class="flex items-center gap-3">
      <span class="bg-[#8B5A2B] text-white rounded-xl p-2.5"><LayoutDashboard class="w-5 h-5" /></span>
      <div>
        <h1 class="font-display font-bold text-2xl text-[#1A3A4A]">Backoffice</h1>
        <p class="text-sm text-[#6B7280]">Kelola booking, kamar, dan laporan.</p>
      </div>
      <Button label="Refresh" icon="pi pi-refresh" outlined class="!rounded-xl !ml-auto !border-[#8B5A2B] !text-[#8B5A2B]" :loading="loading" @click="fetchBookings(); fetchUnits()" />
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
        <template #content>
          <p class="text-xs uppercase tracking-wide text-[#6B7280] font-semibold flex items-center gap-1.5"><CalendarDays class="w-4 h-4" /> Booking hari ini</p>
          <p class="text-2xl font-bold text-[#1A3A4A] mt-1">{{ bookings.length }}</p>
          <p class="text-xs text-[#2E7D32] mt-1 flex items-center gap-1"><TrendingUp class="w-3.5 h-3.5" /> live</p>
        </template>
      </Card>
      <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
        <template #content>
          <p class="text-xs uppercase tracking-wide text-[#6B7280] font-semibold">Occupancy</p>
          <p class="text-2xl font-bold text-[#1A3A4A] mt-1">{{ roomUnits.filter(u=>u.status==='occupied').length }} / {{ roomUnits.length }}</p>
          <div class="mt-2 h-2 bg-[#F3F4F6] rounded-full overflow-hidden"><div class="h-full bg-[#8B5A2B] rounded-full" :style="{ width: (roomUnits.length? (roomUnits.filter(u=>u.status==='occupied').length/roomUnits.length*100):0)+'%' }"></div></div>
        </template>
      </Card>
      <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
        <template #content>
          <p class="text-xs uppercase tracking-wide text-[#6B7280] font-semibold">Revenue (MTD)</p>
          <p class="text-2xl font-bold text-[#8B5A2B] mt-1">Rp 42.500.000</p>
          <p class="text-xs text-[#6B7280] mt-1">Chart placeholder — Chart.js nanti</p>
        </template>
      </Card>
    </div>

    <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
      <template #title><div class="flex items-center justify-between"><span class="text-[#1A3A4A] font-semibold text-base">Booking terbaru</span><span class="text-xs text-[#6B7280]">{{ bookings.length }} rows</span></div></template>
      <template #content>
        <DataTable :value="bookings" paginator :rows="8" stripedRows class="text-sm" responsiveLayout="scroll" :loading="loading" dataKey="id">
          <Column field="id" header="ID" sortable style="width:80px" />
          <Column field="user_id" header="User" sortable style="width:90px" />
          <Column field="room_type_id" header="Tipe" style="width:80px">
            <template #body="{ data }">#{{ data.room_type_id }}</template>
          </Column>
          <Column field="check_in" header="Check-in" sortable>
            <template #body="{ data }">{{ fmtDate(data.check_in) }}</template>
          </Column>
          <Column field="check_out" header="Check-out">
            <template #body="{ data }">{{ fmtDate(data.check_out) }}</template>
          </Column>
          <Column field="status" header="Status">
            <template #body="{ data }"><Tag :value="statusLabel(data.status)" :severity="statusSeverity(data.status)" rounded class="capitalize" /></template>
          </Column>
          <Column field="total_price" header="Total">
            <template #body="{ data }"><span class="font-semibold text-[#8B5A2B]">{{ fmt(data.total_price) }}</span></template>
          </Column>
          <Column header="Proof">
            <template #body="{ data }">
              <span v-if="data.proof_url" class="text-xs text-[#8B5A2B] underline cursor-pointer" @click="openDetail(data)">lihat</span>
              <span v-else class="text-xs text-[#9CA3AF]">-</span>
            </template>
          </Column>
          <Column header="Aksi" style="min-width:280px">
            <template #body="{ data }">
              <div class="flex flex-wrap gap-1.5">
                <Button v-if="data.status==='waiting_verification'" label="Verify" size="small" class="!py-1 !px-2.5 !text-xs !bg-green-600 !border-green-600 hover:!bg-green-700 !rounded-full" :loading="actionLoading===`verify-${data.id}`" @click="doVerify(data.id,'verified')" />
                <Button v-if="data.status==='waiting_verification'" label="Reject" size="small" severity="danger" outlined class="!py-1 !px-2.5 !text-xs !rounded-full" :loading="actionLoading===`verify-${data.id}`" @click="doVerify(data.id,'rejected')" />
                <Button v-if="data.status==='verified'" label="Check-In" size="small" class="!py-1 !px-2.5 !text-xs !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-full" :loading="actionLoading===`checkin-${data.id}`" @click="doCheckIn(data.id)" />
                <Button v-if="data.status==='checked_in'" label="Check-Out" size="small" severity="info" class="!py-1 !px-2.5 !text-xs !rounded-full" :loading="actionLoading===`checkout-${data.id}`" @click="doCheckOut(data.id)" />
                <Button icon="pi pi-eye" size="small" text rounded class="!text-[#6B7280]" @click="openDetail(data)" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <!-- Room units -->
    <Card v-if="roomUnits.length" class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
      <template #title><span class="text-[#1A3A4A] font-semibold text-base">Room Units</span></template>
      <template #content>
        <DataTable :value="roomUnits" paginator :rows="8" stripedRows class="text-sm" responsiveLayout="scroll">
          <Column field="id" header="ID" sortable style="width:70px" />
          <Column field="code" header="Kode" sortable />
          <Column field="room_type_id" header="Type" sortable style="width:80px" />
          <Column field="status" header="Status">
            <template #body="{ data }"><Tag :value="data.status" :severity="unitSeverity(data.status)" rounded class="capitalize" /></template>
          </Column>
          <Column header="Ubah Status" style="min-width:260px">
            <template #body="{ data }">
              <div class="flex gap-2">
                <Select v-model="unitStatusMap[data.id]" :options="unitStatusOptions" optionLabel="label" optionValue="value" placeholder="Pilih status" class="w-36 !text-xs" size="small" />
                <Button label="Set" size="small" class="!py-1 !px-3 !text-xs !bg-[#8B5A2B] !border-[#8B5A2B]" :disabled="!unitStatusMap[data.id]" @click="updateUnitStatus(data)" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB] bg-[#FDF6EC]/60">
      <template #content>
        <p class="text-sm font-semibold text-[#1A3A4A] flex items-center gap-2"><TrendingUp class="w-4 h-4 text-[#8B5A2B]" /> Chart placeholder</p>
        <div class="mt-3 h-40 bg-white rounded-xl border border-dashed border-[#C9A86A]/40 flex items-center justify-center text-sm text-[#6B7280]">Occupancy & revenue chart (Chart.js) — Fase 2</div>
      </template>
    </Card>

    <!-- Detail dialog with proof preview -->
    <Dialog v-model:visible="showDetail" modal header="Detail Booking" :style="{ width:'560px' }" class="!rounded-xl">
      <div v-if="selected" class="space-y-3 text-sm">
        <p><span class="font-semibold">ID:</span> {{ selected.id }} · <Tag :value="statusLabel(selected.status)" :severity="statusSeverity(selected.status)" rounded /></p>
        <p><span class="font-semibold">User:</span> {{ selected.user_id }} · Room Type #{{ selected.room_type_id }} <span v-if="selected.room_unit_id">→ Unit #{{ selected.room_unit_id }}</span></p>
        <p><span class="font-semibold">Dates:</span> {{ fmtDate(selected.check_in) }} → {{ fmtDate(selected.check_out) }} · {{ selected.guests }} tamu</p>
        <p><span class="font-semibold">Total:</span> <span class="text-[#8B5A2B] font-bold">{{ fmt(selected.total_price) }}</span></p>
        <p v-if="selected.reject_reason"><span class="font-semibold">Reject reason:</span> {{ selected.reject_reason }}</p>
        <div v-if="selected.proof_url" class="space-y-2">
          <p class="font-semibold">Bukti pembayaran:</p>
          <p class="text-xs text-[#6B7280] break-all">{{ selected.proof_url }}</p>
          <img v-if="!selected.proof_url.endsWith('.pdf')" :src="`http://localhost:8080/storage/uploads/${selected.proof_url}`" alt="proof" class="w-full rounded-xl border max-h-80 object-contain bg-gray-50" @error="(e)=>e.target.style.display='none'" />
          <div v-else class="border rounded-xl p-6 text-center text-sm text-[#6B7280]">PDF preview: <a :href="`http://localhost:8080/storage/uploads/${selected.proof_url}`" target="_blank" class="text-[#8B5A2B] underline">Buka PDF</a></div>
        </div>
        <Message v-else severity="info" class="text-xs">Belum ada bukti pembayaran.</Message>
      </div>
      <template #footer><Button label="Tutup" class="!rounded-xl !bg-[#8B5A2B] !border-[#8B5A2B]" @click="showDetail=false" /></template>
    </Dialog>

    <Dialog v-model:visible="showReject" modal header="Tolak Booking" :style="{ width:'420px' }">
      <div class="space-y-3">
        <p class="text-sm text-[#6B7280]">Berikan alasan penolakan untuk booking #{{ rejectId }}</p>
        <InputText v-model="rejectReason" placeholder="Alasan ditolak..." class="w-full" />
      </div>
      <template #footer>
        <Button label="Batal" text @click="showReject=false" />
        <Button label="Tolak" severity="danger" :loading="actionLoading===`verify-${rejectId}`" @click="confirmReject" />
      </template>
    </Dialog>
  </div>
</template>
