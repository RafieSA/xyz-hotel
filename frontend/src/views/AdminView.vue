<script setup>
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Tag from 'primevue/tag'
import Card from 'primevue/card'
import { LayoutDashboard, TrendingUp, CalendarDays } from 'lucide-vue-next'

const bookings = [
  { id: 'BK-001', guest: 'Rafie S.', room: 'Deluxe', check_in: '2026-10-10', check_out: '2026-10-12', status: 'verified', total: 1100000 },
  { id: 'BK-002', guest: 'Sinta A.', room: 'Family', check_in: '2026-10-11', check_out: '2026-10-13', status: 'pending_payment', total: 1700000 },
  { id: 'BK-003', guest: 'Budi W.', room: 'Standard', check_in: '2026-10-12', check_out: '2026-10-13', status: 'waiting_verification', total: 350000 },
  { id: 'BK-004', guest: 'Maya L.', room: 'Suite', check_in: '2026-10-13', check_out: '2026-10-15', status: 'checked_in', total: 2500000 },
  { id: 'BK-005', guest: 'Andi K.', room: 'Deluxe', check_in: '2026-10-14', check_out: '2026-10-16', status: 'cancelled', total: 1100000 },
]

const statusSeverity = (s) => ({ verified: 'success', checked_in: 'info', pending_payment: 'warn', waiting_verification: 'warn', cancelled: 'danger', rejected: 'danger' }[s] || 'secondary')
const statusLabel = (s) => s.replace('_', ' ')
const fmt = (n) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(n)
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 py-8 space-y-6">
    <div class="flex items-center gap-3">
      <span class="bg-[#8B5A2B] text-white rounded-xl p-2.5"><LayoutDashboard class="w-5 h-5" /></span>
      <div>
        <h1 class="font-display font-bold text-2xl text-[#1A3A4A]">Backoffice</h1>
        <p class="text-sm text-[#6B7280]">Kelola booking, kamar, dan laporan.</p>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
        <template #content>
          <p class="text-xs uppercase tracking-wide text-[#6B7280] font-semibold flex items-center gap-1.5"><CalendarDays class="w-4 h-4" /> Booking hari ini</p>
          <p class="text-2xl font-bold text-[#1A3A4A] mt-1">12</p>
          <p class="text-xs text-[#2E7D32] mt-1 flex items-center gap-1"><TrendingUp class="w-3.5 h-3.5" /> +18% vs kemarin</p>
        </template>
      </Card>
      <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
        <template #content>
          <p class="text-xs uppercase tracking-wide text-[#6B7280] font-semibold">Occupancy</p>
          <p class="text-2xl font-bold text-[#1A3A4A] mt-1">68%</p>
          <div class="mt-2 h-2 bg-[#F3F4F6] rounded-full overflow-hidden"><div class="h-full bg-[#8B5A2B] rounded-full" style="width:68%"></div></div>
        </template>
      </Card>
      <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
        <template #content>
          <p class="text-xs uppercase tracking-wide text-[#6B7280] font-semibold">Revenue (MTD)</p>
          <p class="text-2xl font-bold text-[#8B5A2B] mt-1">Rp 42.500.000</p>
          <p class="text-xs text-[#6B7280] mt-1">Chart placeholder — Chart.js nanti</p>
        </template>
      </Card>
    </div>

    <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB]">
      <template #title><span class="text-[#1A3A4A] font-semibold text-base">Booking terbaru</span></template>
      <template #content>
        <DataTable :value="bookings" paginator :rows="5" stripedRows class="text-sm" responsiveLayout="scroll">
          <Column field="id" header="ID" sortable />
          <Column field="guest" header="Tamu" sortable />
          <Column field="room" header="Tipe" />
          <Column field="check_in" header="Check-in" sortable />
          <Column field="check_out" header="Check-out" />
          <Column field="status" header="Status">
            <template #body="{ data }"><Tag :value="statusLabel(data.status)" :severity="statusSeverity(data.status)" rounded class="capitalize" /></template>
          </Column>
          <Column field="total" header="Total">
            <template #body="{ data }"><span class="font-semibold text-[#8B5A2B]">{{ fmt(data.total) }}</span></template>
          </Column>
        </DataTable>
      </template>
    </Card>

    <Card class="!rounded-xl !shadow-sm !border !border-[#E5E7EB] bg-[#FDF6EC]/60">
      <template #content>
        <p class="text-sm font-semibold text-[#1A3A4A] flex items-center gap-2"><TrendingUp class="w-4 h-4 text-[#8B5A2B]" /> Chart placeholder</p>
        <div class="mt-3 h-40 bg-white rounded-xl border border-dashed border-[#C9A86A]/40 flex items-center justify-center text-sm text-[#6B7280]">Occupancy & revenue chart (Chart.js) — Fase 2</div>
      </template>
    </Card>
  </div>
</template>
