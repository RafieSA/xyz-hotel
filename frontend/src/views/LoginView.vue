<script setup>
import Card from 'primevue/card'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import { LogIn } from 'lucide-vue-next'
import { ref } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useToast } from 'primevue/usetoast'

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const router = useRouter()
const auth = useAuthStore()
const toast = useToast()

async function onLogin() {
  error.value = ''
  if (!email.value || !password.value) {
    error.value = 'Enter your email and password'
    return
  }
  loading.value = true
  try {
    await auth.loginRequest(email.value, password.value)
    toast.add({ severity: 'success', summary: 'Welcome back', detail: 'You are now signed in', life: 2500 })
    router.push('/')
  } catch (e) {
    error.value = e?.response?.data?.message || e.message || 'Could not sign you in. Check your email and password.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-[70vh] flex items-center justify-center px-4 py-12 bg-[#F9FAFB]">
    <Card class="w-full max-w-md !rounded-2xl !shadow-md !border !border-[#E5E7EB]">
      <template #title><span class="font-display font-bold text-xl text-[#1A3A4A] flex items-center gap-2"><LogIn class="w-5 h-5 text-[#8B5A2B]" /> Welcome back</span></template>
      <template #subtitle><span class="text-sm text-[#6B7280]">Sign in to manage your bookings</span></template>
      <template #content>
        <div class="space-y-4 mt-2">
          <Message v-if="error" severity="error" class="text-sm">{{ error }}</Message>
          <div><label class="text-sm font-medium text-[#1F2937]">Email address</label><InputText v-model="email" placeholder="you@example.com" class="w-full mt-1" @keyup.enter="onLogin" /></div>
          <div><label class="text-sm font-medium text-[#1F2937]">Password</label><InputText v-model="password" type="password" placeholder="••••••••" class="w-full mt-1" @keyup.enter="onLogin" /></div>
          <Button :label="loading ? 'Signing in...' : 'Sign In to Your Account'" :loading="loading" class="w-full !bg-[#8B5A2B] !border-[#8B5A2B] hover:!bg-[#6F4620] !rounded-xl !py-3 font-semibold" @click="onLogin" />
          <p class="text-xs text-center text-[#6B7280]">No account yet? <RouterLink to="/register" class="text-[#8B5A2B] font-semibold hover:underline">Create one</RouterLink></p>
        </div>
      </template>
    </Card>
  </div>
</template>
