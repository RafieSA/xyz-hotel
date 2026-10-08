<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Card from 'primevue/card'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Message from 'primevue/message'
import { Heart, Trash2, Bed, HeartOff, ArrowRight, Sparkles } from 'lucide-vue-next'
import { useToast } from 'primevue/usetoast'
import client from '../api/client'
import { useAuthStore } from '../stores/auth'

const toast = useToast()
const router = useRouter()
const auth = useAuthStore()
const wishlist = ref([])
const loading = ref(false)

const typeMeta = {
  1: { type: 'Standard', price: 350000, img: 'https://images.unsplash.com/photo-1566665797739-1674de7a421a?w=600&q=80&auto=format&fit=crop' },
  2: { type: 'Deluxe', price: 550000, img: 'https://images.unsplash.com/photo-1611892440504-42a792e24d32?w=600&q=80&auto=format&fit=crop' },
  3: { type: 'Family', price: 850000, img: 'https://images.unsplash.com/photo-1590490360182-c33d57733427?w=600&q=80&auto=format&fit=crop' },
  4: { type: 'Suite', price: 1250000, img: 'https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=600&q=80&auto=format&fit=crop' },
}

const fmt = (n) => new Intl.NumberFormat('id-ID').format(n)

function resolveRoom(item) {
  // backend may return {room_type_id, room_type:{...}} or flat
  if (item.room_type) return item.room_type
  if (item.roomType) return item.roomType
  const id = item.room_type_id || item.roomTypeId || item.id || item.type_id
  const meta = typeMeta[id] || { type: `Type #${id}`, price: item.price || 0, img: typeMeta[1].img }
  return { id, name: item.name || meta.type, type: item.type || meta.type, price: item.price || meta.price, capacity: item.capacity || 2, image: item.image || meta.img, facility: item.facility || '' }
}

async function fetchWishlist() {
  if (!auth.isAuthenticated) { wishlist.value = []; return }
  loading.value = true
  try {
    const { data } = await client.get('/api/wishlist')
    const list = data.data || data
    wishlist.value = Array.isArray(list) ? list : []
  } catch (e) {
    if (e?.response?.status !== 401) {
      toast.add({ severity: 'error', summary: 'Could not load wishlist', detail: e?.response?.data?.message || e.message, life: 3000 })
    }
    wishlist.value = []
  } finally { loading.value = false }
}

async function removeWish(item) {
  const roomTypeId = item.room_type_id || item.roomTypeId || item.id || resolveRoom(item).id
  try {
    await client.post('/api/wishlist/toggle', { room_type_id: roomTypeId })
    toast.add({ severity: 'success', summary: 'Removed from wishlist', detail: 'Saved room removed', life: 2000 })
    await fetchWishlist()
  } catch (e) {
    // fallback try DELETE
    try {
      await client.delete(`/api/wishlist/${roomTypeId}`)
      toast.add({ severity: 'success', summary: 'Removed from wishlist', life: 2000 })
      await fetchWishlist()
    } catch (e2) {
      toast.add({ severity: 'error', summary: 'Could not remove', detail: e?.response?.data?.message || e.message, life: 3000 })
    }
  }
}

function goBook(roomId) {
  router.push('/')
  setTimeout(() => {
    const el = document.getElementById('rooms')
    if (el) el.scrollIntoView({ behavior: 'smooth' })
  }, 300)
}

