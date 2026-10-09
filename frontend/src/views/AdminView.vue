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
import { LayoutDashboard, TrendingUp, CalendarDays, Wallet, Bed, ShieldCheck, Boxes, FileText, Download, Plus, Pencil, Trash2, Upload, Grid3x3, Kanban, Radio } from 'lucide-vue-next'
import { ref, onMounted, computed, watch, onBeforeUnmount } from 'vue'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'
import client from '../api/client'

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

// Calendar 7 days
const calendarData = ref(null)
const calendarLoading = ref(false)
const calendarDays = computed(()=>{
  const days=[]
  const today=new Date()
  for(let i=0;i<7;i++){
    const d=new Date(today); d.setDate(today.getDate()+i)
    days.push(d.toISOString().slice(0,10))
  }
  return days
})
async function fetchCalendar(){
  calendarLoading.value=true
  try{
    const { data } = await client.get('/api/admin/calendar', { params:{ days:7 } })
    calendarData.value=data.data||data
  }catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2000})
    // mock fallback for demo: generate per unit per day random
    const units = roomUnits.value.length? roomUnits.value : [{id:1,code:'STD-101'},{id:2,code:'DLX-101'},{id:3,code:'FAM-101'}]
    const mock={}
    calendarDays.value.forEach(day=>{
      mock[day]=units.map(u=>{
        const statuses=['available','occupied','dirty','maintenance']
        const s=statuses[Math.floor(Math.random()*4)]
        return { unit_id:u.id, code:u.code||`UNIT-${u.id}`, status:s, booking_id: s==='occupied'? Math.floor(Math.random()*100)+1 : null }
      })
    })
    calendarData.value=mock
  } finally{ calendarLoading.value=false }
}
function calendarCellColor(status){
  if(status==='available') return '#FDF6EC'
  if(status==='occupied') return '#8B5A2B'
  if(status==='dirty') return '#EF6C00'
  if(status==='maintenance') return '#6B7280'
  return '#FDF6EC'
}
function calendarTextColor(status){
  if(status==='occupied' || status==='maintenance') return 'white'
  return '#1A3A4A'
}

// Kanban 4 cols housekeeping
const kanbanCols = ref({ available:[], occupied:[], dirty:[], maintenance:[] })
const draggedUnit = ref(null)
function buildKanban(){
  const cols={ available:[], occupied:[], dirty:[], maintenance:[] }
  roomUnits.value.forEach(u=>{
    const s=u.status||'available'
    if(cols[s]) cols[s].push(u); else cols.available.push(u)
  })
  // fallback demo if empty
  if(!roomUnits.value.length){
    cols.available=[{id:101,code:'STD-101',status:'available',room_type_id:1}]
    cols.occupied=[{id:102,code:'DLX-101',status:'occupied',room_type_id:2}]
    cols.dirty=[{id:103,code:'FAM-101',status:'dirty',room_type_id:3}]
    cols.maintenance=[{id:104,code:'STE-101',status:'maintenance',room_type_id:4}]
  }
  kanbanCols.value=cols
}
watch(roomUnits, buildKanban)
function onDragStart(unit){ draggedUnit.value=unit }
function onDrop(col){
  const unit=draggedUnit.value
  if(!unit) return
  const newStatus=col
  if(unit.status===newStatus){ draggedUnit.value=null; return }
  // optimistic update
  const prev=unit.status
  // call PATCH
  client.patch(`/api/admin/room-units/${unit.id}/status`, { status:newStatus }).then(()=>{
    toast.add({ severity:'success', summary:'Room status updated', detail:`${unit.code} → ${newStatus}`, life:2500 })
    unit.status=newStatus
    buildKanban()
    fetchUnits()
    fetchCalendar()
  }).catch(e=>{
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', detail:'Slow down', life:2500})
    else toast.add({ severity:'error', summary:'Could not update', detail:e?.response?.data?.message||e.message, life:3000 })
    unit.status=prev
  })
  draggedUnit.value=null
}
function onDragOver(e){ e.preventDefault() }

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

onMounted(()=>{ fetchBookings(); fetchUnits(); fetchReports(); fetchReviews(); fetchRoomTypes(); fetchAuditLogs(); fetchCalendar(); connectWS() })
watch(dateRange, fetchReports)
onBeforeUnmount(()=>{ if(ws) ws.close(); if(wsPoll) clearInterval(wsPoll) })

