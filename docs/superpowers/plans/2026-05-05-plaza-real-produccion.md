# Plaza Real CMS — Production Readiness Plan v2

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Each task file is self-contained under `docs/superpowers/plans/tasks/`. Never read this index as context — dispatch subagents with the full text of each task file.

**Goal:** Migrar cms-plazareal a templ + templUI, implementar diseño Liquid Glass con paleta Plaza Real, corregir sistema de auth, agregar Locales Disponibles + Promociones con Reservas + Analytics, limpiar todo código legacy (JCP, escuela, Ollama, propiedades).

**Architecture:** Go 1.23 + Fiber v2 + PocketBase embebido + `a-h/templ` + templUI + HTMX 1.9 + Alpine.js. Templates compilados como Go code. Fragmentos HTMX son templ components renderizados directamente al ResponseWriter de Fiber. Admin usa layout compartido con composición de componentes. No hay build frontend step (no Node.js en producción).

**Tech Stack:** Go 1.23, Fiber v2.52, PocketBase v0.25, `a-h/templ`, templUI, HTMX 1.9, Alpine.js CDN, Cloudflare R2, Fly.io

**Design Direction:** Liquid Glass (frosted/translucent surfaces, blur, dark neutral backgrounds) + ultra-modern. Paleta fija: `#d71055` (rojo), `#06a0e0` (azul), `#acc60d` (lima). Fonts: Montserrat (headings, 900 italic), Geist (body). Material Symbols Outlined (icons).

**Content Model:**
- `tiendas` — colección de tiendas; `gal` ∈ {"placa-comercial", "torre-flamenco"}
- `content_blocks` — categorías: `NOTICIA`, `COMUNICADO`, `PROMOCION`
- `locales_disponibles` — espacios del mall en arriendo
- `reservas` — reservas para PROMOCION con tipo=evento

---

## Task Index

| # | Título | Archivo | Dep |
|---|--------|---------|-----|
| 00 | Foundation: módulo + templ + templUI + render helper | [task-00-foundation.md](tasks/task-00-foundation.md) | — |
| 01 | Legacy cleanup: propiedades + ollama + categories | [task-01-legacy-cleanup.md](tasks/task-01-legacy-cleanup.md) | 00 |
| 02 | Design System: Liquid Glass CSS + layout/base.templ + layout/admin.templ | [task-02-design-system.md](tasks/task-02-design-system.md) | 00 |
| 03 | R2 Upload Widget: endpoint + templ component | [task-03-r2-upload.md](tasks/task-03-r2-upload.md) | 00 |
| 04 | Public Pages: index, tiendas, tienda-detail, noticias, comunicados → templ | [task-04-public-pages.md](tasks/task-04-public-pages.md) | 02 |
| 05 | Admin Shell + Core Sections: login, dashboard, tiendas, eventos → templ | [task-05-admin-templates.md](tasks/task-05-admin-templates.md) | 02, 03 |
| 06 | Content Blocks: NOTICIA / COMUNICADO / PROMOCION (admin + público) | [task-06-content-blocks.md](tasks/task-06-content-blocks.md) | 05 |
| 07 | Locales Disponibles + Reservas: colecciones, admin, páginas públicas | [task-07-locales-reservas.md](tasks/task-07-locales-reservas.md) | 05 |
| 08 | Fix Auth: login real PocketBase + user CRUD | [task-08-auth.md](tasks/task-08-auth.md) | 05 |
| 09 | Analytics: page_views, leads, dashboard stats, reports CSV | [task-09-analytics.md](tasks/task-09-analytics.md) | 05 |
| 10 | Scraper: CLI para sitio antiguo → JSON importable | [task-10-scraper.md](tasks/task-10-scraper.md) | 00 |
| 11 | Smoke test + deploy Fly.io | [task-11-deploy.md](tasks/task-11-deploy.md) | 01–10 |
| 12 | Site Settings: hero bg home + bg buscador (admin + public) | [task-12-site-settings.md](tasks/task-12-site-settings.md) | 05, 03 |

---

## Dependency Order

