# Task 08 — Analytics: tracking de visitas, leads y panel de informes

**Depends on:** Task 01, Task 02, Task 05 (locales), Task 06 (eventos y reservas)
**Estimated complexity:** media-alta — nuevas colecciones + middleware + nueva página admin

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Colecciones PocketBase existentes: users, media, content_blocks, multimedia,
  playlists, playlist_items, devices, form_responses, whatsapp_logs, tiendas,
  locales_disponibles, reservas
```

**Problema:** El dashboard muestra solo estadísticas de tiendas. No existe tracking de visitas web ni registro de leads (clics en WhatsApp, consultas de locales, reservas de eventos).

`DashboardStats` actual cuenta únicamente tiendas:
```go
// CÓDIGO ACTUAL — solo tiendas:
func DashboardStats(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
    return func(c *fiber.Ctx) error {
        all, _ := pb.FindRecordsByFilter("tiendas", "", "", 1000, 0)
        // ... cuenta publicadas, destacadas, restaurantes, farmacias
    }
}
```

**Colecciones a crear:** `page_views`, `leads`

**Archivos a crear:** `internal/middleware/tracking.go`, `internal/templates/admin/pages/reports.html`

**Archivos a modificar:** `internal/auth/collections.go`, `internal/handlers/admin/handlers.go`, `internal/handlers/api/handlers.go`, `cmd/server/main.go`, `internal/templates/admin/pages/dashboard.html`

---

## Objetivo

Agregar tracking ligero de visitas (middleware fire-and-forget), colección de leads, dashboard con métricas reales y página de informes con exportación CSV.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Modificar | `internal/auth/collections.go` — agregar `page_views` y `leads` |
| Crear | `internal/middleware/tracking.go` |
| Modificar | `cmd/server/main.go` — registrar middleware + nuevas rutas |
| Modificar | `internal/handlers/admin/handlers.go` — `DashboardStats`, `ReportsPage`, `ExportLeadsCSV` |
| Modificar | `internal/handlers/api/handlers.go` — `TrackLead` |
| Crear | `internal/templates/admin/pages/reports.html` |
| Modificar | `internal/templates/admin/pages/dashboard.html` — agregar link "Informes" al sidebar |

---

## Implementación

- [ ] **Step 1: Agregar colecciones page_views y leads en collections.go**

Abrir `internal/auth/collections.go` y localizar el final de `ensureCollections` (justo antes del `return nil` final):
```bash
grep -n "return nil" internal/auth/collections.go | tail -3
```

Insertar **antes** del `return nil` final:

```go
// ── PAGE_VIEWS ──
if _, err := app.FindCollectionByNameOrId("page_views"); err != nil {
    col := core.NewBaseCollection("page_views")
    col.Fields.Add(
        &core.TextField{Name: "page"},
        &core.TextField{Name: "ref"},
        &core.TextField{Name: "ua"},
        &core.TextField{Name: "ip"},
    )
    if err := app.Save(col); err != nil {
        return err
    }
    log.Println("  ✅ Collection 'page_views' created")
}

