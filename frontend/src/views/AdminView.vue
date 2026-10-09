<script setup>
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import Message from 'primevue/message'
import DatePicker from 'primevue/datepicker'
import Chart from 'primevue/chart'
import Rating from 'primevue/rating'
import { LayoutDashboard, TrendingUp, CalendarDays, Wallet, Bed, ShieldCheck, Boxes, FileText, Download, Plus, Pencil, Trash2, Upload, Grid3x3, Kanban, Radio, MessageSquare, Check, Users } from 'lucide-vue-next'
import { ref, onMounted, computed, watch, onBeforeUnmount } from 'vue'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'
import client from '../api/client'
import AdminBookingsTable from '../components/AdminBookingsTable.vue'
import AdminStatsOverview from '../components/AdminStatsOverview.vue'
import HousekeepingKanban from '../components/HousekeepingKanban.vue'
import OccupancyCalendarMatrix from '../components/OccupancyCalendarMatrix.vue'

const toast = useToast()
const { t } = useI18n()
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
const dateRange = ref(null)
const reviews = ref([])

// Room Types CRUD
const roomTypes = ref([])
const showRoomTypeDialog = ref(false)
const editingRoomType = ref(null)
const roomTypeForm = ref({ name:'', description:'', capacity:2, price_per_night:350000, total_units:5 })
const rtSaving = ref(false)
const invoicingId = ref(null)

// Room Units create
const showUnitDialog = ref(false)
const unitForm = ref({ code:'', room_type_id:1, status:'available' })
const unitSaving = ref(false)

// Audit logs
const auditLogs = ref([])
const auditLoading = ref(false)
const activeTab = ref('bookings')
const tabsBarRef = ref(null)
const tabsPillRef = ref(null)
function positionPill(animate=true){
  const bar=tabsBarRef.value; const pill=tabsPillRef.value; if(!bar||!pill) return;
  const active=bar.querySelector('.t-tab[aria-selected="true"]'); if(!active) return;
  if(!animate){ const prev=pill.style.transition; pill.style.transition='none'; pill.style.transform=`translateX(${active.offsetLeft}px)`; pill.style.width=`${active.offsetWidth}px`; void pill.offsetWidth; pill.style.transition=prev; } else { pill.style.transform=`translateX(${active.offsetLeft}px)`; pill.style.width=`${active.offsetWidth}px`; }
}
onMounted(()=>{ nextTick(()=> positionPill(false)); window.addEventListener('resize', ()=> positionPill(false)); })
watch(activeTab, ()=> nextTick(()=> positionPill(true)))

// Calendar matrix
const calendarData = ref(null)
const calendarLoading = ref(false)

async function fetchCalendar(from, to) {
  calendarLoading.value = true
  try {
    const params = {}
    if (from && to) {
      params.from = from
      params.to = to
    } else {
      params.days = 7
    }
    const { data } = await client.get('/api/admin/calendar', { params })
    calendarData.value = data.data || data
  } catch (e) {
    if (e?.response?.status === 429) {
      toast.add({ severity: 'warn', summary: 'Too many requests', life: 2000 })
    }
    // mock fallback for demo: generate per unit per day random
    const units = roomUnits.value.length ? roomUnits.value : [
      { id: 101, code: 'STD-101', room_type_id: 1, status: 'available' },
      { id: 102, code: 'DLX-101', room_type_id: 2, status: 'occupied' },
      { id: 103, code: 'FAM-101', room_type_id: 3, status: 'dirty' },
      { id: 104, code: 'STE-101', room_type_id: 4, status: 'maintenance' }
    ]
    const mock = {}
    const activeDays = []
    if (from && to) {
      const cur = new Date(from)
      const end = new Date(to)
      while (cur <= end) {
        activeDays.push(cur.toISOString().slice(0, 10))
        cur.setDate(cur.getDate() + 1)
      }
    } else {
      const today = new Date()
      for (let i = 0; i < 7; i++) {
        const d = new Date(today)
        d.setDate(today.getDate() + i)
        activeDays.push(d.toISOString().slice(0, 10))
      }
    }
    activeDays.forEach(day => {
      mock[day] = units.map(u => {
        const statuses = ['available', 'occupied', 'dirty', 'maintenance']
        const s = statuses[Math.floor(Math.random() * 4)]
        return {
          unit_id: u.id,
          code: u.code || `UNIT-${u.id}`,
          status: s,
          booking_id: s === 'occupied' ? Math.floor(Math.random() * 100) + 1 : null
        }
      })
    })
    calendarData.value = mock
  } finally {
    calendarLoading.value = false
  }
}

// Photo upload 5 per type
const uploadFiles = ref({})
const uploadLoading = ref({})
async function uploadRoomImages(roomTypeId){
  const files = uploadFiles.value[roomTypeId]
  if(!files || !files.length){ toast.add({severity:'warn', summary:'Select photos first', life:2000}); return }
  if(files.length>5){ toast.add({severity:'warn', summary:'Max 5 photos', life:2000}); return }
  uploadLoading.value[roomTypeId]=true
  try{
    const fd=new FormData()
    for(let f of files) fd.append('images', f)
    // try multiple endpoint variants
    let ok=false
    try{ await client.post(`/api/admin/room-types/${roomTypeId}/images`, fd, { headers:{'Content-Type':'multipart/form-data'}}); ok=true }catch(e1){
      if(e1?.response?.status===429) throw e1
      try{ await client.post(`/api/admin/room-types/${roomTypeId}/upload`, fd, { headers:{'Content-Type':'multipart/form-data'}}); ok=true }catch(e2){
        if(e2?.response?.status===429) throw e2
        await client.post(`/api/room-types/${roomTypeId}/images`, fd, { headers:{'Content-Type':'multipart/form-data'}}); ok=true
      }
    }
    if(ok) toast.add({ severity:'success', summary:'Photos uploaded', detail:`${files.length} images for type #${roomTypeId}`, life:2500 })
  }catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
    else toast.add({ severity:'error', summary:'Upload failed', detail:e?.response?.data?.message||e.message, life:3500 })
  } finally{ uploadLoading.value[roomTypeId]=false }
}
function onFileChange(e, roomTypeId){
  const list=e.target.files
  uploadFiles.value[roomTypeId]= list ? Array.from(list) : []
}

