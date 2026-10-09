<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import Dialog from 'primevue/dialog'
import { useToast } from 'primevue/usetoast'
import { useI18n } from 'vue-i18n'
import client from '../api/client'
import { Upload, Eye, Calendar, Receipt, Download, XCircle, Heart, Trash2, Bed } from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'

const toast = useToast()
const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const bookings = ref([])
const loading = ref(false)
const uploadingId = ref(null)
const selectedFile = ref({})
const cancellingId = ref(null)
const invoicingId = ref(null)
const showCancel = ref(false)
const cancelTarget = ref(null)
const wishlist = ref([])
const wishlistLoading = ref(false)

const statusSeverity = (s) => ({ pending_payment: 'warn', waiting_verification: 'warn', verified: 'success', checked_in: 'info', checked_out: 'secondary', rejected: 'danger', expired: 'danger', cancelled: 'danger' }[s] || 'secondary')
const fmt = (n) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n)
const fmtDate = (d) => { try { return new Date(d).toLocaleDateString('id-ID') } catch { return d } }
const canCancel = (s) => ['pending_payment','waiting_verification'].includes(s)
const canInvoice = (s) => ['verified','checked_in','checked_out'].includes(s)

async function fetchBookings() {
  loading.value = true
  try {
    const { data } = await client.get('/api/bookings')
    const list = data.data || data
    bookings.value = Array.isArray(list) ? list : []
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Could not load bookings', detail: e?.response?.data?.message || e.message, life: 3000 })
  } finally { loading.value = false }
}

function onFileSelect(event, bookingId) {
  const file = event.files?.[0] || event.target?.files?.[0]
  if (file) selectedFile.value[bookingId] = file
}

async function uploadProof(booking) {
  const file = selectedFile.value[booking.id]
  if (!file) { toast.add({ severity: 'warn', summary: 'Select a file first', life: 2000 }); return }
  if (file.size > 5 * 1024 * 1024) { toast.add({ severity: 'error', summary: 'File too large', detail: 'Maximum size is 5MB', life: 2500 }); return }
  const ext = '.' + file.name.split('.').pop().toLowerCase()
  if (!['.jpg','.jpeg','.png','.pdf'].includes(ext)) { toast.add({ severity: 'error', summary: 'Invalid file type', detail: 'Use JPG, PNG, or PDF', life: 2500 }); return }
  uploadingId.value = booking.id
  try {
    const form = new FormData()
    form.append('proof', file)
    const { data } = await client.post(`/api/bookings/${booking.id}/proof`, form, { headers: { 'Content-Type': 'multipart/form-data' } })
    toast.add({ severity: 'success', summary: 'Proof uploaded', detail: data.message || 'We are now verifying your payment', life: 3000 })
    await fetchBookings()
    selectedFile.value[booking.id] = null
  } catch (e) {
    toast.add({ severity: 'error', summary: 'Upload failed', detail: e?.response?.data?.message || e.message, life: 3500 })
  } finally { uploadingId.value = null }
}

function confirmCancel(booking){ cancelTarget.value=booking; showCancel.value=true }
async function doCancel(){
  const b = cancelTarget.value
  if(!b) return
  cancellingId.value=b.id
  try{
    await client.patch(`/api/bookings/${b.id}/cancel`)
    toast.add({ severity:'success', summary:'Booking cancelled', detail:`Booking #${b.id} has been cancelled`, life:3000 })
    showCancel.value=false
    await fetchBookings()
  }catch(e){
    toast.add({ severity:'error', summary:'Could not cancel booking', detail:e?.response?.data?.message || e.message, life:3500 })
  }finally{ cancellingId.value=null }
}

