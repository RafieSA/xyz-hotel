<script setup>
import Card from 'primevue/card'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import { LogIn } from 'lucide-vue-next'
import { ref } from 'vue'

const email = ref('')
const password = ref('')
</script>

<template>
  <div class="min-h-[70vh] flex items-center justify-center px-4 py-12 bg-[#F9FAFB]">
    <Card class="w-full max-w-md !rounded-2xl !shadow-md !border !border-[#E5E7EB]">
      <template #title><span class="font-display font-bold text-xl text-[#1A3A4A] flex items-center gap-2"><LogIn class="w-5 h-5 text-[#8B5A2B]" /> Masuk</span></template>
      <template #subtitle><span class="text-sm text-[#6B7280]">Karyawan & admin xyz-hotel</span></template>
      <template #content>
        <div class="space-y-4 mt-2">
          <div><label class="text-sm font-medium text-[#1F2937]">Email *</label><InputText v-model="email" placeholder="nama@xyz-hotel.id" class="w-full mt-1" /></div>
          <div><label class="text-sm font-medium text-[#1F2937]">Password *</label><InputText v-model="password" type="password" placeholder="••••••••" class="w-full mt-1" /></div>
          <Button label="Masuk" class="w-full !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl !py-3 font-semibold" />
          <p class="text-xs text-center text-[#6B7280]">Belum punya akun? Hubungi owner.</p>
        </div>
      </template>
    </Card>
  </div>
</template>