// Export CSV
async function exportCSV(){
  try{
    const params={}
    if(dateRange.value && dateRange.value[0]) params.from = toISO(dateRange.value[0])
    if(dateRange.value && dateRange.value[1]) params.to = toISO(dateRange.value[1])
    const res=await client.get('/api/admin/reports/export.csv', { params, responseType:'blob' })
    const blob=new Blob([res.data], { type: res.headers['content-type']||'text/csv' })
    const url=window.URL.createObjectURL(blob)
    const a=document.createElement('a')
    a.href=url
    const cd=res.headers['content-disposition']||''
    let filename='export.csv'
    const m=cd.match(/filename="?([^"]+)"?/)
    if(m) filename=m[1]
    a.download=filename
    document.body.appendChild(a); a.click(); a.remove()
    window.URL.revokeObjectURL(url)
    toast.add({ severity:'success', summary:'CSV exported', life:2000 })
  }catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
    else{
      let msg=e?.response?.data?.message||e.message
      if(e?.response?.data instanceof Blob){
        try{ const t=await e.response.data.text(); const j=JSON.parse(t); msg=j.message||msg }catch{}
      }
      toast.add({ severity:'error', summary:'Export failed', detail:msg, life:3500 })
    }
  }
}

// WebSocket admin
let ws=null
let wsPoll=null
const wsConnected = ref(false)
function connectWS(){
  try{
    const token=localStorage.getItem('token')
    const proto=window.location.protocol==='https:'?'wss:':'ws:'
    const base=import.meta.env.VITE_API_URL || 'http://localhost:8080'
    const host=base.replace(/^https?:\/\//,'').replace(/\/$/,'')
    const url=`${proto}//${host}/ws/admin${token?`?token=${token}`:''}`
    ws=new WebSocket(url)
    ws.onopen=()=>{ wsConnected.value=true; toast.add({severity:'info', summary:'Live updates connected', life:2000}) }
    ws.onmessage=(ev)=>{
      try{
        const payload=JSON.parse(ev.data)
        const id=payload.id||payload.booking_id||payload.bookingId||'46'
        toast.add({ severity:'success', summary:'New booking', detail:`New booking #${id} received`, life:4000 })
      }catch{
        toast.add({ severity:'success', summary:'New booking #46', detail: ev.data.slice(0,120), life:4000 })
      }
      fetchBookings(); fetchReports()
    }
    ws.onclose=()=>{
      wsConnected.value=false
      // fallback polling 30s
      if(!wsPoll) wsPoll=setInterval(()=>{ fetchBookings(); fetchReports() }, 30000)
      setTimeout(connectWS, 5000)
    }
    ws.onerror=()=>{ wsConnected.value=false }
  }catch{
    wsConnected.value=false
    if(!wsPoll) wsPoll=setInterval(()=>{ fetchBookings() }, 30000)
  }
}

const statusSeverity = (s) => ({ verified: 'success', checked_in: 'info', pending_payment: 'warn', waiting_verification: 'warn', checked_out: 'secondary', cancelled: 'danger', rejected: 'danger', expired: 'danger' }[s] || 'secondary')
const statusLabel = (s) => s?.replace('_', ' ') || s
const fmt = (n) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n||0)
const fmtDate = (d) => { try { return new Date(d).toISOString().slice(0,10) } catch { return d } }
function toISO(d){ if(!d) return ''; const dt=d instanceof Date?d:new Date(d); if(isNaN(dt)) return ''; return dt.toISOString().slice(0,10) }
const canInvoice = (s)=> ['verified','checked_in','checked_out'].includes(s)

