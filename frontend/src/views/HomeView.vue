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
import RoomCard from '../components/RoomCard.vue'
import HeroSearch from '../components/HeroSearch.vue'
import ResortGallery from '../components/ResortGallery.vue'
import AddonsLoyaltySection from '../components/AddonsLoyaltySection.vue'
import GuestReviewsSection from '../components/GuestReviewsSection.vue'
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
watch(voucherError, v=>{ if(v){ voucherShake.value=true; const el=document.querySelector('.t-input-wrap .t-input'); if(el){ el.classList.remove('is-shaking'); void el.offsetWidth; el.classList.add('is-shaking'); setTimeout(()=> el.classList.remove('is-shaking'), 300) } setTimeout(()=> voucherShake.value=false, 3000) } })

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
const ourSpacesRef = ref(null)
const ourSpacesShown = ref(false)
let ourSpacesObserver = null
const lightboxOpen = ref(false)
const lightboxImages = ref([])
const lightboxIndex = ref(0)
const carouselIndex = ref({ 1:0, 2:0, 3:0, 4:0 })
const TILT_MAX = 10
function onCardTilt(e){
  if(window.matchMedia("(prefers-reduced-motion: reduce)").matches) return
  const wrap = e.currentTarget
  const card = wrap.querySelector(".t-tilt-card")
  if(!card) return
  const r = wrap.getBoundingClientRect()
  const px = Math.min(1, Math.max(0, (e.clientX - r.left)/r.width))
  const py = Math.min(1, Math.max(0, (e.clientY - r.top)/r.height))
  wrap.classList.add("is-hover")
  card.classList.add("is-tilting")
  card.style.setProperty("--tilt-ry", ((px-0.5)*TILT_MAX).toFixed(2)+"deg")
  card.style.setProperty("--tilt-rx", ((0.5-py)*TILT_MAX).toFixed(2)+"deg")
  card.style.setProperty("--tilt-gx", (px*100).toFixed(1)+"%")
  card.style.setProperty("--tilt-gy", (py*100).toFixed(1)+"%")
}
function onCardLeave(e){
  const wrap = e.currentTarget
  const card = wrap.querySelector(".t-tilt-card")
  wrap.classList.remove("is-hover")
  if(card){ card.classList.remove("is-tilting"); card.style.setProperty("--tilt-rx","0deg"); card.style.setProperty("--tilt-ry","0deg") }
}
const addonsOpen = ref(false)
const wishBurst = ref({})
const voucherShake = ref(false)

function openLightbox(roomId, idx){
  const imgs = galleryImages.value[roomId] || fallbackGallery[roomId] || []
  lightboxImages.value = imgs
  lightboxIndex.value = idx
  lightboxOpen.value = true
}
function nextLightbox(){ lightboxIndex.value = (lightboxIndex.value+1)%lightboxImages.value.length }
function burstLike(id){ wishBurst.value[id]=true; setTimeout(()=> wishBurst.value[id]=false, 650) }
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
})
watch(()=>auth.isAuthenticated, ()=>{ fetchWishlist(); fetchLoyalty(); fetchMyBookings() })

// handle keyboard lightbox
function onKey(e){
  if(!lightboxOpen.value) return
  if(e.key==='Escape') lightboxOpen.value=false
  if(e.key==='ArrowRight') nextLightbox()
  if(e.key==='ArrowLeft') prevLightbox()
}
onMounted(()=>{ window.addEventListener('keydown', onKey); nextTick(()=>{ if(ourSpacesRef.value){ ourSpacesShown.value = false; void ourSpacesRef.value.offsetHeight; requestAnimationFrame(()=>{ ourSpacesShown.value = true }); if("IntersectionObserver" in window){ ourSpacesObserver = new IntersectionObserver((entries)=>{ entries.forEach(ent=>{ if(ent.isIntersecting) ourSpacesShown.value = true }) }, { threshold: 0.2 }); ourSpacesObserver.observe(ourSpacesRef.value) } else { ourSpacesShown.value = true } } }) })
onBeforeUnmount(()=>{ window.removeEventListener('keydown', onKey); if(ourSpacesObserver) ourSpacesObserver.disconnect() })

</script>

