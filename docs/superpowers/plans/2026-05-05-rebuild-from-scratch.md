# Rebuild From Scratch Implementation Plan

> **For agentic workers:** This is a **comparison-quality plan**, not execution-ready. It's intended to be weighed against the strip-down plan (`2026-05-05-strip-down-legacy-cruft.md`) so the user can pick a path. Tasks list deliverables and rough hours; per-step code blocks are deferred until/unless this plan is selected for execution. If selected, expand each task into TDD-style steps with full code per the writing-plans skill before starting.

**Goal:** Construct a new Plaza Real CMS in a clean git history, re-implementing only the features production currently relies on, with no legacy code from prior CMS generations carried over.

**Architecture:** Greenfield Go module. Templ-first (no static HTML fallback). Clean directory tree designed before any code is written. Reuses the production PocketBase data volume so live records survive the rebuild. Same external integrations as today (R2, Fly, Cloudflare).

**Tech Stack:** Go 1.24, Fiber v2, templ, templUI, Tailwind, PocketBase, R2 (S3-compatible).

**Time budget:** **30–60 hours of focused work (~1 week full-time, 2–3 weeks part-time).** This is **12–25× the strip-down**. The honest comparison appears at the end of this document.

---

## Up-front honest assessment

The user's hypothesis was: *"talvez es mucho más simple"*. Reading the current codebase reveals this is unlikely:

| Subsystem | Existing in repo | Effort to rewrite |
|---|---|---|
| PocketBase collection definitions | ~700 lines of `auth/collections.go` | 4–6 h to reproduce schemas + validation + relations correctly |
| JWT auth + cookie session + role middleware | `auth/jwt.go` + `middleware/auth.go` | 2–3 h, with security review |
| Real-time hub + WebSocket broadcast | `realtime/hub.go`, `handlers/ws/websocket.go` | 3–4 h, hard to test |
| R2 upload widget (drag-drop, presigned, HTMX) | `services/r2/`, `view/components/upload_field.templ`, `handlers/admin.UploadFile` | 3–4 h |
| Analytics middleware (skip rules, fail-open, async) | `middleware/tracking.go` | 2 h |
| 30+ admin handlers (tiendas, content, locales, reservas, users, settings, reports) | `handlers/admin/handlers.go` (kept portion ≈ 1500 lines) | 12–15 h |
| Public pages + fragment handlers | `handlers/web/handlers.go` + `handlers/fragments/*` + `view/pages/public/*` + `view/fragments/*` | 5–8 h |
| Liquid Glass design system (CSS, tokens) | `internal/view/layout/*.templ`, `web/static/*.css` | 1–2 h to copy + verify |
| Scraper CLI (legacy JSON → tiendas format) | `cmd/scraper/` | 2 h |
| Templ component library setup (templUI) | `components/` | 1 h |

Even with aggressive copy-paste of the parts that don't need rewrite, this floor is roughly 30 hours. Realistic ceiling with debugging, edge cases, and re-discovering subtleties (like `gal` field semantics or content_blocks date normalization) is 60+ hours.

**The strip-down plan reaches the same end state — Plaza Real-only code, clean repo — in 2.5 hours.** The "from scratch" feeling can also be achieved with the **orphan-branch variant** described at the end of this document (3–5 h), without re-implementing anything.

If after reading this you still want true greenfield rebuild, the task list follows.

---

## Phase 0: Scaffold — ~2 h

**Goal:** Empty new project that builds and serves a "Hello Plaza Real" page.

- [ ] **Task 0.1: Decide on repo strategy** *(0.25 h)*
  - **Option A (recommended):** new branch `v2-rebuild` in same repo, eventual force-push to `main` when ready.
  - **Option B:** new GitHub repo `cms-plazareal-v2`. More isolation, but Fly app + DNS reconfiguration needed.
  - **Option C:** orphan branch `git checkout --orphan v2`. Same repo, no parent commit.

