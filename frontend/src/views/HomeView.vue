<script setup>
import Card from 'primevue/card'
import Button from 'primevue/button'
import DatePicker from 'primevue/datepicker'
import Tag from 'primevue/tag'
import Rating from 'primevue/rating'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import { Bed, Users, Calendar, Star, MapPin, Wifi, Coffee, Waves, Heart, Search, Award, Utensils, Car, BedDouble, X, ChevronLeft, ChevronRight, Image as ImageIcon } from 'lucide-vue-next'
import { ref, watch, computed, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'
import client from '../api/client'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const toast = useToast()
const auth = useAuthStore()
const { t, locale } = useI18n()

const checkIn = ref(null)
const checkOut = ref(null)
const guests = ref(2)
const availMap = ref({})
const loadingAvail = ref(false)
const bookingLoading = ref('')
const availError = ref('')
const voucherCode = ref('')
const voucherValidating = ref(false)
const voucherInfo = ref(null)
const voucherError = ref('')

// Search
const searchQuery = ref('')
const searchDebounce = ref(null)
const filteredRoomsApi = ref(null)
const searching = ref(false)

// Rooms base static + api merge
const rooms = [
  { id: 1, type: 'Standard', price: 350000, cap: 2, facility: 'Smart TV · Breakfast', img: 'https://images.unsplash.com/photo-1566665797739-1674de7a421a?w=600&q=80&auto=format&fit=crop', rating: 4.6, icon: Bed },
  { id: 2, type: 'Deluxe', price: 550000, cap: 2, facility: 'Balkon · Mini fridge', img: 'https://images.unsplash.com/photo-1611892440504-42a792e24d32?w=600&q=80&auto=format&fit=crop', rating: 4.8, icon: Coffee },
  { id: 3, type: 'Family', price: 850000, cap: 4, facility: '2 Bedroom · Kitchen', img: 'https://images.unsplash.com/photo-1590490360182-c33d57733427?w=600&q=80&auto=format&fit=crop', rating: 4.9, icon: Users },
  { id: 4, type: 'Suite', price: 1250000, cap: 3, facility: 'Living room · Bathtub', img: 'https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=600&q=80&auto=format&fit=crop', rating: 5.0, icon: Waves },
]

const fallbackGallery = {
  1: [
    'https://images.unsplash.com/photo-1566665797739-1674de7a421a?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1560448204-e02f11c3d0e2?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1582719478250-c89cae4dc85b?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1571896349842-33c89424de2d?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1520250497591-112f2f40a3f4?w=800&q=80&auto=format&fit=crop',
  ],
  2: [
    'https://images.unsplash.com/photo-1611892440504-42a792e24d32?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1591088398332-8a7791972843?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1618773928121-c32242e63f39?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1566665797739-1674de7a421a?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1551882547-b79c417633b4?w=800&q=80&auto=format&fit=crop',
  ],
  3: [
    'https://images.unsplash.com/photo-1590490360182-c33d57733427?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1586023492125-27b2c045efd7?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1566669437688-88b0d1683f2a?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1631049307264-da0ec9d70304?w=800&q=80&auto=format&fit=crop',
  ],
  4: [
    'https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1582719478250-c89cae4dc85b?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1611892440504-42a792e24d32?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1566073771259-6a8506099945?w=800&q=80&auto=format&fit=crop',
    'https://images.unsplash.com/photo-1590490360182-c33d57733427?w=800&q=80&auto=format&fit=crop',
  ],
}

// Gallery state
const galleryImages = ref({}) // id -> [urls]
const lightboxOpen = ref(false)
const lightboxImages = ref([])
const lightboxIndex = ref(0)
const carouselIndex = ref({ 1:0, 2:0, 3:0, 4:0 })

function openLightbox(roomId, idx){
  const imgs = galleryImages.value[roomId] || fallbackGallery[roomId] || []
  lightboxImages.value = imgs
  lightboxIndex.value = idx
  lightboxOpen.value = true
}
function nextLightbox(){ lightboxIndex.value = (lightboxIndex.value+1)%lightboxImages.value.length }
function prevLightbox(){ lightboxIndex.value = (lightboxIndex.value-1+lightboxImages.value.length)%lightboxImages.value.length }
async function fetchGallery(){
  for(const r of rooms){
    try{
      const { data } = await client.get(`/api/room-types/${r.id}/images`)
      const list = data.data || data
      let urls=[]
      if(Array.isArray(list)) urls = list.map(i=> i.url || i.image_url || i.src || i).filter(Boolean)
      if(urls.length>=1){
        const padded = [...urls]
        while(padded.length<5) padded.push(fallbackGallery[r.id][padded.length%5])
        galleryImages.value[r.id]=padded.slice(0,5)
      } else {
        galleryImages.value[r.id]=fallbackGallery[r.id]
      }
    }catch{
      galleryImages.value[r.id]=fallbackGallery[r.id]
    }
  }
}

// Loyalty
const loyaltyPoints = ref(0)
const useLoyalty = ref(false)
async function fetchLoyalty(){
  if(!auth.isAuthenticated){ loyaltyPoints.value=0; return }
  try{
    const { data } = await client.get('/api/loyalty/points')
    const p = data.data || data
    loyaltyPoints.value = p.points ?? p.loyalty_points ?? p.total ?? 0
  }catch{
    try{
      const { data } = await client.get('/api/loyalty')
      const p = data.data || data
      loyaltyPoints.value = p.points ?? 0
    }catch{ loyaltyPoints.value = 0 }
  }
}
const loyaltyProgress = computed(()=> Math.min(loyaltyPoints.value,100))
const loyaltyCanRedeem = computed(()=> loyaltyPoints.value>=100)

// Add-ons
const addons = ref([])
const selectedAddons = ref(new Set())
async function fetchAddons(){
  try{
    const { data } = await client.get('/api/addons')
    const list = data.data || data
    if(Array.isArray(list) && list.length) addons.value=list.map(a=>({ id:a.id, name:a.name, price:Number(a.price||a.price_per_night||0), description:a.description||'' }))
    else throw new Error('empty')
  }catch{
    addons.value=[
      { id:1, name:'Breakfast', price:50000, description:'Daily breakfast for 2' },
      { id:2, name:'Airport transfer', price:150000, description:'Private car to Ngurah Rai' },
      { id:3, name:'Extra bed', price:100000, description:'Extra single bed' },
    ]
  }
}
function toggleAddon(id){
  const s=new Set(selectedAddons.value)
  if(s.has(id)) s.delete(id); else s.add(id)
  selectedAddons.value=s
}
const addonsTotal = computed(()=> {
  let sum=0
  selectedAddons.value.forEach(id=>{
    const a=addons.value.find(x=>x.id===id)
    if(a) sum+=a.price
  })
  return sum
})

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
function isWeekendDate(d){
  const day = d.getDay() // 0 Sun, 5 Fri, 6 Sat
  return day===5 || day===6
}
function weekendMultiplier(ci, co){
  if(!ci||!co) return 1
  // if any night falls on Fri or Sat, +20% on whole stay (simple)
  try{
    const start=new Date(ci); const end=new Date(co)
    for(let dt=new Date(start); dt<end; dt.setDate(dt.getDate()+1)){
      if(isWeekendDate(dt)) return 1.2
    }
  }catch{}
  return 1
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
    availError.value = 'Check-out must be after check-in'
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
        if(e?.response?.status===429){
          toast.add({ severity:'warn', summary:'Too many requests', detail:'Slow down, try again in a minute', life:3000 })
        }
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
    toast.add({ severity: 'info', summary: 'Availability updated', detail: `${Object.keys(availMap.value).length} room types checked`, life: 2000 })
  }
}

