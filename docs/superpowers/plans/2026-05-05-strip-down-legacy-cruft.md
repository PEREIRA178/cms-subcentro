# Strip-Down Legacy Cruft Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove all code and templates from two legacy app generations (devices/playlists/multimedia/whatsapp CMS, propiedades/JCP intermediate phase) leaving only Plaza Real templ-based code.

**Architecture:** Pure deletion. No new functionality. After each phase, `go build ./...` must pass. Plaza Real golden paths (public homepage, admin login/dashboard/tiendas/locales/reservas/content/reports/settings) must keep working. Production deploy via `fly deploy` at the end.

**Tech Stack:** Go (Fiber), templ, templUI, PocketBase, R2, Tailwind. No tests in repo — verification = `go build ./...` + `templ generate` + manual smoke test.

**Time budget:** ~2.5 hours.

**What gets deleted (high level):**
- Admin handlers: `Multimedia*`, `Events*`, `News*`, `Playlist*`, `Device*`, `WhatsAppLogs`, plus helpers `eventFormHTML`, `newsFormHTML`, `playlistOptsHTML`, `savePlItems`, `buildContentPool`, `playlistEditorHTML`
- Web handlers: `DeviceDisplay`, `TotemDisplay`, `WhatsAppWebhook`, `NoticiaHandler`
- API handlers: `DevicePlaylist`, `UpcomingEvents` (entire `internal/handlers/api/` package)
- WebSocket: `DeviceSocket` (keep web `WebSocket` for content broadcasts)
- Realtime: `BroadcastToDevice` + device/playlist/multimedia PB hooks (keep web broadcast)
- Services: `internal/services/playlist.go`, `internal/services/whatsapp/`
- Auth/collections: `SeedDevicesAndPlaylists`, `migrateDevicesCurrentView`, `migratePlaylistItemsContentBlockID`, `migrateMultimediaStartTime`
- Templates: 7 admin HTML pages + `internal/templates/devices/` + `internal/templates/web/noticia.html` + `web/tienda-individual.html`
- Routes in `cmd/server/main.go` for everything above
- 5 stale remote branches `origin/claude/*`

**Out of scope (separate plans):**
- Bringing back categorías chips (regression noted but not addressed here)
- Migrating remaining `web/*.html` static pages (`index.html`, `buscador-tiendas.html`, `noticias.html`) fully to templ
- Migrating Tiendas admin form helper `tiendaFormHTML` to templ
- Cleaning up legacy PocketBase collections (`devices`, `playlists`, `playlist_items`, `multimedia`) from `pb_data` — DB cleanup is manual via PocketBase admin UI

---

## Task 1: Setup & safety net

**Files:** none modified — this is git-only setup.

- [ ] **Step 1.1: Confirm clean working tree**

Run: `git status`
Expected: `nothing to commit, working tree clean` on branch `main`.

- [ ] **Step 1.2: Create rollback tag**

Run:
```bash
git tag pre-stripdown-2026-05-05 -m "Snapshot before legacy cruft strip-down"
git push origin pre-stripdown-2026-05-05
```
Expected: tag created locally and pushed to origin. To rollback later: `git reset --hard pre-stripdown-2026-05-05 && git push --force-with-lease`.

- [ ] **Step 1.3: Run baseline build**

Run:
```bash
templ generate && go build ./...
```
Expected: both commands exit 0. If they fail, **stop** — the baseline is broken and the plan can't safely proceed.

- [ ] **Step 1.4: Delete 5 stale remote PR branches**

Run:
```bash
git push origin --delete claude/add-persistent-volume-mount-5kgg2 claude/bulk-store-upload-Ee4UH claude/fix-stores-tab-emoji-i856k claude/update-galleries-branding-YwgbE claude/update-go-jwt-QUlb8
git remote prune origin
```
Expected: 5 `deleted` lines, prune confirms locally. No local commit produced (this is remote-only).

---

## Task 2: Strip legacy admin handlers (multimedia + events + news)

**Files:**
- Modify: `internal/handlers/admin/handlers.go` (delete ~600 lines covering 3 legacy sections)

- [ ] **Step 2.1: Delete Multimedia section**

