# 08 — Dependencies & Packages (LOCKED 2026-10-08)

> Keputusan final Rafie: **Backend Go + Frontend Vue — prioritas #1 BAGUS DIMATA (cantik profesional)**. Bosen shadcn, tidak butuh engine cepat. Stack ini hasil deep dive internet 2026 + skill frontend-design.

## Stack Final Locked

```
Frontend: Vue 3 + Vite + Tailwind CSS + PrimeVue Aura + Lucide Vue
Backend:  Go 1.26.3 + Fiber + pgx/sqlx + golang-migrate + validator + jwt
DB:       Postgres 18.4 (local, tanpa Docker)
```

## Frontend — Vue 3

| Kebutuhan | Library | Versi | Alasan Cantik | Install |
|-----------|---------|-------|---------------|---------|
| **Framework** | `vue` | ^3.5 | Vue 3 composition API, paling membumi | `npm create vue@latest` |
| **Build** | `vite` | ^6 | Super cepat, HMR instant | bawaan `create-vue` |
| **Routing** | `vue-router` | ^4 | Frontoffice vs Backoffice pisah route | `npm i vue-router` |
| **State** | `pinia` | ^3 | Auth & booking state clean, SRP | `npm i pinia` |
| **Styling** | `tailwindcss` | ^3.4 | Canvas cantik 100% custom `#8B5A2B`, bukan template | `npm i -D tailwindcss postcss autoprefixer` |
| **UI Library** | `primevue` | ^4.3 | **Aura theme paling cantik 2026** — rounded-xl, shadow soft, DataTable & Calendar premium untuk hotel | `npm i primevue` |
| **Theme** | `@primevue/themes` | via primevue | Aura preset custom ke `#8B5A2B` + `#C9A86A` | bawaan primevue |
| **Icons** | `lucide-vue-next` | latest | Garis tipis premium, pasangan PrimeVue, 1000+ icon | `npm i lucide-vue-next` |
| **HTTP** | `axios` | ^1.7 | Call Go API `http://localhost:8080` | `npm i axios` |
| **Form Validasi** | `zod` + `vee-validate` | latest | Validasi Zod di Vue, jangan percaya input user | `npm i zod vee-validate` |

### PrimeVue Aura — Kunci Cantik

**Jangan pakai warna default biru.** Override di `frontend/src/theme/aura.js`:

```js
import Aura from '@primevue/themes/aura'
import { definePreset } from '@primevue/themes'

export const WarmAura = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#FDF6EC', 500: '#8B5A2B', 600: '#6F4620'
    },
    colorScheme: {
      light: { primary: { color: '{primary.500}' } }
    }
  }
})
// main.js: app.use(PrimeVue, { theme: { preset: WarmAura } })
```

**Komponen yang dipakai untuk hotel:**

| Fitur Hotel | Komponen PrimeVue | Kenapa Cantik |
|-------------|-------------------|---------------|
| Card kamar | `Card` | rounded-xl + shadow-md + hover lift 2px |
| Kalender availability | `DatePicker` | Aura rounded-xl cream, range highlight `#FDF6EC` |
| Tabel booking (backoffice) | `DataTable` + `Column` + `Tag` | Filter, pagination, Tag status warna sesuai `07-design.md` |
| Dialog booking | `Dialog` | rounded-2xl, overlay `#1A3A4A` 40% |
| Toast notif | `Toast` | verified hijau, rejected merah |
| Chart laporan | `Chart` (Chart.js) | Occupancy & revenue |

### Tailwind Config — Token Cantik

```js
// tailwind.config.js
export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: {
    extend: {
      colors: {
        primary: '#8B5A2B', 'primary-hover': '#6F4620',
        cream: '#FDF6EC', teal: '#1A3A4A', gold: '#C9A86A'
      },
      borderRadius: { xl: '12px', '2xl': '16px' },
      fontFamily: { display: ['Playfair Display', 'serif'], body: ['Inter', 'sans-serif'] }
    }
  }
}
```

## Backend — Go 1.26.3