// ── LEADS ──
if _, err := app.FindCollectionByNameOrId("leads"); err != nil {
    col := core.NewBaseCollection("leads")
    col.Fields.Add(
        &core.TextField{Name: "tipo", Required: true},
        &core.TextField{Name: "nombre"},
        &core.TextField{Name: "email"},
        &core.TextField{Name: "telefono"},
        &core.TextField{Name: "source"},
        &core.TextField{Name: "source_url"},
        &core.TextField{Name: "mensaje"},
        &core.TextField{Name: "status"},
    )
    if err := app.Save(col); err != nil {
        return err
    }
    log.Println("  ✅ Collection 'leads' created")
}
```

- [ ] **Step 2: Crear internal/middleware/tracking.go**

```go
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// PageTracking registra visitas en page_views. Solo páginas públicas, fire-and-forget.
func PageTracking(pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Ignorar rutas no-públicas y assets
		if strings.HasPrefix(path, "/admin") ||
			strings.HasPrefix(path, "/fragments") ||
			strings.HasPrefix(path, "/api") ||
			strings.HasPrefix(path, "/static") ||
			strings.HasPrefix(path, "/ws") ||
			strings.HasPrefix(path, "/display") ||
			strings.HasPrefix(path, "/totem") ||
			strings.HasSuffix(path, ".ico") ||
			strings.HasSuffix(path, ".xml") {
			return c.Next()
		}

		col, err := pb.FindCollectionByNameOrId("page_views")
		if err != nil {
			return c.Next()
		}

		// Anonimizar IP: conservar solo los primeros 3 octetos
		ip := c.IP()
		parts := strings.Split(ip, ".")
		if len(parts) == 4 {
			ip = parts[0] + "." + parts[1] + "." + parts[2] + ".0"
		}

		ua := c.Get("User-Agent")
		if len(ua) > 200 {
			ua = ua[:200]
		}

		r := core.NewRecord(col)
		r.Set("page", path)
		r.Set("ref", c.Get("Referer"))
		r.Set("ua", ua)
		r.Set("ip", ip)
		pb.Save(r) // fire-and-forget

		return c.Next()
	}
}
```

- [ ] **Step 3: Registrar middleware en cmd/server/main.go**

Buscar el bloque de middlewares globales:
```bash
grep -n "app.Use(cors\|app.Use(logger\|app.Use(recover" cmd/server/main.go
```

Agregar **después** del `app.Use(cors.New(...))`:
```go
// Page tracking — solo rutas públicas
app.Use(middleware.PageTracking(pb))
```

Agregar también las nuevas rutas en la sección API y admin. Buscar:
```bash
grep -n "// ── PUBLIC API\|api :=\|// ── ADMIN" cmd/server/main.go
```

En la sección `api`:
```go
api.Post("/leads", apiHandlers.TrackLead(pb))
```

En la sección `adm`:
```go
adm.Get("/reports", admin.ReportsPage(cfg, pb))
adm.Get("/reports/export-leads", admin.ExportLeadsCSV(cfg, pb))
```

- [ ] **Step 4: Reemplazar DashboardStats en handlers/admin/handlers.go**

Buscar la función actual:
```bash
grep -n "func DashboardStats" internal/handlers/admin/handlers.go
```

Reemplazar completamente con:

```go
func DashboardStats(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Tiendas
		allTiendas, _ := pb.FindRecordsByFilter("tiendas", "", "", 2000, 0)
		totalTiendas, pubTiendas := len(allTiendas), 0
		for _, r := range allTiendas {
			if r.GetString("status") == "publicado" {
				pubTiendas++
			}
		}

		// Reservas pendientes
		reservasPend, _ := pb.FindRecordsByFilter("reservas", "status = 'pendiente'", "", 1000, 0)

		// Locales disponibles
		localesDisp, _ := pb.FindRecordsByFilter("locales_disponibles", "status = 'disponible'", "", 1000, 0)

		// Visitas hoy
		hoy := time.Now().Format("2006-01-02")
		visitasHoy, _ := pb.FindRecordsByFilter("page_views",
			fmt.Sprintf("created >= '%s 00:00:00'", hoy), "", 5000, 0)

		// Visitas últimos 7 días
		hace7 := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		visitasSemana, _ := pb.FindRecordsByFilter("page_views",
			fmt.Sprintf("created >= '%s 00:00:00'", hace7), "", 50000, 0)

		// Leads nuevos
		leadsNuevos, _ := pb.FindRecordsByFilter("leads", "status = 'nuevo'", "", 1000, 0)

		html := fmt.Sprintf(`
<div class="stat-card accent">
  <div class="stat-card-label">Tiendas publicadas</div>
  <div class="stat-card-value">%d</div>
  <div class="stat-card-delta">%d total · %d borradores</div>
</div>
<div class="stat-card">
  <div class="stat-card-label">Visitas hoy</div>
  <div class="stat-card-value">%d</div>
  <div class="stat-card-delta">%d esta semana</div>
</div>
<div class="stat-card warn">
  <div class="stat-card-label">Reservas pendientes</div>
  <div class="stat-card-value">%d</div>
  <div class="stat-card-delta"><a href="/admin/reservas" style="color:inherit">Ver reservas →</a></div>
</div>
<div class="stat-card">
  <div class="stat-card-label">Locales disponibles</div>
  <div class="stat-card-value">%d</div>
  <div class="stat-card-delta"><a href="/admin/locales" style="color:inherit">Gestionar →</a></div>
</div>
<div class="stat-card warn">
  <div class="stat-card-label">Leads nuevos</div>
  <div class="stat-card-value">%d</div>
  <div class="stat-card-delta"><a href="/admin/reports" style="color:inherit">Ver informes →</a></div>
</div>`,
			pubTiendas, totalTiendas, totalTiendas-pubTiendas,
			len(visitasHoy), len(visitasSemana),
			len(reservasPend),
			len(localesDisp),
			len(leadsNuevos),
		)

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}
```

Verificar que el import `"time"` ya esté en el bloque de imports del archivo (debería estar). Si no, agregarlo.

- [ ] **Step 5: Agregar TrackLead en internal/handlers/api/handlers.go**

Buscar el final del archivo:
```bash
tail -20 internal/handlers/api/handlers.go
```

Agregar al final:

```go
// TrackLead registra un lead (click WhatsApp, consulta local, etc.) desde el frontend público.
func TrackLead(pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		col, err := pb.FindCollectionByNameOrId("leads")
		if err != nil {
			return c.SendStatus(200) // silencioso
		}

		tipo := c.FormValue("tipo")
		if tipo == "" {
			return c.SendStatus(200)
		}

		r := core.NewRecord(col)
		r.Set("tipo", tipo)
		r.Set("source", c.FormValue("source"))
		r.Set("source_url", c.Get("Referer"))
		r.Set("nombre", c.FormValue("nombre"))
		r.Set("email", c.FormValue("email"))
		r.Set("telefono", c.FormValue("telefono"))
		r.Set("mensaje", c.FormValue("mensaje"))
		r.Set("status", "nuevo")
		pb.Save(r) // fire-and-forget

		return c.SendStatus(200)
	}
}
```

Verificar que el archivo tenga el import `"github.com/pocketbase/pocketbase/core"`. Si no existe:
```bash
grep -n '"github.com/pocketbase/pocketbase/core"' internal/handlers/api/handlers.go
```
Agregar al bloque de imports si falta.

- [ ] **Step 6: Agregar ReportsPage y ExportLeadsCSV en handlers/admin/handlers.go**

Agregar al final del archivo:

```go
// ReportsPage muestra el panel de informes con filtro por rango de fechas.
func ReportsPage(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		desde := c.Query("desde", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
		hasta := c.Query("hasta", time.Now().Format("2006-01-02"))
		filter := fmt.Sprintf("created >= '%s 00:00:00' && created <= '%s 23:59:59'", desde, hasta)

		views, _ := pb.FindRecordsByFilter("page_views", filter, "-created", 10000, 0)
		pageCounts := map[string]int{}
		for _, v := range views {
			pageCounts[v.GetString("page")]++
		}

		leads, _ := pb.FindRecordsByFilter("leads", filter, "-created", 5000, 0)
		leadTipos := map[string]int{}
		for _, l := range leads {
			leadTipos[l.GetString("tipo")]++
		}

		reservas, _ := pb.FindRecordsByFilter("reservas", filter, "-created", 1000, 0)

		type pageRow struct {
			Pagina  string
			Visitas int
		}
		var topPages []pageRow
		for pg, cnt := range pageCounts {
			topPages = append(topPages, pageRow{pg, cnt})
		}
		// sort simple por visitas desc
		for i := 0; i < len(topPages); i++ {
			for j := i + 1; j < len(topPages); j++ {
				if topPages[j].Visitas > topPages[i].Visitas {
					topPages[i], topPages[j] = topPages[j], topPages[i]
				}
			}
		}
		if len(topPages) > 10 {
			topPages = topPages[:10]
		}

		tmpl, err := template.ParseFiles("./internal/templates/admin/pages/reports.html")
		if err != nil {
			return c.Status(500).SendString("Template error: " + err.Error())
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.Execute(c, map[string]any{
			"Desde":       desde,
			"Hasta":       hasta,
			"TotalVistas": len(views),
			"TopPages":    topPages,
			"LeadTipos":   leadTipos,
			"Reservas":    reservas,
			"TotalLeads":  len(leads),
		})
	}
}

// ExportLeadsCSV exporta leads del período como archivo CSV descargable.
func ExportLeadsCSV(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		desde := c.Query("desde", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
		hasta := c.Query("hasta", time.Now().Format("2006-01-02"))
		filter := fmt.Sprintf("created >= '%s 00:00:00' && created <= '%s 23:59:59'", desde, hasta)

		leads, _ := pb.FindRecordsByFilter("leads", filter, "-created", 10000, 0)

		var sb strings.Builder
		sb.WriteString("Fecha,Tipo,Nombre,Email,Telefono,Source,Mensaje,Status\n")
		for _, l := range leads {
			sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s\n",
				l.GetDateTime("created").Time().Format("2006-01-02 15:04"),
				csvEscape(l.GetString("tipo")),
				csvEscape(l.GetString("nombre")),
				csvEscape(l.GetString("email")),
				csvEscape(l.GetString("telefono")),
				csvEscape(l.GetString("source")),
				csvEscape(l.GetString("mensaje")),
				csvEscape(l.GetString("status")),
			))
		}

		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set("Content-Disposition",
			fmt.Sprintf(`attachment; filename="leads-%s-%s.csv"`, desde, hasta))
		return c.SendString(sb.String())
	}
}