In `internal/handlers/admin/handlers.go`, delete the entire block from `func MultimediaList(` (around line 149) through the closing `}` of `func MultimediaDelete(` (around line 334). The block contains 6 functions: `MultimediaList`, `MultimediaForm`, `MultimediaCreate`, `MultimediaEdit`, `MultimediaUpdate`, `MultimediaDelete`.

Use a single Edit replacing the full block with an empty string. Use the line preceding `MultimediaList` (last line of `DashboardStats`) as the unique anchor on the upper boundary, and the line following `MultimediaDelete`'s closing brace (start of `EventsList`'s comment or signature) as the lower boundary.

- [ ] **Step 2.2: Build to confirm Multimedia is severed cleanly**

Run: `go build ./...`
Expected: errors will appear in `cmd/server/main.go` referencing `admin.MultimediaList` (and the other 5 deleted funcs). Those errors are expected — they get fixed in Task 8 when we strip routes from `main.go`. **Do NOT proceed past this step until you've read the build output and confirmed the only errors are unresolved references to deleted functions.** Any other error means the deletion took too much code.

- [ ] **Step 2.3: Delete Events section**

Delete the block from `func EventsList(` (around line 336) through the closing `}` of `func eventFormHTML(` (around line 590). Includes: `EventsList`, `EventForm`, `EventCreate`, `EventEdit`, `EventUpdate`, `EventDelete`, `EventPublish`, helper `eventFormHTML`.

- [ ] **Step 2.4: Delete News section**

Delete the block from `func newsFormHTML(` (around line 592) through the closing `}` of `func NewsDelete(` (around line 749). Includes: helper `newsFormHTML`, `NewsList`, `NewsForm`, `NewsCreate`, `NewsEdit`, `NewsUpdate`, `NewsDelete`.

- [ ] **Step 2.5: Verify build still has only the expected reference errors**

Run: `go build ./... 2>&1 | head -40`
Expected: only `undefined: admin.MultimediaList`, `admin.EventsList`, `admin.NewsList`, etc. — no other errors.

- [ ] **Step 2.6: Commit**

```bash
git add internal/handlers/admin/handlers.go
git commit -m "refactor: strip legacy multimedia/events/news admin handlers

Plaza Real uses content_blocks (NOTICIA/COMUNICADO/PROMOCION) via
ContentList* handlers — the standalone Multimedia, Events, and News
sections were leftover from the original CMS app and never wired
into the new templ admin. Routes removed in a follow-up commit."
```

---

## Task 3: Strip legacy admin handlers (playlists + devices + whatsapp + helpers)

**Files:**
- Modify: `internal/handlers/admin/handlers.go` (delete ~750 more lines)

- [ ] **Step 3.1: Delete Playlists section**

Delete the block from `func PlaylistList(` (around line 751) through the closing `}` of `func PlaylistReorder(` (around line 925). Includes 7 functions.

- [ ] **Step 3.2: Delete Devices section**

Delete the block from `func DeviceList(` (around line 928) through the closing `}` of `func DeviceAssignPlaylist(` (around line 1185). Includes 7 functions.

- [ ] **Step 3.3: Delete WhatsAppLogs**

Delete `func WhatsAppLogs(cfg *config.Config) fiber.Handler { ... }` (around lines 1333-1339, ~7 lines).

- [ ] **Step 3.4: Delete playlist/content-pool helpers**

Delete the block from `func playlistOptsHTML(` (around line 1682) through the closing `}` of `func playlistEditorHTML(` (around line 1986). Includes: `playlistOptsHTML`, `savePlItems`, `buildContentPool`, `playlistEditorHTML`. **Verify** before deleting that none of the kept admin handlers (`TiendasList`, `Content*`, `Locales*`, `Reservas*`, `Reports*`, `Settings*`) call these helpers — grep first:

```bash
grep -nE "playlistOptsHTML|savePlItems|buildContentPool|playlistEditorHTML" internal/handlers/admin/handlers.go
```

Expected: only the function definitions themselves match, plus possibly references inside the *to-be-deleted* Playlist/Device sections (which by this point are already gone, so the only matches should be the definitions). If any other handler references them, **stop and update this plan** — those callers are also legacy and must be added to the strip list.