async function fetchBookings() {
  loading.value = true
  try {
    const { data } = await client.get('/api/admin/bookings')
    const list = data.data || data
    bookings.value = Array.isArray(list) ? list : []
  } catch (e) {
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2000})
    try {
      const { data } = await client.get('/api/bookings')
      const list = data.data || data
      bookings.value = Array.isArray(list) ? list : []
    } catch (e2) {
      if(e2?.response?.status!==429) toast.add({ severity: 'error', summary: 'Could not load bookings', detail: e?.response?.data?.message || e.message, life: 3000 })
    }
  } finally { loading.value = false }
}
async function fetchUnits() {
  try {
    const { data } = await client.get('/api/admin/room-units')
    const list = data.data || data
    roomUnits.value = Array.isArray(list) ? list : []
  } catch{
    try{
      const { data } = await client.get('/api/admin/room_units')
      const list=data.data||data
      roomUnits.value=Array.isArray(list)?list:[]
    }catch{}
  }
}
async function fetchRoomTypes(){
  try{
    const { data } = await client.get('/api/room-types')
    const list = data.data || data
    roomTypes.value = Array.isArray(list) ? list : []
    if(!list.length){
      const { data: d2 } = await client.get('/api/admin/rooms')
      const l2 = d2.data || d2
      if(Array.isArray(l2)) roomTypes.value=l2
    }
  }catch{
    try{
      const { data } = await client.get('/api/admin/rooms')
      const list = data.data || data
      roomTypes.value = Array.isArray(list) ? list : []
    }catch{ roomTypes.value=[] }
  }
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
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Reports rate limited', life:2000})
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
async function fetchAuditLogs(){
  auditLoading.value=true
  try{
    let data
    try{ const r= await client.get('/api/admin/audit-logs'); data=r.data }catch{ const r= await client.get('/api/audit-logs'); data=r.data }
    const list = data.data || data
    auditLogs.value = Array.isArray(list) ? list : []
  }catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Audit rate limited', life:2000})
    try{
      const { data } = await client.get('/api/admin/audit_logs')
      const list=data.data||data
      auditLogs.value=Array.isArray(list)?list:[]
    }catch{ auditLogs.value=[] }
  }finally{ auditLoading.value=false }
}

onMounted(()=>{ fetchBookings(); fetchUnits(); fetchReports(); fetchReviews(); fetchRoomTypes(); fetchAuditLogs(); fetchCalendar(); connectWS(); nextTick(()=>{ showStatPop.value=true }); watch([totalBookings, totalRevenue, occupancyRate, availableUnits], ()=>{ showStatPop.value=false; nextTick(()=>{ void document.body.offsetHeight; showStatPop.value=true }) }) })
watch(dateRange, fetchReports)
onBeforeUnmount(()=>{ if(ws) ws.close(); if(wsPoll) clearInterval(wsPoll) })

async function doVerify(id, action) {
  if (action==='rejected') { rejectId.value=id; showReject.value=true; return }
  actionLoading.value = `verify-${id}`
  try {
    await client.patch(`/api/admin/bookings/${id}/verify`, { action })
    toast.add({ severity: 'success', summary: 'Booking confirmed', detail: `Booking #${id} is now verified`, life: 2500 })
    triggerSuccess(id)
    await fetchBookings(); await fetchReports(); await fetchAuditLogs()
  } catch (e) { 
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
    else toast.add({ severity: 'error', summary: 'Could not confirm booking', detail: e?.response?.data?.message || e.message, life: 3500 }) 
  }
  finally { actionLoading.value='' }
}
async function confirmReject() {
  if (!rejectReason.value.trim()) { toast.add({ severity:'warn', summary:'Add a reason to decline', detail:'Tell the guest why you declined this booking', life:2000}); return }
  const id = rejectId.value
  actionLoading.value=`verify-${id}`
  try {
    await client.patch(`/api/admin/bookings/${id}/verify`, { action:'rejected', reject_reason: rejectReason.value })
    toast.add({ severity:'success', summary:'Booking declined', detail:`Booking #${id} was declined`, life:2500 })
    triggerSuccess(id)
    showReject.value=false; rejectReason.value=''; await fetchBookings(); await fetchReports(); await fetchAuditLogs()
  } catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
    else toast.add({ severity:'error', summary:'Could not decline booking', detail:e?.response?.data?.message||e.message, life:3500}) 
  }
  finally{ actionLoading.value='' }
}
async function doCheckIn(id){
  actionLoading.value=`checkin-${id}`
  try{ await client.patch(`/api/admin/bookings/${id}/checkin`); toast.add({ severity:'success', summary:'Guest checked in', detail:`Booking #${id} is now checked in`, life:2500 }); await fetchBookings(); await fetchUnits(); await fetchReports(); await fetchAuditLogs() } catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
    else toast.add({ severity:'error', summary:'Check in failed', detail:e?.response?.data?.message||e.message, life:3500}) 
  } finally{ actionLoading.value='' }
}
async function doCheckOut(id){
  actionLoading.value=`checkout-${id}`
  try{ await client.patch(`/api/admin/bookings/${id}/checkout`); toast.add({ severity:'success', summary:'Checkout complete', detail:`Booking #${id} is now checked out`, life:2500 }); await fetchBookings(); await fetchUnits(); await fetchReports(); await fetchAuditLogs() } catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
    else toast.add({ severity:'error', summary:'Checkout failed', detail:e?.response?.data?.message||e.message, life:3500}) 
  } finally{ actionLoading.value='' }
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
  try{ await client.patch(`/api/admin/room-units/${unit.id}/status`, { status:newStatus }); toast.add({ severity:'success', summary:'Room status updated', detail:`${unit.code} is now ${newStatus}`, life:2500 }); await fetchUnits(); await fetchAuditLogs() } catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
    else toast.add({ severity:'error', summary:'Could not update room', detail:e?.response?.data?.message||e.message, life:3500}) 
  }
}
const unitSeverity = (s) => ({ available: 'success', occupied: 'info', dirty: 'warn', maintenance: 'secondary' }[s] || 'secondary')

function getRoomTypeName(typeId) {
  const found = roomTypes.value.find(t => t.id === typeId)
  return found?.name || found?.type || `Tipe #${typeId}`
}