- [ ] **Task 0.2: Init module + base directory tree** *(0.5 h)*
  Final tree (per current production layout, slightly tightened):
  ```
  cmd/
    server/main.go
    scraper/main.go
  internal/
    auth/         # collections, hooks, jwt
    config/
    handlers/
      admin/      # one file per resource (tiendas.go, content.go, locales.go, …)
      web/        # public pages
      fragments/  # htmx fragments
      ws/
    middleware/
    realtime/
    services/
      r2/
    view/
      layout/
      components/
      fragments/
      pages/
        admin/
        public/
  web/static/
  components/     # templUI vendored components (kept verbatim)
  ```
  `go mod init cms-plazareal`. Add Fiber, templ, PocketBase, R2 SDK dependencies.

- [ ] **Task 0.3: Tooling: Tailwind + templ + Makefile + hot-reload** *(0.5 h)*
  - `tailwind.config.js`, `postcss.config.js`, `package.json` for tailwind compile
  - `Makefile` with `dev`, `build`, `templ`, `tailwind` targets
  - `templ generate --watch` integration

- [ ] **Task 0.4: Hello world** *(0.25 h)*
  Minimal `cmd/server/main.go` that starts Fiber + PocketBase, serves a templ "Plaza Real CMS" page on `/`. Build, run, hit `localhost:3000`. Commit.

- [ ] **Task 0.5: Dockerfile + fly.toml carry-over** *(0.5 h)*
  Copy from current repo. Verify image builds locally.

---

## Phase 1: Config + PocketBase collections — ~4 h

**Goal:** PB starts, all 8 collections exist with the right schema. Seed data optional.

- [ ] **Task 1.1: Config loader** *(0.5 h)*
  Env-driven config (port, R2 keys, JWT secret, CORS origins, etc.). Single struct.

- [ ] **Task 1.2: PB init + RegisterPBHooks** *(0.5 h)*
  Wire PocketBase startup. `auth.RegisterPBHooks(pb)` skeleton.

- [ ] **Task 1.3: Define collections** *(2.5 h)*
  Eight collections, each with field list, indexes, validation, access rules. Translate from current `internal/auth/collections.go`:
  - `users` (extends PB `users`): adds `nombre`, `role` (`superadmin|director|admin|editor`)
  - `tiendas`: nombre, slug, categoria, gal (galería A–F), local_num, logo, photos[], tags[], descripcion, about, about2, payment_methods, similar_stores, whatsapp, telefono, rating, horarios, status, destacada
  - `content_blocks`: title, body, category (NOTICIA/COMUNICADO/PROMOCION), status, published_at, expires_at, image, slug, template
  - `locales_disponibles`: numero, area_m2, estado, descripcion, photos
  - `reservas`: nombre, contacto, local_id (rel), fecha, estado (pendiente/confirmada/cancelada), notas
  - `site_settings`: hero_bg, search_bg, mall_open_hours, contact info
  - `page_views`: path, referrer, user_agent, ip_hash, ts, session_id
  - `leads`: source, name, email, phone, message, ts

- [ ] **Task 1.4: Seed dev data** *(0.5 h)*
  Idempotent seed for tiendas (~10 sample stores) + 1 superadmin user. Skip if PB already has records.

---

## Phase 2: Auth — ~3 h

**Goal:** Login → cookie session → admin routes gated by role.

- [ ] **Task 2.1: JWT generation + validation** *(1 h)*
  `internal/auth/jwt.go`. Issuer `cms-plazareal`. 24h TTL. HS256.

- [ ] **Task 2.2: Login templ page** *(0.5 h)*
  `view/pages/admin/login.templ`. Liquid Glass dark theme. Email + password fields. HTMX submit.

- [ ] **Task 2.3: Login handler** *(0.5 h)*
  POST `/admin/login`. Verify against PB `users` collection. Issue JWT, set HttpOnly cookie `pr_token`. Redirect via `HX-Redirect` header.

- [ ] **Task 2.4: Auth middleware** *(0.5 h)*
  `middleware.AuthRequired(cfg)` reads cookie, validates, attaches claims to ctx. Redirects to `/admin/login` on failure.