- [ ] **Step 3.5: Verify build (errors should still only be unresolved admin references in main.go)**

Run: `go build ./... 2>&1 | head -60`
Expected: errors are `undefined: admin.PlaylistList`, `admin.DeviceList`, `admin.WhatsAppLogs`, etc. plus the Multimedia/Events/News from Task 2. Nothing else.

- [ ] **Step 3.6: Verify line count reduction**

Run: `wc -l internal/handlers/admin/handlers.go`
Expected: file dropped from ~2622 lines to ~1480 lines (~1140 deleted).

- [ ] **Step 3.7: Commit**

```bash
git add internal/handlers/admin/handlers.go
git commit -m "refactor: strip legacy playlist/device/whatsapp admin handlers

Plaza Real has no signage/playlist/device features — these were
artifacts of the original CMS. Also drops the playlist content-pool
helpers (playlistOptsHTML, savePlItems, buildContentPool,
playlistEditorHTML). Admin handlers file: 2622 → ~1480 lines."
```

---

## Task 4: Strip legacy web + api + ws handlers

**Files:**
- Modify: `internal/handlers/web/handlers.go`
- Delete: `internal/handlers/api/handlers.go` (entire file)
- Modify: `internal/handlers/ws/websocket.go` (remove `DeviceSocket`, keep `WebSocket`)

- [ ] **Step 4.1: Delete legacy web handlers**

In `internal/handlers/web/handlers.go`, delete:
- `func DeviceDisplay(` block (around lines 177-186)
- `func TotemDisplay(` block (around lines 188-197)
- `func WhatsAppWebhook(` block (around lines 230-235)
- `func NoticiaHandler(` block (around lines 237-303) — uses deleted noticia.html template

Keep: `IndexHandler`, `TiendasPageHandler`, `NoticiasPageHandler`, `ComunicadosPageHandler`, `LocalesPageHandler`, `PromocionesPageHandler`, `TiendaDetailHandler`, `PageHandler`, `RSSFeed`, plus helpers `splitCSV`, `sanitizeSlug`, `computeOpenStatus`.

- [ ] **Step 4.2: Verify imports in web/handlers.go are still used**

After the deletions, check which imports may have become unused (e.g., if only `NoticiaHandler` used a particular import). Run:

```bash
go build ./internal/handlers/web/ 2>&1
```

If the compiler reports unused imports, remove them with Edit. Otherwise proceed.

- [ ] **Step 4.3: Delete entire api/handlers.go file**

`api/handlers.go` contains only `DevicePlaylist` and `UpcomingEvents` — both legacy. Delete the file:

```bash
git rm internal/handlers/api/handlers.go
```

The `internal/handlers/api/` directory will be empty afterward — leave it (harmless) or `rmdir`. The import in `cmd/server/main.go` (`apiHandlers "cms-plazareal/internal/handlers/api"`) will fail to resolve; that gets removed in Task 8.

- [ ] **Step 4.4: Strip DeviceSocket from ws/websocket.go**

In `internal/handlers/ws/websocket.go`, delete `func DeviceSocket(hub *realtime.Hub) func(*websocket.Conn) { ... }` (around lines 13-65). Keep `func WebSocket(hub *realtime.Hub) func(*websocket.Conn)`. Verify no other code in the file is removed.

- [ ] **Step 4.5: Verify build**

Run: `go build ./... 2>&1 | head -40`
Expected: errors are `undefined: ws.DeviceSocket`, `apiHandlers.DevicePlaylist`, `apiHandlers.UpcomingEvents`, `web.DeviceDisplay`, `web.TotemDisplay`, `web.WhatsAppWebhook`, `web.NoticiaHandler` — all in `main.go`. Nothing else.

- [ ] **Step 4.6: Commit**

```bash
git add internal/handlers/
git commit -m "refactor: strip legacy device/whatsapp/api handlers

- web/handlers.go: drop DeviceDisplay, TotemDisplay,
  WhatsAppWebhook, NoticiaHandler (the latter served per-article
  HTML view, superseded conceptually by templ public pages).
- api/handlers.go: deleted entirely (DevicePlaylist, UpcomingEvents
  served the totem players that no longer exist).
- ws/websocket.go: drop DeviceSocket; keep web WebSocket for
  content-update broadcasts."
```