async function downloadInvoice(booking){
  invoicingId.value=booking.id
  try{
    const res = await client.get(`/api/bookings/${booking.id}/invoice`, { responseType:'blob' })
    const blob = new Blob([res.data], { type: res.headers['content-type'] || 'application/pdf' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href=url
    const cd = res.headers['content-disposition'] || ''
    let filename = `invoice-${booking.id}.pdf`
    const m = cd.match(/filename="?([^"]+)"?/)
    if(m) filename=m[1]
    a.download=filename
    document.body.appendChild(a)
    a.click()
    a.remove()
    window.URL.revokeObjectURL(url)
    toast.add({ severity:'success', summary:'Invoice downloaded', detail:`Invoice for booking #${booking.id} is ready`, life:2500 })
  }catch(e){
    const msg = e?.response?.data?.message || e.message || 'Could not download invoice'
    // if blob error, try parse json
    if(e?.response?.data instanceof Blob){
      try{ const t=await e.response.data.text(); const j=JSON.parse(t); toast.add({severity:'error', summary:'Invoice not available', detail:j.message||msg, life:3500}); invoicingId.value=null; return }catch{}
    }
    toast.add({ severity:'error', summary:'Invoice not available', detail:msg, life:3500 })
  }finally{ invoicingId.value=null }
}

// Wishlist section inside MyBookings
const typeMeta = {
  1: { type: 'Standard', price: 350000, img: 'https://images.unsplash.com/photo-1566665797739-1674de7a421a?w=600&q=80&auto=format&fit=crop' },
  2: { type: 'Deluxe', price: 550000, img: 'https://images.unsplash.com/photo-1611892440504-42a792e24d32?w=600&q=80&auto=format&fit=crop' },
  3: { type: 'Family', price: 850000, img: 'https://images.unsplash.com/photo-1590490360182-c33d57733427?w=600&q=80&auto=format&fit=crop' },
  4: { type: 'Suite', price: 1250000, img: 'https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=600&q=80&auto=format&fit=crop' },
}
function resolveRoom(item){
  if(item.room_type) return item.room_type
  if(item.roomType) return item.roomType
  const id=item.room_type_id||item.roomTypeId||item.id
  const meta=typeMeta[id]||{type:`Type #${id}`, price:0, img:typeMeta[1].img}
  return { id, name: item.name||meta.type, type: item.type||meta.type, price:item.price||meta.price, img: item.image||meta.img }
}
async function fetchWishlist(){
  if(!auth.isAuthenticated) return
  wishlistLoading.value=true
  try{
    const { data } = await client.get('/api/wishlist')
    const list=data.data||data
    wishlist.value=Array.isArray(list)?list:[]
  }catch{ wishlist.value=[] }
  finally{ wishlistLoading.value=false }
}
async function removeWish(item){
  const roomTypeId=item.room_type_id||item.roomTypeId||item.id||resolveRoom(item).id
  try{
    await client.post('/api/wishlist/toggle', { room_type_id: roomTypeId })
    toast.add({ severity:'info', summary:'Removed from wishlist', life:2000 })
    await fetchWishlist()
    if(typeof window!=='undefined') window.dispatchEvent(new Event('wishlist:updated'))
  }catch(e){ toast.add({severity:'error', summary:'Could not remove', detail:e?.response?.data?.message||e.message, life:3000}) }
}

onMounted(()=>{ fetchBookings(); fetchWishlist() })
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 space-y-10">
    <!-- Header hierarchy -->
    <div class="space-y-3">
      <div class="flex items-center gap-3">
        <span class="bg-[#8B5A2B] text-white rounded-xl p-2.5 shadow-sm"><Receipt class="w-5 h-5" /></span>
        <div>
          <h1 class="font-display font-bold text-3xl text-[#1A3A4A] dark:text-[#FDF6EC] tracking-tight" style="font-family:'Playfair Display',serif">{{ t('bookings.title') }}</h1>
          <p class="text-sm text-[#6B7280] mt-1">{{ t('bookings.desc') }}</p>
        </div>
        <Button label="Refresh" icon="pi pi-refresh" outlined class="!rounded-xl !ml-auto !border-[#8B5A2B] !text-[#8B5A2B]" :loading="loading" @click="fetchBookings" />
      </div>
      <div class="h-px bg-gradient-to-r from-[#C9A86A] via-[#C9A86A]/40 to-transparent"></div>
    </div>

    <Message v-if="!bookings.length && !loading" severity="info" class="!rounded-xl">{{ t('bookings.noBookings') }}</Message>

    <div v-for="b in bookings" :key="b.id" class="border border-[#E5E7EB] dark:border-[#4A5568] rounded-2xl bg-white dark:bg-[#2D3748] shadow-sm p-5 md:p-6 space-y-4 hover:shadow-md transition-shadow">
      <div class="flex flex-wrap gap-3 items-start justify-between">
        <div>
          <p class="font-bold text-[#1A3A4A] flex items-center gap-2 flex-wrap">Booking #{{ b.id }} <Tag :value="b.status.replace('_',' ')" :severity="statusSeverity(b.status)" rounded class="capitalize text-xs" /></p>
          <p class="text-sm text-[#6B7280] flex items-center gap-1.5 mt-1"><Calendar class="w-4 h-4 text-[#8B5A2B]" /> {{ fmtDate(b.check_in) }} → {{ fmtDate(b.check_out) }} · {{ b.guests }} guests · Room Type #{{ b.room_type_id }}</p>
          <p class="text-sm font-semibold text-[#8B5A2B] mt-1">{{ fmt(b.total_price) }}</p>
          <p v-if="b.voucher_id" class="text-xs text-[#6B7280] mt-1">Voucher applied: #{{ b.voucher_id }}</p>
        </div>
        <div v-if="b.proof_url" class="text-right">
          <p class="text-xs text-[#6B7280] mb-1">Payment proof</p>
          <img v-if="b.proof_url.endsWith('.pdf')" src="https://cdn-icons-png.flaticon.com/512/337/337946.png" alt="pdf" class="w-20 h-20 object-contain border rounded-xl" />
          <img v-else :src="`http://localhost:8080/storage/uploads/${b.proof_url}`" :alt="'proof-'+b.id" class="w-28 h-20 object-cover rounded-xl border" @error="(e)=>e.target.style.display='none'" />
          <p class="text-xs text-[#6B7280] mt-1 break-all max-w-[160px]">{{ b.proof_url }}</p>
        </div>
      </div>

      <!-- Upload proof only when pending_payment -->
      <div v-if="b.status==='pending_payment'" class="bg-[#FDF6EC] dark:bg-[#1A3A4A] border border-[#E5E7EB] dark:border-[#4A5568] rounded-xl p-4 space-y-3">
        <p class="text-sm font-semibold text-[#1A3A4A] flex items-center gap-2"><Upload class="w-4 h-4 text-[#8B5A2B]" /> Upload payment proof (JPG, PNG, or PDF, max 5MB)</p>
        <div class="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
          <input type="file" accept=".jpg,.jpeg,.png,.pdf" class="text-sm file:mr-3 file:py-2 file:px-4 file:rounded-full file:border-0 file:bg-[#8B5A2B] file:text-white file:text-sm hover:file:bg-[#6F4620] file:cursor-pointer" @change="e=>onFileSelect({target:e}, b.id)" />
          <span v-if="selectedFile[b.id]" class="text-xs text-[#6B7280]">{{ selectedFile[b.id].name }} ({{ (selectedFile[b.id].size/1024).toFixed(1) }} KB)</span>
          <Button :label="uploadingId===b.id ? 'Uploading...' : 'Upload Proof Now'" icon="pi pi-upload" :loading="uploadingId===b.id" class="!bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl !ml-auto" @click="uploadProof(b)" />
        </div>
        <p class="text-xs text-[#6B7280]">Your file is stored securely and only visible to our team.</p>
      </div>

      <div v-else-if="b.status==='waiting_verification'" class="text-sm text-amber-800 bg-amber-50 border border-amber-200 rounded-xl px-3 py-2">Awaiting verification. We will confirm your payment shortly.</div>
      <div v-else-if="b.status==='verified'" class="text-sm text-green-800 bg-green-50 border border-green-200 rounded-xl px-3 py-2">Payment verified. See you at check in on your scheduled date.</div>
      <div v-else-if="b.status==='rejected'" class="text-sm text-red-800 bg-red-50 border border-red-200 rounded-xl px-3 py-2">Payment declined<span v-if="b.reject_reason">: {{ b.reject_reason }}</span></div>
      <div v-else-if="b.status==='checked_in'" class="text-sm text-blue-800 bg-blue-50 border border-blue-200 rounded-xl px-3 py-2 flex items-center gap-2"><Eye class="w-4 h-4" /> Checked in<span v-if="b.room_unit_id"> · Room unit #{{ b.room_unit_id }}</span></div>
      <div v-else-if="b.status==='checked_out'" class="text-sm text-gray-700 bg-gray-50 border border-gray-200 rounded-xl px-3 py-2">Checked out. Thanks for staying with us.</div>
      <div v-else-if="b.status==='cancelled'" class="text-sm text-red-800 bg-red-50 border border-red-200 rounded-xl px-3 py-2">Cancelled.</div>

      <!-- Actions row: Cancel + Invoice -->
      <div class="flex flex-wrap gap-2 pt-2 border-t border-[#F3F4F6]">
        <Button v-if="canCancel(b.status)" label="Cancel booking" icon="pi pi-times" severity="danger" outlined class="!rounded-xl !py-2 text-sm" :loading="cancellingId===b.id" @click="confirmCancel(b)">
          <template #icon><XCircle class="w-4 h-4" /></template>
        </Button>
        <Button v-if="canInvoice(b.status)" label="Download Invoice" class="!bg-[#1A3A4A] !border-[#1A3A4A] hover:!bg-[#0F2A36] !rounded-xl !py-2 text-sm" :loading="invoicingId===b.id" @click="downloadInvoice(b)">
          <template #icon><Download class="w-4 h-4" /></template>
        </Button>
      </div>
    </div>

    <!-- Wishlist section -->
    <section class="space-y-4">
      <div class="flex items-center gap-3">
        <span class="bg-[#FDF6EC] border border-[#8B5A2B]/15 text-[#8B5A2B] rounded-xl p-2"><Heart class="w-5 h-5 fill-[#8B5A2B]" /></span>
        <div>
          <h2 class="font-display font-bold text-xl text-[#1A3A4A]" style="font-family:'Playfair Display',serif">{{ t('bookings.savedRooms') }}</h2>
          <p class="text-sm text-[#6B7280]">Your wishlist from the home page. Remove or book directly.</p>
        </div>
        <Button label="View wishlist" outlined class="!ml-auto !rounded-xl !border-[#8B5A2B] !text-[#8B5A2B]" @click="router.push('/wishlist')" />
      </div>
      <div class="h-px bg-gradient-to-r from-[#C9A86A]/60 via-[#C9A86A]/30 to-transparent"></div>
      <div v-if="wishlistLoading" class="text-sm text-[#6B7280] py-6 text-center">Loading wishlist...</div>
      <div v-else-if="!wishlist.length" class="bg-white border border-dashed border-[#E5E7EB] rounded-2xl p-8 text-center">
        <Heart class="w-8 h-8 text-[#C9A86A] mx-auto" />
        <p class="text-sm font-semibold text-[#1A3A4A] mt-3">{{ t('bookings.noSaved') }}</p>
        <p class="text-xs text-[#6B7280] mt-1">Tap the heart on any room card to save it here.</p>
      </div>
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card v-for="(item, idx) in wishlist" :key="item.id || idx" class="!rounded-xl !border !border-[#E5E7EB] dark:!border-[#4A5568] dark:!bg-[#2D3748] !shadow-sm overflow-hidden">
          <template #header>
            <div class="relative">
              <img :src="resolveRoom(item).img" :alt="resolveRoom(item).type" class="w-full aspect-[16/10] object-cover" />
              <span class="absolute top-2 left-2 bg-white/90 text-[#1A3A4A] text-xs font-bold rounded-full px-2 py-1">{{ resolveRoom(item).type }}</span>
            </div>
          </template>
          <template #title><span class="text-sm font-semibold text-[#1A3A4A]">{{ resolveRoom(item).type }}</span></template>
          <template #content><p class="text-sm font-bold text-[#8B5A2B]">Rp {{ new Intl.NumberFormat('id-ID').format(resolveRoom(item).price) }} / night</p></template>
          <template #footer>
            <div class="flex gap-2">
              <Button label="Remove" severity="danger" outlined class="!rounded-xl flex-1 !py-1.5 text-xs" @click="removeWish(item)">
                <template #icon><Trash2 class="w-3.5 h-3.5" /></template>
              </Button>
              <Button label="Browse" class="!bg-[#8B5A2B] !border-[#8B5A2B] !rounded-xl flex-1 !py-1.5 text-xs" @click="router.push('/')" />
            </div>
          </template>
        </Card>
      </div>
    </section>

    <Dialog v-model:visible="showCancel" modal header="Cancel booking?" :style="{ width:'420px' }" class="!rounded-2xl">
      <p class="text-sm text-[#6B7280]">This will cancel booking <span class="font-semibold text-[#1A3A4A]">#{{ cancelTarget?.id }}</span>. This cannot be undone. Only pending bookings can be cancelled.</p>
      <template #footer>
        <Button label="Keep booking" text class="!rounded-xl" @click="showCancel=false" />
        <Button label="Yes, cancel" severity="danger" class="!rounded-xl" :loading="cancellingId===cancelTarget?.id" @click="doCancel" />
      </template>
    </Dialog>
  </div>
</template>