// Pricing helpers with weekend + voucher + loyalty + addons
function discountedPrice(room) {
  if (!voucherInfo.value) return room.price
  const disc = voucherInfo.value.discount_percent ?? voucherInfo.value.discount ?? 0
  return Math.round(room.price * (1 - disc/100))
}
function baseNightPrice(room){
  const ci=toISO(checkIn.value); const co=toISO(checkOut.value)
  const mult = canSearch.value ? weekendMultiplier(checkIn.value, checkOut.value) : 1
  let p = discountedPrice(room) * mult
  return Math.round(p)
}
function nightsCount() {
  if (!checkIn.value || !checkOut.value) return 1
  const ci = toISO(checkIn.value); const co = toISO(checkOut.value)
  if (!ci || !co) return 1
  const n = Math.ceil((new Date(co)-new Date(ci))/(1000*60*60*24))
  return n>0? n:1
}
function totalForRoom(room){
  const nights=nightsCount()
  let total = baseNightPrice(room)*nights + addonsTotal.value
  if(useLoyalty.value && loyaltyCanRedeem.value) total = Math.max(0, total-100000)
  return total
}

function availText(room) {
  const a = availMap.value[room.id]
  if (!a) return ''
  return `${a.available} left of ${a.total_units} units`
}
function availSeverity(room) {
  const a = availMap.value[room.id]
  if (!a) return 'secondary'
  if (a.available <= 0) return 'danger'
  if (a.available <= 2) return 'warn'
  return 'success'
}

async function onBooking(room) {
  if (!canSearch.value) {
    toast.add({ severity: 'warn', summary: 'Select your dates', detail: 'Choose check in and check out first', life: 2500 })
    return
  }
  if (!auth.isAuthenticated) {
    toast.add({ severity: 'warn', summary: 'Sign in to continue', detail: 'Sign in to reserve your room', life: 2500 })
    router.push('/login')
    return
  }
  const ci = toISO(checkIn.value)
  const co = toISO(checkOut.value)
  const avail = availMap.value[room.id]
  if (avail && avail.available <= 0) {
    toast.add({ severity: 'error', summary: 'Fully booked', detail: `${room.type} is fully booked for those dates`, life: 2500 })
    return
  }
  bookingLoading.value = room.type
  try {
    const payload = {
      room_type_id: room.id,
      check_in: ci,
      check_out: co,
      guests: guests.value,
    }
    if (voucherCode.value && voucherInfo.value) payload.voucher_code = voucherCode.value.trim()
    else if (voucherCode.value) payload.voucher_code = voucherCode.value.trim()
    if(selectedAddons.value.size) payload.addon_ids = Array.from(selectedAddons.value)
    if(useLoyalty.value && loyaltyCanRedeem.value) payload.use_loyalty = true
    const { data } = await client.post('/api/bookings', payload)
    const booking = data.data || data
    // if loyalty used, call redeem (if backend expects separate call)
    if(useLoyalty.value && loyaltyCanRedeem.value){
      try{ await client.post('/api/loyalty/redeem', { booking_id: booking.id, points:100 }) }catch{}
      useLoyalty.value=false
      await fetchLoyalty()
      window.dispatchEvent(new Event('loyalty:updated'))
    }
    toast.add({ severity: 'success', summary: 'Room reserved', detail: `Booking #${booking.id || ''} is pending payment. Total Rp ${fmt(booking.total_price || totalForRoom(room))}`, life: 4000 })
    await fetchAvailability()
    await fetchLoyalty()
  } catch (e) {
    if(e?.response?.status===429){
      toast.add({ severity:'warn', summary:'Too many requests', detail:'You are booking too fast, try again in a minute', life:3500 })
    } else {
      const msg = e?.response?.data?.message || e.message || 'Reservation failed'
      toast.add({ severity: 'error', summary: 'Reservation failed', detail: msg, life: 3500 })
    }
  } finally {
    bookingLoading.value = ''
  }
}

async function validateVoucher() {
  const code = voucherCode.value.trim()
  if (!code) { voucherError.value='Enter a voucher code'; return }
  if (!canSearch.value) { voucherError.value='Select dates so we can validate minimum nights'; }
  voucherValidating.value=true
  voucherError.value=''
  voucherInfo.value=null
  try {
    let res
    try {
      const ci = toISO(checkIn.value)
      const co = toISO(checkOut.value)
      const params = { code }
      if (ci) params.check_in = ci
      if (co) params.check_out = co
      if (ci && co) {
        const nights = Math.ceil((new Date(co)-new Date(ci))/(1000*60*60*24))
        if (nights>0) params.nights = nights
      }
      res = await client.get('/api/vouchers/validate', { params })
    } catch (e1) {
      if(e1?.response?.status===429) throw e1
      res = await client.post('/api/vouchers/validate', { code, check_in: toISO(checkIn.value), check_out: toISO(checkOut.value) })
    }
    const payload = res.data.data || res.data
    voucherInfo.value = payload
    toast.add({ severity:'success', summary:'Voucher applied', detail:`You save ${payload.discount_percent || payload.discount || ''}% on this stay`, life:2500 })
  } catch (e) {
    if(e?.response?.status===429){
      voucherError.value='Too many requests, try again in a minute'
      toast.add({ severity:'warn', summary:'Rate limited', detail:'Too many voucher checks, slow down', life:3000 })
    } else {
      const msg = e?.response?.data?.message || e.message || 'Voucher not valid'
      voucherError.value = msg
      toast.add({ severity:'error', summary:'Voucher not applied', detail:msg, life:3000 })
    }
  } finally { voucherValidating.value=false }
}

// Search ILIKE debounced
function onSearchInput(v){
  searchQuery.value=v
  if(searchDebounce.value) clearTimeout(searchDebounce.value)
  searchDebounce.value=setTimeout(async()=>{
    const q=v.trim()
    if(!q){ filteredRoomsApi.value=null; return }
    searching.value=true
    try{
      const { data } = await client.get('/api/room-types', { params:{ q } })
      const list = data.data || data
      if(Array.isArray(list)){
        filteredRoomsApi.value=list
      } else {
        filteredRoomsApi.value=null
      }
    }catch(e){
      if(e?.response?.status===429) toast.add({ severity:'warn', summary:'Too many requests', life:2000 })
      // fallback client filter
      filteredRoomsApi.value=null
    } finally{ searching.value=false }
  }, 400)
}
const displayedRooms = computed(()=>{
  const q=searchQuery.value.trim().toLowerCase()
  if(!q) {
    // if backend returned filtered list, use mapping
    if(filteredRoomsApi.value && filteredRoomsApi.value.length){
      // map api ids to rooms
      const ids=new Set(filteredRoomsApi.value.map(r=> r.id))
      return rooms.filter(r=> ids.has(r.id))
    }
    return rooms
  }
  // if we have api filtered, use it else client filter
  if(filteredRoomsApi.value!==null){
    const ids=new Set(filteredRoomsApi.value.map(r=> r.id))
    const filtered = rooms.filter(r=> ids.has(r.id))
    if(filtered.length) return filtered
    // if api returned empty, show empty state
    return []
  }
  return rooms.filter(r=> r.type.toLowerCase().includes(q) || r.facility.toLowerCase().includes(q))
})