| Kebutuhan | Library | Alasan | Install |
|-----------|---------|--------|---------|
| **Framework** | `github.com/gofiber/fiber/v2` | Paling ngebut, syntax Express-like, docs jelas | `go get -u github.com/gofiber/fiber/v2` |
| **DB Driver** | `github.com/jackc/pgx/v5` + `github.com/jmoiron/sqlx` | pgx driver Postgres tercepat, sqlx query manual tapi tidak ribet | `go get github.com/jackc/pgx/v5 github.com/jmoiron/sqlx` |
| **Migrasi** | `github.com/golang-migrate/migrate/v4` | File `001_*.sql` terurut, ACID | `go get github.com/golang-migrate/migrate/v4` |
| **Validasi** | `github.com/go-playground/validator/v10` | `validate:"required,gt=0"` cegah invalid | `go get github.com/go-playground/validator/v10` |
| **JWT** | `github.com/golang-jwt/jwt/v5` | Access 15m + refresh httpOnly, bcrypt hash | `go get github.com/golang-jwt/jwt/v5` |
| **Env** | `github.com/joho/godotenv` | Load `.env` simple | `go get github.com/joho/godotenv` |
| **CORS** | `github.com/gofiber/fiber/v2/middleware/cors` | Frontend Vue `localhost:5173` ↔ Go `8080` | bawaan fiber |
| **Logging** | `log/slog` (stdlib Go 1.21+) | Terstruktur, tidak perlu lib | stdlib |

### Struktur Backend (akan di-scaffold Fase 1)

```
backend/
├── cmd/server/main.go          # Fiber app + route
├── internal/
│   ├── handler/                # HTTP only (SRP)
│   ├── service/                # Business + transaction FOR UPDATE
│   ├── middleware/             # JWT + RBAC (owner/manager/receptionist/customer)
│   ├── model/                  # struct User, RoomType, Booking
│   └── repo/                   # sqlx query ($1,$2 anti-SQLi)
├── migrations/001_init.sql     # users, room_types, room_units, bookings, vouchers, audit_logs
├── .env.example
└── go.mod
```

## Versi Terkunci (Faktual M4 2026-10-08)

| Tool | Versi | Cek |
|------|-------|-----|
| Go | 1.26.3 darwin/arm64 | `go version` ✅ |
| Node | 26.7.0 | `node -v` ✅ |
| Postgres | 18.4 | `psql --version` ✅ |
| Git | 2.53.0 | `git --version` ✅ |

## Best Practice Cantik (Wajib)

- **1 warna primary:** `#8B5A2B` untuk semua CTA, harga, badge primary — jangan campur biru template.
- **1 radius:** `rounded-xl` card, `rounded-full` button — jangan random.
- **1 shadow:** `shadow-sm` normal, `shadow-md` hover — jangan `shadow-lg` norak.
- **Icon 20px** konsisten `w-5 h-5`, warna `text-[#8B5A2B]` atau `text-[#6B7280]`.
- **Validasi di 2 tempat:** Zod di Vue + validator di Go — jangan percaya frontend saja.
- **No handwritten sia-sia:** Pakai Fiber, sqlx, PrimeVue — jangan bikin router/table dari 0.

## DoD Dependencies

- [x] Dipilih & didokumentasikan (file ini)
- [ ] `frontend/` scaffold: `npm create vue@latest` + tailwind + primevue + lucide
- [ ] `backend/` scaffold: `go mod init` + fiber + pgx/sqlx + migrate
- [ ] `07-design.md` token sync ke `tailwind.config.js` & `aura.js`

> Next: Fase 1 scaffolding — `backend/` + `frontend/` hello world + migrasi + seed 18 unit.

## Referensi Preview (Buka di Browser)

- PrimeVue Aura: https://primevue.org/
- PrimeVue DatePicker: https://primevue.org/datepicker
- PrimeVue DataTable: https://primevue.org/datatable
- Nuxt UI (alternatif yang tidak dipilih): https://ui.nuxt.com/
- Lucide Icons: https://lucide.dev/icons/
- UnoCSS (tidak dipilih, karena prioritas cantik bukan speed): https://unocss.dev/

> Keputusan locked. Jangan ganti tanpa diskusi. Prioritas tetap #1 BAGUS DIMATA.