```
Task 00 (foundation)
  ├── Task 01 (legacy cleanup)   — puede ir en paralelo con 02/03
  ├── Task 02 (design system)    — bloquea 04, 05
  │     ├── Task 04 (public pages)
  │     └── Task 05 (admin shell)
  │           ├── Task 06 (content blocks)
  │           ├── Task 07 (locales + reservas)
  │           ├── Task 08 (auth)
  │           └── Task 09 (analytics)
  └── Task 03 (R2 upload)        — necesario antes de Task 05
Task 10 (scraper)                — independiente después de Task 00
Task 12 (site settings)          — en paralelo con 06/07/08/09
Task 11 (deploy)                 — último, todo completo
```

---

## Render Helper (patrón global)

Todos los handlers usan este helper en `internal/helpers/render.go`:

```go
package helpers

import (
    "github.com/a-h/templ"
    "github.com/gofiber/fiber/v2"
)

func Render(c *fiber.Ctx, component templ.Component) error {
    c.Set("Content-Type", "text/html; charset=utf-8")
    return component.Render(c.UserContext(), c.Response().BodyWriter())
}
```

---

## View File Structure

```
internal/view/
  layout/
    base.templ         — public: head + navbar + footer
    admin.templ        — admin: sidebar + topbar + content slot
  components/
    upload_field.templ — R2 upload widget
    toast.templ
    badge.templ
  pages/
    public/
      index.templ
      tiendas.templ
      tienda_detail.templ
      noticias.templ
      comunicados.templ
      locales.templ
      promociones.templ
    admin/
      login.templ
      dashboard.templ
      tiendas_page.templ
      content_page.templ    — shared para NOTICIA/COMUNICADO/PROMOCION
      locales_page.templ
      reservas_page.templ
      multimedia_page.templ
      devices_page.templ
      users_page.templ
      reports_page.templ
  fragments/
    hero.templ
    tiendas_grid.templ
    tiendas_marquee.templ
    noticias_cards.templ
    comunicados_cards.templ
    promos_cards.templ
    locales_grid.templ
    reserva_form.templ
```

---

## File Map

| Acción | Archivos |
|--------|---------|
| **Crear** | `internal/view/**/*.templ`, `web/static/css/app.css`, `web/static/css/admin.css`, `internal/helpers/render.go`, `Makefile`, `cmd/scraper/main.go` |
| **Modificar** | `go.mod`, `go.sum`, todos `*.go` (import path), `internal/auth/collections.go`, `internal/handlers/admin/handlers.go`, `internal/handlers/fragments/*.go`, `internal/handlers/web/handlers.go`, `internal/handlers/api/handlers.go`, `cmd/server/main.go`, `internal/config/config.go`, `internal/auth/jwt.go`, `internal/middleware/auth.go` |
| **Eliminar** | `web/propiedades.html`, `internal/templates/admin/pages/propiedades.html`, `internal/handlers/fragments/propiedades.go`, `internal/services/ollama/ollama.go`, `jcp-gestioninmobiliaria.service`, `DESIGN_AUDIT.md`, `package.json`, `package-lock.json`, `internal/templates/admin/pages/*.html` (reemplazados por templ) |

---

## Spec Coverage

| Requisito | Task |
|-----------|------|
| Módulo renombrado a cms-plazareal | 00 |
| templ + templUI instalados | 00 |
| Propiedades + Ollama eliminados | 01 |
| Categorías content_blocks: NOTICIA/COMUNICADO/PROMOCION | 01 |
| Diseño Liquid Glass con paleta Plaza Real | 02 |
| Layout base.templ + admin.templ compartidos | 02 |
| Upload R2 desde admin (drag-drop, auto-URL) | 03 |
| Páginas públicas en templ | 04 |
| Tienda-detail server-side `/tiendas/:slug` | 04 |
| Admin en templ (login, dashboard, CRUD) | 05 |
| NOTICIA/COMUNICADO/PROMOCION admin + público | 06 |
| Galerías: placa-comercial / torre-flamenco | 06 |
| Locales Disponibles (colección + admin + público) | 07 |
| Reservas para PROMOCION tipo=evento | 07 |
| Login real contra PocketBase users | 08 |
| Cookie pr_token, JWT issuer cms-plazareal | 08 |
| Page tracking (page_views) | 09 |
| Leads collection + API | 09 |
| Dashboard stats expandido | 09 |
| Informes con filtro fecha + CSV export | 09 |
| Scraper → JSON para importación masiva | 10 |
| Smoke test + fly deploy | 11 |