// Reviews
const roomTypesApi = ref([])
const reviewsByType = ref({})
const myBookings = ref([])
const reviewForm = ref({ booking_id: null, rating: 0, comment: '' })
const submittingReview = ref(false)
async function fetchRoomTypes(){
  try{
    const { data } = await client.get('/api/room-types')
    const list = data.data || data
    roomTypesApi.value = Array.isArray(list)? list : []
  }catch{ roomTypesApi.value=[] }
}
async function fetchReviews(){
  try{
    const { data } = await client.get('/api/reviews')
    const list = data.data || data
    const arr = Array.isArray(list)? list : []
    const map={}
    arr.forEach(r=>{ const k=r.room_type_id||r.roomTypeId; if(k){ (map[k]=map[k]||[]).push(r) } })
    reviewsByType.value = map
  }catch{ reviewsByType.value={} }
}
async function fetchMyBookings(){
  if(!auth.isAuthenticated) return
  try{ const {data}=await client.get('/api/bookings'); const list=data.data||data; myBookings.value=Array.isArray(list)?list:[] }catch{}
}
function avgRating(roomId){
  const api = roomTypesApi.value.find(r=> r.id===roomId)
  if(api?.avg_rating != null) return Number(api.avg_rating)
  if(api?.avgRating != null) return Number(api.avgRating)
  const revs = reviewsByType.value[roomId]||[]
  if(!revs.length) return rooms.find(r=>r.id===roomId)?.rating || 0
  return (revs.reduce((s,r)=>s+Number(r.rating),0)/revs.length)
}
function ratingCount(roomId){ return (reviewsByType.value[roomId]||[]).length }
const eligibleBookings = computed(()=>{
  return myBookings.value.filter(b=> b.status==='checked_out')
})
async function submitReview(){
  if(!reviewForm.value.booking_id) { toast.add({severity:'warn', summary:'Select a booking', life:2000}); return }
  if(!reviewForm.value.rating) { toast.add({severity:'warn', summary:'Add a rating', life:2000}); return }
  submittingReview.value=true
  try{
    await client.post('/api/reviews', { booking_id: reviewForm.value.booking_id, rating: reviewForm.value.rating, comment: reviewForm.value.comment })
    toast.add({severity:'success', summary:'Review submitted', detail:'Thanks for sharing your stay', life:2500})
    reviewForm.value={ booking_id:null, rating:0, comment:''}
    await fetchReviews(); await fetchRoomTypes()
  }catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', detail:'Slow down', life:2500})
    else toast.add({severity:'error', summary:'Could not submit review', detail:e?.response?.data?.message||e.message, life:3500})
  }
  finally{ submittingReview.value=false }
}
const wishlistIds = ref(new Set())
const wishLoading = ref(null)
async function fetchWishlist(){
  if(!auth.isAuthenticated){ wishlistIds.value=new Set(); return }
  try{
    const { data } = await client.get('/api/wishlist')
    const list = data.data || data
    const arr = Array.isArray(list) ? list : []
    const ids = arr.map(i=> i.room_type_id || i.roomTypeId || i.id || i.room_type?.id).filter(Boolean)
    wishlistIds.value = new Set(ids)
  }catch{ wishlistIds.value=new Set() }
}
function isWished(roomId){ return wishlistIds.value.has(roomId) }
async function toggleWishlist(room){
  if(!auth.isAuthenticated){ toast.add({severity:'warn', summary:'Sign in to save', detail:'Create an account to use wishlist', life:2500}); router.push('/login'); return }
  const id = room.id
  const wasWished = isWished(id)
  wishLoading.value=id
  try{
    const { data } = await client.post('/api/wishlist/toggle', { room_type_id: id })
    const wished = data.data?.wished ?? data.wished ?? !wasWished
    const next = new Set(wishlistIds.value)
    if(wished) next.add(id); else next.delete(id)
    wishlistIds.value = next
    if(wished) toast.add({severity:'success', summary:'Added to wishlist', detail:`${room.type} saved`, life:2000})
    else toast.add({severity:'info', summary:'Removed from wishlist', detail:`${room.type} removed`, life:2000})
    if(typeof window !== 'undefined') window.dispatchEvent(new Event('wishlist:updated'))
  }catch(e){
    if(e?.response?.status===429) toast.add({severity:'warn', summary:'Too many requests', life:2000})
    else toast.add({severity:'error', summary:'Could not update wishlist', detail:e?.response?.data?.message || e.message, life:3000})
  }finally{ wishLoading.value=null }
}

