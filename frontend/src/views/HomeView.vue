<script setup>
import Card from 'primevue/card'
import Button from 'primevue/button'
import DatePicker from 'primevue/datepicker'
import { Bed, Users, Calendar, Star, MapPin, Wifi, Coffee, Waves } from 'lucide-vue-next'
import { ref } from 'vue'

const checkIn = ref(null)
const checkOut = ref(null)
const guests = ref(2)

const rooms = [
  { type: 'Standard', price: 350000, cap: 2, facility: 'Smart TV · Breakfast', img: 'https://images.unsplash.com/photo-1566665797739-1674de7a421a?w=600&q=80&auto=format&fit=crop', rating: 4.6, icon: Bed },
  { type: 'Deluxe', price: 550000, cap: 2, facility: 'Balkon · Mini fridge', img: 'https://images.unsplash.com/photo-1611892440504-42a792e24d32?w=600&q=80&auto=format&fit=crop', rating: 4.8, icon: Coffee },
  { type: 'Family', price: 850000, cap: 4, facility: '2 Bedroom · Kitchen', img: 'https://images.unsplash.com/photo-1590490360182-c33d57733427?w=600&q=80&auto=format&fit=crop', rating: 4.9, icon: Users },
  { type: 'Suite', price: 1250000, cap: 3, facility: 'Living room · Bathtub', img: 'https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=600&q=80&auto=format&fit=crop', rating: 5.0, icon: Waves },
]

const fmt = (n) => new Intl.NumberFormat('id-ID').format(n)
</script>