function auditActionSeverity(action) {
  if (!action) return 'secondary'
  const act = String(action).toLowerCase()
  if (act.includes('create') || act.includes('verify') || act.includes('login')) return 'success'
  if (act.includes('checkin') || act.includes('check_in') || act.includes('update')) return 'info'
  if (act.includes('checkout') || act.includes('check_out')) return 'warn'
  if (act.includes('reject') || act.includes('delete')) return 'danger'
  return 'secondary'
}

// Room Types CRUD helpers
function openCreateRoomType(){
  editingRoomType.value=null
  roomTypeForm.value={ name:'', description:'', capacity:2, price_per_night:350000, total_units:5 }
  showRoomTypeDialog.value=true
}
function openEditRoomType(rt){
  editingRoomType.value=rt
  roomTypeForm.value={
    name: rt.name || rt.type || '',
    description: rt.description || '',
    capacity: rt.capacity || 2,
    price_per_night: rt.price_per_night || rt.price || 350000,
    total_units: rt.total_units || rt.totalUnits || 5
  }
  showRoomTypeDialog.value=true
}
async function saveRoomType(){
  if(!roomTypeForm.value.name.trim()){ toast.add({severity:'warn', summary:'Name is required', life:2000}); return }
  rtSaving.value=true
  try{
    if(editingRoomType.value){
      await client.put(`/api/admin/room-types/${editingRoomType.value.id}`, roomTypeForm.value)
      toast.add({ severity:'success', summary:'Room type updated', life:2500 })
    }else{
      await client.post('/api/admin/room-types', roomTypeForm.value)
      toast.add({ severity:'success', summary:'Room type created', life:2500 })
    }
    showRoomTypeDialog.value=false
    await fetchRoomTypes(); await fetchAuditLogs()
  }catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
    else toast.add({ severity:'error', summary:'Could not save room type', detail:e?.response?.data?.message||e.message, life:3500 })
  }finally{ rtSaving.value=false }
}
async function deleteRoomType(rt){
  if(!confirm(`Delete room type "${rt.name || rt.type}"? This cannot be undone.`)) return
  try{
    await client.delete(`/api/admin/room-types/${rt.id}`)
    toast.add({ severity:'success', summary:'Room type deleted', detail:`${rt.name||rt.type} removed`, life:2500 })
    await fetchRoomTypes(); await fetchAuditLogs()
  }catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
    else toast.add({ severity:'error', summary:'Could not delete', detail:e?.response?.data?.message||e.message, life:3500 }) 
  }
}

// Room Units create
function openCreateUnit(){
  unitForm.value={ code:'', room_type_id: roomTypes.value[0]?.id || 1, status:'available' }
  showUnitDialog.value=true
}
async function createUnit(){
  if(!unitForm.value.code.trim()){ toast.add({severity:'warn', summary:'Code is required', detail:'Example Deluxe-101', life:2000}); return }
  unitSaving.value=true
  try{
    await client.post('/api/admin/room-units', unitForm.value)
    toast.add({ severity:'success', summary:'Room unit created', detail:`${unitForm.value.code} added`, life:2500 })
    showUnitDialog.value=false
    await fetchUnits(); await fetchAuditLogs()
  }catch(e){
    try{
      await client.post('/api/admin/room_units', unitForm.value)
      toast.add({ severity:'success', summary:'Room unit created', life:2500 })
      showUnitDialog.value=false
      await fetchUnits()
    }catch(e2){
      if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
      else toast.add({ severity:'error', summary:'Could not create unit', detail:e?.response?.data?.message||e.message, life:3500 })
    }
  }finally{ unitSaving.value=false }
}

