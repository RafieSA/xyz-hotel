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
import DatePicker from 'primevue/datepicker'
import Chart from 'primevue/chart'
import Rating from 'primevue/rating'
import Textarea from 'primevue/textarea'
import { LayoutDashboard, TrendingUp, CalendarDays, Wallet, Bed } from 'lucide-vue-next'
import { ref, onMounted, computed, watch } from 'vue'
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
const unitStatusMap = ref({})
const reports = ref(null)
const reportsLoading = ref(false)
const dateRange = ref(null) // [Date, Date]
const reviews = ref([])

const statusSeverity = (s) => ({ verified: 'success', checked_in: 'info', pending_payment: 'warn', waiting_verification: 'warn', checked_out: 'secondary', cancelled: 'danger', rejected: 'danger', expired: 'danger' }[s] || 'secondary')
const statusLabel = (s) => s?.replace('_', ' ') || s
const fmt = (n) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n||0)
const fmtDate = (d) => { try { return new Date(d).toISOString().slice(0,10) } catch { return d } }
function toISO(d){ if(!d) return ''; const dt=d instanceof Date?d:new Date(d); if(isNaN(dt)) return ''; return dt.toISOString().slice(0,10) }

async function fetchBookings() {
  loading.value = true
  try {
    const { data } = await client.get('/api/admin/bookings')
    const list = data.data || data
    bookings.value = Array.isArray(list) ? list : []
  } catch (e) {
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
async function fetchReports(){
  reportsLoading.value=true
  try{
    const params={}
    if(dateRange.value && dateRange.value[0]) params.from = toISO(dateRange.value[0])
    if(dateRange.value && dateRange.value[1]) params.to = toISO(dateRange.value[1])
    const { data } = await client.get('/api/admin/reports/summary', { params })
    reports.value = data.data || data
  }catch(e){
    // keep previous or fallback to computed from bookings
    reports.value = null
  }finally{ reportsLoading.value=false }
}
async function fetchReviews(){
  try{
    const { data } = await client.get('/api/reviews')
    const list = data.data || data
    reviews.value = Array.isArray(list)? list : []
  }catch{ reviews.value=[] }
}

onMounted(()=>{ fetchBookings(); fetchUnits(); fetchReports(); fetchReviews() })
watch(dateRange, fetchReports)

async function doVerify(id, action) {
  if (action==='rejected') { rejectId.value=id; showReject.value=true; return }
  actionLoading.value = `verify-${id}`
  try {
    await client.patch(`/api/admin/bookings/${id}/verify`, { action })
    toast.add({ severity: 'success', summary: 'Verified', detail: `Booking #${id} verified`, life: 2500 })
    await fetchBookings(); await fetchReports()
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
    showReject.value=false; rejectReason.value=''; await fetchBookings(); await fetchReports()
  } catch(e){ toast.add({ severity:'error', summary:'Reject gagal', detail:e?.response?.data?.message||e.message, life:3500}) }
  finally{ actionLoading.value='' }
}
async function doCheckIn(id){
  actionLoading.value=`checkin-${id}`
  try{ await client.patch(`/api/admin/bookings/${id}/checkin`); toast.add({ severity:'success', summary:'Check-In berhasil', detail:`Booking #${id} checked_in`, life:2500 }); await fetchBookings(); await fetchUnits(); await fetchReports() } catch(e){ toast.add({ severity:'error', summary:'Check-In gagal', detail:e?.response?.data?.message||e.message, life:3500}) } finally{ actionLoading.value='' }
}
async function doCheckOut(id){
  actionLoading.value=`checkout-${id}`
  try{ await client.patch(`/api/admin/bookings/${id}/checkout`); toast.add({ severity:'success', summary:'Check-Out berhasil', detail:`Booking #${id} checked_out`, life:2500 }); await fetchBookings(); await fetchUnits(); await fetchReports() } catch(e){ toast.add({ severity:'error', summary:'Check-Out gagal', detail:e?.response?.data?.message||e.message, life:3500}) } finally{ actionLoading.value='' }
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

// derived stats with fallback
const totalBookings = computed(()=> reports.value?.total_bookings ?? reports.value?.bookings_count ?? bookings.value.length)
const totalRevenue = computed(()=> reports.value?.total_revenue ?? reports.value?.revenue ?? bookings.value.filter(b=>['verified','checked_in','checked_out'].includes(b.status)).reduce((s,b)=>s+Number(b.total_price||0),0))
const occupancyRate = computed(()=>{
  if(reports.value?.occupancy_rate != null) return reports.value.occupancy_rate
  if(reports.value?.occupancyRate != null) return reports.value.occupancyRate
  if(roomUnits.value.length) return Math.round(roomUnits.value.filter(u=>u.status==='occupied').length/roomUnits.value.length*100)
  return 0
})
const availableUnits = computed(()=>{
  if(reports.value?.available_units != null) return reports.value.available_units
  return roomUnits.value.filter(u=>u.status==='available').length || (roomUnits.value.length - roomUnits.value.filter(u=>u.status==='occupied').length)
})
const bookingsByStatus = computed(()=> reports.value?.bookings_by_status || reports.value?.by_status || null)

// Chart data
const WA = { brown:'#8B5A2B', gold:'#C9A86A', navy:'#1A3A4A', cream:'#FDF6EC' }

const revenueChartData = computed(()=>{
  const perDay = reports.value?.revenue_per_day || reports.value?.revenuePerDay || null
  if(perDay && Array.isArray(perDay) && perDay.length){
    return {
      labels: perDay.map(r=> r.date || r.day),
      datasets:[{ label:'Revenue', data: perDay.map(r=> Number(r.revenue||r.total||0)), borderColor:WA.brown, backgroundColor:'rgba(139,90,43,0.15)', tension:0.35, fill:true, pointBackgroundColor:WA.brown }]
    }
  }
  // fallback: last 7 days from bookings
  const map={}
  bookings.value.forEach(b=>{
    const d=fmtDate(b.created_at || b.check_in)
    map[d]=(map[d]||0)+Number(b.total_price||0)
  })
  const labels=Object.keys(map).sort().slice(-7)
  return { labels: labels.length? labels: ['No data'], datasets:[{ label:'Revenue', data: labels.length? labels.map(l=>map[l]):[0], borderColor:WA.brown, backgroundColor:'rgba(139,90,43,0.15)', tension:0.35, fill:true, pointBackgroundColor:WA.brown }] }
})
const revenueChartOptions = { responsive:true, maintainAspectRatio:false, plugins:{ legend:{ display:false } }, scales:{ y:{ beginAtZero:true, ticks:{ callback:(v)=> 'Rp '+(v/1000)+'k' } } } }

const barChartData = computed(()=>{
  const byType = reports.value?.bookings_by_room_type || reports.value?.by_room_type
  if(byType && Array.isArray(byType) && byType.length){
    return { labels: byType.map(r=> r.room_type || r.name || r.type), datasets:[{ label:'Bookings', data: byType.map(r=> r.count||r.total||0), backgroundColor:[WA.brown, WA.gold, WA.navy, '#D4A76A'] }] }
  }
  const counts={}
  bookings.value.forEach(b=>{ const k='Type '+(b.room_type_id||'?'); counts[k]=(counts[k]||0)+1 })
  const labels=Object.keys(counts)
  return { labels: labels.length? labels:['No data'], datasets:[{ label:'Bookings', data: labels.length? labels.map(l=>counts[l]):[0], backgroundColor:[WA.brown, WA.gold, WA.navy, '#D4A76A'] }] }
})
const barChartOptions = { responsive:true, maintainAspectRatio:false, plugins:{ legend:{ display:false } }, scales:{ y:{ beginAtZero:true, ticks:{ stepSize:1 } } } }

const doughnutData = computed(()=>{
  const total=roomUnits.value.length || 10
  const occupied=roomUnits.value.filter(u=>u.status==='occupied').length || Math.round(occupancyRate.value/100*total)
  const available=Math.max(0, total-occupied)
  return { labels:['Occupied','Available'], datasets:[{ data:[occupied, available], backgroundColor:[WA.brown, WA.gold], borderWidth:0 }] }
})
const doughnutOptions = { responsive:true, maintainAspectRatio:false, cutout:'65%', plugins:{ legend:{ position:'bottom', labels:{ usePointStyle:true, boxWidth:10 } } } }
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center gap-3">
      <div class="flex items-center gap-3">
        <span class="bg-[#8B5A2B] text-white rounded-xl p-2.5"><LayoutDashboard class="w-5 h-5" /></span>
        <div>
          <h1 class="font-display font-bold text-2xl text-[#1A3A4A]">Backoffice</h1>
          <p class="text-sm text-[#6B7280]">Kelola booking, kamar, dan laporan.</p>
        </div>
      </div>
      <div class="sm:ml-auto flex items-center gap-2 flex-wrap">
        <DatePicker v-model="dateRange" selectionMode="range" :manualInput="false" placeholder="Filter tanggal" showIcon class="min-w-[220px]" />
        <Button label="Refresh" icon="pi pi-refresh" outlined class="!rounded-xl !border-[#8B5A2B] !text-[#8B5A2B]" :loading="loading || reportsLoading" @click="fetchBookings(); fetchUnits(); fetchReports()" />
      </div>
    </div>

    <!-- 4 stats cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] hover:!shadow-md transition-shadow">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs uppercase tracking-widest text-[#6B7280] font-semibold">Total Bookings</p>
              <p class="text-3xl font-bold text-[#1A3A4A] mt-1">{{ totalBookings }}</p>
              <p class="text-xs text-[#6B7280] mt-1">Semua status</p>
            </div>
            <span class="bg-[#FDF6EC] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-xl p-2.5"><CalendarDays class="w-5 h-5" /></span>
          </div>
        </template>
      </Card>
      <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] hover:!shadow-md transition-shadow">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs uppercase tracking-widest text-[#6B7280] font-semibold">Revenue</p>
              <p class="text-2xl font-bold text-[#8B5A2B] mt-1">{{ fmt(totalRevenue) }}</p>
              <p class="text-xs text-[#6B7280] mt-1">Verified + checked</p>
            </div>
            <span class="bg-[#FDF6EC] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-xl p-2.5"><Wallet class="w-5 h-5" /></span>
          </div>
        </template>
      </Card>
      <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] hover:!shadow-md transition-shadow">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs uppercase tracking-widest text-[#6B7280] font-semibold">Occupancy Rate</p>
              <p class="text-3xl font-bold text-[#1A3A4A] mt-1">{{ occupancyRate }}%</p>
              <div class="mt-2 h-1.5 w-24 bg-[#F3F4F6] rounded-full overflow-hidden"><div class="h-full bg-[#8B5A2B] rounded-full" :style="{ width: occupancyRate+'%' }"></div></div>
            </div>
            <span class="bg-[#FDF6EC] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-xl p-2.5"><TrendingUp class="w-5 h-5" /></span>
          </div>
        </template>
      </Card>
      <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] hover:!shadow-md transition-shadow">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs uppercase tracking-widest text-[#6B7280] font-semibold">Available Units</p>
              <p class="text-3xl font-bold text-[#1A3A4A] mt-1">{{ availableUnits }} <span class="text-sm font-normal text-[#6B7280]">/ {{ roomUnits.length || '-' }}</span></p>
              <p class="text-xs text-[#2E7D32] mt-1">Ready to book</p>
            </div>
            <span class="bg-[#FDF6EC] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-xl p-2.5"><Bed class="w-5 h-5" /></span>
          </div>
        </template>
      </Card>
    </div>

    <!-- Charts -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB] lg:col-span-2">
        <template #title><span class="text-[#1A3A4A] font-semibold text-sm">Revenue per Hari</span></template>
        <template #content>
          <div class="h-[260px]"><Chart type="line" :data="revenueChartData" :options="revenueChartOptions" /></div>
        </template>
      </Card>
      <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
        <template #title><span class="text-[#1A3A4A] font-semibold text-sm">Occupancy vs Available</span></template>
        <template #content>
          <div class="h-[260px] flex items-center justify-center"><Chart type="doughnut" :data="doughnutData" :options="doughnutOptions" /></div>
          <p class="text-center text-xs text-[#6B7280] mt-2">{{ occupancyRate }}% terisi · {{ 100-occupancyRate }}% tersedia</p>
        </template>
      </Card>
    </div>
    <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
      <template #title><span class="text-[#1A3A4A] font-semibold text-sm">Bookings by Room Type</span></template>
      <template #content>
        <div class="h-[260px]"><Chart type="bar" :data="barChartData" :options="barChartOptions" /></div>
      </template>
    </Card>

    <Card v-if="bookingsByStatus" class="!rounded-xl !shadow-sm !border !border-[#E5E7EB] bg-[#FDF6EC]/40">
      <template #content>
        <div class="flex flex-wrap gap-2 text-xs">
          <span v-for="(v,k) in bookingsByStatus" :key="k" class="bg-white border border-[#E5E7EB] rounded-full px-3 py-1"><span class="font-semibold text-[#1A3A4A] capitalize">{{ k.replace('_',' ') }}</span> <span class="text-[#8B5A2B] font-bold">{{ v }}</span></span>
        </div>
      </template>
    </Card>

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
          <Column field="code" header="Code" sortable />
          <Column field="room_type_id" header="Type" sortable><template #body="{ data }">#{{ data.room_type_id }}</template></Column>
          <Column field="status" header="Status"><template #body="{ data }"><Tag :value="data.status" :severity="unitSeverity(data.status)" rounded /></template></Column>
          <Column header="Ubah Status" style="min-width:280px">
            <template #body="{ data }">
              <div class="flex gap-2">
                <Select v-model="unitStatusMap[data.id]" :options="unitStatusOptions" optionLabel="label" optionValue="value" placeholder="Pilih" class="w-full !text-xs" />
                <Button label="Save" size="small" class="!bg-[#8B5A2B] !border-[#8B5A2B] !rounded-full" @click="updateUnitStatus(data)" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <!-- Reviews -->
    <Card v-if="reviews.length" class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
      <template #title><span class="text-[#1A3A4A] font-semibold text-base">Reviews ({{ reviews.length }})</span></template>
      <template #content>
        <DataTable :value="reviews" paginator :rows="5" stripedRows class="text-sm">
          <Column field="id" header="ID" style="width:70px" />
          <Column field="booking_id" header="Booking" style="width:90px" />
          <Column field="rating" header="Rating" style="width:140px"><template #body="{ data }"><Rating :modelValue="data.rating" readonly :stars="5" /></template></Column>
          <Column field="comment" header="Comment" />
          <Column field="created_at" header="Date"><template #body="{ data }">{{ fmtDate(data.created_at) }}</template></Column>
        </DataTable>
      </template>
    </Card>

    <Dialog v-model:visible="showDetail" modal header="Detail Booking" :style="{ width:'560px' }" class="!rounded-xl">
      <div v-if="selected" class="space-y-3 text-sm">
        <p><span class="font-semibold">ID:</span> {{ selected.id }} — {{ selected.status }}</p>
        <p><span class="font-semibold">Total:</span> {{ fmt(selected.total_price) }}</p>
        <p v-if="selected.proof_url" class="break-all"><span class="font-semibold">Proof:</span> <a :href="selected.proof_url" target="_blank" class="text-[#8B5A2B] underline">{{ selected.proof_url }}</a></p>
        <img v-if="selected.proof_url && !selected.proof_url.endsWith('.pdf')" :src="selected.proof_url" alt="proof" class="max-h-64 rounded-lg border" />
      </div>
      <template #footer><Button label="Tutup" class="!rounded-xl !bg-[#8B5A2B] !border-[#8B5A2B]" @click="showDetail=false" /></template>
    </Dialog>

    <Dialog v-model:visible="showReject" modal header="Tolak Booking" :style="{ width:'420px' }">
      <div class="space-y-3">
        <p class="text-sm text-[#6B7280]">Alasan penolakan akan dikirim ke tamu.</p>
        <InputText v-model="rejectReason" placeholder="Alasan reject" class="w-full" />
      </div>
      <template #footer>
        <Button label="Batal" text @click="showReject=false" />
        <Button label="Reject" severity="danger" :loading="actionLoading===`verify-${rejectId}`" @click="confirmReject" />
      </template>
    </Dialog>
  </div>
</template>