<template>
  <div class="bg-[#FDF6EC] dark:bg-[#1A3A4A] transition-colors duration-200">
    <!-- Hero -->
    <section class="relative overflow-hidden bg-[#1A3A4A]">
      <img src="https://images.unsplash.com/photo-1566073771259-6a8506099945?w=1400&q=80&auto=format&fit=crop" alt="Hotel lobby warm" class="absolute inset-0 w-full h-full object-cover opacity-50" />
      <div class="absolute inset-0 bg-gradient-to-t from-[#1A3A4A]/80 via-[#1A3A4A]/30 to-transparent"></div>
      <div class="relative max-w-7xl mx-auto px-4 sm:px-6 py-16 md:py-24 text-center text-white">
        <p class="inline-flex items-center gap-2 bg-white/15 backdrop-blur rounded-full px-4 py-1.5 text-xs tracking-widest uppercase"><MapPin class="w-3.5 h-3.5 text-[#C9A86A]" /> {{ t('hero.badge') }}</p>
        <h1 class="font-display font-bold text-4xl md:text-5xl leading-tight mt-4" style="font-family:'Playfair Display',serif">{{ t('hero.title') }}<br /><span class="t-shimmer-text t-shimmer--hero text-[#C9A86A]" :data-text="t('hero.subtitle')">{{ t('hero.subtitle') }}</span></h1>
        <p class="mt-4 text-white/80 max-w-2xl mx-auto text-base">{{ t('hero.desc') }}</p>

        <!-- Integrated Floating Hero Search -->
        <div class="mt-8">
          <HeroSearch
            v-model:checkIn="checkIn"
            v-model:checkOut="checkOut"
            v-model:guests="guests"
            v-model:voucherCode="voucherCode"
            :loading-avail="loadingAvail"
            :voucher-validating="voucherValidating"
            :voucher-info="voucherInfo"
            :voucher-error="voucherError"
            :avail-error="availError"
            @search="onSearch"
            @validate-voucher="validateVoucher"
            @clear-voucher="voucherInfo = null; voucherCode = ''"
          />
        </div>
      </div>
    </section>

    <!-- Rooms -->
    <section id="rooms" class="max-w-7xl mx-auto px-4 sm:px-6 py-12 md:py-16 scroll-mt-20">
      <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <h2 class="font-display font-bold text-3xl text-[#1A3A4A] dark:text-[#FDF6EC]" style="font-family:'Playfair Display',serif; font-size:28px">{{ t('filter.title') }}</h2>
          <div class="h-px w-16 bg-[#C9A86A] mt-2"></div>
          <p class="text-[#6B7280] dark:text-[#FDF6EC]/70 text-base mt-2">{{ t('filter.subtitle') }}</p>
        </div>
        <div class="flex items-center gap-3">
          <span class="inline-flex items-center gap-1.5 text-xs bg-[#FDF6EC] dark:bg-[#2D3748] text-[#8B5A2B] border border-[#8B5A2B]/20 rounded-full px-3 py-1.5 font-semibold">
            <Wifi class="w-3.5 h-3.5" /> Free Wi-Fi · Breakfast
          </span>
        </div>
      </div>

      <!-- Integrated Keyword Filter Bar -->
      <div class="mt-6 bg-white dark:bg-[#2D3748] dark:border-[#4A5568] rounded-2xl shadow-sm border border-[#E5E7EB] p-3 sm:p-4 flex flex-col sm:flex-row gap-3 items-center">
        <div class="relative flex-1 w-full">
          <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-[#9CA3AF]" />
          <input
            :value="searchQuery"
            @input="onSearchInput($event.target.value)"
            placeholder="Cari spesifikasi: Standard, Deluxe, balkon, pemandangan kebun, sarapan..."
            class="w-full pl-10 pr-4 py-2 text-sm bg-transparent border-0 focus:outline-none focus:ring-0 placeholder:text-[#9CA3AF] text-[#1A3A4A] dark:text-white"
          />
        </div>
        <div class="flex items-center gap-2 shrink-0">
          <span v-if="searching" class="text-xs text-[#6B7280]">Mencari...</span>
          <span v-else-if="searchQuery" class="text-xs text-[#8B5A2B] font-semibold">{{ displayedRooms.length }} tipe ditemukan</span>
          <Button v-if="searchQuery" label="Reset" text size="small" class="!rounded-full !text-xs !py-1" @click="onSearchInput('')" />
        </div>
      </div>

      <div v-if="canSearch && loadingAvail" class="mt-6 text-center text-sm text-[#6B7280]">Memeriksa ketersediaan kamar...</div>
      <div v-if="displayedRooms.length===0" class="mt-8 text-center py-12 bg-white rounded-2xl border border-dashed border-[#E5E7EB]">
        <p class="text-sm text-[#6B7280] dark:text-[#FDF6EC]">{{ t('search.noMatch', { query: searchQuery }) }}</p>
      </div>

      <div v-if="!displayedRooms.length && !searchQuery" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mt-8">
        <div v-for="n in 4" :key="n" class="t-skel rounded-2xl border border-[#E5E7EB] bg-white overflow-hidden">
          <div class="t-skel-skeleton is-pulsing p-0">
            <div class="t-skel-bar w-full aspect-[16/10] !rounded-none !h-auto"></div>
            <div class="p-4 space-y-3"><div class="t-skel-bar w-3/4"></div><div class="t-skel-bar w-1/2"></div><div class="t-skel-bar w-full"></div></div>
          </div>
        </div>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mt-8">
        <RoomCard
          v-for="r in displayedRooms"
          :key="r.id"
          :room="r"
          :avail="availMap[r.id]"
          :loading-avail="loadingAvail"
          :is-wished="isWished(r.id)"
          :wish-loading="wishLoading === r.id"
          :booking-loading="bookingLoading === r.type"
          :nights="nightsCount()"
          :can-search="canSearch"
          :discount-percent="voucherInfo?.discount_percent ?? voucherInfo?.discount ?? 0"
          :weekend-multiplier="weekendMultiplier(checkIn, checkOut)"
          :avg-rating="avgRating(r.id)"
          :rating-count="ratingCount(r.id)"
          :is-selected="selectedRoomId === r.id"
          @toggle-wishlist="toggleWishlist"
          @view-details="(room) => { carouselIndex[room.id] = (carouselIndex[room.id] + 1) % (galleryImages[room.id]?.length || 5); openLightbox(room.id, carouselIndex[room.id]) }"
          @book="onBooking"
        />
      </div>
    </section>

    <!-- Resort Showcase Gallery -->
    <ResortGallery
      :rooms="rooms"
      :gallery-images="galleryImages"
      :fallback-gallery="fallbackGallery"
    />

    <!-- Map & Nearby -->
    <section id="location" class="max-w-7xl mx-auto px-4 sm:px-6 py-12 scroll-mt-20">
      <div class="text-center">
        <h2 class="font-bold text-[#1A3A4A] dark:text-[#FDF6EC]" style="font-family:'Playfair Display',serif; font-size:28px">{{ t('map.title') }}</h2>
        <div class="h-px w-16 bg-[#C9A86A] mx-auto mt-3"></div>
        <p class="text-[#6B7280] dark:text-[#C9A86A]/80 text-base mt-3">{{ t('map.desc') }}</p>
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mt-8">
        <div class="lg:col-span-2 bg-white dark:bg-[#2D3748] rounded-2xl shadow-sm border border-[#E5E7EB] dark:border-[#4A5568] overflow-hidden relative z-0 isolate">
          <div v-if="!mapFailed" ref="mapRef" id="map" class="w-full h-[400px] bg-[#FDF6EC] dark:bg-[#2D3748] rounded-2xl relative z-0"></div>
          <div v-else class="w-full h-[400px] bg-[#FDF6EC] dark:bg-[#2D3748] flex flex-col items-center justify-center p-6 text-center">
            <img src="https://images.unsplash.com/photo-1520250497591-112f2f40a3f4?w=800&q=80&auto=format&fit=crop" alt="Map fallback Ubud" class="w-full h-48 object-cover rounded-2xl" />
            <p class="mt-4 font-semibold text-[#1A3A4A] dark:text-white">Jl. Suweta No. 8B, Ubud, Gianyar, Bali</p>
            <p class="text-sm text-[#6B7280] dark:text-gray-400">-8.519, 115.263 · Buka di OpenStreetMap</p>
            <a href="https://www.openstreetmap.org/?mlat=-8.519&mlon=115.263#map=15/-8.519/115.263" target="_blank" class="mt-3 inline-flex items-center gap-1.5 bg-[#8B5A2B] text-white rounded-full px-4 py-2 text-sm font-semibold"><MapPin class="w-4 h-4" /> Buka peta</a>
            <a href="https://wa.me/6281234567890" class="mt-2 text-sm text-[#8B5A2B] underline">WhatsApp: +62 812-3456-7890</a>
          </div>
          <div class="p-4 flex items-center gap-2 text-xs text-[#6B7280] dark:text-gray-400 border-t border-[#E5E7EB] dark:border-[#4A5568]">
            <MapPin class="w-4 h-4 text-[#8B5A2B]" /> Pusat Wisata Ubud · Dekat Puri Saren Agung, Pasar Seni, & Monkey Forest
          </div>
        </div>
        <div class="space-y-4">
          <div class="bg-[#FDF6EC] dark:bg-[#2D3748] rounded-2xl border border-[#E5E7EB] dark:border-[#4A5568] p-5">
            <h3 class="font-semibold text-[#1A3A4A] dark:text-white flex items-center gap-2"><MapPin class="w-4 h-4 text-[#8B5A2B]" /> Destinasi Sekitar</h3>
            <div class="mt-4 space-y-3">
              <div class="bg-white dark:bg-[#1A3A4A] rounded-2xl p-4 border border-[#E5E7EB] dark:border-[#4A5568] flex items-center justify-between">
                <div>
                  <p class="font-semibold text-[#1A3A4A] dark:text-white text-sm">Hutan Monyet (Monkey Forest)</p>
                  <p class="text-xs text-[#6B7280] dark:text-gray-400">Jalan santai di cagar alam</p>
                </div>
                <span class="bg-[#8B5A2B] text-white text-xs font-bold rounded-full px-3 py-1">0.5km</span>
              </div>
              <div class="bg-white dark:bg-[#1A3A4A] rounded-2xl p-4 border border-[#E5E7EB] dark:border-[#4A5568] flex items-center justify-between">
                <div>
                  <p class="font-semibold text-[#1A3A4A] dark:text-white text-sm">Puri Saren Agung & Pasar Seni</p>
                  <p class="text-xs text-[#6B7280] dark:text-gray-400">Pusat seni & pertunjukan tari</p>
                </div>
                <span class="bg-[#8B5A2B] text-white text-xs font-bold rounded-full px-3 py-1">0.2km</span>
              </div>
              <div class="bg-white dark:bg-[#1A3A4A] rounded-2xl p-4 border border-[#E5E7EB] dark:border-[#4A5568] flex items-center justify-between">
                <div>
                  <p class="font-semibold text-[#1A3A4A] dark:text-white text-sm">Ubud Traditional Spa & Wellness</p>
                  <p class="text-xs text-[#6B7280] dark:text-gray-400">Pijat herbal & terapi relaksasi</p>
                </div>
                <span class="bg-[#8B5A2B] text-white text-xs font-bold rounded-full px-3 py-1">0.3km</span>
              </div>
            </div>
            <p class="text-xs text-[#6B7280] dark:text-gray-400 mt-4">Seluruh destinasi wisata favorit di atas dapat ditempuh dengan berjalan kaki santai dari resort.</p>
          </div>
          <div class="bg-[#1A3A4A] dark:bg-[#0F2A36] rounded-2xl p-5 text-white">
            <p class="font-semibold">Menuju ke Resort</p>
            <p class="text-sm text-white/70 mt-1">45 menit berkendara dari Bandara Internasional Ngurah Rai (DPS). Layanan jemputan privat tersedia saat pemesanan.</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Add-ons & Loyalty Rewards Section -->
    <div id="amenities" class="scroll-mt-20">
      <AddonsLoyaltySection
        :addons="addons"
        :selected-addons="selectedAddons"
        :is-authenticated="auth.isAuthenticated"
        :loyalty-points="loyaltyPoints"
        :use-loyalty="useLoyalty"
        :loyalty-can-redeem="loyaltyCanRedeem"
        :addons-total="addonsTotal"
        @toggle-addon="toggleAddon"
        @update:use-loyalty="useLoyalty = $event"
      />
    </div>

    <!-- Guest Reviews Section -->
    <div id="reviews" class="scroll-mt-20">
      <GuestReviewsSection
        :rooms="rooms"
        :reviews-by-type="reviewsByType"
        :avg-rating="avgRating"
        :rating-count="ratingCount"
        :is-authenticated="auth.isAuthenticated"
        :eligible-bookings="eligibleBookings"
        :review-form="reviewForm"
        :submitting-review="submittingReview"
        @submit-review="submitReview"
      />
    </div>

    <section class="bg-[#FDF6EC] border-y border-[#E5E7EB]/60">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 flex flex-col md:flex-row gap-6 justify-between text-sm">
        <p class="text-[#1A3A4A] font-semibold">Need help? <span class="font-normal text-[#6B7280]">WhatsApp 0812-3456-7890. Check in 14:00, check out 12:00. One booking per room type.</span></p>
        <p class="text-[#6B7280]">Your price is locked at booking. Voucher optional.</p>
      </div>
    </section>

  </div>
</template>

<style>
@keyframes kenburns { from { transform: scale(1) } to { transform: scale(1.08) } }
</style>