onMounted(fetchWishlist)
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 py-10 space-y-8">
    <!-- Header hierarchy WarmAura -->
    <div class="space-y-3">
      <div class="flex items-center gap-3">
        <span class="bg-[#8B5A2B] text-white rounded-xl p-2.5 shadow-sm"><Heart class="w-5 h-5 fill-white" /></span>
        <div>
          <h1 class="font-display font-bold text-3xl text-[#1A3A4A] tracking-tight" style="font-family:'Playfair Display',serif">Wishlist</h1>
          <p class="text-sm text-[#6B7280] mt-1">Rooms you saved. Tap the heart again on the home page to toggle.</p>
        </div>
        <Button label="Browse rooms" icon="pi pi-arrow-right" class="!ml-auto !bg-[#8B5A2B] !border-[#8B5A2B] !rounded-xl" @click="router.push('/')" />
      </div>
      <div class="h-px bg-gradient-to-r from-[#C9A86A] via-[#C9A86A]/40 to-transparent"></div>
    </div>

    <Message v-if="!auth.isAuthenticated" severity="warn" class="rounded-xl">Sign in to save and view your wishlist.</Message>

    <div v-else>
      <div v-if="loading" class="text-center py-16 text-sm text-[#6B7280]">Loading your saved rooms...</div>

      <div v-else-if="!wishlist.length" class="text-center py-16 bg-white rounded-2xl border border-[#E5E7EB] shadow-sm">
        <div class="w-16 h-16 mx-auto rounded-full bg-[#FDF6EC] border border-[#E5E7EB] flex items-center justify-center"><HeartOff class="w-7 h-7 text-[#8B5A2B]" /></div>
        <h3 class="font-display font-semibold text-lg text-[#1A3A4A] mt-4" style="font-family:'Playfair Display',serif">No saved rooms yet</h3>
        <p class="text-sm text-[#6B7280] mt-1 max-w-md mx-auto">Tap the heart on any room card at the home page. Your favorites stay here for quick booking.</p>
        <Button label="Explore rooms" icon="pi pi-search" class="!mt-6 !bg-[#1A3A4A] !border-[#1A3A4A] !rounded-xl" @click="router.push('/')" />
      </div>

      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        <Card v-for="(item, idx) in wishlist" :key="item.id || idx" class="overflow-hidden !rounded-xl !shadow-sm hover:!shadow-md transition-all !border !border-[#E5E7EB] group">
          <template #header>
            <div class="relative overflow-hidden">
              <img :src="resolveRoom(item).image || resolveRoom(item).img" :alt="resolveRoom(item).type" class="w-full aspect-[16/10] object-cover group-hover:scale-[1.02] transition-transform duration-300" />
              <span class="absolute top-3 left-3 bg-white/95 backdrop-blur text-[#1A3A4A] text-xs font-bold rounded-full px-2.5 py-1 shadow-sm flex items-center gap-1"><Sparkles class="w-3 h-3 text-[#C9A86A]" /> Saved</span>
              <button class="absolute top-3 right-3 w-9 h-9 rounded-full bg-white shadow flex items-center justify-center hover:scale-105 transition" @click="removeWish(item)" aria-label="Remove from wishlist">
                <Heart class="w-4 h-4 fill-[#8B5A2B] text-[#8B5A2B]" />
              </button>
            </div>
          </template>
          <template #title><span class="font-display font-semibold text-[#1A3A4A]">{{ resolveRoom(item).type || resolveRoom(item).name }}</span></template>
          <template #subtitle><span class="text-xs text-[#6B7280] flex items-center gap-1"><Bed class="w-3.5 h-3.5" /> Sleeps {{ resolveRoom(item).capacity || 2 }} · {{ resolveRoom(item).facility || 'WarmAura comfort' }}</span></template>
          <template #content>
            <p class="font-bold text-[#8B5A2B] text-lg">Rp {{ fmt(resolveRoom(item).price || 0) }} <span class="font-normal text-sm text-[#6B7280]">/ night</span></p>
            <Tag v-if="item.created_at" :value="new Date(item.created_at).toLocaleDateString('id-ID')" severity="secondary" rounded class="mt-2 text-xs" />
          </template>
          <template #footer>
            <div class="flex gap-2">
              <Button label="Remove" icon="pi pi-trash" outlined severity="danger" class="!rounded-xl flex-1" @click="removeWish(item)" />
              <Button label="Book now" class="!bg-[#8B5A2B] !border-[#8B5A2B] !rounded-xl flex-1" @click="goBook(resolveRoom(item).id)">
                <template #icon><ArrowRight class="w-4 h-4" /></template>
              </Button>
            </div>
          </template>
        </Card>
      </div>
    </div>
  </div>
</template>
