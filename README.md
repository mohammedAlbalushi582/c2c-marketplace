# الأمجاد للأعمال والعقارات — Al-Amjad Marketplace

An OpenSooq-style classifieds marketplace for **real-estate/land** and **website-template**
sales, for an Arabic-speaking (RTL) audience in Oman.

## Stack
| Layer | Tech |
|-------|------|
| Backend | Go 1.26 · chi · pgx/pgxpool · sqlc · golang-migrate · JWT · bcrypt |
| Frontend | Next.js 14 (App Router, standalone) · Tailwind · RTL Arabic |
| CMS | Strapi 5 (own Postgres DB) — static/marketing content only |
| DB | PostgreSQL 16 |
| Infra | Docker Compose |

## Monorepo layout
```
backend/      Go API — clean architecture (domain/usecase/repository/delivery)
frontend/     Next.js 14 RTL app
cms/          Strapi 5 (separate database)
migrations/   golang-migrate raw SQL (up/down) + seed + demo content
deploy/       Postgres init (creates Strapi's separate DB)
docker-compose.yml
```

## Run the whole thing
```bash
cp .env.example .env
cd cms && cp .env.example .env 2>/dev/null || true; cd ..   # Strapi secrets (generated on scaffold)

docker compose up --build
```
Then open:

| Service | URL | Notes |
|---------|-----|-------|
| **Frontend** | http://localhost:3000 | the marketplace |
| Backend API | http://localhost:8080/api/v1 | REST, versioned |
| Health | http://localhost:8080/healthz | |
| Strapi admin | http://localhost:1337/admin | create the first admin on load |

Migrations (schema + seed + demo listings) run automatically via the `migrate` service.

### Demo accounts (seeded)
| Role | Login | Password |
|------|-------|----------|
| Admin | `admin@amjad.om` | `password123` |
| Seller | `seller@amjad.om` | `password123` |

Admins see a **مراجعة الإعلانات** (moderation) tab in the dashboard to approve/reject
pending listings.

## Local dev (without Docker)
```bash
docker compose up -d postgres      # just the DB
make migrate-up                    # apply migrations (golang-migrate via Docker)
make backend-run                   # Go API on :8080  (reads ./.env)
cd frontend && npm install && npm run dev   # Next.js on :3000
```
`make sqlc` regenerates the type-safe query layer after editing `backend/.../queries/*.sql`.

## API surface (`/api/v1`)
```
POST   /auth/register            POST /auth/login          POST /auth/refresh    GET /auth/me
GET    /categories               GET  /categories/{slug}   GET  /categories/{id}/fields
GET    /locations
GET    /listings (search+filter) GET  /listings/{id}
POST   /listings                 PUT  /listings/{id}       DELETE /listings/{id}
POST   /listings/{id}/images     POST/DELETE /listings/{id}/favorite
GET    /me/listings              GET  /me/favorites
# admin
GET    /admin/listings?status=pending      PATCH /admin/listings/{id}/status
POST   /admin/categories ...               POST  /admin/categories/{id}/fields
```

## Database schema (8 migrations)
| # | Migration | Contents |
|---|-----------|----------|
| 1 | extensions_and_helpers | `set_updated_at()` trigger |
| 2 | users_auth | `users`, `auth_providers` (SSO-ready), `sessions` |
| 3 | locations | Oman governorate → wilayat (self-ref) |
| 4 | categories | `categories` (tree) + `category_fields` (dynamic attrs) |
| 5 | listings | `listings` (soft-delete) + `listing_images` + `listing_attributes` (EAV) |
| 6 | favorites_reports | `favorites`, `reports` |
| 7 | seed | 2 categories + custom fields + Oman locations |
| 8 | demo_content | demo users + 6 active listings with attributes |

Design highlights:
- **Extensible categories** — admin adds subcategories + per-category custom fields; a leaf
  category resolves fields from itself **and its ancestors** (recursive CTE).
- **EAV attributes** with typed columns keep per-field filtering indexable.
- **JWT auth** — stateless access tokens + hashed refresh tokens (rotated on use); Phase-2
  OAuth SSO drops into `auth_providers` with no schema change.
- **Moderation** — new listings start `pending`; an admin approves to `active`.
- **No messaging in v1** — contact via `contact_phone` / WhatsApp; a `messages` table can be
  added later without touching existing tables.

## Notes
- Strapi runs on a **separate `strapi` database** (created by `deploy/postgres/init` on a fresh
  volume). For a clean slate: `docker compose down -v && docker compose up --build`.
- File uploads use local storage behind a `Storage` interface (`backend/.../platform/storage`),
  swappable for S3 / Oracle Object Storage. Served at `/uploads/*`.

## Build status — v1 complete ✅
1. ✅ Backend skeleton, DB connection, migration setup
2. ✅ Users + auth (register/login/JWT + refresh rotation)
3. ✅ Categories + dynamic attributes (+ locations)
4. ✅ Listings CRUD (images, search/filter, EAV attributes, favorites, moderation)
5. ✅ Frontend auth pages + RTL Tailwind theme
6. ✅ Listing browse/search with filters
7. ✅ Post-a-listing dynamic form (category-driven fields)
8. ✅ Strapi on its own DB
9. ✅ Full Docker Compose wiring — verified end-to-end
