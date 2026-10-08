<script setup>
import Card from 'primevue/card'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import { UserPlus } from 'lucide-vue-next'
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useToast } from 'primevue/usetoast'

const name = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const router = useRouter()
const auth = useAuthStore()
const toast = useToast()

async function onRegister() {
  error.value = ''
  if (!name.value || !email.value || !password.value) {
    error.value = 'Semua field wajib diisi'
    return
  }
  loading.value = true
  try {
    await auth.registerRequest(name.value, email.value, password.value)
    toast.add({ severity: 'success', summary: 'Berhasil', detail: 'Registrasi sukses — otomatis login', life: 2500 })
    router.push('/')
  } catch (e) {
    error.value = e?.response?.data?.message || e.message || 'Registrasi gagal'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-[70vh] flex items-center justify-center px-4 py-12 bg-[#F9FAFB]">
    <Card class="w-full max-w-md !rounded-2xl !shadow-md !border !border-[#E5E7EB]">
      <template #title><span class="font-display font-bold text-xl text-[#1A3A4A] flex items-center gap-2"><UserPlus class="w-5 h-5 text-[#8B5A2B]" /> Daftar</span></template>
      <template #subtitle><span class="text-sm text-[#6B7280]">Buat akun customer xyz-hotel</span></template>
      <template #content>
        <div class="space-y-4 mt-2">
          <Message v-if="error" severity="error" class="text-sm">{{ error }}</Message>
          <div><label class="text-sm font-medium text-[#1F2937]">Nama *</label><InputText v-model="name" placeholder="Nama lengkap" class="w-full mt-1" /></div>
          <div><label class="text-sm font-medium text-[#1F2937]">Email *</label><InputText v-model="email" placeholder="nama@email.com" class="w-full mt-1" /></div>
          <div><label class="text-sm font-medium text-[#1F2937]">Password *</label><InputText v-model="password" type="password" placeholder="••••••••" class="w-full mt-1" @keyup.enter="onRegister" /></div>
          <Button :label="loading ? 'Memproses...' : 'Daftar'" :loading="loading" class="w-full !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl !py-3 font-semibold" @click="onRegister" />
          <p class="text-xs text-center text-[#6B7280]">Sudah punya akun? <RouterLink to="/login" class="text-[#8B5A2B] font-semibold hover:underline">Masuk</RouterLink></p>
        </div>
      </template>
    </Card>
  </div>
</template>