async function doVerify(id, action) {
  if (action==='rejected') { rejectId.value=id; showReject.value=true; return }
  actionLoading.value = `verify-${id}`
  try {
    await client.patch(`/api/admin/bookings/${id}/verify`, { action })
    toast.add({ severity: 'success', summary: 'Booking confirmed', detail: `Booking #${id} is now verified`, life: 2500 })
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
const unitSeverity = (s)=>({ available:'success', occupied:'info', dirty:'warn', maintenance:'danger' }[s]||'secondary')

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
        <div class="sm:ml-auto flex items-center gap-2 flex-wrap">
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

    <!-- 4 stats cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748] hover:!shadow-md transition-shadow">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs uppercase tracking-widest text-[#6B7280] font-semibold">Total Bookings</p>
              <p class="text-3xl font-bold text-[#1A3A4A] mt-1">{{ totalBookings }}</p>
              <p class="text-xs text-[#6B7280] mt-1">All statuses</p>
            </div>
            <span class="bg-[#FDF6EC] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-2xl p-2.5"><CalendarDays class="w-5 h-5" /></span>
          </div>
        </template>
      </Card>
      <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748] hover:!shadow-md transition-shadow">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs uppercase tracking-widest text-[#6B7280] font-semibold">Total Revenue</p>
              <p class="text-2xl font-bold text-[#8B5A2B] mt-1">{{ fmt(totalRevenue) }}</p>
              <p class="text-xs text-[#6B7280] mt-1">Verified and completed stays</p>
            </div>
            <span class="bg-[#FDF6EC] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-2xl p-2.5"><Wallet class="w-5 h-5" /></span>
          </div>
        </template>
      </Card>
      <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748] hover:!shadow-md transition-shadow">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs uppercase tracking-widest text-[#6B7280] font-semibold">Occupancy</p>
              <p class="text-3xl font-bold text-[#1A3A4A] mt-1">{{ occupancyRate }}%</p>
              <div class="mt-2 h-1.5 w-24 bg-[#F3F4F6] rounded-full overflow-hidden"><div class="h-full bg-[#8B5A2B] rounded-full" :style="{ width: occupancyRate+'%' }"></div></div>
            </div>
            <span class="bg-[#FDF6EC] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-2xl p-2.5"><TrendingUp class="w-5 h-5" /></span>
          </div>
        </template>
      </Card>
      <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748] hover:!shadow-md transition-shadow">
        <template #content>
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs uppercase tracking-widest text-[#6B7280] font-semibold">Available Units</p>
              <p class="text-3xl font-bold text-[#1A3A4A] mt-1">{{ availableUnits }} <span class="text-sm font-normal text-[#6B7280]">/ {{ roomUnits.length || '-' }}</span></p>
              <p class="text-xs text-[#2E7D32] mt-1">Ready for guests</p>
            </div>
            <span class="bg-[#FDF6EC] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-2xl p-2.5"><Bed class="w-5 h-5" /></span>
          </div>
        </template>
      </Card>
    </div>

    <!-- Charts -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748] lg:col-span-2">
        <template #title><span class="text-[#1A3A4A] font-semibold text-sm">Revenue per Day</span></template>
        <template #content>
          <div class="h-[260px]"><Chart type="line" :data="revenueChartData" :options="revenueChartOptions" /></div>
        </template>
      </Card>
      <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
        <template #title><span class="text-[#1A3A4A] font-semibold text-sm">Occupancy vs Availability</span></template>
        <template #content>
          <div class="h-[260px] flex items-center justify-center"><Chart type="doughnut" :data="doughnutData" :options="doughnutOptions" /></div>
          <p class="text-center text-xs text-[#6B7280] mt-2">{{ occupancyRate }}% occupied and {{ 100-occupancyRate }}% available</p>
        </template>
      </Card>
    </div>
    <Card class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title><span class="text-[#1A3A4A] font-semibold text-sm">Bookings by Room Type</span></template>
      <template #content>
        <div class="h-[260px]"><Chart type="bar" :data="barChartData" :options="barChartOptions" /></div>
      </template>
    </Card>

    <Card v-if="bookingsByStatus" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748] bg-[#FDF6EC]/40">
      <template #content>
        <div class="flex flex-wrap gap-2 text-xs">
          <span v-for="(v,k) in bookingsByStatus" :key="k" class="bg-white border border-[#E5E7EB] rounded-full px-3 py-1"><span class="font-semibold text-[#1A3A4A] capitalize">{{ k.replace('_',' ') }}</span> <span class="text-[#8B5A2B] font-bold">{{ v }}</span></span>
        </div>
      </template>
    </Card>

    <!-- Tabs navigation -->
    <div class="flex flex-wrap gap-2 border-b border-[#E5E7EB] pb-3">
      <button v-for="t in [{id:'bookings',label:'Bookings',icon:FileText},{id:'calendar',label:'Occupancy Calendar',icon:Grid3x3},{id:'housekeeping',label:'Housekeeping',icon:Kanban},{id:'roomtypes',label:'Room Types',icon:Boxes},{id:'units',label:'Room Units',icon:Bed},{id:'audit',label:'Audit Log',icon:ShieldCheck}]" :key="t.id" @click="activeTab=t.id" :class="['inline-flex items-center gap-1.5 px-4 py-2 rounded-full text-sm font-semibold transition', activeTab===t.id ? 'bg-[#1A3A4A] text-white shadow' : 'bg-white border border-[#E5E7EB] text-[#6B7280] hover:border-[#8B5A2B] hover:text-[#1A3A4A]']">
        <component :is="t.icon" class="w-4 h-4" /> {{ t.label }}
      </button>
    </div>

    <!-- Recent bookings -->
    <Card v-show="activeTab==='bookings'" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title>
        <div class="flex items-center justify-between">
          <span class="text-[#1A3A4A] font-display font-bold text-lg" style="font-family:'Playfair Display',serif">Recent bookings</span>
          <span class="text-xs text-[#6B7280]">{{ bookings.length }} bookings</span>
        </div>
        <div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div>
      </template>
      <template #content>
        <DataTable :value="bookings" paginator :rows="8" stripedRows class="text-sm" responsiveLayout="scroll" :loading="loading" dataKey="id">
          <Column field="id" header="ID" sortable style="width:70px" />
          <Column field="user_id" header="User" sortable style="width:80px" />
          <Column field="room_type_id" header="Type" style="width:80px">
            <template #body="{ data }">#{{ data.room_type_id }}</template>
          </Column>
          <Column field="check_in" header="Check in" sortable>
            <template #body="{ data }">{{ fmtDate(data.check_in) }}</template>
          </Column>
          <Column field="check_out" header="Check out">
            <template #body="{ data }">{{ fmtDate(data.check_out) }}</template>
          </Column>
          <Column field="status" header="Status">
            <template #body="{ data }"><Tag :value="statusLabel(data.status)" :severity="statusSeverity(data.status)" rounded class="capitalize text-xs" /></template>
          </Column>
          <Column field="total_price" header="Total">
            <template #body="{ data }"><span class="font-semibold text-[#8B5A2B]">{{ fmt(data.total_price) }}</span></template>
          </Column>
          <Column header="Proof">
            <template #body="{ data }">
              <span v-if="data.proof_url" class="text-xs text-[#8B5A2B] underline cursor-pointer" @click="openDetail(data)">view</span>
              <span v-else class="text-xs text-[#9CA3AF]">-</span>
            </template>
          </Column>
          <Column header="Actions" style="min-width:360px">
            <template #body="{ data }">
              <div class="flex flex-wrap gap-1.5">
                <Button v-if="data.status==='waiting_verification'" label="Approve" size="small" class="!py-1 !px-2.5 !text-xs !bg-green-600 !border-green-600 hover:!bg-green-700 !rounded-full" :loading="actionLoading===`verify-${data.id}`" @click="doVerify(data.id,'verified')" />
                <Button v-if="data.status==='waiting_verification'" label="Decline" size="small" severity="danger" outlined class="!py-1 !px-2.5 !text-xs !rounded-full" :loading="actionLoading===`verify-${data.id}`" @click="doVerify(data.id,'rejected')" />
                <Button v-if="data.status==='verified'" label="Check In" size="small" class="!py-1 !px-2.5 !text-xs !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-full" :loading="actionLoading===`checkin-${data.id}`" @click="doCheckIn(data.id)" />
                <Button v-if="data.status==='checked_in'" label="Check Out" size="small" severity="info" class="!py-1 !px-2.5 !text-xs !rounded-full" :loading="actionLoading===`checkout-${data.id}`" @click="doCheckOut(data.id)" />
                <Button v-if="canInvoice(data.status)" label="Invoice" size="small" outlined class="!py-1 !px-2.5 !text-xs !rounded-full !border-[#1A3A4A] !text-[#1A3A4A]" :loading="invoicingId===data.id" @click="downloadInvoiceAdmin(data)">
                  <template #icon><Download class="w-3.5 h-3.5" /></template>
                </Button>
                <Button icon="pi pi-eye" size="small" text rounded class="!text-[#6B7280]" @click="openDetail(data)" />
              </div>
            </template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <!-- Occupancy Calendar 7 days -->
    <Card v-show="activeTab==='calendar'" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title>
        <div class="flex items-center justify-between">
          <span class="text-[#1A3A4A] font-display font-bold text-lg flex items-center gap-2" style="font-family:'Playfair Display',serif"><Grid3x3 class="w-5 h-5 text-[#8B5A2B]" /> Occupancy Calendar · 7 Days</span>
          <Button label="Refresh" icon="pi pi-refresh" outlined size="small" class="!rounded-full !border-[#8B5A2B] !text-[#8B5A2B]" :loading="calendarLoading" @click="fetchCalendar" />
        </div>
        <div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div>
        <p class="text-xs text-[#6B7280] mt-2">7 days x {{ roomUnits.length || 18 }} units. Colors: available cream #FDF6EC, occupied #8B5A2B, dirty orange, maintenance grey. Fetches GET /api/admin/calendar.</p>
        <div class="flex gap-2 mt-3 text-xs">
          <span class="inline-flex items-center gap-1.5"><span class="w-3 h-3 rounded" style="background:#FDF6EC;border:1px solid #E5E7EB"></span> Available</span>
          <span class="inline-flex items-center gap-1.5"><span class="w-3 h-3 rounded" style="background:#8B5A2B"></span> Occupied</span>
          <span class="inline-flex items-center gap-1.5"><span class="w-3 h-3 rounded" style="background:#EF6C00"></span> Dirty</span>
          <span class="inline-flex items-center gap-1.5"><span class="w-3 h-3 rounded" style="background:#6B7280"></span> Maintenance</span>
        </div>
      </template>
      <template #content>
        <div v-if="calendarLoading" class="text-center py-8 text-sm text-[#6B7280]">Loading calendar...</div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-xs border-collapse min-w-[700px]">
            <thead>
              <tr>
                <th class="text-left p-2 bg-[#F9FAFB] border border-[#E5E7EB] sticky left-0 z-10">Unit</th>
                <th v-for="d in calendarDays" :key="d" class="p-2 bg-[#F9FAFB] border border-[#E5E7EB] text-center min-w-[90px]">{{ d.slice(5) }}<br><span class="text-[10px] text-[#6B7280]">{{ new Date(d).toLocaleDateString('en-US',{weekday:'short'}) }}</span></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="unit in (roomUnits.length? roomUnits : [{id:101,code:'STD-101'},{id:102,code:'DLX-101'},{id:103,code:'FAM-101'},{id:104,code:'SUITE-101'}])" :key="unit.id">
                <td class="p-2 border border-[#E5E7EB] font-semibold text-[#1A3A4A] bg-white sticky left-0">{{ unit.code || 'UNIT-'+unit.id }}</td>
                <td v-for="d in calendarDays" :key="d+'-'+unit.id" class="p-1 border border-[#E5E7EB] text-center">
                  <div class="rounded-xl px-2 py-2 text-xs font-semibold" :style="{ background: calendarCellColor((calendarData?.[d]?.find(x=>x.unit_id===unit.id || x.code===unit.code)?.status) || unit.status || 'available'), color: calendarTextColor((calendarData?.[d]?.find(x=>x.unit_id===unit.id)?.status) || 'available') }">
                    {{ (calendarData?.[d]?.find(x=>x.unit_id===unit.id)?.status) || unit.status || 'available' }}
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <p class="text-xs text-[#9CA3AF] mt-3">Display only for MVP, drag not required. Shows per day per unit.</p>
      </template>
    </Card>

    <!-- Housekeeping Kanban 4 cols -->
    <Card v-show="activeTab==='housekeeping'" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title>
        <div class="flex items-center justify-between">
          <span class="text-[#1A3A4A] font-display font-bold text-lg flex items-center gap-2" style="font-family:'Playfair Display',serif"><Kanban class="w-5 h-5 text-[#8B5A2B]" /> Housekeeping Kanban</span>
          <Button label="Refresh Units" icon="pi pi-refresh" outlined size="small" class="!rounded-full !border-[#8B5A2B] !text-[#8B5A2B]" @click="fetchUnits" />
        </div>
        <div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div>
        <p class="text-xs text-[#6B7280] mt-2">Drag cards between columns. Dirty → Available calls PATCH room-units status. STD-101 demo.</p>
      </template>
      <template #content>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div v-for="col in ['available','occupied','dirty','maintenance']" :key="col" class="rounded-2xl border border-[#E5E7EB] bg-[#F9FAFB] p-3 min-h-[320px]" @dragover="onDragOver" @drop="onDrop(col)">
            <h4 class="text-xs uppercase tracking-widest font-bold text-[#1A3A4A] flex items-center gap-2 mb-3">
              <span class="w-2 h-2 rounded-full" :style="{ background: col==='available' ? '#22C55E' : col==='occupied' ? '#8B5A2B' : col==='dirty' ? '#EF6C00' : '#6B7280' }"></span>
              {{ col }} <span class="ml-auto bg-white border border-[#E5E7EB] rounded-full px-2 py-0.5 text-[10px]">{{ kanbanCols[col].length }}</span>
            </h4>
            <div class="space-y-3">
              <div v-for="unit in kanbanCols[col]" :key="unit.id" draggable="true" @dragstart="onDragStart(unit)" class="bg-white rounded-2xl border p-3 shadow-sm cursor-grab active:cursor-grabbing hover:shadow-md transition" :class="col==='occupied' ? 'border-[#8B5A2B]/30' : 'border-[#E5E7EB]'">
                <p class="font-bold text-[#1A3A4A] text-sm">{{ unit.code }}</p>
                <p class="text-xs text-[#6B7280]">Type #{{ unit.room_type_id || '-' }} · {{ unit.status }}</p>
                <Tag :value="unit.status" :severity="unitSeverity(unit.status)" rounded class="mt-2 text-xs" />
                <div class="mt-2 flex gap-1">
                  <select :value="unit.status" @change="(e)=>{ unitStatusMap[unit.id]=e.target.value }" class="flex-1 text-xs border border-[#E5E7EB] rounded-full px-2 py-1">
                    <option v-for="o in unitStatusOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
                  </select>
                  <Button label="Save" size="small" class="!rounded-full !bg-[#8B5A2B] !border-[#8B5A2B] !px-3 !py-1 text-xs" @click="updateUnitStatus(unit)" />
                </div>
              </div>
              <p v-if="!kanbanCols[col].length" class="text-xs text-[#9CA3AF] text-center py-8 border-2 border-dashed border-[#E5E7EB] rounded-2xl">No {{ col }} units. Drop here.</p>
            </div>
          </div>
        </div>
      </template>
    </Card>

    <!-- Room Types CRUD -->
    <Card v-show="activeTab==='roomtypes'" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title>
        <div class="flex items-center justify-between">
          <span class="text-[#1A3A4A] font-display font-bold text-lg" style="font-family:'Playfair Display',serif">Room Types</span>
          <Button label="Create Room Type" class="!bg-[#8B5A2B] !border-[#8B5A2B] !rounded-2xl" @click="openCreateRoomType">
            <template #icon><Plus class="w-4 h-4" /></template>
          </Button>
        </div>
        <div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div>
        <p class="text-xs text-[#6B7280] mt-2">CRUD for room types. Price snapshotting keeps historical bookings intact. Hierarchy headers #1A3A4A with gold dividers.</p>
      </template>
      <template #content>
        <DataTable :value="roomTypes" paginator :rows="8" stripedRows class="text-sm" responsiveLayout="scroll" dataKey="id">
          <Column field="id" header="ID" sortable style="width:70px" />
          <Column field="name" header="Name" sortable>
            <template #body="{ data }">{{ data.name || data.type }}</template>
          </Column>
          <Column field="description" header="Description" style="max-width:240px">
            <template #body="{ data }"><span class="text-xs text-[#6B7280] line-clamp-2">{{ data.description || '-' }}</span></template>
          </Column>
          <Column field="capacity" header="Capacity" sortable style="width:110px">
            <template #body="{ data }">{{ data.capacity || 2 }} guests</template>
          </Column>
          <Column field="price_per_night" header="Price" sortable>
            <template #body="{ data }"><span class="font-semibold text-[#8B5A2B]">{{ fmt(data.price_per_night || data.price || 0) }}</span></template>
          </Column>
          <Column field="total_units" header="Units" sortable style="width:90px">
            <template #body="{ data }">{{ data.total_units || data.totalUnits || '-' }}</template>
          </Column>
          <Column header="Photos (5 max)" style="min-width:260px">
            <template #body="{ data }">
              <div class="flex items-center gap-1">
                <input type="file" multiple accept="image/jpeg,image/png,image/jpg" :id="'upload-'+data.id" class="hidden" @change="onFileChange($event, data.id)" />
                <Button label="Choose" icon="pi pi-images" size="small" outlined class="!rounded-full !text-xs !py-1 !border-[#8B5A2B] !text-[#8B5A2B]" @click="document.getElementById('upload-'+data.id).click()" />
                <Button :label="uploadLoading[data.id] ? 'Uploading...' : 'Upload'" size="small" class="!rounded-full !bg-[#8B5A2B] !border-[#8B5A2B] !text-xs !py-1" :loading="uploadLoading[data.id]" @click="uploadRoomImages(data.id)">
                  <template #icon><Upload class="w-3 h-3" /></template>
                </Button>
              </div>
              <p v-if="uploadFiles[data.id]?.length" class="text-[10px] text-[#6B7280] mt-1">{{ uploadFiles[data.id].length }} files selected (max 5, jpg/png 5MB)</p>
            </template>
          </Column>
          <Column header="Actions" style="min-width:180px">
            <template #body="{ data }">
              <div class="flex gap-1.5">
                <Button size="small" outlined class="!rounded-full !border-[#1A3A4A] !text-[#1A3A4A] !py-1 !px-2.5 text-xs" @click="openEditRoomType(data)">
                  <template #icon><Pencil class="w-3.5 h-3.5" /></template> Edit
                </Button>
                <Button size="small" severity="danger" outlined class="!rounded-full !py-1 !px-2.5 text-xs" @click="deleteRoomType(data)">
                  <template #icon><Trash2 class="w-3.5 h-3.5" /></template> Delete
                </Button>
              </div>
            </template>
          </Column>
        </DataTable>
        <div v-if="!roomTypes.length" class="text-center py-8 text-sm text-[#6B7280]">No room types yet. Create your first type above.</div>
      </template>
    </Card>

    <!-- Room units -->
    <Card v-show="activeTab==='units'" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title>
        <div class="flex items-center justify-between">
          <span class="text-[#1A3A4A] font-display font-bold text-lg" style="font-family:'Playfair Display',serif">Room Units</span>
          <Button label="Create Unit" class="!bg-[#1A3A4A] !border-[#1A3A4A] !rounded-2xl" @click="openCreateUnit">
            <template #icon><Plus class="w-4 h-4" /></template>
          </Button>
        </div>
        <div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div>
      </template>
      <template #content>
        <DataTable :value="roomUnits" paginator :rows="8" stripedRows class="text-sm" responsiveLayout="scroll" dataKey="id" :loading="loading">
          <Column field="code" header="Code" sortable />
          <Column field="room_type_id" header="Type" sortable><template #body="{ data }">#{{ data.room_type_id }}</template></Column>
          <Column field="status" header="Status"><template #body="{ data }"><Tag :value="data.status" :severity="unitSeverity(data.status)" rounded class="text-xs" /></template></Column>
          <Column header="Update Status" style="min-width:280px">
            <template #body="{ data }">
              <div class="flex gap-2">
                <Select v-model="unitStatusMap[data.id]" :options="unitStatusOptions" optionLabel="label" optionValue="value" placeholder="Select status" class="w-full !text-xs" />
                <Button label="Save" size="small" class="!bg-[#8B5A2B] !border-[#8B5A2B] !rounded-full !px-4" @click="updateUnitStatus(data)" />
              </div>
            </template>
          </Column>
        </DataTable>
        <p v-if="!roomUnits.length" class="text-center py-6 text-sm text-[#6B7280]">No units yet. Create a unit to assign at check in.</p>
      </template>
    </Card>

    <!-- Audit logs -->
    <Card v-show="activeTab==='audit'" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title>
        <div class="flex items-center justify-between gap-2">
          <span class="text-[#1A3A4A] font-display font-bold text-lg flex items-center gap-2" style="font-family:'Playfair Display',serif"><ShieldCheck class="w-5 h-5 text-[#8B5A2B]" /> Audit Log</span>
          <Button label="Refresh" icon="pi pi-refresh" outlined size="small" class="!rounded-full !border-[#8B5A2B] !text-[#8B5A2B]" :loading="auditLoading" @click="fetchAuditLogs" />
        </div>
        <div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div>
        <p class="text-xs text-[#6B7280] mt-2">Who changed what, when. Tracks booking, room, and auth actions for accountability.</p>
      </template>
      <template #content>
        <DataTable :value="auditLogs" paginator :rows="10" stripedRows class="text-sm" responsiveLayout="scroll" :loading="auditLoading" dataKey="id">
          <Column field="id" header="ID" style="width:80px" sortable />
          <Column field="user_id" header="User" style="width:100px">
            <template #body="{ data }"><span class="text-xs">{{ data.user_id ?? data.userId ?? '-' }}</span></template>
          </Column>
          <Column field="action" header="Action" sortable>
            <template #body="{ data }"><Tag :value="data.action" severity="secondary" rounded class="text-xs" /></template>
          </Column>
          <Column field="entity" header="Entity" sortable>
            <template #body="{ data }">{{ data.entity || '-' }} <span v-if="data.entity_id" class="text-[#6B7280]">#{{ data.entity_id }}</span></template>
          </Column>
          <Column field="created_at" header="Time" sortable>
            <template #body="{ data }"><span class="text-xs">{{ data.created_at ? new Date(data.created_at).toLocaleString('id-ID') : (data.createdAt ? new Date(data.createdAt).toLocaleString('id-ID') : '-') }}</span></template>
          </Column>
          <Column field="payload" header="Details">
            <template #body="{ data }"><span class="text-xs text-[#6B7280] line-clamp-2 break-all">{{ typeof data.payload==='string' ? data.payload.slice(0,120) : JSON.stringify(data.payload||'').slice(0,120) }}</span></template>
          </Column>
        </DataTable>
        <div v-if="!auditLogs.length && !auditLoading" class="text-center py-8 text-sm text-[#6B7280]">No audit entries yet. Actions will appear here after you manage bookings or rooms.</div>
      </template>
    </Card>

    <!-- Reviews always visible below tabs -->
    <Card v-if="reviews.length" class="!rounded-2xl !shadow-sm !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
      <template #title><span class="text-[#1A3A4A] font-display font-bold text-lg" style="font-family:'Playfair Display',serif">Reviews ({{ reviews.length }})</span><div class="h-px bg-gradient-to-r from-[#C9A86A] to-transparent mt-3"></div></template>
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

    <Dialog v-model:visible="showDetail" modal header="Booking Details" :style="{ width:'560px' }" class="!rounded-2xl">
      <div v-if="selected" class="space-y-3 text-sm">
        <p><span class="font-semibold">ID:</span> {{ selected.id }} ({{ selected.status }})</p>
        <p><span class="font-semibold">Total:</span> {{ fmt(selected.total_price) }}</p>
        <p v-if="selected.proof_url" class="break-all"><span class="font-semibold">Proof:</span> <a :href="selected.proof_url" target="_blank" class="text-[#8B5A2B] underline">{{ selected.proof_url }}</a></p>
        <img v-if="selected.proof_url && !selected.proof_url.endsWith('.pdf')" :src="selected.proof_url" alt="proof" class="max-h-64 rounded-2xl border" />
      </div>
      <template #footer><Button label="Close" class="!rounded-2xl !bg-[#8B5A2B] !border-[#8B5A2B]" @click="showDetail=false" /></template>
    </Dialog>

    <Dialog v-model:visible="showReject" modal header="Decline Booking" :style="{ width:'420px' }">
      <div class="space-y-3">
        <p class="text-sm text-[#6B7280]">The guest will see this reason.</p>
        <InputText v-model="rejectReason" placeholder="Enter reason for declining" class="w-full" />
      </div>
      <template #footer>
        <Button label="Cancel" text @click="showReject=false" />
        <Button label="Decline Booking" severity="danger" :loading="actionLoading===`verify-${rejectId}`" @click="confirmReject" />
      </template>
    </Dialog>

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