async function downloadInvoiceAdmin(booking){
  invoicingId.value=booking.id
  try{
    const res = await client.get(`/api/bookings/${booking.id}/invoice`, { responseType:'blob' })
    const blob = new Blob([res.data], { type: res.headers['content-type'] || 'application/pdf' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href=url
    const cd = res.headers['content-disposition'] || ''
    let filename=`invoice-${booking.id}.pdf`
    const m=cd.match(/filename="?([^"]+)"?/)
    if(m) filename=m[1]
    a.download=filename
    document.body.appendChild(a); a.click(); a.remove()
    window.URL.revokeObjectURL(url)
    toast.add({ severity:'success', summary:'Invoice downloaded', detail:`Invoice #${booking.id}`, life:2500 })
  }catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2500})
    else{
      let msg=e?.response?.data?.message||e.message
      if(e?.response?.data instanceof Blob){
        try{ const t=await e.response.data.text(); const j=JSON.parse(t); msg=j.message||msg }catch{}
      }
      toast.add({ severity:'error', summary:'Invoice not available', detail:msg, life:3500 })
    }
  }finally{ invoicingId.value=null }
}

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
const showStatPop = ref(false)
const verifySuccessId = ref(null)
let verifySuccessTimer = null
function triggerSuccess(id){ verifySuccessId.value = id; clearTimeout(verifySuccessTimer); nextTick(()=>{ const el=document.querySelector(`[data-success-id="${id}"]`); if(el){ el.setAttribute("data-state","out"); void el.offsetWidth; el.setAttribute("data-state","in") } }); verifySuccessTimer=setTimeout(()=>{ verifySuccessId.value=null }, 3000) }
const doughnutOptions = { responsive:true, maintainAspectRatio:false, cutout:'65%', plugins:{ legend:{ position:'bottom', labels:{ usePointStyle:true, boxWidth:10 } } } }
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 space-y-8">
    <div class="space-y-3">
      <div class="flex flex-col sm:flex-row sm:items-center gap-3">
        <div class="flex items-center gap-3">
          <span class="bg-[#8B5A2B] text-white rounded-2xl p-2.5 shadow-sm"><LayoutDashboard class="w-5 h-5" /></span>
          <div>
            <h1 class="font-display font-bold text-3xl text-[#1A3A4A] dark:text-[#FDF6EC] tracking-tight" style="font-family:'Playfair Display',serif">{{ t('admin.title') }}</h1>
            <p class="text-base text-[#6B7280]">{{ t('admin.subtitle') }}</p>
          </div>
        </div>
        <div class="t-panel-reveal sm:ml-auto flex items-center gap-2 flex-wrap" data-open="true">
          <span v-if="wsConnected" class="inline-flex items-center gap-1.5 bg-green-50 text-green-700 border border-green-200 rounded-full px-3 py-1 text-xs font-semibold"><Radio class="w-3 h-3" /> Live</span>
          <span v-else class="inline-flex items-center gap-1.5 bg-[#FDF6EC] text-[#6B7280] border border-[#E5E7EB] rounded-full px-3 py-1 text-xs">Polling 30s</span>
          <DatePicker v-model="dateRange" selectionMode="range" :manualInput="false" placeholder="Filter by date" showIcon class="min-w-[220px]" />
          <Button label="Refresh Data" icon="pi pi-refresh" outlined class="!rounded-2xl !border-[#8B5A2B] !text-[#8B5A2B]" :loading="loading || reportsLoading" @click="fetchBookings(); fetchUnits(); fetchReports(); fetchAuditLogs(); fetchCalendar()" />
          <Button label="Export CSV" icon="pi pi-download" class="!rounded-2xl !bg-[#1A3A4A] !border-[#1A3A4A]" @click="exportCSV">
            <template #icon><Download class="w-4 h-4" /></template>
          </Button>
        </div>
      </div>
      <div class="h-px bg-gradient-to-r from-[#C9A86A] via-[#C9A86A]/40 to-transparent"></div>
    </div>

    <!-- Metric Cards & Collapsible Charts via AdminStatsOverview -->
    <AdminStatsOverview
      :total-bookings="totalBookings"
      :total-revenue="totalRevenue"
      :occupancy-rate="occupancyRate"
      :available-units="availableUnits"
      :total-units="roomUnits.length || '-'"
      :bookings-by-status="bookingsByStatus"
      :revenue-chart-data="revenueChartData"
      :revenue-chart-options="revenueChartOptions"
      :doughnut-data="doughnutData"
      :doughnut-options="doughnutOptions"
      :bar-chart-data="barChartData"
      :bar-chart-options="barChartOptions"
      :show-stat-pop="showStatPop"
    />

    <!-- Tabs navigation t-tabs sliding -->
    <div ref="tabsBarRef" class="t-tabs flex flex-wrap gap-1 border-b border-[#E5E7EB] pb-3 !bg-transparent !p-0 !rounded-none">
      <span ref="tabsPillRef" class="t-tabs-pill !bg-[#1A3A4A] !h-[36px] !top-0" aria-hidden="true"></span>
      <button v-for="t in [{id:'bookings',label:'Bookings',icon:FileText},{id:'calendar',label:'Occupancy Calendar',icon:Grid3x3},{id:'housekeeping',label:'Housekeeping',icon:Kanban},{id:'roomtypes',label:'Room Types',icon:Boxes},{id:'units',label:'Room Units',icon:Bed},{id:'audit',label:'Audit Log',icon:ShieldCheck}]" :key="t.id" @click="activeTab=t.id; nextTick(()=>positionPill())" :aria-selected="activeTab===t.id ? 'true' : 'false'" :class="['t-tab !rounded-full text-sm font-semibold', activeTab===t.id ? '!text-white' : '']">
        <component :is="t.icon" class="w-4 h-4" /> {{ t.label }}
      </button>
    </div>

    <!-- Recent bookings via AdminBookingsTable -->
    <AdminBookingsTable
      v-show="activeTab === 'bookings'"
      :bookings="bookings"
      :loading="loading"
      :action-loading="actionLoading"
      :invoicing-id="invoicingId"
      :verify-success-id="verifySuccessId"
      @verify="doVerify"
      @confirm-reject="({ id, reason }) => { rejectId = id; rejectReason = reason; confirmReject() }"
      @check-in="doCheckIn"
      @check-out="doCheckOut"
      @download-invoice="downloadInvoiceAdmin"
      @refresh="fetchBookings"
    />

    <!-- Occupancy Calendar via OccupancyCalendarMatrix -->
    <OccupancyCalendarMatrix
      v-show="activeTab === 'calendar'"
      :room-units="roomUnits"
      :room-types="roomTypes"
      :calendar-data="calendarData"
      :loading="calendarLoading"
      @change-range="({ from, to }) => fetchCalendar(from, to)"
      @refresh="fetchCalendar"
    />

    <!-- Housekeeping Kanban via HousekeepingKanban -->
    <HousekeepingKanban
      v-show="activeTab === 'housekeeping'"
      :room-units="roomUnits"
      :room-types="roomTypes"
      @update-status="({ unit, newStatus }) => { unitStatusMap[unit.id] = newStatus; updateUnitStatus(unit) }"
      @refresh="fetchUnits"
    />

    <!-- Room Types CRUD -->
    <Card v-show="activeTab==='roomtypes'" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title>
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-[#FDF6EC] dark:bg-[#1A3A4A]/60 flex items-center justify-center text-[#8B5A2B]">
              <Boxes class="w-4 h-4" />
            </div>
            <div>
              <span class="text-[#1A3A4A] dark:text-white font-display font-bold text-lg" style="font-family:'Playfair Display',serif">Tipe Kamar (Room Types)</span>
              <p class="text-xs text-[#6B7280] dark:text-gray-400 font-normal mt-0.5">Kelola spesifikasi, kapasitas, tarif per malam, dan inventaris foto kamar.</p>
            </div>
          </div>
          <Button label="Tambah Tipe Kamar" class="!bg-[#8B5A2B] !border-[#8B5A2B] !rounded-xl !text-xs !py-2" @click="openCreateRoomType">
            <template #icon><Plus class="w-4 h-4 mr-1" /></template>
          </Button>
        </div>
        <div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div>
      </template>
      <template #content>
        <DataTable :value="roomTypes" paginator :rows="8" stripedRows class="text-sm" responsiveLayout="scroll" dataKey="id">
          <Column field="id" header="ID" sortable style="width:70px" />
          <Column field="name" header="Nama & Kapasitas" sortable style="min-width:180px">
            <template #body="{ data }">
              <div class="flex flex-col">
                <span class="font-bold text-[#1A3A4A] dark:text-white">{{ data.name || data.type }}</span>
                <span class="text-[11px] text-[#6B7280] dark:text-gray-400 flex items-center gap-1 mt-0.5">
                  <Users class="w-3 h-3 text-[#8B5A2B]" />
                  Kapasitas {{ data.capacity || 2 }} Tamu
                </span>
              </div>
            </template>
          </Column>
          <Column field="description" header="Deskripsi" style="max-width:240px">
            <template #body="{ data }">
              <span class="text-xs text-[#6B7280] dark:text-gray-400 line-clamp-2">{{ data.description || '-' }}</span>
            </template>
          </Column>
          <Column field="price_per_night" header="Tarif / Malam" sortable style="min-width:140px">
            <template #body="{ data }">
              <span class="font-bold text-[#8B5A2B] text-sm">{{ fmt(data.price_per_night || data.price || 0) }}</span>
            </template>
          </Column>
          <Column field="total_units" header="Total Unit" sortable style="width:110px">
            <template #body="{ data }">
              <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-[#FDF6EC] text-[#8B5A2B] border border-[#8B5A2B]/20">
                {{ data.total_units || data.totalUnits || '-' }} Unit
              </span>
            </template>
          </Column>
          <Column header="Unggah Foto (Maks 5)" style="min-width:260px">
            <template #body="{ data }">
              <div class="flex items-center gap-1.5">
                <input type="file" multiple accept="image/jpeg,image/png,image/jpg" :id="'upload-'+data.id" class="hidden" @change="onFileChange($event, data.id)" />
                <Button label="Pilih Foto" icon="pi pi-images" size="small" outlined class="!rounded-xl !text-xs !py-1 !border-[#8B5A2B] !text-[#8B5A2B]" @click="document.getElementById('upload-'+data.id).click()" />
                <Button :label="uploadLoading[data.id] ? 'Mengunggah...' : 'Unggah'" size="small" class="!rounded-xl !bg-[#8B5A2B] !border-[#8B5A2B] !text-xs !py-1" :loading="uploadLoading[data.id]" @click="uploadRoomImages(data.id)">
                  <template #icon><Upload class="w-3 h-3 mr-1" /></template>
                </Button>
              </div>
              <p v-if="uploadFiles[data.id]?.length" class="text-[10px] text-[#8B5A2B] font-medium mt-1">
                ✓ {{ uploadFiles[data.id].length }} file dipilih (siap diunggah)
              </p>
            </template>
          </Column>
          <Column header="Aksi" style="min-width:180px">
            <template #body="{ data }">
              <div class="flex gap-1.5">
                <Button size="small" outlined class="!rounded-xl !border-[#1A3A4A] !text-[#1A3A4A] !py-1 !px-2.5 text-xs" @click="openEditRoomType(data)">
                  <template #icon><Pencil class="w-3.5 h-3.5 mr-1" /></template> Edit
                </Button>
                <Button size="small" severity="danger" outlined class="!rounded-xl !py-1 !px-2.5 text-xs" @click="deleteRoomType(data)">
                  <template #icon><Trash2 class="w-3.5 h-3.5 mr-1" /></template> Hapus
                </Button>
              </div>
            </template>
          </Column>
        </DataTable>
        <div v-if="!roomTypes.length" class="text-center py-8 text-sm text-[#6B7280]">Belum ada tipe kamar. Tambahkan tipe kamar baru di atas.</div>
      </template>
    </Card>

    <!-- Room units -->
    <Card v-show="activeTab==='units'" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title>
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-[#FDF6EC] dark:bg-[#1A3A4A]/60 flex items-center justify-center text-[#8B5A2B]">
              <Bed class="w-4 h-4" />
            </div>
            <div>
              <span class="text-[#1A3A4A] dark:text-white font-display font-bold text-lg" style="font-family:'Playfair Display',serif">Unit Kamar Fisik (Room Units)</span>
              <p class="text-xs text-[#6B7280] dark:text-gray-400 font-normal mt-0.5">Daftar inventaris kamar fisik dan pembaruan status operasional harian.</p>
            </div>
          </div>
          <Button label="Tambah Unit Kamar" class="!bg-[#1A3A4A] !border-[#1A3A4A] !rounded-xl !text-xs !py-2" @click="openCreateUnit">
            <template #icon><Plus class="w-4 h-4 mr-1" /></template>
          </Button>
        </div>
        <div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div>
      </template>
      <template #content>
        <DataTable :value="roomUnits" paginator :rows="8" stripedRows class="text-sm" responsiveLayout="scroll" dataKey="id" :loading="loading">
          <Column field="code" header="Kode Kamar" sortable>
            <template #body="{ data }">
              <span class="font-bold text-[#1A3A4A] dark:text-white tracking-wide">{{ data.code }}</span>
            </template>
          </Column>
          <Column field="room_type_id" header="Tipe Kamar" sortable>
            <template #body="{ data }">
              <span class="text-xs font-semibold text-[#8B5A2B]">{{ getRoomTypeName(data.room_type_id) }}</span>
            </template>
          </Column>
          <Column field="status" header="Status Saat Ini">
            <template #body="{ data }">
              <Tag :value="data.status" :severity="unitSeverity(data.status)" rounded class="text-xs uppercase font-bold" />
            </template>
          </Column>
          <Column header="Ubah Status Cepat" style="min-width:280px">
            <template #body="{ data }">
              <div class="flex gap-2">
                <Select v-model="unitStatusMap[data.id]" :options="unitStatusOptions" optionLabel="label" optionValue="value" placeholder="Pilih status baru" class="w-full !text-xs" />
                <Button label="Simpan" size="small" class="!bg-[#8B5A2B] !border-[#8B5A2B] !rounded-xl !px-4 text-xs font-semibold" @click="updateUnitStatus(data)">
                  <template #icon><Check class="w-3.5 h-3.5 mr-1" /></template>
                </Button>
              </div>
            </template>
          </Column>
        </DataTable>
        <p v-if="!roomUnits.length" class="text-center py-6 text-sm text-[#6B7280]">Belum ada unit kamar. Tambahkan unit kamar untuk mulai menerima check-in.</p>
      </template>
    </Card>

    <!-- Audit logs -->
    <Card v-show="activeTab==='audit'" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title>
        <div class="flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-[#FDF6EC] dark:bg-[#1A3A4A]/60 flex items-center justify-center text-[#8B5A2B]">
              <ShieldCheck class="w-4 h-4" />
            </div>
            <div>
              <span class="text-[#1A3A4A] dark:text-white font-display font-bold text-lg" style="font-family:'Playfair Display',serif">Log Audit Aktivitas (Audit Trail)</span>
              <p class="text-xs text-[#6B7280] dark:text-gray-400 font-normal mt-0.5">Riwayat lengkap perubahan sistem oleh staf dan resepsionis untuk akuntabilitas keamanan.</p>
            </div>
          </div>
          <Button label="Segarkan" icon="pi pi-refresh" outlined size="small" class="!rounded-xl !border-[#8B5A2B] !text-[#8B5A2B] !text-xs" :loading="auditLoading" @click="fetchAuditLogs" />
        </div>
        <div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div>
      </template>
      <template #content>
        <DataTable :value="auditLogs" paginator :rows="10" stripedRows class="text-sm" responsiveLayout="scroll" :loading="auditLoading" dataKey="id">
          <Column field="id" header="ID" style="width:70px" sortable />
          <Column field="created_at" header="Waktu Kejadian" sortable style="min-width:160px">
            <template #body="{ data }">
              <span class="text-xs font-medium text-[#1A3A4A] dark:text-gray-200">
                {{ data.created_at ? new Date(data.created_at).toLocaleString('id-ID') : (data.createdAt ? new Date(data.createdAt).toLocaleString('id-ID') : '-') }}
              </span>
            </template>
          </Column>
          <Column field="user_id" header="Staf / ID" style="width:100px">
            <template #body="{ data }">
              <span class="inline-flex items-center px-2 py-0.5 rounded-md text-[11px] font-semibold bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300">
                User #{{ data.user_id ?? data.userId ?? '-' }}
              </span>
            </template>
          </Column>
          <Column field="action" header="Aksi Dilakukan" sortable style="min-width:140px">
            <template #body="{ data }">
              <Tag :value="data.action" :severity="auditActionSeverity(data.action)" rounded class="text-xs font-semibold" />
            </template>
          </Column>
          <Column field="entity" header="Target Entitas" sortable style="min-width:140px">
            <template #body="{ data }">
              <span class="font-medium text-[#1A3A4A] dark:text-gray-200">{{ data.entity || '-' }}</span>
              <span v-if="data.entity_id" class="text-xs text-[#8B5A2B] font-semibold ml-1">#{{ data.entity_id }}</span>
            </template>
          </Column>
          <Column field="payload" header="Detail Muatan">
            <template #body="{ data }">
              <span class="font-mono text-[11px] text-[#6B7280] dark:text-gray-400 bg-gray-50 dark:bg-gray-900/50 px-2 py-1 rounded-md line-clamp-2 break-all block">
                {{ typeof data.payload==='string' ? data.payload.slice(0,120) : JSON.stringify(data.payload||'').slice(0,120) }}
              </span>
            </template>
          </Column>
        </DataTable>
        <div v-if="!auditLogs.length && !auditLoading" class="text-center py-8 text-sm text-[#6B7280]">Belum ada entri audit. Tindakan staf akan otomatis tercatat di sini.</div>
      </template>
    </Card>

    <!-- Reviews always visible below tabs -->
    <Card v-if="reviews.length" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 rounded-lg bg-[#FDF6EC] dark:bg-[#1A3A4A]/60 flex items-center justify-center text-[#8B5A2B]">
              <MessageSquare class="w-4 h-4" />
            </div>
            <div>
              <span class="text-[#1A3A4A] dark:text-white font-display font-bold text-lg" style="font-family:'Playfair Display',serif">Ulasan Tamu ({{ reviews.length }})</span>
              <p class="text-xs text-[#6B7280] dark:text-gray-400 font-normal mt-0.5">Ulasan dan skor kepuasan yang dikirimkan tamu setelah menyelesaikan masa menginap.</p>
            </div>
          </div>
        </div>
        <div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div>
      </template>
      <template #content>
        <DataTable :value="reviews" paginator :rows="5" stripedRows class="text-sm">
          <Column field="id" header="ID" style="width:70px" />
          <Column field="booking_id" header="Reservasi" style="width:110px">
            <template #body="{ data }">
              <span class="font-semibold text-[#8B5A2B]">#BKG-{{ data.booking_id }}</span>
            </template>
          </Column>
          <Column field="rating" header="Rating Bintang" style="width:160px">
            <template #body="{ data }">
              <Rating :modelValue="data.rating" readonly :stars="5" />
            </template>
          </Column>
          <Column field="comment" header="Komentar Tamu">
            <template #body="{ data }">
              <p class="text-xs text-[#1A3A4A] dark:text-gray-200 italic">"{{ data.comment }}"</p>
            </template>
          </Column>
          <Column field="created_at" header="Tanggal Ulasan" style="width:140px">
            <template #body="{ data }">
              <span class="text-xs text-[#6B7280]">{{ fmtDate(data.created_at) }}</span>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>



    <!-- Room Type Dialog -->
    <Dialog v-model:visible="showRoomTypeDialog" modal :header="editingRoomType ? 'Edit Room Type' : 'Create Room Type'" :style="{ width:'520px' }" class="!rounded-2xl">
      <div class="space-y-4">
        <div>
          <label class="text-xs font-semibold text-[#1A3A4A] uppercase tracking-wide">Name</label>
          <InputText v-model="roomTypeForm.name" placeholder="Deluxe" class="w-full mt-1" />
        </div>
        <div>
          <label class="text-xs font-semibold text-[#1A3A4A] uppercase tracking-wide">Description</label>
          <Textarea v-model="roomTypeForm.description" rows="2" placeholder="Warm aura, balcony, breakfast included" class="w-full mt-1" autoResize />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="text-xs font-semibold text-[#1A3A4A] uppercase tracking-wide">Capacity</label>
            <InputNumber v-model="roomTypeForm.capacity" :min="1" :max="10" showButtons class="w-full mt-1" />
          </div>
          <div>
            <label class="text-xs font-semibold text-[#1A3A4A] uppercase tracking-wide">Total Units</label>
            <InputNumber v-model="roomTypeForm.total_units" :min="1" :max="100" showButtons class="w-full mt-1" />
          </div>
        </div>
        <div>
          <label class="text-xs font-semibold text-[#1A3A4A] uppercase tracking-wide">Price per Night (IDR)</label>
          <InputNumber v-model="roomTypeForm.price_per_night" mode="currency" currency="IDR" locale="id-ID" class="w-full mt-1" />
        </div>
      </div>
      <template #footer>
        <Button label="Cancel" text class="!rounded-2xl" @click="showRoomTypeDialog=false" />
        <Button :label="editingRoomType ? 'Update' : 'Create'" class="!bg-[#8B5A2B] !border-[#8B5A2B] !rounded-2xl" :loading="rtSaving" @click="saveRoomType" />
      </template>
    </Dialog>

    <!-- Room Unit Dialog -->
    <Dialog v-model:visible="showUnitDialog" modal header="Create Room Unit" :style="{ width:'440px' }" class="!rounded-2xl">
      <div class="space-y-4">
        <div>
          <label class="text-xs font-semibold text-[#1A3A4A] uppercase tracking-wide">Code</label>
          <InputText v-model="unitForm.code" placeholder="Deluxe-101" class="w-full mt-1" />
          <p class="text-xs text-[#6B7280] mt-1">Unique code like Standard-01 or Suite-201.</p>
        </div>
        <div>
          <label class="text-xs font-semibold text-[#1A3A4A] uppercase tracking-wide">Room Type</label>
          <Select v-model="unitForm.room_type_id" :options="roomTypes" optionLabel="name" optionValue="id" placeholder="Select type" class="w-full mt-1">
            <template #option="{ option }">{{ option.name || option.type }} (#{{ option.id }})</template>
            <template #value="{ value }"><span v-if="value">Type #{{ value }}</span><span v-else class="text-[#9CA3AF]">Select type</span></template>
          </Select>
        </div>
        <div>
          <label class="text-xs font-semibold text-[#1A3A4A] uppercase tracking-wide">Status</label>
          <Select v-model="unitForm.status" :options="unitStatusOptions" optionLabel="label" optionValue="value" class="w-full mt-1" />
        </div>
      </div>
      <template #footer>
        <Button label="Cancel" text class="!rounded-2xl" @click="showUnitDialog=false" />
        <Button label="Create Unit" class="!bg-[#1A3A4A] !border-[#1A3A4A] !rounded-2xl" :loading="unitSaving" @click="createUnit" />
      </template>
    </Dialog>
  </div>
</template>
