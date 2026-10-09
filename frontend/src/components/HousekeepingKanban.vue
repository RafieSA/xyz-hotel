<script setup>
import { ref, computed } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import {
  Kanban,
  Sparkles,
  CheckCircle2,
  Wrench,
  KeyRound,
  AlertTriangle,
  RotateCw,
  Bed,
  Check
} from 'lucide-vue-next'

const props = defineProps({
  roomUnits: {
    type: Array,
    required: true,
  },
  roomTypes: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['update-status', 'refresh'])

// 4 Kolom Kanban
const columns = [
  { key: 'available', title: 'Siap Huni', color: '#2E7D32', bg: 'bg-emerald-50 dark:bg-emerald-950/30', border: 'border-emerald-200 dark:border-emerald-800' },
  { key: 'occupied', title: 'Sedang Terisi', color: '#8B5A2B', bg: 'bg-[#FDF6EC] dark:bg-[#1A3A4A]/50', border: 'border-[#8B5A2B]/30 dark:border-[#8B5A2B]/50' },
  { key: 'dirty', title: 'Perlu Dibersihkan', color: '#EF6C00', bg: 'bg-amber-50 dark:bg-amber-950/30', border: 'border-amber-200 dark:border-amber-800' },
  { key: 'maintenance', title: 'Perawatan / Rusak', color: '#6B7280', bg: 'bg-gray-50 dark:bg-gray-800/40', border: 'border-gray-200 dark:border-gray-700' },
]

// Drag and drop state
const draggedUnit = ref(null)

function onDragStart(unit) {
  draggedUnit.value = unit
}

function onDragOver(e) {
  e.preventDefault()
}

function onDrop(targetCol) {
  if (!draggedUnit.value) return
  if (draggedUnit.value.status !== targetCol) {
    emit('update-status', { unit: draggedUnit.value, newStatus: targetCol })
  }
  draggedUnit.value = null
}

// Group units by column
const kanbanData = computed(() => {
  const result = { available: [], occupied: [], dirty: [], maintenance: [] }
  const units = props.roomUnits && props.roomUnits.length ? props.roomUnits : []
  units.forEach(u => {
    const s = u.status || 'available'
    if (result[s]) result[s].push(u)
    else result.available.push(u)
  })
  return result
})

// Helper nama tipe kamar
function getRoomTypeName(typeId) {
  const found = props.roomTypes.find(rt => rt.id === typeId)
  return found?.name || found?.type || `Tipe #${typeId || 1}`
}

function getColumnIcon(key) {
  switch (key) {
    case 'available': return CheckCircle2
    case 'occupied': return KeyRound
    case 'dirty': return AlertTriangle
    case 'maintenance': return Wrench
    default: return Bed
  }
}
</script>

<template>
  <div class="bg-white dark:bg-[#2D3748] rounded-3xl p-5 md:p-6 border border-[#E5E7EB] dark:border-gray-700 shadow-sm">
    <!-- Header Kanban Board -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-5 border-b border-[#F3F4F6] dark:border-gray-700">
      <div>
        <div class="flex items-center gap-2">
          <Kanban class="w-5 h-5 text-[#8B5A2B]" />
          <h2 class="font-display font-bold text-xl text-[#1A3A4A] dark:text-white" style="font-family:'Playfair Display',serif">
            Papan Kendali Kebersihan & Status Kamar (*Housekeeping*)
          </h2>
        </div>
        <p class="text-xs text-[#6B7280] dark:text-gray-400 mt-1">
          Pantau kesiapan unit kamar secara langsung. Geser kartu antar-kolom (*drag & drop*) atau gunakan tombol cepat 1-klik.
        </p>
      </div>

      <Button
        label="Segarkan Unit"
        icon="pi pi-refresh"
        outlined
        size="small"
        class="!rounded-xl !text-xs !border-[#8B5A2B] !text-[#8B5A2B] self-start sm:self-auto whitespace-nowrap"
        @click="emit('refresh')"
      />
    </div>

    <!-- 4 Kolom Kanban Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mt-6">
      <div
        v-for="col in columns"
        :key="col.key"
        class="rounded-2xl border p-4 flex flex-col min-h-[380px] transition-colors"
        :class="[col.bg, col.border]"
        @dragover="onDragOver"
        @drop="onDrop(col.key)"
      >
        <!-- Header Kolom -->
        <div class="flex items-center justify-between pb-3 mb-3 border-b border-black/5 dark:border-white/10">
          <div class="flex items-center gap-2">
            <component :is="getColumnIcon(col.key)" class="w-4 h-4" :style="{ color: col.color }" />
            <h3 class="text-xs font-bold uppercase tracking-wider text-[#1A3A4A] dark:text-white">
              {{ col.title }}
            </h3>
          </div>
          <span class="px-2 py-0.5 rounded-full text-[11px] font-bold bg-white dark:bg-[#1A3A4A] border border-black/5 dark:border-white/10 text-gray-700 dark:text-gray-200">
            {{ kanbanData[col.key]?.length || 0 }}
          </span>
        </div>

        <!-- Daftar Kartu Unit di Kolom -->
        <div class="space-y-3 flex-1 overflow-y-auto max-h-[500px] pr-0.5">
          <div
            v-for="unit in kanbanData[col.key]"
            :key="'unit-' + unit.id"
            draggable="true"
            @dragstart="onDragStart(unit)"
            class="bg-white dark:bg-[#1A3A4A] p-3.5 rounded-2xl border border-gray-200 dark:border-gray-700 shadow-xs hover:shadow-md transition-all cursor-grab active:cursor-grabbing select-none group"
          >
            <!-- Nomor Kamar & Tipe -->
            <div class="flex items-start justify-between gap-2">
              <div>
                <p class="font-bold text-sm text-[#1A3A4A] dark:text-white flex items-center gap-1.5">
                  <Bed class="w-3.5 h-3.5 text-[#8B5A2B]" />
                  {{ unit.code }}
                </p>
                <p class="text-[11px] text-[#6B7280] dark:text-gray-400 mt-0.5">
                  {{ getRoomTypeName(unit.room_type_id) }}
                </p>
              </div>

              <!-- Lencana Status -->
              <span
                class="w-2.5 h-2.5 rounded-full"
                :style="{ backgroundColor: col.color }"
                :title="col.title"
              ></span>
            </div>

            <!-- Tombol Aksi Cepat 1-Klik untuk Staf -->
            <div class="mt-3 pt-2.5 border-t border-gray-100 dark:border-gray-800 flex flex-col gap-1.5">
              <!-- Jika kamar kotor (dirty): Tombol 1-klik "Tandai Bersih" -->
              <button
                v-if="unit.status === 'dirty'"
                type="button"
                @click.stop="emit('update-status', { unit, newStatus: 'available' })"
                class="w-full flex items-center justify-center gap-1.5 py-1.5 px-3 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-semibold shadow-xs transition cursor-pointer"
              >
                <Sparkles class="w-3.5 h-3.5" />
                <span>Tandai Bersih & Siap</span>
              </button>

              <!-- Jika kamar siap huni (available): Opsi setel maintenance jika ada kerusakan -->
              <button
                v-if="unit.status === 'available'"
                type="button"
                @click.stop="emit('update-status', { unit, newStatus: 'maintenance' })"
                class="w-full flex items-center justify-center gap-1 py-1 px-2.5 rounded-xl text-[11px] font-medium text-gray-500 hover:text-amber-700 hover:bg-amber-50 dark:hover:bg-amber-950/30 transition cursor-pointer"
              >
                <Wrench class="w-3 h-3" />
                <span>Jadwalkan Servis</span>
              </button>

              <!-- Jika sedang maintenance: Tombol 1-klik "Selesai Servis" -->
              <button
                v-if="unit.status === 'maintenance'"
                type="button"
                @click.stop="emit('update-status', { unit, newStatus: 'available' })"
                class="w-full flex items-center justify-center gap-1.5 py-1.5 px-3 rounded-xl bg-[#8B5A2B] hover:bg-[#6F4620] text-white text-xs font-semibold shadow-xs transition cursor-pointer"
              >
                <Check class="w-3.5 h-3.5" />
                <span>Selesai Servis & Siap</span>
              </button>

              <!-- Jika sedang terisi tamu (occupied) -->
              <div v-if="unit.status === 'occupied'" class="text-[11px] text-[#8B5A2B] dark:text-[#C9A86A] font-medium flex items-center justify-center gap-1 py-0.5">
                <KeyRound class="w-3 h-3" />
                <span>Kamar dihuni tamu</span>
              </div>
            </div>
          </div>

          <!-- Empty State Tiap Kolom -->
          <div
            v-if="!kanbanData[col.key]?.length"
            class="py-10 text-center border-2 border-dashed border-black/10 dark:border-white/10 rounded-2xl text-xs text-gray-400"
          >
            <span>Tidak ada kamar</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