<template>
  <div>
    <!-- Hero -->
    <section class="relative overflow-hidden bg-[#1A3A4A]">
      <img src="https://images.unsplash.com/photo-1566073771259-6a8506099945?w=1400&q=80&auto=format&fit=crop" alt="Hotel lobby warm" class="absolute inset-0 w-full h-full object-cover opacity-50" />
      <div class="absolute inset-0 bg-gradient-to-t from-[#1A3A4A]/80 via-[#1A3A4A]/30 to-transparent"></div>
      <div class="relative max-w-7xl mx-auto px-4 sm:px-6 py-16 md:py-24 text-center text-white">
        <p class="inline-flex items-center gap-2 bg-white/15 backdrop-blur rounded-full px-4 py-1.5 text-xs tracking-widest uppercase"><MapPin class="w-3.5 h-3.5 text-[#C9A86A]" /> Ubud · Bali · Since 2024</p>
        <h1 class="font-display font-bold text-4xl md:text-5xl leading-tight mt-4">Menginap hangat,<br /><span class="text-[#C9A86A]">seperti di rumah.</span></h1>
        <p class="mt-4 text-white/80 max-w-2xl mx-auto">4 tipe kamar curated — Standard hingga Suite. Harga transparan, booking 3 langkah, konfirmasi real-time.</p>

        <!-- Search box -->
        <div class="mt-8 bg-white rounded-2xl shadow-lg p-4 md:p-5 flex flex-col md:flex-row gap-3 items-stretch md:items-end text-left max-w-4xl mx-auto">
          <div class="flex-1">
            <label class="text-xs font-semibold text-[#6B7280] uppercase tracking-wide flex items-center gap-1.5"><Calendar class="w-3.5 h-3.5" /> Check-in</label>
            <DatePicker v-model="checkIn" placeholder="Pilih tanggal" dateFormat="dd/mm/yy" showIcon class="w-full mt-1" />
          </div>
          <div class="flex-1">
            <label class="text-xs font-semibold text-[#6B7280] uppercase tracking-wide flex items-center gap-1.5"><Calendar class="w-3.5 h-3.5" /> Check-out</label>
            <DatePicker v-model="checkOut" placeholder="Pilih tanggal" dateFormat="dd/mm/yy" showIcon class="w-full mt-1" />
          </div>
          <div class="w-full md:w-36">
            <label class="text-xs font-semibold text-[#6B7280] uppercase tracking-wide flex items-center gap-1.5"><Users class="w-3.5 h-3.5" /> Tamu</label>
            <select v-model="guests" class="mt-1 w-full border border-[#E5E7EB] rounded-xl px-3 py-2.5 text-sm text-[#1F2937] bg-white focus:outline-none focus:ring-2 focus:ring-[#8B5A2B]/30 focus:border-[#8B5A2B]">
              <option :value="1">1 Tamu</option>
              <option :value="2">2 Tamu</option>
              <option :value="3">3 Tamu</option>
              <option :value="4">4 Tamu</option>
            </select>
          </div>
          <Button label="Cari Kamar" icon="pi pi-search" class="md:w-auto w-full !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl !px-8 !py-3 font-semibold whitespace-nowrap" />
        </div>
        <p class="mt-3 text-xs text-white/60">Free cancellation · Bayar di hotel · No hidden fee</p>
      </div>
    </section>

    <!-- Rooms -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 py-12 md:py-16">
      <div class="flex items-end justify-between gap-4">
        <div>
          <h2 class="font-display font-bold text-2xl md:text-3xl text-[#1A3A4A]">Pilih tipe kamar</h2>
          <p class="text-[#6B7280] text-sm mt-1">4 tipe locked — harga per malam, sudah termasuk pajak.</p>
        </div>
        <span class="hidden md:inline-flex items-center gap-1.5 text-xs bg-[#FDF6EC] text-[#8B5A2B] border border-[#8B5A2B]/20 rounded-full px-3 py-1.5 font-semibold"><Wifi class="w-3.5 h-3.5" /> Free Wi-Fi · Breakfast</span>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mt-8">
        <Card v-for="r in rooms" :key="r.type" class="overflow-hidden !rounded-xl !shadow-sm hover:!shadow-md hover:-translate-y-0.5 transition-all duration-200 !border !border-[#E5E7EB]">
          <template #header>
            <div class="relative">
              <img :src="r.img" :alt="'Kamar ' + r.type" class="w-full aspect-[16/10] object-cover" />
              <span class="absolute top-3 left-3 bg-white/95 backdrop-blur text-[#1A3A4A] text-xs font-bold rounded-full px-2.5 py-1 flex items-center gap-1 shadow-sm"><Star class="w-3.5 h-3.5 text-[#C9A86A] fill-[#C9A86A]" /> {{ r.rating }}</span>
              <span class="absolute top-3 right-3 bg-[#8B5A2B] text-white text-xs font-semibold rounded-full px-2.5 py-1">{{ r.type }}</span>
            </div>
          </template>
          <template #title><span class="text-[#1A3A4A] font-display font-semibold">{{ r.type }}</span></template>
          <template #subtitle><span class="text-xs text-[#6B7280] flex items-center gap-1.5"><Bed class="w-3.5 h-3.5" /> Kapasitas {{ r.cap }} orang · {{ r.facility }}</span></template>
          <template #content>
            <p class="font-bold text-[#8B5A2B] text-lg leading-none">Rp {{ fmt(r.price) }} <span class="font-normal text-sm text-[#6B7280]">/ malam</span></p>
          </template>
          <template #footer>
            <div class="flex gap-2 pt-1">
              <Button label="Lihat Detail" outlined class="!rounded-xl !text-[#8B5A2B] !border-[#8B5A2B] flex-1 !py-2 text-sm" />
              <Button label="Booking" class="!bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl flex-1 !py-2 text-sm font-semibold" />
            </div>
          </template>
        </Card>
      </div>
    </section>

    <section class="bg-[#FDF6EC] border-y border-[#E5E7EB]/60">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 flex flex-col md:flex-row gap-6 justify-between text-sm">
        <span class="flex items-center gap-2 text-[#1A3A4A]"><span class="w-8 h-8 rounded-full bg-white border flex items-center justify-center"><Waves class="w-4 h-4 text-[#8B5A2B]" /></span> Kolam renang & spa</span>
        <span class="flex items-center gap-2 text-[#1A3A4A]"><span class="w-8 h-8 rounded-full bg-white border flex items-center justify-center"><Coffee class="w-4 h-4 text-[#8B5A2B]" /></span> Restaurant & café</span>
        <span class="flex items-center gap-2 text-[#1A3A4A]"><span class="w-8 h-8 rounded-full bg-white border flex items-center justify-center"><Users class="w-4 h-4 text-[#8B5A2B]" /></span> Family friendly</span>
        <span class="flex items-center gap-2 text-[#1A3A4A]"><span class="w-8 h-8 rounded-full bg-white border flex items-center justify-center"><MapPin class="w-4 h-4 text-[#8B5A2B]" /></span> 5 menit ke Monkey Forest</span>
      </div>
    </section>
  </div>
</template>