---

## Task 5: Strip legacy realtime hooks + delete legacy services

**Files:**
- Modify: `internal/realtime/hub.go` (remove device-targeted broadcast + device/playlist/multimedia PB hooks)
- Delete: `internal/services/playlist.go`
- Delete: `internal/services/whatsapp/` (entire directory)

- [ ] **Step 5.1: Audit realtime/hub.go for device-specific code**

Read `internal/realtime/hub.go` end-to-end to identify exactly what's device-specific. Specifically:
- `BroadcastToDevice` method (around line 162) — DELETE
- The portion of `RegisterPBHooks` (around line 177) that subscribes to `devices`, `playlists`, `playlist_items`, `multimedia` collections — DELETE that block, keep web-content broadcast hooks (for `content_blocks`, `tiendas`, etc. if any)
- `BroadcastAll`, `BroadcastWeb`, `Run`, `Register`, `Unregister`, `NewHub`, `SetHubInstance` — KEEP

If any device-specific message types are defined (constants, structs), delete those too — but only after confirming with grep that no kept code references them.

- [ ] **Step 5.2: Apply realtime hub edits**

Edit `internal/realtime/hub.go` per the audit. Show the engineer doing this: after editing, run `go vet ./internal/realtime/` to catch typos, then `go build ./...` and confirm no NEW errors appear in `realtime/` — only the still-pending `main.go` errors from earlier tasks.

- [ ] **Step 5.3: Delete services/playlist.go**

```bash
git rm internal/services/playlist.go
```