// Map Leaflet interactive 4 pins
const mapRef = ref(null)
const mapFailed = ref(false)
let leafletMap=null
const selectedRoomId = ref(null)
function scrollToRoom(roomId){
  selectedRoomId.value = roomId
  const el = document.getElementById('room-card-' + roomId) || document.getElementById('rooms')
  if(el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  // highlight effect
  setTimeout(()=>{ selectedRoomId.value = null }, 2500)
}
function initMap(){
  const lat=-8.519, lng=115.263
  if(!mapRef.value) return
  if(typeof window==='undefined' || !window.L){
    mapFailed.value=true
    return
  }
  try{
    const L=window.L
    leafletMap = L.map(mapRef.value, { scrollWheelZoom: false }).setView([lat,lng], 13)
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', { attribution:'© OpenStreetMap' }).addTo(leafletMap)
    const pins = [
      { id:1, type:'Standard', price:350000, lat:-8.519, lng:115.263 },
      { id:2, type:'Deluxe', price:550000, lat:-8.521, lng:115.265 },
      { id:3, type:'Family', price:850000, lat:-8.517, lng:115.261 },
      { id:4, type:'Suite', price:1250000, lat:-8.523, lng:115.267 },
    ]
    pins.forEach(p=>{
      const pinIcon = L.divIcon({ className:'custom-pin', html:`<div style="background:#8B5A2B;width:32px;height:32px;border-radius:50% 50% 50% 0;transform:rotate(-45deg);display:flex;align-items:center;justify-content:center;box-shadow:0 4px 12px rgba(0,0,0,0.3);border:3px solid white"><span style="transform:rotate(45deg);color:white;font-size:14px;font-weight:bold;">${p.id}</span></div>`, iconSize:[32,32], iconAnchor:[16,32] })
      const popupHtml = `<div style="font-family:Inter,sans-serif;text-align:center;min-width:140px"><b style="color:#1A3A4A">${p.type}</b><br><span style="color:#8B5A2B;font-weight:700">Rp ${new Intl.NumberFormat('id-ID').format(p.price)} / night</span><br><button data-room="${p.id}" class="reserve-btn" style="margin-top:8px;background:#8B5A2B;color:white;border:none;border-radius:9999px;padding:6px 14px;font-size:12px;font-weight:600;cursor:pointer">Reserve</button></div>`
      const marker = L.marker([p.lat,p.lng], { icon: pinIcon }).addTo(leafletMap)
      marker.bindPopup(popupHtml)
      marker.on('popupopen', ()=>{
        setTimeout(()=>{
          const btn = document.querySelector(`.reserve-btn[data-room="${p.id}"]`)
          if(btn) btn.addEventListener('click', ()=>{ leafletMap.closePopup(); scrollToRoom(p.id) })
        }, 100)
      })
    })
    // expose for fallback
    window._scrollToRoom = scrollToRoom
  }catch(e){
    mapFailed.value=true
  }
}
function loadLeaflet(){
  if(typeof document==='undefined') return
  if(window.L){ nextTick(initMap); return }
  // css already in index.html
  const existingCss = document.querySelector('link[href*="leaflet.css"]')
  if(!existingCss){
    const link=document.createElement('link')
    link.rel='stylesheet'
    link.href='https://unpkg.com/leaflet@1.9.4/dist/leaflet.css'
    document.head.appendChild(link)
  }
  const script=document.createElement('script')
  script.src='https://unpkg.com/leaflet@1.9.4/dist/leaflet.js'
  script.onload=()=> nextTick(initMap)
  script.onerror=()=> mapFailed.value=true
  document.head.appendChild(script)
  setTimeout(()=>{ if(!window.L) mapFailed.value=true }, 5000)
}

onMounted(()=>{
  fetchRoomTypes(); fetchReviews(); fetchMyBookings(); fetchWishlist(); fetchGallery(); fetchLoyalty(); fetchAddons(); loadLeaflet()
  // auto carousel rotation
  setInterval(()=>{
    rooms.forEach(r=>{
      const imgs=galleryImages.value[r.id]||fallbackGallery[r.id]
      if(imgs) carouselIndex.value[r.id]=(carouselIndex.value[r.id]+1)%imgs.length
    })
  }, 4000)
})
watch(()=>auth.isAuthenticated, ()=>{ fetchWishlist(); fetchLoyalty(); fetchMyBookings() })

// handle keyboard lightbox
function onKey(e){
  if(!lightboxOpen.value) return
  if(e.key==='Escape') lightboxOpen.value=false
  if(e.key==='ArrowRight') nextLightbox()
  if(e.key==='ArrowLeft') prevLightbox()
}
onMounted(()=> window.addEventListener('keydown', onKey))
onBeforeUnmount(()=> window.removeEventListener('keydown', onKey))

</script>

<template>
  <div class="bg-[#FDF6EC] dark:bg-[#1A3A4A] transition-colors duration-200">
    <!-- Hero -->
    <section class="relative overflow-hidden bg-[#1A3A4A]">
      <img src="https://images.unsplash.com/photo-1566073771259-6a8506099945?w=1400&q=80&auto=format&fit=crop" alt="Hotel lobby warm" class="absolute inset-0 w-full h-full object-cover opacity-50" />
      <div class="absolute inset-0 bg-gradient-to-t from-[#1A3A4A]/80 via-[#1A3A4A]/30 to-transparent"></div>
      <div class="relative max-w-7xl mx-auto px-4 sm:px-6 py-16 md:py-24 text-center text-white">
        <p class="inline-flex items-center gap-2 bg-white/15 backdrop-blur rounded-full px-4 py-1.5 text-xs tracking-widest uppercase"><MapPin class="w-3.5 h-3.5 text-[#C9A86A]" /> {{ t('hero.badge') }}</p>
        <h1 class="font-display font-bold text-4xl md:text-5xl leading-tight mt-4" style="font-family:'Playfair Display',serif">{{ t('hero.title') }}<br /><span class="text-[#C9A86A]">{{ t('hero.subtitle') }}</span></h1>
        <p class="mt-4 text-white/80 max-w-2xl mx-auto text-base">{{ t('hero.desc') }}</p>

        <!-- Search box -->
        <div class="mt-8 bg-white dark:bg-[#2D3748] dark:border dark:border-[#4A5568] rounded-2xl shadow-lg p-4 md:p-5 flex flex-col md:flex-row gap-3 items-stretch md:items-end text-left max-w-4xl mx-auto">
          <div class="flex-1">
            <label class="text-xs font-semibold text-[#6B7280] dark:text-[#C9A86A] uppercase tracking-wide flex items-center gap-1.5"><Calendar class="w-3.5 h-3.5" /> {{ t('search.checkin') }}</label>
            <DatePicker v-model="checkIn" :placeholder="t('search.selectDate')" dateFormat="dd/mm/yy" showIcon class="w-full mt-1" />
          </div>
          <div class="flex-1">
            <label class="text-xs font-semibold text-[#6B7280] dark:text-[#C9A86A] uppercase tracking-wide flex items-center gap-1.5"><Calendar class="w-3.5 h-3.5" /> {{ t('search.checkout') }}</label>
            <DatePicker v-model="checkOut" :placeholder="t('search.selectDate')" dateFormat="dd/mm/yy" showIcon class="w-full mt-1" />
          </div>
          <div class="w-full md:w-36">
            <label class="text-xs font-semibold text-[#6B7280] dark:text-[#C9A86A] uppercase tracking-wide flex items-center gap-1.5"><Users class="w-3.5 h-3.5" /> {{ t('search.guests') }}</label>
            <select v-model="guests" class="mt-1 w-full border border-[#E5E7EB] rounded-2xl px-3 py-2.5 text-sm text-[#1F2937] bg-white focus:outline-none focus:ring-2 focus:ring-[#8B5A2B]/30 focus:border-[#8B5A2B]">
              <option :value="1">{{ t('search.guest1') }}</option>
              <option :value="2">{{ t('search.guest2') }}</option>
              <option :value="3">{{ t('search.guest3') }}</option>
              <option :value="4">{{ t('search.guest4') }}</option>
            </select>
          </div>
          <Button :label="loadingAvail ? 'Checking...' : 'Check Availability'" :loading="loadingAvail" icon="pi pi-search" class="md:w-auto w-full !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-2xl !px-8 !py-3 font-semibold whitespace-nowrap" @click="onSearch" />
        </div>
        <!-- Voucher input -->
        <div class="mt-4 bg-white dark:bg-[#2D3748] dark:border dark:border-[#4A5568] rounded-2xl shadow p-4 flex flex-col sm:flex-row gap-3 items-stretch sm:items-end max-w-4xl mx-auto text-left">
          <div class="flex-1">
            <label class="text-xs font-semibold text-[#6B7280] dark:text-[#C9A86A] uppercase tracking-wide">{{ t('voucher.label') }}</label>
            <InputText v-model="voucherCode" placeholder="Enter code, e.g. HEMAT20" :placeholder="t('voucher.placeholder')" class="w-full mt-1 !rounded-xl" />
          </div>
          <Button :label="voucherValidating ? t('voucher.applied') : t('voucher.apply')" :loading="voucherValidating" icon="pi pi-ticket" class="!bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl !px-6 whitespace-nowrap" @click="validateVoucher" />
          <div v-if="voucherInfo" class="flex flex-col justify-center text-left sm:text-right">
            <span class="text-xs text-[#6B7280]">Discount</span>
            <span class="font-bold text-[#2E7D32] text-lg">{{ voucherInfo.discount_percent ?? voucherInfo.discount }}% OFF</span>
            <span class="text-xs text-[#6B7280]">Min {{ voucherInfo.min_nights }} nights. You selected {{ nightsCount() }} nights</span>
          </div>
        </div>
        <Message v-if="voucherError" severity="error" class="max-w-4xl mx-auto mt-2 text-left text-xs">{{ voucherError }}</Message>
        <Message v-if="voucherInfo" severity="success" class="max-w-4xl mx-auto mt-2 text-left text-xs">Voucher {{ voucherCode }} applied. Your room price drops {{ voucherInfo.discount_percent ?? voucherInfo.discount }}% at checkout.</Message>
        <Message v-if="availError" severity="error" class="max-w-4xl mx-auto mt-3 text-left">{{ availError }}</Message>
        <p class="mt-3 text-xs text-white/60">Free cancellation. Pay at the hotel. No hidden fees. Weekend Fri Sat +20%.</p>
      </div>
    </section>

    <!-- Search ILIKE -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 pt-10">
      <div class="bg-white dark:bg-[#2D3748] dark:border-[#4A5568] rounded-2xl shadow-sm border border-[#E5E7EB] p-4 flex flex-col sm:flex-row gap-3 items-center">
        <div class="relative flex-1 w-full">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#9CA3AF]" />
          <input :value="searchQuery" @input="onSearchInput($event.target.value)" placeholder="Search rooms: Standard, Deluxe, breakfast, balcony..." class="w-full pl-10 pr-4 py-2.5 rounded-xl border border-[#E5E7EB] text-sm focus:outline-none focus:ring-2 focus:ring-[#8B5A2B]/20 focus:border-[#8B5A2B] placeholder:text-[#9CA3AF]" />
        </div>
        <span v-if="searching" class="text-xs text-[#6B7280]">Searching...</span>
        <span v-else-if="searchQuery" class="text-xs text-[#6B7280]">{{ displayedRooms.length }} results for "{{ searchQuery }}"</span>
        <Button v-if="searchQuery" label="Clear" text size="small" class="!rounded-full" @click="onSearchInput('')" />
      </div>
    </section>

    <!-- Rooms -->
    <section id="rooms" class="max-w-7xl mx-auto px-4 sm:px-6 py-12 md:py-16">
      <div class="flex items-end justify-between gap-4">
        <div>
          <h2 class="font-display font-bold text-3xl text-[#1A3A4A] dark:text-[#FDF6EC]" style="font-family:'Playfair Display',serif; font-size:28px">{{ t('filter.title') }}</h2>
          <div class="h-px w-16 bg-[#C9A86A] mt-2"></div>
          <p class="text-[#6B7280] dark:text-[#FDF6EC]/70 text-base mt-2">{{ t('filter.subtitle') }}</p>
        </div>
        <span class="hidden md:inline-flex items-center gap-1.5 text-xs bg-[#FDF6EC] text-[#8B5A2B] border border-[#8B5A2B]/20 rounded-full px-3 py-1.5 font-semibold"><Wifi class="w-3.5 h-3.5" /> Free Wi-Fi · Breakfast</span>
      </div>

      <div v-if="canSearch && loadingAvail" class="mt-6 text-center text-sm text-[#6B7280]">Checking availability...</div>
      <div v-if="displayedRooms.length===0" class="mt-8 text-center py-12 bg-white rounded-2xl border border-dashed border-[#E5E7EB]">
        <p class="text-sm text-[#6B7280] dark:text-[#FDF6EC]">{{ t('search.noMatch', { query: searchQuery }) }}</p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mt-8">
        <Card v-for="r in displayedRooms" :key="r.type" :id="'room-card-' + r.id" :class="selectedRoomId===r.id ? 'ring-2 ring-[#8B5A2B] ring-offset-2 dark:ring-[#C9A86A]' : ''" class="overflow-hidden !rounded-2xl !shadow-sm hover:!shadow-md hover:-translate-y-0.5 transition-all duration-200 !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748]">
          <template #header>
            <div class="relative">
              <img :src="r.img" :alt="r.type + ' room'" class="w-full aspect-[16/10] object-cover" />
              <span class="absolute top-3 left-3 bg-white/95 backdrop-blur text-[#1A3A4A] text-xs font-bold rounded-full px-2.5 py-1 flex items-center gap-1 shadow-sm"><Star class="w-3.5 h-3.5 text-[#C9A86A] fill-[#C9A86A]" /> {{ avgRating(r.id).toFixed(1) }} <span v-if="ratingCount(r.id)" class="font-normal text-[#6B7280]">({{ ratingCount(r.id) }})</span></span>
              <span class="absolute top-3 right-3 bg-[#8B5A2B] text-white text-xs font-semibold rounded-full px-2.5 py-1">{{ r.type }}</span>
              <button @click="toggleWishlist(r)" :disabled="wishLoading===r.id" class="absolute bottom-3 right-3 w-9 h-9 rounded-full bg-white/95 backdrop-blur shadow-md flex items-center justify-center hover:scale-105 transition border border-white" :aria-label="isWished(r.id) ? 'Remove from wishlist' : 'Add to wishlist'">
                <Heart class="w-5 h-5 transition-colors" :class="isWished(r.id) ? 'fill-[#8B5A2B] text-[#8B5A2B]' : 'text-[#9CA3AF]'" />
              </button>
            </div>
          </template>
          <template #title><span class="text-[#1A3A4A] font-display font-semibold">{{ r.type }}</span></template>
          <template #subtitle><span class="text-xs text-[#6B7280] flex items-center gap-1.5"><Bed class="w-3.5 h-3.5" /> Sleeps {{ r.cap }} · {{ r.facility }}</span><div class="mt-1"><Rating :modelValue="Math.round(avgRating(r.id))" readonly :stars="5" class="!gap-0.5" /></div></template>
          <template #content>
            <div class="space-y-1">
              <p class="font-bold text-[#8B5A2B] text-lg leading-none">Rp {{ fmt(r.price) }} <span class="font-normal text-sm text-[#6B7280]">/ night</span></p>
              <p v-if="weekendMultiplier(checkIn && checkIn.value, checkOut && checkOut.value)>1" class="text-xs font-semibold text-[#8B5A2B]">Weekend rate +20% applied</p>
              <p v-if="voucherInfo" class="text-sm font-semibold text-[#2E7D32]">Rp {{ fmt(discountedPrice(r)) }} / night <span class="text-xs font-normal text-[#6B7280]">Save {{ voucherInfo.discount_percent ?? voucherInfo.discount }}%</span></p>
              <p v-if="canSearch" class="text-xs text-[#6B7280]">Base {{ nightsCount() }} nights: <span class="font-semibold text-[#1A3A4A]">Rp {{ fmt(baseNightPrice(r)*nightsCount()) }}</span> <span v-if="voucherInfo" class="line-through text-[#9CA3AF] ml-1">Rp {{ fmt(r.price*nightsCount()) }}</span></p>
              <p v-if="addonsTotal>0" class="text-xs text-[#6B7280]">Add-ons: <span class="font-semibold text-[#1A3A4A]">Rp {{ fmt(addonsTotal) }}</span></p>
              <p v-if="useLoyalty && loyaltyCanRedeem" class="text-xs font-semibold text-[#C9A86A]">Loyalty -Rp 100.000</p>
              <p v-if="canSearch" class="text-sm font-bold text-[#1A3A4A] dark:text-[#FDF6EC] border-t border-[#F3F4F6] dark:border-[#4A5568] pt-1 mt-1">Total: Rp {{ fmt(totalForRoom(r)) }}</p>
            </div>
            <div v-if="availMap[r.id]" class="mt-2">
              <Tag :value="availText(r)" :severity="availSeverity(r)" rounded class="text-xs" />
              <p class="text-xs text-[#6B7280] mt-1">Occupied: {{ availMap[r.id].occupied }} · Available: {{ availMap[r.id].available }}</p>
            </div>
            <p v-else-if="canSearch" class="text-xs text-[#9CA3AF] mt-2">Select dates to check availability</p>
          </template>
          <template #footer>
            <div class="flex gap-2 pt-1">
              <Button label="View Details" outlined class="!rounded-xl !text-[#8B5A2B] !border-[#8B5A2B] flex-1 !py-2 text-sm" @click="carouselIndex[r.id]=(carouselIndex[r.id]+1)%(galleryImages[r.id]?.length||5)" />
              <Button :label="bookingLoading === r.type ? 'Reserving...' : 'Reserve This Room'" :loading="bookingLoading === r.type" :disabled="availMap[r.id] && availMap[r.id].available <= 0" class="!bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl flex-1 !py-2 text-sm font-semibold disabled:!bg-gray-300 disabled:!border-gray-300" @click="onBooking(r)" />
            </div>
          </template>
        </Card>
      </div>
    </section>

    <!-- Gallery premium Our Spaces -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 py-12">
      <div class="text-center">
        <h2 class="font-bold text-[#1A3A4A]" style="font-family:'Playfair Display',serif; font-size:28px">Our Spaces</h2>
        <div class="h-px w-16 bg-[#C9A86A] mx-auto mt-3"></div>
        <p class="text-[#6B7280] text-base mt-3 max-w-2xl mx-auto">Five handpicked views per room. Warm light, woven textures, Ubud calm. Tap any image to enter lightbox.</p>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mt-10">
        <div v-for="r in rooms" :key="'gal-'+r.id" class="bg-white dark:bg-[#2D3748] rounded-2xl shadow-sm border border-[#E5E7EB] dark:border-[#4A5568] overflow-hidden hover:shadow-md transition-shadow">
          <div class="p-4 flex items-center justify-between">
            <h3 class="font-semibold text-[#1A3A4A] flex items-center gap-2"><component :is="r.icon" class="w-4 h-4 text-[#8B5A2B]" /> {{ r.type }}</h3>
            <span class="text-xs text-[#6B7280]">5 photos</span>
          </div>
          <div class="relative group overflow-hidden bg-[#FDF6EC]" style="aspect-ratio:16/10">
            <img
              :src="(galleryImages[r.id]||fallbackGallery[r.id])[carouselIndex[r.id]]"
              :alt="r.type + ' gallery'"
              class="w-full h-full object-cover cursor-zoom-in transition-transform duration-[4000ms] group-hover:scale-105"
              style="animation: kenburns 8s ease-in-out infinite alternate"
              @click="openLightbox(r.id, carouselIndex[r.id])"
              @error="$event.target.src=fallbackGallery[r.id][0]"
            />
            <div class="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent pointer-events-none"></div>
            <button @click="carouselIndex[r.id]=(carouselIndex[r.id]-1+(galleryImages[r.id]?.length||5))%(galleryImages[r.id]?.length||5)" class="absolute left-3 top-1/2 -translate-y-1/2 bg-white/90 hover:bg-white rounded-full p-2 shadow-md transition opacity-80 group-hover:opacity-100">
              <ChevronLeft class="w-4 h-4 text-[#1A3A4A]" />
            </button>
            <button @click="carouselIndex[r.id]=(carouselIndex[r.id]+1)%(galleryImages[r.id]?.length||5)" class="absolute right-3 top-1/2 -translate-y-1/2 bg-white/90 hover:bg-white rounded-full p-2 shadow-md transition opacity-80 group-hover:opacity-100">
              <ChevronRight class="w-4 h-4 text-[#1A3A4A]" />
            </button>
            <div class="absolute bottom-3 left-1/2 -translate-x-1/2 flex gap-1.5">
              <span v-for="(_, idx) in (galleryImages[r.id]||fallbackGallery[r.id])" :key="idx" class="w-2 h-2 rounded-full transition-colors" :style="{ background: idx===carouselIndex[r.id] ? '#8B5A2B' : '#E5E7EB', border: '1px solid rgba(255,255,255,0.8)' }"></span>
            </div>
          </div>
          <div class="grid grid-cols-5 gap-2 p-3 bg-[#FDF6EC]">
            <button v-for="(img, idx) in (galleryImages[r.id]||fallbackGallery[r.id])" :key="idx" @click="carouselIndex[r.id]=idx; openLightbox(r.id, idx)" class="relative rounded-xl overflow-hidden aspect-square border-2 transition" :class="idx===carouselIndex[r.id] ? 'border-[#8B5A2B] shadow-sm' : 'border-transparent hover:border-[#E5E7EB]'">
              <img :src="img" :alt="r.type+' thumb '+idx" class="w-full h-full object-cover" loading="lazy" @error="$event.target.src=fallbackGallery[r.id][idx%5]" />
            </button>
          </div>
          <p class="px-4 pb-4 text-xs text-[#6B7280] flex items-center gap-1"><ImageIcon class="w-3 h-3" /> {{ r.type }} gallery: tap to enlarge, swipe dots #8B5A2B</p>
        </div>
      </div>
    </section>

    <!-- Map & Nearby -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 py-12">
      <div class="text-center">
        <h2 class="font-bold text-[#1A3A4A] dark:text-[#FDF6EC]" style="font-family:'Playfair Display',serif; font-size:28px">{{ t('map.title') }}</h2>
        <div class="h-px w-16 bg-[#C9A86A] mx-auto mt-3"></div>
        <p class="text-[#6B7280] dark:text-[#C9A86A]/80 text-base mt-3">{{ t('map.desc') }}</p>
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mt-8">
        <div class="lg:col-span-2 bg-white rounded-2xl shadow-sm border border-[#E5E7EB] overflow-hidden">
          <div v-if="!mapFailed" ref="mapRef" id="map" class="w-full h-[400px] bg-[#FDF6EC] dark:bg-[#2D3748] rounded-2xl"></div>
          <div v-else class="w-full h-[400px] bg-[#FDF6EC] dark:bg-[#2D3748] flex flex-col items-center justify-center p-6 text-center">
            <img src="https://images.unsplash.com/photo-1520250497591-112f2f40a3f4?w=800&q=80&auto=format&fit=crop" alt="Map fallback Ubud" class="w-full h-48 object-cover rounded-2xl" />
            <p class="mt-4 font-semibold text-[#1A3A4A]">Jl. Hangat No. 8B, Ubud, Bali</p>
            <p class="text-sm text-[#6B7280]">-8.519, 115.263 · Open in Google Maps</p>
            <a href="https://www.openstreetmap.org/?mlat=-8.519&mlon=115.263#map=15/-8.519/115.263" target="_blank" class="mt-3 inline-flex items-center gap-1.5 bg-[#8B5A2B] text-white rounded-full px-4 py-2 text-sm font-semibold"><MapPin class="w-4 h-4" /> Open map</a>
            <a href="https://wa.me/6281234567890" class="mt-2 text-sm text-[#8B5A2B] underline">WhatsApp 0812-3456-7890</a>
          </div>
          <div class="p-4 flex items-center gap-2 text-xs text-[#6B7280] border-t border-[#E5E7EB]">
            <MapPin class="w-4 h-4 text-[#8B5A2B]" /> Ubud center · tiles OpenStreetMap · custom pin #8B5A2B
          </div>
        </div>
        <div class="space-y-4">
          <div class="bg-[#FDF6EC] dark:bg-[#2D3748] rounded-2xl border border-[#E5E7EB] dark:border-[#4A5568] p-5">
            <h3 class="font-semibold text-[#1A3A4A] flex items-center gap-2"><MapPin class="w-4 h-4 text-[#8B5A2B]" /> Nearby</h3>
            <div class="mt-4 space-y-3">
              <div class="bg-white rounded-2xl p-4 border border-[#E5E7EB] flex items-center justify-between">
                <div>
                  <p class="font-semibold text-[#1A3A4A] text-sm">Beach</p>
                  <p class="text-xs text-[#6B7280]">Sunset stroll</p>
                </div>
                <span class="bg-[#8B5A2B] text-white text-xs font-bold rounded-full px-3 py-1">0.5km</span>
              </div>
              <div class="bg-white rounded-2xl p-4 border border-[#E5E7EB] flex items-center justify-between">
                <div>
                  <p class="font-semibold text-[#1A3A4A] text-sm">Cafe</p>
                  <p class="text-xs text-[#6B7280]">Specialty coffee</p>
                </div>
                <span class="bg-[#8B5A2B] text-white text-xs font-bold rounded-full px-3 py-1">0.2km</span>
              </div>
              <div class="bg-white rounded-2xl p-4 border border-[#E5E7EB] flex items-center justify-between">
                <div>
                  <p class="font-semibold text-[#1A3A4A] text-sm">Spa</p>
                  <p class="text-xs text-[#6B7280]">Warm stone therapy</p>
                </div>
                <span class="bg-[#8B5A2B] text-white text-xs font-bold rounded-full px-3 py-1">0.3km</span>
              </div>
            </div>
            <p class="text-xs text-[#6B7280] mt-4">All distances walking. Cream cards #FDF6EC, gold hierarchy.</p>
          </div>
          <div class="bg-[#1A3A4A] rounded-2xl p-5 text-white">
            <p class="font-semibold">Getting here</p>
            <p class="text-sm text-white/70 mt-1">From Ngurah Rai Airport 45 min. Airport transfer 150k available at checkout.</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Loyalty card -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 py-8">
      <Card class="!rounded-2xl !border !border-[#C9A86A]/30 !shadow-sm bg-gradient-to-br from-white to-[#FDF6EC]">
        <template #content>
          <div class="flex flex-col md:flex-row gap-6 items-center">
            <div class="flex-1">
              <h3 class="font-bold text-[#1A3A4A] flex items-center gap-2" style="font-family:'Playfair Display',serif; font-size:18px"><Award class="w-5 h-5 text-[#C9A86A]" /> Loyalty Rewards</h3>
              <div class="h-px w-12 bg-[#C9A86A] mt-2"></div>
              <p class="text-sm text-[#6B7280] mt-2">Earn 10 points per night. 100 points = <span class="font-bold text-[#8B5A2B]">IDR 100k</span> off. Points after check out.</p>
              <div v-if="auth.isAuthenticated" class="mt-4">
                <div class="flex items-center justify-between text-xs">
                  <span class="font-semibold text-[#1A3A4A]">{{ loyaltyPoints }} points</span>
                  <span class="text-[#6B7280]">80 / 100 to free 100k</span>
                </div>
                <div class="mt-2 h-2.5 bg-[#E5E7EB] rounded-full overflow-hidden">
                  <div class="h-full bg-gradient-to-r from-[#C9A86A] to-[#8B5A2B] rounded-full transition-all" :style="{ width: Math.min(loyaltyPoints,100)+ '%' }"></div>
                </div>
                <div class="mt-2 h-2.5 bg-[#F3F4F6] rounded-full overflow-hidden hidden">
                  <div class="h-full bg-[#8B5A2B]" :style="{ width: (loyaltyProgress)+ '%' }"></div>
                </div>
                <p class="text-xs text-[#6B7280] mt-1">Progress 80/100 demo: stay 8 nights to redeem. Fill 100 to unlock.</p>
                <p v-if="loyaltyPoints>=100" class="text-xs font-semibold text-[#2E7D32] mt-1">You can redeem now! Tick the checkbox in booking.</p>
              </div>
              <p v-else class="text-xs text-[#6B7280] mt-3">Sign in to track points. Gold badge #C9A86A in navbar.</p>
            </div>
            <div class="w-full md:w-64 bg-white rounded-2xl border border-[#E5E7EB] p-4 shadow-sm">
              <p class="text-xs font-semibold text-[#6B7280] uppercase tracking-wide">Your progress</p>
              <p class="text-3xl font-bold text-[#1A3A4A] mt-1">{{ loyaltyPoints }} <span class="text-base font-normal text-[#6B7280]">/ 100</span></p>
              <div class="mt-3 h-2 bg-[#F3F4F6] rounded-full overflow-hidden"><div class="h-full bg-[#C9A86A] rounded-full" :style="{ width: Math.min(loyaltyPoints,100)+'%' }"></div></div>
              <p class="text-xs text-[#6B7280] mt-2">Example 80/100 to next reward</p>
            </div>
          </div>
        </template>
      </Card>
    </section>

    <!-- Add-ons booking -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 py-8">
      <div class="bg-white rounded-2xl shadow-sm border border-[#E5E7EB] p-6">
        <h3 class="font-bold text-[#1A3A4A]" style="font-family:'Playfair Display',serif; font-size:20px">Add-ons for your stay</h3>
        <div class="h-px w-12 bg-[#C9A86A] mt-2"></div>
        <p class="text-sm text-[#6B7280] mt-2">Pick extras before you reserve. Total updates live with voucher and loyalty.</p>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mt-6">
          <label v-for="a in addons" :key="a.id" class="flex gap-3 p-4 rounded-2xl border-2 cursor-pointer transition" :class="selectedAddons.has(a.id) ? 'border-[#8B5A2B] bg-[#FDF6EC]' : 'border-[#E5E7EB] hover:border-[#C9A86A]/40 bg-white'">
            <input type="checkbox" :checked="selectedAddons.has(a.id)" @change="toggleAddon(a.id)" class="mt-1 accent-[#8B5A2B] w-4 h-4" />
            <div class="flex-1">
              <p class="font-semibold text-[#1A3A4A] text-sm flex items-center gap-1.5"><Utensils v-if="a.name==='Breakfast'" class="w-3.5 h-3.5" /><Car v-if="a.name==='Airport transfer'" class="w-3.5 h-3.5" /><BedDouble v-if="a.name==='Extra bed'" class="w-3.5 h-3.5" /> {{ a.name }}</p>
              <p class="text-xs text-[#6B7280]">{{ a.description }}</p>
              <p class="text-sm font-bold text-[#8B5A2B] mt-1">+ Rp {{ fmt(a.price) }}</p>
            </div>
          </label>
        </div>
        <div class="mt-6 flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between bg-[#FDF6EC] rounded-2xl p-4 border border-[#E5E7EB]">
          <div class="flex items-center gap-2">
            <input type="checkbox" id="useLoyalty" v-model="useLoyalty" :disabled="!loyaltyCanRedeem" class="w-4 h-4 accent-[#8B5A2B]" />
            <label for="useLoyalty" class="text-sm" :class="!loyaltyCanRedeem ? 'text-[#9CA3AF]' : 'text-[#1A3A4A] font-semibold'">Apply loyalty 100 points = Rp 100.000 off</label>
            <span v-if="!loyaltyCanRedeem" class="text-xs text-[#9CA3AF]">(need 100, you have {{ loyaltyPoints }})</span>
          </div>
          <div class="text-right">
            <p class="text-xs text-[#6B7280]">Add-ons total</p>
            <p class="font-bold text-[#1A3A4A]">Rp {{ fmt(addonsTotal) }}</p>
            <p v-if="useLoyalty && loyaltyCanRedeem" class="text-xs text-[#C9A86A] font-semibold">Loyalty applied -100k</p>
          </div>
        </div>
        <p class="text-xs text-[#6B7280] mt-3">We send addon_ids and use_loyalty on booking. Weekend pricing already included in totals above.</p>
      </div>
    </section>

    <!-- Reviews per type -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 py-10">
      <h3 class="font-display font-bold text-2xl text-[#1A3A4A]" style="font-family:'Playfair Display',serif">Guest reviews</h3>
      <div class="h-px w-12 bg-[#C9A86A] mt-2"></div>
      <p class="text-base text-[#6B7280] mt-2">Average rating per room type. Real reviews from guests who completed their stay.</p>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mt-6">
        <Card v-for="r in rooms" :key="'rev-'+r.id" class="!rounded-2xl !border !border-[#E5E7EB] !shadow-sm">
          <template #title><span class="text-sm font-semibold text-[#1A3A4A]">{{ r.type }} · <span class="text-[#8B5A2B]">{{ avgRating(r.id).toFixed(1) }} <Star class="inline w-3 h-3 text-[#C9A86A] fill-[#C9A86A] -mt-0.5" /></span> <span class="text-xs text-[#6B7280]">({{ ratingCount(r.id) }} ulasan)</span></span></template>
          <template #content>
            <div v-if="(reviewsByType[r.id]||[]).length" class="space-y-3">
              <div v-for="rev in (reviewsByType[r.id]||[]).slice(0,3)" :key="rev.id" class="border-b border-[#F3F4F6] pb-2 last:border-0">
                <Rating :modelValue="rev.rating" readonly :stars="5" />
                <p class="text-sm text-[#1A3A4A] mt-1">{{ rev.comment || '-' }}</p>
                <p class="text-xs text-[#9CA3AF]">{{ rev.created_at ? new Date(rev.created_at).toLocaleDateString('en-GB') : '' }}</p>
              </div>
            </div>
            <p v-else class="text-xs text-[#9CA3AF]">No reviews for this room yet.</p>
          </template>
        </Card>
      </div>
      <!-- Review form -->
      <Card v-if="auth.isAuthenticated" class="mt-6 !rounded-2xl !border !border-[#E5E7EB] !shadow-sm max-w-2xl">
        <template #title><span class="text-base font-semibold text-[#1A3A4A]">Write a review</span></template>
        <template #subtitle><span class="text-xs text-[#6B7280]">Only for stays you have completed. One review per booking.</span></template>
        <template #content>
          <div class="space-y-3">
            <div>
              <label class="text-xs font-semibold text-[#6B7280] uppercase">Booking</label>
              <Select v-model="reviewForm.booking_id" :options="eligibleBookings" optionLabel="id" optionValue="id" placeholder="Select a completed booking" class="w-full mt-1" :emptyMessage="'No completed bookings'">
                <template #option="{ option }">#{{ option.id }} · Type {{ option.room_type_id }} · {{ option.check_in }} → {{ option.check_out }}</template>
                <template #value="{ value }"><span v-if="value">#{{ value }}</span><span v-else class="text-[#9CA3AF]">Select a booking</span></template>
              </Select>
            </div>
            <div>
              <label class="text-xs font-semibold text-[#6B7280] uppercase">Rating</label>
              <Rating v-model="reviewForm.rating" :stars="5" class="mt-1" />
            </div>
            <div>
              <label class="text-xs font-semibold text-[#6B7280] uppercase">Comment</label>
              <Textarea v-model="reviewForm.comment" rows="3" placeholder="How was your stay? What stood out?" class="w-full mt-1" autoResize />
            </div>
            <Button label="Submit Your Review" icon="pi pi-send" class="!bg-[#8B5A2B] !border-[#8B5A2B] !rounded-xl" :loading="submittingReview" @click="submitReview" />
          </div>
        </template>
      </Card>
      <p v-else class="text-sm text-[#6B7280] mt-4">Sign in to write a review after you check out.</p>
    </section>

    <section class="bg-[#FDF6EC] border-y border-[#E5E7EB]/60">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 flex flex-col md:flex-row gap-6 justify-between text-sm">
        <p class="text-[#1A3A4A] font-semibold">Need help? <span class="font-normal text-[#6B7280]">WhatsApp 0812-3456-7890. Check in 14:00, check out 12:00. One booking per room type.</span></p>
        <p class="text-[#6B7280]">Your price is locked at booking. Voucher optional.</p>
      </div>
    </section>

    <!-- Lightbox -->
    <div v-if="lightboxOpen" class="fixed inset-0 z-50 bg-black/85 flex items-center justify-center p-4" @click.self="lightboxOpen=false">
      <button @click="lightboxOpen=false" class="absolute top-4 right-4 bg-white/10 hover:bg-white/20 backdrop-blur rounded-full p-2 text-white"><X class="w-6 h-6" /></button>
      <button @click="prevLightbox" class="absolute left-4 top-1/2 -translate-y-1/2 bg-white/90 hover:bg-white rounded-full p-3 shadow-lg"><ChevronLeft class="w-5 h-5 text-[#1A3A4A]" /></button>
      <img :src="lightboxImages[lightboxIndex]" alt="Lightbox" class="max-w-full max-h-[85vh] rounded-2xl shadow-2xl object-contain" />
      <button @click="nextLightbox" class="absolute right-4 top-1/2 -translate-y-1/2 bg-white/90 hover:bg-white rounded-full p-3 shadow-lg"><ChevronRight class="w-5 h-5 text-[#1A3A4A]" /></button>
      <div class="absolute bottom-6 left-1/2 -translate-x-1/2 flex gap-2">
        <span v-for="(_, i) in lightboxImages" :key="i" class="w-2 h-2 rounded-full" :style="{ background: i===lightboxIndex ? '#C9A86A' : 'rgba(255,255,255,0.5)' }"></span>
      </div>
    </div>
  </div>
</template>

<style>
@keyframes kenburns { from { transform: scale(1) } to { transform: scale(1.08) } }
</style>
