# 03: Tech Stack Decision (No Next.js)

> Grilling Q1: answered directly and precisely. Why we skipped Next.js and why Go plus Vue fits this hotel.

## Why Was Next.js Previously Recommended?
Honestly, for 3 reasons (not fanboyism):

| Reason | Down-to-Earth Explanation |
|--------|---------------------------|
| Hotel SEO needs SSR | Hotels need Google to find their room pages. Next.js excels at SSR. But **it's not the only option**, Laravel SSR, Nuxt, and SvelteKit can do it too. |
| Large ecosystem | Many ready-made Next.js hotel templates exist. But Laravel & Nuxt templates are also plentiful. |
| Common reference | Most JS hotel portfolios use Next. But you said **big no**, so we respect that, your preference comes first. |

**Conclusion:** Next.js is not mandatory. We can achieve equally good results without it. Forcing a stack you dislike is unacceptable.

## Replacement Options (Without Next.js), Honest Tradeoffs

### Option A, Laravel 12 + Inertia + Vue 3 + MySQL/Postgres ⭐ RECOMMENDED
**Overview:** A single Laravel codebase serving as both backend + frontend (Vue inside Laravel via Inertia). Like `xyz-haircut` but in PHP.

| Aspect | Assessment |
|--------|------------|
| **Pros** | Your PHP 8.2 is ready, MySQL/Postgres is ready, **no Docker needed**. Laravel includes auth, CSRF, safe ORM (anti-SQLi), RBAC (Gates/Policies), queue, mail, storage for proof uploads, all built-in. Inertia gives SPA UX without the hassle of a separate API. Suitable for a hotel portfolio (many real hotels run on PHP). Fast locally on M4. |
| **Cons** | Need to learn Inertia if you haven't used it before. Not pure JS. |
| **Scalability** | High, scaling Laravel to multi-hotel is straightforward. |
| **Security** | Laravel defaults already prevent SQLi, XSS, CSRF, IDOR via Policies. |
| **Example** | `app/Http/Controllers/BookingController.php` handles availability with a DB transaction. `resources/js/Pages/Booking/Create.vue` displays the calendar. |

**Good fit if:** You want to ship fast, local-first, with a portfolio that's easy to deploy to cheap shared hosting/VPS.

### Option B, Nuxt 3 (Vue) + Nitro + Postgres (via Prisma/Drizzle)
**Overview:** The Vue cousin of Next.js. Also great at SSR.

| Aspect | Assessment |
|--------|------------|
| **Pros** | Vue is more approachable than React, SSR is solid for hotel SEO, large community. |
| **Cons** | Auth must be set up manually (not as complete as Laravel). Requires a separate backend or Nitro. Still runs on M4 without Docker, but deployment needs a Node server. |
| **Scalability** | High. |
| **Security** | Must handle CSRF/SQLi manually. |

**Good fit if:** You love Vue and want full JS.

### Option C, SvelteKit + Postgres
**Overview:** The lightest & fastest, with simple syntax.

| Aspect | Assessment |
|--------|------------|
| **Pros** | Lightweight, fast, small bundle, good SSR. |
| **Cons** | Fewer hotel templates, smaller community than Laravel/Nuxt. |
| **Scalability** | High but requires manual setup. |

### Option D, Dart Frog + Flutter Web (Consistent with xyz-haircut)
**Overview:** Align with the previous project.

| Aspect | Assessment |
|--------|------------|
| **Pros** | Consistent, already familiar to you. |
| **Cons** | Flutter Web has weak SEO (hotels need SEO), large bundle, few hotel templates. **Not recommended for a hotel website.** |

## ✅ DECISION LOCKED, 2026-10-08, Golang Backend + Vue.js Frontend (G1)

**Rafie's final choice: Golang Backend (Go 1.26.3) + Vue 3 Frontend + Postgres 18.4.**

### Final Architecture
```
[ Vue 3 + Vite + Tailwind #8B5A2B ]  -- REST API JSON (JWT) -->  [ Go API, Fiber/Gin + sqlx ]  --> [ Postgres 18.4 ]
        |                                                       |                              |
   Frontoffice: search, booking, upload                    RBAC 4 roles (owner/manager/       tables: users, room_types,
   Backoffice: dashboard, verification, reports             receptionist/customer)             room_units, bookings, vouchers,
   Design tokens from docs/07-design.md                    Transaction FOR UPDATE             audit_logs
                                                           Upload proof (5MB, jpg/png/pdf)
                                                           Expiry 12 hours (ticker)
```

### Stack Details
| Layer | Technology | Version | Reason |
|-------|------------|---------|--------|
| Backend | **Go + Fiber** (alternative Gin) | Go 1.26.3 | Fiber is the fastest, Express-like syntax, lightweight on M4 |
| DB | **Postgres** | 18.4 | Solid ACID + `FOR UPDATE`, already available locally |
| Frontend | **Vue 3 + Vite + Tailwind** | Node 26 | Vue is approachable, Vite is super fast, Tailwind for the `#8B5A2B` design |
| Auth | JWT access 15 min + refresh (httpOnly) + bcrypt |, | Prevents Broken Authentication |
| Upload | `storage/uploads` local, MIME validation |, | Manual transfer proof |
| Repo | Monorepo `backend/` + `frontend/` + `docs/` |, | Local `main` branch, no worktrees |

### Why Go + Vue Is a Good Fit for xyz-hotel
1. **Factual M4:** Go 1.26.3 & Node 26 & Postgres are ready, verified via `go version` on 2026-10-08.
2. **Go portfolio:** Rare & high-value, showcases "can do Go + concurrency", standing out from 100 Laravel applicants.
3. **Complete level still achievable:** Vouchers, reports, reviews, all possible in Go, just more manual (no magic ORM).
4. **Security:** Must be disciplined with `$1,$2` placeholders (anti-SQLi), JWT + role middleware (anti IDOR/BOLA/BFLA), Vue auto-escaping (anti-XSS).
5. **No wasteful handwriting from scratch:** Use Fiber (official), `sqlx`, `golang-migrate`, `go-playground/validator`, don't build a router from zero.

### Agreed Tradeoffs (Honest)
| Go + Vue | Consequence |
|----------|-------------|
| Dev takes 2,3× longer vs Laravel | Auth/RBAC/upload/validation written manually, but no deadline, so it's fine |
| Separate backend/frontend | Double setup (`go run` + `npm run dev`), but clean separation |
| Few Go hotel templates | Vue frontend built from scratch using design.md |
| Fastest performance | Handles booking race conditions with `FOR UPDATE` while staying ACID |

### Mandatory Best Practices (Go + Vue)
- Clean Code, SRP: `handler` handles HTTP only, `service` handles transactions, `middleware` handles auth/roles.
- Validation: `validator` in Go + `Zod` in Vue, never trust user input.
- Transaction: `BEGIN; SELECT ... FOR UPDATE; INSERT; COMMIT;` for bookings.
- Logging: structured `log/slog` + `AuditLog` table.
- Testable: `AvailabilityService` can be unit-tested without HTTP.

> Grilling Q1 locked. Q3/Q5/Q6 were already locked earlier. Ready for Phase 1 scaffolding.