- [ ] **Step 5.4: Delete services/whatsapp/**

```bash
git rm -r internal/services/whatsapp/
```

- [ ] **Step 5.5: Verify build**

Run: `go build ./... 2>&1 | head -40`
Expected: still the cluster of `main.go` errors from earlier tasks. No new errors in `realtime/`, `services/`, or anywhere else.

- [ ] **Step 5.6: Commit**

```bash
git add internal/realtime/hub.go internal/services/
git commit -m "refactor: strip legacy realtime device hooks + services

- realtime/hub.go: drop BroadcastToDevice and PB hooks for
  devices/playlists/playlist_items/multimedia. Web content
  broadcast hooks retained.
- services/playlist.go: deleted (consumed by deleted playlist
  admin handlers).
- services/whatsapp/: deleted (consumed by deleted whatsapp
  webhook + admin logs)."
```

---

## Task 6: Strip legacy collection setup from auth/collections.go

**Files:**
- Modify: `internal/auth/collections.go` (remove device/playlist/multimedia seeding + obsolete migrations)

- [ ] **Step 6.1: Identify what to delete**

In `internal/auth/collections.go`, the following functions are tied to legacy collections:
- `SeedDevicesAndPlaylists` (line 635) — DELETE
- `migrateDevicesCurrentView` (line 550) — DELETE
- `migratePlaylistItemsContentBlockID` (line 565) — DELETE
- `migrateMultimediaStartTime` (line 580) — DELETE

KEEP:
- `RegisterPBHooks` (will need its body edited to stop calling deleted functions)
- `ensureCollections` (will need edits to stop ensuring legacy collections — but only if it currently does)
- `ensureSiteSettings`, `ensurePageViews`, `ensureLeads`, `ensureLocalesDisponibles`, `ensureReservas` — Plaza Real collections
- `migrateContentBlocks`, `seedContentBlocks`, `migrateContentBlocksTemplate`, `migrateTiendasStatusHorario`, `seedTiendas` — Plaza Real data

- [ ] **Step 6.2: Read collections.go and locate the call sites**

Open the file and locate:
1. The body of `RegisterPBHooks` — find calls to the 4 functions above and remove those calls.
2. The body of `ensureCollections` — find any code that creates `devices`, `playlists`, `playlist_items`, `multimedia` collections. Plaza Real doesn't need them. Remove if present. **Caution:** if these collections already exist in production `pb_data` (they do), removing the `ensureCollections` code does NOT drop the collections — it only stops re-creating them on a fresh deploy. That's safe.

- [ ] **Step 6.3: Delete the 4 legacy migration/seed functions**

Use Edit to remove each function block. Order doesn't matter.

- [ ] **Step 6.4: Verify build**

```bash
go build ./internal/auth/ && go build ./...
```
Expected: `auth` package builds clean. `main.go` errors persist (handled in Task 8).

- [ ] **Step 6.5: Commit**

```bash
git add internal/auth/collections.go
git commit -m "refactor: drop legacy collection seeding + migrations

Removes SeedDevicesAndPlaylists and migrations for
devices/playlists/multimedia/playlist_items collections.
Existing pb_data records remain — cleanup is via PocketBase admin
UI when convenient."
```

---

## Task 7: Delete orphan HTML templates

**Files:**
- Delete: 12 HTML template files

- [ ] **Step 7.1: Delete templates with confirmed no Go reference**

These have been confirmed orphan in the prior discovery (no Go file references them):
```bash
git rm internal/templates/admin/pages/login.html
git rm internal/templates/admin/pages/users.html
git rm internal/templates/admin/pages/dashboard.html
git rm internal/templates/admin/pages/tiendas.html
git rm web/tienda-individual.html
```

(Note: `dashboard.html` and `tiendas.html` ARE referenced in `admin/handlers.go`, but those references live inside handlers that have either been migrated to templ or that we've already deleted. After Tasks 2 and 3, the only references that remain are dead string literals — confirm with grep below.)

- [ ] **Step 7.2: Verify dashboard.html / tiendas.html really are dead references**

```bash
grep -nE 'dashboard\.html|tiendas\.html' internal/handlers/admin/handlers.go
```

If grep finds matches, read those lines: if they're inside handlers that load the file (`os.ReadFile`, `c.SendFile`, `template.ParseFiles`), the handler is still using them — **stop, do not delete those HTML files**, and instead migrate the handler to use the templ counterpart (`Dashboard()` should render `dashboard.templ`, etc.). This is real migration work; flag for a separate task. If grep finds no matches, proceed.

- [ ] **Step 7.3: Delete templates whose handlers are being deleted in Task 8**

These will become orphan once we strip routes/handlers in Task 8. Delete now:
```bash
git rm internal/templates/admin/pages/devices.html
git rm internal/templates/admin/pages/playlists.html
git rm internal/templates/admin/pages/multimedia.html
git rm internal/templates/admin/pages/whatsapp-logs.html
git rm internal/templates/admin/pages/events.html
git rm internal/templates/admin/pages/news.html
git rm -r internal/templates/devices/
git rm internal/templates/web/noticia.html
```

- [ ] **Step 7.4: Verify only Plaza Real-relevant HTML remains**

```bash
git ls-files | grep -E '\.html$'
```
Expected output (5 files):
```
internal/templates/web/noticia.html  ← if not deleted, that's a problem
web/buscador-tiendas.html
web/index.html
web/noticias.html
```

If `noticia.html` survived, delete again. If `tienda-individual.html` survived, delete. If admin/pages HTML survived, delete.

- [ ] **Step 7.5: Commit**

```bash
git add -A
git commit -m "chore: delete 12 orphan HTML templates

All admin pages migrated to templ (login, dashboard, users,
tiendas) or attached to deleted legacy handlers (devices,
playlists, multimedia, events, news, whatsapp-logs).
Devices subsystem templates (display.html, totem.html) gone.
tienda-individual.html → tienda_detail.templ.
noticia.html (per-article view) deferred to a future templ task."
```

---

## Task 8: Strip legacy routes from cmd/server/main.go

**Files:**
- Modify: `cmd/server/main.go`

- [ ] **Step 8.1: Read the file**

Re-read `cmd/server/main.go` to confirm route layout matches what's in this plan (lines may have shifted slightly with no source changes — this file hasn't been edited in earlier tasks).

- [ ] **Step 8.2: Remove legacy import**

In the import block (lines 9-16 region), delete the line:
```go
apiHandlers "cms-plazareal/internal/handlers/api"
```

- [ ] **Step 8.3: Remove `eventos.html` alias route**

Delete this single line (around line 96) — it's a leftover alias to `PromocionesPageHandler`:
```go
app.Get("/eventos.html", web.PromocionesPageHandler(cfg))
```
(The canonical route `/promociones` and the templ page survive.)

- [ ] **Step 8.4: Remove device API routes**

Delete the API group block (around lines 123-126):
```go
// ── PUBLIC API ──
api := app.Group("/api")
api.Get("/devices/:code/playlist", apiHandlers.DevicePlaylist(cfg, pb))
api.Get("/events/upcoming", apiHandlers.UpcomingEvents(cfg, pb))
```

- [ ] **Step 8.5: Remove device/totem display routes + device WebSocket**

In the DEVICE / WS block (lines 128-138), delete:
```go
app.Get("/display/:code", web.DeviceDisplay(cfg))
app.Get("/totem/:code", web.TotemDisplay(cfg))
```
And:
```go
app.Get("/ws/device/:code", gows.New(ws.DeviceSocket(hub)))
```

KEEP the rest of the WS block: the `app.Use("/ws", ...)` upgrade gate and `app.Get("/ws/web", gows.New(ws.WebSocket(hub)))`. Update the section comment from `// ── DEVICE / WS ──` to `// ── WS ──`.

- [ ] **Step 8.6: Remove admin Multimedia routes**

Delete (around lines 159-165):
```go
// Multimedia
adm.Get("/multimedia", admin.MultimediaList(cfg, pb))
adm.Get("/multimedia/new", admin.MultimediaForm(cfg))
adm.Post("/multimedia", admin.MultimediaCreate(cfg, pb))
adm.Get("/multimedia/:id/edit", admin.MultimediaEdit(cfg, pb))
adm.Put("/multimedia/:id", admin.MultimediaUpdate(cfg, pb))
adm.Delete("/multimedia/:id", admin.MultimediaDelete(cfg, pb))
```

- [ ] **Step 8.7: Remove admin Events routes**

Delete (around lines 167-174):
```go
// Events (content_blocks excl. NOTICIA)
adm.Get("/events", admin.EventsList(cfg, pb))
adm.Get("/events/new", admin.EventForm(cfg))
adm.Post("/events", admin.EventCreate(cfg, pb))
adm.Get("/events/:id/edit", admin.EventEdit(cfg, pb))
adm.Put("/events/:id", admin.EventUpdate(cfg, pb))
adm.Delete("/events/:id", admin.EventDelete(cfg, pb))
adm.Post("/events/:id/publish", admin.EventPublish(cfg, pb))
```

- [ ] **Step 8.8: Remove admin News routes**

Delete (around lines 176-182):
```go
// News (content_blocks category=NOTICIA)
adm.Get("/news", admin.NewsList(cfg, pb))
adm.Get("/news/new", admin.NewsForm(cfg))
adm.Post("/news", admin.NewsCreate(cfg, pb))
adm.Get("/news/:id/edit", admin.NewsEdit(cfg, pb))
adm.Put("/news/:id", admin.NewsUpdate(cfg, pb))
adm.Delete("/news/:id", admin.NewsDelete(cfg, pb))
```

- [ ] **Step 8.9: Remove admin Playlists routes**

Delete (around lines 198-205):
```go
// Playlists
adm.Get("/playlists", admin.PlaylistList(cfg, pb))
adm.Get("/playlists/new", admin.PlaylistForm(cfg, pb))
adm.Post("/playlists", admin.PlaylistCreate(cfg, pb))
adm.Get("/playlists/:id/edit", admin.PlaylistEdit(cfg, pb))
adm.Put("/playlists/:id", admin.PlaylistUpdate(cfg, pb))
adm.Delete("/playlists/:id", admin.PlaylistDelete(cfg, pb))
adm.Post("/playlists/:id/items/reorder", admin.PlaylistReorder(cfg, pb))
```

- [ ] **Step 8.10: Remove admin Devices routes**

Delete (around lines 207-214):
```go
// Devices
adm.Get("/devices", admin.DeviceList(cfg, pb))
adm.Get("/devices/new", admin.DeviceForm(cfg, pb))
adm.Post("/devices", admin.DeviceCreate(cfg, pb))
adm.Get("/devices/:id/edit", admin.DeviceEdit(cfg, pb))
adm.Put("/devices/:id", admin.DeviceUpdate(cfg, pb))
adm.Delete("/devices/:id", admin.DeviceDelete(cfg, pb))
adm.Post("/devices/:id/assign-playlist", admin.DeviceAssignPlaylist(cfg, pb))
```

- [ ] **Step 8.11: Remove admin WhatsApp logs route**

Delete (around line 224):
```go
adm.Get("/whatsapp-logs", admin.WhatsAppLogs(cfg))
```

- [ ] **Step 8.12: Remove WhatsApp webhook route**

Delete (around line 256):
```go
app.Post("/webhook/whatsapp", web.WhatsAppWebhook(cfg))
```

- [ ] **Step 8.13: Remove `/locales.html` alias route (orphan)**

Delete (around line 93):
```go
app.Get("/locales.html", web.LocalesPageHandler(cfg, pb))
```
The canonical routes `/locales` and `/locales-disponibles` survive. (Earlier discovery showed no `web/locales.html` file exists, so this route was already serving the dynamic templ via `LocalesPageHandler`. The alias adds no value.)

- [ ] **Step 8.14: Remove `NoticiaHandler` route**

Delete (around line 120):
```go
app.Get("/noticias/:id", web.NoticiaHandler(cfg, pb))
```

- [ ] **Step 8.15: Verify `RSSFeed` route is still referenced**

The line `app.Get("/rss.xml", web.RSSFeed(cfg))` (around line 121) should remain. Confirm.

- [ ] **Step 8.16: Final build**

```bash
templ generate && go build ./...
```
Expected: **both succeed with exit 0**. This is the moment of truth — if anything still fails, an unresolved reference slipped through; locate via the error output and either restore or finish stripping.

- [ ] **Step 8.17: Run go vet for unused imports**

```bash
go vet ./...
```
If anything reports unused imports in `cmd/server/main.go`, remove them. (`gows` and `apiHandlers` are the prime candidates; `apiHandlers` should already be removed in 8.2.)

- [ ] **Step 8.18: Verify route count makes sense**

```bash
grep -cE "(app|adm|api|frag)\.(Get|Post|Put|Delete|Use)" cmd/server/main.go
```
Expected: roughly 35-45 routes (down from ~70). Plaza Real-only.

- [ ] **Step 8.19: Commit**

```bash
git add cmd/server/main.go
git commit -m "refactor: drop legacy route registrations from main.go

Removes route blocks for multimedia, events, news, playlists,
devices, whatsapp-logs, /webhook/whatsapp, /display, /totem,
/ws/device, /api/devices, /api/events. Drops orphan aliases
/eventos.html, /locales.html. NoticiaHandler /noticias/:id
removed (template deleted in prior commit). Plaza Real public
+ admin templ surface unchanged."
```

---

## Task 9: Smoke test locally

**Files:** none modified.

- [ ] **Step 9.1: Start the server**

```bash
templ generate && go run ./cmd/server
```

Wait for log line: `🏢 Plaza Real CMS en http://localhost:3000`. PocketBase should also start on `:8090`.

- [ ] **Step 9.2: Public site smoke**

In a browser (or with `curl -I`), confirm 200 on each:
- `http://localhost:3000/` (or `/index.html`)
- `http://localhost:3000/buscador-tiendas`
- `http://localhost:3000/noticias`
- `http://localhost:3000/comunicados`
- `http://localhost:3000/locales`
- `http://localhost:3000/promociones`
- `http://localhost:3000/tiendas/<any-existing-slug>`

For each, verify the page renders without the browser console showing 500/404. Screenshots optional but useful.

- [ ] **Step 9.3: Admin smoke**

In a browser:
1. `http://localhost:3000/admin/login` — log in with an existing user
2. After login, hit `/admin/dashboard` — stats render
3. `/admin/tiendas` — list renders
4. `/admin/noticias` — content_blocks NOTICIA list renders (also test creating one)
5. `/admin/comunicados`, `/admin/promociones` — both render
6. `/admin/locales` — list renders, create one
7. `/admin/reservas` — list renders
8. `/admin/users` (as superadmin) — renders
9. `/admin/reports` — renders
10. `/admin/settings` — renders

For each: open DevTools Network tab and verify no 404s on hx-get/hx-post fragments, no 500s on form submits.

- [ ] **Step 9.4: Confirm legacy routes 404**

```bash
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:3000/admin/devices
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:3000/admin/playlists
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:3000/admin/multimedia
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:3000/admin/events
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:3000/admin/news
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:3000/admin/whatsapp-logs
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:3000/display/x
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:3000/totem/x
curl -o /dev/null -s -w "%{http_code}\n" http://localhost:3000/eventos.html
```
Expected: all return **404**. (They would have hit the `/admin` AuthRequired middleware first, so unauthenticated calls might 302 — that's also acceptable as long as the route itself no longer exists post-auth.)

- [ ] **Step 9.5: Stop the server**

`Ctrl+C` in the terminal running `go run`.

---

## Task 10: Push and deploy

**Files:** none modified.

- [ ] **Step 10.1: Push branch**

```bash
git push origin main
```
Expected: fast-forward of `origin/main` from `78a76e2` to whatever the new HEAD is (~7 new commits from this plan).

- [ ] **Step 10.2: Deploy to Fly**

```bash
fly deploy
```

Wait for `--> v26 deployed successfully` (release number will increment from current v25). This builds the Docker image fresh; expect ~2-4 min.

- [ ] **Step 10.3: Production smoke**

In a browser, repeat the public smoke list from Step 9.2 against `https://cms-plazareal.fly.dev/` instead of localhost. Confirm site loads, fragments load (DevTools Network), favicon shows Plaza Real, no console errors.

- [ ] **Step 10.4: Production admin smoke**

Hit `https://cms-plazareal.fly.dev/admin/login` and walk through 4-5 admin pages from Step 9.3. Spot check.

- [ ] **Step 10.5: Confirm Fly release**

```bash
fly releases -a cms-plazareal | head -5
```
Expected: `v26` (or `v27` if there were prior failed builds), status `complete`, recent date.

- [ ] **Step 10.6: Final repo state**

```bash
git log --oneline pre-stripdown-2026-05-05..HEAD
wc -l internal/handlers/admin/handlers.go cmd/server/main.go
git ls-files | grep -E '\.html$' | wc -l
git ls-files '*.templ' | wc -l
```
Expected:
- ~7 commits since the tag
- `admin/handlers.go` ~1480 lines (was 2622)
- `main.go` ~200 lines (was 271)
- ~5 HTML files (was 17)
- ~30+ templ files (unchanged)

---

## Rollback (if anything goes wrong post-deploy)

If the Fly deploy is broken in a way you can't fix in 10 minutes:

```bash
fly releases rollback v25 -a cms-plazareal
```

If the local repo state is bad:

```bash
git reset --hard pre-stripdown-2026-05-05
git push --force-with-lease origin main
```

The tag is your safety net. Don't delete it for at least a week.

---

## Follow-up plans (out of scope here)

After this strip-down lands and feels stable, write separate plans for:

1. **Bring back categorías chips** — the regression noted during planning. Likely involves a templ component for category chips on the public homepage and/or a categorías section in admin. Reference `origin/claude/update-galleries-branding-YwgbE` (still on origin) for the conceptual implementation.

2. **Migrate remaining static HTML to templ** — `web/index.html`, `web/buscador-tiendas.html`, `web/noticias.html` are still served by `IndexHandler` etc. as static. Move to `internal/view/pages/public/index.templ` (which already exists but isn't wired) and have the handler render templ instead.

3. **Templ-ify Tiendas admin form** — the helper `tiendaFormHTML` (admin/handlers.go ~line 1534) builds a form via string concatenation. Replace with `internal/view/pages/admin/tienda_form.templ` (already exists).

4. **Bulk CSV upload for tiendas** — useful feature buried in `origin/claude/bulk-store-upload-Ee4UH`. Re-implement on top of templ admin.
