# 08: Dependencies and Packages (Locked 2026-10-08)

> Locked decision: Go backend plus Vue frontend. Priority 1 is that it looks great and feels premium. This stack comes from a 2026 deep dive and it runs lean on the M4.

## Final Stack Locked

```
Frontend: Vue 3 + Vite + Tailwind CSS + PrimeVue Aura + Lucide Vue
Backend:  Go 1.26.3 + Fiber + pgx/sqlx + golang-migrate + validator + jwt
DB:       Postgres 18.4 (local, without Docker)
```

## Frontend, Vue 3

| Need | Library | Version | Why It Looks Good | Install |
|------|---------|---------|-------------------|---------|
| **Framework** | `vue` | ^3.5 | Vue 3 Composition API, most approachable | `npm create vue@latest` |
| **Build** | `vite` | ^6 | Super fast, instant HMR | included with `create-vue` |
| **Routing** | `vue-router` | ^4 | Separate routes for Frontoffice vs Backoffice | `npm i vue-router` |
| **State** | `pinia` | ^3 | Clean auth & booking state, SRP | `npm i pinia` |
| **Styling** | `tailwindcss` | ^3.4 | 100% custom beautiful canvas `#8B5A2B`, not a template | `npm i -D tailwindcss postcss autoprefixer` |
| **UI Library** | `primevue` | ^4.3 | **Most beautiful Aura theme in 2026**, rounded-xl, soft shadows, premium DataTable & Calendar for a hotel | `npm i primevue` |
| **Theme** | `@primevue/themes` | via primevue | Aura preset customized to `#8B5A2B` + `#C9A86A` | included with primevue |
| **Icons** | `lucide-vue-next` | latest | Thin premium lines, perfect match for PrimeVue, 1000+ icons | `npm i lucide-vue-next` |
| **HTTP** | `axios` | ^1.7 | Calls Go API `http://localhost:8080` | `npm i axios` |
| **Form Validation** | `zod` + `vee-validate` | latest | Zod validation in Vue, never trust user input | `npm i zod vee-validate` |

### PrimeVue Aura, The Key to Beauty

**Do not use the default blue.** Override in `frontend/src/theme/aura.js`:

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

**Components used for the hotel:**

| Hotel Feature | PrimeVue Component | Why It Looks Good |
|---------------|--------------------|-------------------|
| Room card | `Card` | rounded-xl + shadow-md + 2px hover lift |
| Availability calendar | `DatePicker` | Aura rounded-xl cream, range highlight `#FDF6EC` |
| Booking table (backoffice) | `DataTable` + `Column` + `Tag` | Filters, pagination, status Tag colors per `07-design.md` |
| Booking dialog | `Dialog` | rounded-2xl, overlay `#1A3A4A` 40% |
| Toast notification | `Toast` | verified green, rejected red |
| Report chart | `Chart` (Chart.js) | Occupancy & revenue |

### Tailwind Config, Beautiful Tokens

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

## Backend, Go 1.26.3

| Need | Library | Reason | Install |
|------|---------|--------|---------|
| **Framework** | `github.com/gofiber/fiber/v2` | Fastest, Express-like syntax, clear docs | `go get -u github.com/gofiber/fiber/v2` |
| **DB Driver** | `github.com/jackc/pgx/v5` + `github.com/jmoiron/sqlx` | Fastest Postgres driver (pgx), sqlx for manual but simple queries | `go get github.com/jackc/pgx/v5 github.com/jmoiron/sqlx` |
| **Migration** | `github.com/golang-migrate/migrate/v4` | Ordered `001_*.sql` files, ACID | `go get github.com/golang-migrate/migrate/v4` |
| **Validation** | `github.com/go-playground/validator/v10` | `validate:"required,gt=0"` prevents invalid input | `go get github.com/go-playground/validator/v10` |
| **JWT** | `github.com/golang-jwt/jwt/v5` | Access 15m + refresh httpOnly, bcrypt hash | `go get github.com/golang-jwt/jwt/v5` |
| **Env** | `github.com/joho/godotenv` | Simple `.env` loading | `go get github.com/joho/godotenv` |
| **CORS** | `github.com/gofiber/fiber/v2/middleware/cors` | Vue frontend `localhost:5173` ↔ Go `8080` | included with fiber |
| **Logging** | `log/slog` (Go 1.21+ stdlib) | Structured, no extra lib needed | stdlib |

### Backend Structure (to be scaffolded in Phase 1)

```
backend/
├── cmd/server/main.go          # Fiber app + routes
├── internal/
│   ├── handler/                # HTTP only (SRP)
│   ├── service/                # Business + transaction FOR UPDATE
│   ├── middleware/             # JWT + RBAC (owner/manager/receptionist/customer)
│   ├── model/                  # struct User, RoomType, Booking
│   └── repo/                   # sqlx queries ($1,$2 anti-SQLi)
├── migrations/001_init.sql     # users, room_types, room_units, bookings, vouchers, audit_logs
├── .env.example
└── go.mod
```

## Locked Versions (Factual as of M4 2026-10-08)

| Tool | Version | Check |
|------|---------|-------|
| Go | 1.26.3 darwin/arm64 | `go version` ✅ |
| Node | 26.7.0 | `node -v` ✅ |
| Postgres | 18.4 | `psql --version` ✅ |
| Git | 2.53.0 | `git --version` ✅ |

## Beauty Best Practices (Required)

- **1 primary color:** `#8B5A2B` for all CTAs, prices, primary badges, do not mix in blue template colors.
- **1 radius:** `rounded-xl` for cards, `rounded-full` for buttons, do not randomize.
- **1 shadow:** `shadow-sm` normal, `shadow-md` on hover, no tacky `shadow-lg`.
- **Icons at 20px** consistently `w-5 h-5`, color `text-[#8B5A2B]` or `text-[#6B7280]`.
- **Validate in 2 places:** Zod in Vue + validator in Go, never trust the frontend alone.
- **No pointless hand-coding:** Use Fiber, sqlx, PrimeVue, do not build a router/table from scratch.

## DoD Dependencies

- [x] Selected & documented (this file)
- [ ] `frontend/` scaffold: `npm create vue@latest` + tailwind + primevue + lucide
- [ ] `backend/` scaffold: `go mod init` + fiber + pgx/sqlx + migrate
- [ ] `07-design.md` tokens synced to `tailwind.config.js` & `aura.js`

> Next: Phase 1 scaffolding, `backend/` + `frontend/` hello world + migration + seed 18 units.

## Preview References (Open in Browser)

- PrimeVue Aura: https://primevue.org/
- PrimeVue DatePicker: https://primevue.org/datepicker
- PrimeVue DataTable: https://primevue.org/datatable
- Nuxt UI (alternative not chosen): https://ui.nuxt.com/
- Lucide Icons: https://lucide.dev/icons/
- UnoCSS (not chosen, because priority is beauty not speed): https://unocss.dev/

> Decision locked. Do not change without discussion. Priority remains #1 LOOKS GOOD.