- [ ] **Task 2.5: Role middleware + logout** *(0.5 h)*
  `middleware.RoleRequired("superadmin", ...)`. POST `/admin/logout` clears cookie.

---

## Phase 3: Admin shell — ~3 h

**Goal:** After login, admin layout renders with sidebar nav + topbar; dashboard shows stats.

- [ ] **Task 3.1: Admin layout templ** *(1 h)*
  `view/layout/admin.templ`. Sidebar (Tiendas, Noticias, Comunicados, Promociones, Locales, Reservas, Usuarios, Reportes, Ajustes). Topbar with user menu. Liquid Glass styling.

- [ ] **Task 3.2: Dashboard page** *(1 h)*
  `view/pages/admin/dashboard.templ`. Stat cards: total tiendas, content_blocks last 30d, reservas pending, recent leads.

- [ ] **Task 3.3: DashboardStats fragment** *(0.5 h)*
  HTMX-loaded stats endpoint. Async, refreshable.

- [ ] **Task 3.4: HTMX wiring** *(0.5 h)*
  Toast component, modal helper, error boundary. Empty `/admin/empty` route for safe HTMX modal close.

---

## Phase 4: Admin CRUDs — ~12 h

**Goal:** All admin sections functional with create/edit/delete + photo upload where applicable.

- [ ] **Task 4.1: Tiendas CRUD** *(4 h)*
  - List page (filters: categoria, gal, status, destacada)
  - Form (~25 fields incl. photo array, similar stores, payment methods)
  - R2 photo upload integration
  - Toggle published/draft

- [ ] **Task 4.2: Content blocks CRUD** *(2 h)*
  Single CRUD reused for NOTICIA/COMUNICADO/PROMOCION. Date pickers, status, image upload.

- [ ] **Task 4.3: Locales disponibles CRUD** *(1.5 h)*
  Form: numero, area, estado, fotos.

- [ ] **Task 4.4: Reservas list + actions** *(1 h)*
  Read-only-ish: list, change estado, delete. No new/edit (reservas come from public form).

- [ ] **Task 4.5: Users CRUD (superadmin)** *(1 h)*
  Create/edit users with role, nombre, password.

- [ ] **Task 4.6: Site settings page** *(1 h)*
  Hero bg image, search bg image, mall info. R2 upload.

- [ ] **Task 4.7: Reports + CSV export** *(1.5 h)*
  Stats query + downloadable CSV (page_views, reservas, leads).

---

## Phase 5: Public pages + fragments — ~6 h

**Goal:** All public-facing pages render via templ; HTMX fragments populate dynamic sections.

- [ ] **Task 5.1: Home (`/`)** *(1 h)*
  Hero carousel (from content_blocks + site_settings) + featured tiendas + locales fragment + comunicados fragment.

- [ ] **Task 5.2: Tiendas (`/buscador-tiendas`)** *(1.5 h)*
  Filters (categoria, gal), grid, marquee. Server-side filtering with HTMX fragment swap.

- [ ] **Task 5.3: Tienda detail (`/tiendas/:slug`)** *(1 h)*
  Per-store page with photos carousel, info, similar stores.

- [ ] **Task 5.4: Noticias / Comunicados / Promociones** *(1.5 h)*
  Three list pages, three card-fragment endpoints. Pagination optional.

- [ ] **Task 5.5: Locales disponibles + reserva form** *(0.5 h)*
  List + form that creates a `reservas` record.

- [ ] **Task 5.6: RSS feed** *(0.25 h)*
  `/rss.xml` from content_blocks NOTICIA.

- [ ] **Task 5.7: Static assets + favicon** *(0.25 h)*
  Plaza Real favicon, logos, og-image.

---

## Phase 6: Cross-cutting integrations — ~5 h

- [ ] **Task 6.1: R2 upload service + UploadField widget** *(2 h)*
  R2 client init, `POST /admin/upload`, presigned URL flow, drag-drop templ component with progress bar.

- [ ] **Task 6.2: Analytics middleware** *(1.5 h)*
  Per-request page_views write. Skip `/admin`, `/frag*`, `/static`, `/api`, `/ws`. Skip 4xx/5xx. Async, fail-open.