func csvEscape(s string) string {
	if strings.ContainsAny(s, "\",\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}
```

- [ ] **Step 7: Crear internal/templates/admin/pages/reports.html**

El template sigue el mismo layout del dashboard (misma estructura HTML con sidebar y topbar). Para no repetir todo el HTML del layout, copiar `dashboard.html` como base y reemplazar el `<div class="content" id="main-content">` con el contenido del informe.

Estructura del bloque de contenido que reemplaza el `#main-content`:

```html
<!-- Selector de rango de fechas -->
<form class="reports-filter-bar"
      hx-get="/admin/reports"
      hx-target="#main-content"
      hx-push-url="true"
      style="display:flex;gap:12px;align-items:flex-end;margin-bottom:24px;flex-wrap:wrap">
  <div class="form-group" style="margin:0">
    <label style="font-size:12px;color:var(--md-outline)">Desde</label>
    <input type="date" name="desde" value="{{.Desde}}" class="form-input"/>
  </div>
  <div class="form-group" style="margin:0">
    <label style="font-size:12px;color:var(--md-outline)">Hasta</label>
    <input type="date" name="hasta" value="{{.Hasta}}" class="form-input"/>
  </div>
  <button type="submit" class="topbar-btn topbar-btn-primary">Filtrar</button>
  <a href="/admin/reports/export-leads?desde={{.Desde}}&hasta={{.Hasta}}"
     class="topbar-btn topbar-btn-outline">
    <span class="material-symbols-outlined" style="font-size:16px">download</span>
    Exportar Leads CSV
  </a>
</form>

<!-- Stats row -->
<div class="stats-row" style="margin-bottom:24px">
  <div class="stat-card accent">
    <div class="stat-card-label">Visitas totales</div>
    <div class="stat-card-value">{{.TotalVistas}}</div>
    <div class="stat-card-delta">en el período seleccionado</div>
  </div>
  <div class="stat-card">
    <div class="stat-card-label">Leads generados</div>
    <div class="stat-card-value">{{.TotalLeads}}</div>
    <div class="stat-card-delta">contactos e interacciones</div>
  </div>
  <div class="stat-card">
    <div class="stat-card-label">Reservas en período</div>
    <div class="stat-card-value">{{len .Reservas}}</div>
    <div class="stat-card-delta">eventos del mall</div>
  </div>
</div>

<!-- Top páginas -->
<div class="card" style="margin-bottom:24px">
  <div class="card-header">
    <h2 class="card-title">Top páginas visitadas</h2>
  </div>
  <table class="data-table">
    <thead>
      <tr><th>Página</th><th style="text-align:right">Visitas</th></tr>
    </thead>
    <tbody>
      {{range .TopPages}}
      <tr>
        <td>{{.Pagina}}</td>
        <td style="text-align:right;font-weight:600">{{.Visitas}}</td>
      </tr>
      {{else}}
      <tr><td colspan="2" class="empty-state-cell">Sin datos en este período.</td></tr>
      {{end}}
    </tbody>
  </table>
</div>

<!-- Leads por tipo -->
<div class="card">
  <div class="card-header">
    <h2 class="card-title">Leads por tipo</h2>
    <span style="font-size:12px;color:var(--md-outline)">
      Exportar → CSV incluye detalle completo
    </span>
  </div>
  <table class="data-table">
    <thead>
      <tr><th>Tipo de interacción</th><th style="text-align:right">Cantidad</th></tr>
    </thead>
    <tbody>
      {{range $tipo, $cnt := .LeadTipos}}
      <tr>
        <td>{{$tipo}}</td>
        <td style="text-align:right;font-weight:600">{{$cnt}}</td>
      </tr>
      {{else}}
      <tr><td colspan="2" class="empty-state-cell">Sin leads en este período.</td></tr>
      {{end}}
    </tbody>
  </table>
</div>
```

Para crear `reports.html` con el layout completo: copiar `dashboard.html`, cambiar `<h1 class="topbar-title">Dashboard</h1>` por `<h1 class="topbar-title">Informes</h1>`, y reemplazar el bloque `<div class="content" id="main-content">` con el HTML de arriba.

- [ ] **Step 8: Agregar link "Informes" al sidebar del dashboard**

En `internal/templates/admin/pages/dashboard.html`, buscar la sección del sidebar:
```bash
grep -n "sidebar-section\|Usuarios\|Sistema" internal/templates/admin/pages/dashboard.html
```

En la sección "Sistema" del sidebar, agregar después del link de Usuarios:
```html
<a href="/admin/reports" class="sidebar-link"
   hx-get="/admin/reports"
   hx-target="#main-content"
   hx-push-url="true">
  <span class="material-symbols-outlined">bar_chart</span> Informes
</a>
```

- [ ] **Step 9: Compilar**

```bash
go build ./...
# Esperado: sin errores
```

- [ ] **Step 10: Test manual**

```bash
go run cmd/server/main.go
```

1. Navegar por 3-4 páginas públicas: `/`, `/buscador-tiendas.html`, `/noticias.html`
2. Abrir `/admin/dashboard` → las stat cards deben mostrar "Visitas hoy: N" con N > 0
3. Abrir `/admin/reports` → debe mostrar el período por defecto (últimos 30 días) con datos
4. Click "Exportar Leads CSV" → debe descargar un archivo `.csv`
5. Cambiar rango de fechas → debe recargar con HTMX

- [ ] **Step 11: Commit**

```bash
git add internal/auth/collections.go \
        internal/middleware/tracking.go \
        internal/handlers/admin/handlers.go \
        internal/handlers/api/handlers.go \
        internal/templates/admin/pages/reports.html \
        internal/templates/admin/pages/dashboard.html \
        cmd/server/main.go
git commit -m "feat: add page tracking, leads collection, expanded dashboard stats, reports page with CSV export"
```

---

## Estado del repositorio al finalizar este task

- Colecciones `page_views` y `leads` existen en PocketBase
- Cada visita pública se registra automáticamente (IP anonimizada)
- Dashboard muestra: tiendas, visitas hoy/semana, reservas pendientes, locales disponibles, leads nuevos
- `/admin/reports` muestra top páginas y leads por tipo con filtro de fechas
- Exportación CSV funcional desde el admin