- [ ] **Task 6.3: Real-time hub + WebSocket** *(1 h)*
  Web-only broadcast (no devices). PB hooks on `content_blocks`, `tiendas` → push update event to connected web clients for live admin and (optional) public refresh.

- [ ] **Task 6.4: Scraper CLI** *(0.5 h)*
  Port `cmd/scraper/` from current repo verbatim — it's standalone and already works.

---

## Phase 7: Data migration + cutover — ~3 h

**Goal:** New code reads existing production data without loss.

- [ ] **Task 7.1: Volume strategy** *(0.5 h)*
  Decide: (a) reuse existing Fly volume `pb_data` by deploying v2 to same app, or (b) snapshot + restore to new app. Recommend (a).

- [ ] **Task 7.2: Schema compatibility check** *(1 h)*
  Run new code against a *copy* of production `pb_data`. Verify all collections work, no migration required. If new collection definitions disagree with prod schema, write idempotent migrations in `auth.RegisterPBHooks`.

- [ ] **Task 7.3: Smoke test against prod data** *(0.5 h)*
  Local server pointed at copied `pb_data`. Walk every public + admin page. Verify no missing fields, no panics.

- [ ] **Task 7.4: Deploy** *(0.5 h)*
  `fly deploy`. Watch v26 deploy.

- [ ] **Task 7.5: Production smoke** *(0.5 h)*
  All public + admin pages. Compare to v25 visually. Look for regressions.

---

## Phase 8: Cleanup of old repo — ~1 h

- [ ] **Task 8.1: Force-push v2 to main** *(0.25 h)*
  Either `git push --force-with-lease origin v2:main` (squashing history) or merge v2 → main + force-push.

- [ ] **Task 8.2: Tag old state** *(0.25 h)*
  Tag the pre-rebuild head as `legacy-cms-final` for reference.

- [ ] **Task 8.3: Delete v2 branch** *(0.1 h)*
  After merge.

- [ ] **Task 8.4: Update README, docs** *(0.4 h)*

---

## Total estimate

| Phase | Hours |
|---|---|
| 0 — Scaffold | 2 |
| 1 — Config + PB collections | 4 |
| 2 — Auth | 3 |
| 3 — Admin shell | 3 |
| 4 — Admin CRUDs | 12 |
| 5 — Public pages | 6 |
| 6 — Integrations | 5 |
| 7 — Data + cutover | 3 |
| 8 — Cleanup | 1 |
| **Floor** | **~39 h** |
| **Realistic with debugging** | **50–60 h** |

---

## Risks specific to rebuild

1. **Subtle field semantics get re-derived wrong.** Things like the `gal` letter→galleria mapping in `tiendas`, `content_blocks.template` field, `page_views.session_id` rotation rules, `reservas.estado` transitions — all encoded as institutional knowledge in the current code. Re-deriving them from memory or by re-reading old code defeats the "from scratch" benefit.

2. **PocketBase migration drift.** The current `auth/collections.go` has both the schema AND idempotent migrations that fix prior schema mistakes (`migrateContentBlocks`, `migrateContentBlocksTemplate`, `migrateTiendasStatusHorario`). Production `pb_data` has data shaped by those migrations. If the new code defines collections fresh without the migrations, it may load production records with type errors.

3. **Liquid Glass + templUI overrides.** The CSS files contain ~3000 lines of design tokens, dark-mode rules, accessibility tweaks. Copying them verbatim is the smart move, but at that point you're admitting the rebuild isn't really from scratch.

4. **Auth cookie name + JWT issuer must match.** Existing logged-in admin users have `pr_token` cookies signed with the current issuer/secret. A rebuild changing either will log everyone out and may invalidate refresh flows. Not catastrophic but worth flagging.

5. **Contiguity of production.** The rebuild can't deploy until feature-complete. That's days of branch divergence where the current main keeps getting bug fixes (chips/categorías regression, etc.). Merge cost grows daily.

---

## Comparison with strip-down

| Dimension | Strip-down (`2026-05-05-strip-down-legacy-cruft.md`) | Rebuild from scratch (this plan) |
|---|---|---|
| Time to clean repo | **2.5 h** | 30–60 h |
| Implementation risk | Low (deletes dead code) | High (rewriting working logic re-introduces bugs) |
| Production downtime risk | Near zero (deploy is a small Docker rebuild) | Real risk (auth changes, schema drift) |
| Templ work preserved | Yes | No, must rewrite |
| Liquid Glass design | Yes | Copy or rewrite (1–2 h either way) |
| Scraper CLI | Yes | Copy verbatim |
| Auth + JWT + middleware | Yes | Rewrite (+ user re-login) |
| PB collections | Yes (defined as code) | Rewrite (with drift risk) |
| Real-time hub | Yes (already minus device hooks after strip) | Rewrite |
| Git history | Incremental, traceable diffs across 7 commits | Single squashed commit (clean) or full new repo |
| "Fresh repo feel" | Mostly — directory tree pre-existing | Yes, completely new layout possible |
| Categorías regression | Still must be fixed in either path | Same |
| Reversibility | Trivial (`git reset --hard pre-stripdown-2026-05-05`) | Difficult once new collections exist |

---

## Hybrid: orphan-branch clean slate (~3–5 h)

If the appeal of "from scratch" is the **clean git history feeling**, not actual rewriting, this hybrid achieves it:

- [ ] **Task H.1: Orphan branch** *(0.25 h)*
  ```bash
  git checkout --orphan v2-clean
  git rm -rf .
  ```

- [ ] **Task H.2: Selectively copy keepers from `pre-stripdown-2026-05-05` tag** *(2 h)*
  Use `git checkout pre-stripdown-2026-05-05 -- <path>` for each path on a curated keep-list:
  - `cmd/server/`, `cmd/scraper/`
  - `internal/auth/` (minus legacy seeders — manually edit `collections.go`)
  - `internal/config/`, `internal/middleware/`, `internal/realtime/` (minus device code)
  - `internal/handlers/admin/handlers.go` (only kept funcs — pre-strip via Edit)
  - `internal/handlers/web/handlers.go` (only kept funcs)
  - `internal/handlers/fragments/`, `internal/handlers/ws/`
  - `internal/services/r2/`
  - `internal/view/` (entire tree — already clean)
  - `web/static/`, `web/{index,buscador-tiendas,noticias}.html`
  - `components/`
  - `Makefile`, `Dockerfile`, `fly.toml`, `go.mod`, `go.sum`, `.gitignore`, `tailwind.config.js`, `package.json`

- [ ] **Task H.3: Build + smoke** *(0.5 h)*
  `templ generate && go build ./... && go run ./cmd/server`. Test golden paths.

- [ ] **Task H.4: Single commit** *(0.25 h)*
  ```bash
  git add -A
  git commit -m "Plaza Real CMS — clean slate from production v25"
  ```

- [ ] **Task H.5: Tag old + force-push** *(0.5 h)*
  ```bash
  git tag legacy-cms-final main
  git push origin legacy-cms-final
  git branch -M v2-clean main
  git push --force-with-lease origin main
  ```

- [ ] **Task H.6: Deploy + verify prod** *(0.5–1 h)*

**Total: 3–5 h.** Same end state as strip-down + Plan B's clean-history feel, none of the rewrite cost. Only loss: the 19-commit history with their messages becomes inaccessible from `main` (still reachable via `legacy-cms-final` tag).

---

## Recommendation

For "*horas no semanas*" with this codebase: **strip-down (Plan A)** is the right answer. If you specifically want the clean-history feeling at the end, **add the hybrid (Plan H) as a final step** after strip-down: ~3 h more, single commit on top of an orphan branch, tag the old history.

True from-scratch (this plan, Phases 0–8) only makes sense if the goal is to *change architecture*, not just clean cruft — e.g., moving away from PocketBase, switching from Fiber to another router, dropping templ, etc. None of those motivations are present in the conversation.
