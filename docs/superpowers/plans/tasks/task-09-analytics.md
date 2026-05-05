# Task 09 — Analytics: page_views, leads, dashboard stats, reports CSV

**Depends on:** Task 05 (admin shell), Task 08 (auth — para tener userID en middleware)
**Estimated complexity:** media — middleware + colección + dashboard stats + CSV export

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Colección page_views: puede existir ya, verificar campos
Middleware de auth: ya setea c.Locals("userID", claims.Subject)
Dashboard: ya tiene DashboardStats(d DashboardStatsData) en templ esperando datos reales
```

---

## Objetivo

1. Middleware de tracking que registra cada request público en `page_views`
2. Colección `leads` para capturar interés en locales disponibles
3. Handlers que calculan estadísticas reales para el dashboard
4. Endpoint CSV para exportar datos de analytics e informes

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Verificar/Modificar | `internal/auth/collections.go` — colecciones page_views + leads |
| Crear | `internal/middleware/tracking.go` — middleware de page views |
| Crear | `internal/view/pages/admin/reports_page.templ` |
| Modificar | `internal/handlers/admin/handlers.go` — dashboard stats reales + reports + leads |
| Modificar | `cmd/server/main.go` — registrar middleware + rutas reports |

---

## Implementación

- [ ] **Step 1: Verificar/crear colecciones**

```bash
grep -n "page_views\|leads" internal/auth/collections.go | head -20
```

Si `page_views` no existe, agregar:

```go
// page_views — registro de visitas públicas
// Campos:
//   path       string   "/tiendas/starbucks"
//   referrer   string   (header Referer)
//   user_agent string
//   ip         string   (hashed para privacidad)
//   created    datetime (automático en PocketBase)
```

```go
&migrate.CreateCollection{
    Collection: &core.Collection{
        Name: "page_views",
        Type: core.CollectionTypeBase,
        Fields: core.FieldsList{
            &core.TextField{Name: "path"},
            &core.TextField{Name: "referrer"},
            &core.TextField{Name: "user_agent"},
            &core.TextField{Name: "ip"},
        },
    },
},
```

Si `leads` no existe, agregar:

```go
// leads — interés de potenciales arrendatarios
// Campos:
//   nombre     string
//   email      string
//   telefono   string
//   mensaje    string
//   local_id   string (ID del local_disponible, opcional)
//   estado     select ["nuevo", "contactado", "descartado"]
```

```go
&migrate.CreateCollection{
    Collection: &core.Collection{
        Name: "leads",
        Type: core.CollectionTypeBase,
        Fields: core.FieldsList{
            &core.TextField{Name: "nombre", Required: true},
            &core.EmailField{Name: "email", Required: true},
            &core.TextField{Name: "telefono"},
            &core.TextField{Name: "mensaje"},
            &core.TextField{Name: "local_id"},
            &core.SelectField{Name: "estado", Values: []string{"nuevo", "contactado", "descartado"}},
        },
    },
},
```

- [ ] **Step 2: Crear internal/middleware/tracking.go**

```go
package middleware

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"cms-plazareal/internal/config"
	"github.com/gofiber/fiber/v2"
	pocketbase "github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// TrackPageView registra page views para rutas públicas (no /admin, no /frag, no /static).
func TrackPageView(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		path := c.Path()
		// Solo registrar rutas públicas exitosas
		if strings.HasPrefix(path, "/admin") ||
			strings.HasPrefix(path, "/frag") ||
			strings.HasPrefix(path, "/static") ||
			c.Response().StatusCode() >= 400 {
			return err
		}

		col, colErr := pb.FindCollectionByNameOrId("page_views")
		if colErr != nil {
			return err
		}

		record := core.NewRecord(col)
		record.Set("path", path)
		record.Set("referrer", c.Get("Referer"))
		record.Set("user_agent", c.Get("User-Agent"))
		// Hash IP para privacidad (no guardar IP en crudo)
		h := sha256.Sum256([]byte(c.IP()))
		record.Set("ip", fmt.Sprintf("%x", h[:8]))
		_ = pb.Save(record) // ignorar error — tracking no debe romper respuesta

		return err
	}
}
```

- [ ] **Step 3: Registrar middleware en cmd/server/main.go**

Buscar dónde se registran los middlewares globales:
```bash
grep -n "Use\|middleware\." cmd/server/main.go | head -20
```

Agregar el middleware de tracking **después** de los middlewares de auth/static, **antes** de las rutas:

```go
// Tracking middleware — solo rutas públicas
app.Use(middleware.TrackPageView(cfg, pb))
```

**Nota:** Usar `app.Use` a nivel global es correcto porque el middleware filtra internamente por prefijo de path. Asegurarse de agregarlo antes de definir las rutas, no después.

- [ ] **Step 4: Agregar handler de dashboard stats reales**

Actualizar el handler `DashboardStatsHandler` existente en `handlers.go` para calcular datos reales:

```go
// DashboardStatsHandler — GET /admin/dashboard/stats
// Retorna DashboardStats(d) con datos reales de PocketBase.
func DashboardStatsHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Tiendas publicadas
		tiendasPublicadas, _ := pb.CountRecordsByFilter("tiendas", "status = 'publicado'")

		// Locales disponibles
		localesDisponibles, _ := pb.CountRecordsByFilter("locales_disponibles", "estado = 'disponible'")

		// Reservas pendientes
		reservasPendientes, _ := pb.CountRecordsByFilter("reservas", "estado = 'pendiente'")

		// Visitas hoy
		today := time.Now().Format("2006-01-02")
		visitasHoy, _ := pb.CountRecordsByFilter("page_views", "created >= '"+today+"'")

		// Visitas semana
		weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		visitasSemana, _ := pb.CountRecordsByFilter("page_views", "created >= '"+weekAgo+"'")

		// Nuevos leads (últimos 30 días)
		monthAgo := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
		nuevosLeads, _ := pb.CountRecordsByFilter("leads", "created >= '"+monthAgo+"'")

		return helpers.Render(c, adminView.DashboardStats(adminView.DashboardStatsData{
			TiendasPublicadas:  int(tiendasPublicadas),
			LocalesDisponibles: int(localesDisponibles),
			ReservasPendientes: int(reservasPendientes),
			VisitasHoy:         int(visitasHoy),
			VisitasSemana:      int(visitasSemana),
			NuevosLeads:        int(nuevosLeads),
		}))
	}
}
```

**Verificar si `CountRecordsByFilter` existe en la versión de PocketBase usada:**
```bash
grep -n "CountRecords\|FindRecordsByFilter" go.sum | head -5
grep -rn "CountRecords" --include="*.go" vendor/ 2>/dev/null | head -5
```

Si `CountRecordsByFilter` no existe, usar `FindRecordsByFilter` y contar el slice:
```go
records, _ := pb.FindRecordsByFilter("tiendas", "status = 'publicado'", "", 10000, 0)
tiendasPublicadas := len(records)
```

- [ ] **Step 5: Crear internal/view/pages/admin/reports_page.templ**

```templ
package admin

import "cms-plazareal/internal/view/layout"

type TopPage struct {
	Path  string
	Count int
}

type ReportsPageData struct {
	TotalVisitasSemana int
	TotalVisitasMes    int
	TopPages           []TopPage
	TotalLeads         int
	LeadsNuevos        int
}

templ ReportsPage(d ReportsPageData) {
	@layout.Admin("Informes", "informes", reportsPageBody(d))
}

templ reportsPageBody(d ReportsPageData) {
	<div class="admin-page">
		<div class="page-header">
			<h1 class="page-title">Informes</h1>
			<div class="page-header-actions">
				<a href="/admin/reports/export?type=page_views" class="btn btn-ghost">
					<span class="material-symbols-outlined">download</span>
					Exportar Visitas CSV
				</a>
				<a href="/admin/reports/export?type=leads" class="btn btn-ghost">
					<span class="material-symbols-outlined">download</span>
					Exportar Leads CSV
				</a>
			</div>
		</div>
		<div class="stats-grid stats-grid-4">
			@statCard("Visitas (7 días)", fmt.Sprintf("%d", d.TotalVisitasSemana), "trending_up", "blue")
			@statCard("Visitas (30 días)", fmt.Sprintf("%d", d.TotalVisitasMes), "calendar_month", "blue")
			@statCard("Leads totales", fmt.Sprintf("%d", d.TotalLeads), "contact_mail", "green")
			@statCard("Leads nuevos (30d)", fmt.Sprintf("%d", d.LeadsNuevos), "person_add", "green")
		</div>
		<div class="card glass-card" style="margin-top:2rem;">
			<h2 class="section-title">Páginas más visitadas (últimos 30 días)</h2>
			<table class="admin-table">
				<thead>
					<tr>
						<th>Ruta</th>
						<th>Visitas</th>
					</tr>
				</thead>
				<tbody>
					for _, p := range d.TopPages {
						<tr>
							<td class="font-mono text-sm">{ p.Path }</td>
							<td>{ fmt.Sprintf("%d", p.Count) }</td>
						</tr>
					}
				</tbody>
			</table>
		</div>
	</div>
}
```

- [ ] **Step 6: Agregar handlers de reports en handlers.go**

```go
// ReportsPage — GET /admin/reports
func ReportsPageHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		monthAgo := time.Now().AddDate(0, -1, 0).Format("2006-01-02")

		// Visitas semana
		weekRecords, _ := pb.FindRecordsByFilter("page_views", "created >= '"+weekAgo+"'", "", 50000, 0)
		// Visitas mes
		monthRecords, _ := pb.FindRecordsByFilter("page_views", "created >= '"+monthAgo+"'", "", 50000, 0)

		// Top pages del mes: contar por path
		pathCount := make(map[string]int)
		for _, r := range monthRecords {
			pathCount[r.GetString("path")]++
		}
		topPages := make([]adminView.TopPage, 0, len(pathCount))
		for path, count := range pathCount {
			topPages = append(topPages, adminView.TopPage{Path: path, Count: count})
		}
		sort.Slice(topPages, func(i, j int) bool { return topPages[i].Count > topPages[j].Count })
		if len(topPages) > 20 {
			topPages = topPages[:20]
		}

		// Leads
		allLeads, _ := pb.FindRecordsByFilter("leads", "1=1", "", 10000, 0)
		newLeads, _ := pb.FindRecordsByFilter("leads", "created >= '"+monthAgo+"'", "", 10000, 0)

		data := adminView.ReportsPageData{
			TotalVisitasSemana: len(weekRecords),
			TotalVisitasMes:    len(monthRecords),
			TopPages:           topPages,
			TotalLeads:         len(allLeads),
			LeadsNuevos:        len(newLeads),
		}

		content := adminView.ReportsPage(data)
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, content)
		}
		return helpers.Render(c, layout.Admin("Informes", "informes", content))
	}
}

// ReportsExport — GET /admin/reports/export?type=page_views|leads
func ReportsExport(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		exportType := c.Query("type", "page_views")
		var csvRows [][]string

		switch exportType {
		case "leads":
			records, _ := pb.FindRecordsByFilter("leads", "1=1", "-created", 50000, 0)
			csvRows = append(csvRows, []string{"ID", "Nombre", "Email", "Teléfono", "Mensaje", "Estado", "Fecha"})
			for _, r := range records {
				csvRows = append(csvRows, []string{
					r.Id,
					r.GetString("nombre"),
					r.GetString("email"),
					r.GetString("telefono"),
					r.GetString("mensaje"),
					r.GetString("estado"),
					r.GetString("created"),
				})
			}
		default: // page_views
			records, _ := pb.FindRecordsByFilter("page_views", "1=1", "-created", 50000, 0)
			csvRows = append(csvRows, []string{"Path", "Referrer", "User-Agent", "IP Hash", "Fecha"})
			for _, r := range records {
				csvRows = append(csvRows, []string{
					r.GetString("path"),
					r.GetString("referrer"),
					r.GetString("user_agent"),
					r.GetString("ip"),
					r.GetString("created"),
				})
			}
		}

		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set("Content-Disposition", "attachment; filename="+exportType+"-"+time.Now().Format("2006-01-02")+".csv")

		w := csv.NewWriter(c.Response().BodyWriter())
		_ = w.WriteAll(csvRows)
		w.Flush()
		return nil
	}
}
```

**Agregar imports en handlers.go:**
```go
import (
    "encoding/csv"
    "sort"
)
```

- [ ] **Step 7: Registrar rutas en cmd/server/main.go**

```go
// Reports / Informes
adm.Get("/reports", middleware.RoleRequired("superadmin","director"), admin.ReportsPageHandler(cfg, pb))
adm.Get("/reports/export", middleware.RoleRequired("superadmin","director"), admin.ReportsExport(cfg, pb))

// Dashboard stats
adm.Get("/dashboard/stats", middleware.RoleRequired("superadmin","director","admin","editor"), admin.DashboardStatsHandler(cfg, pb))
```

- [ ] **Step 8: Compilar y verificar**

```bash
make generate
go build ./...
```

- [ ] **Step 9: Test manual**

```bash
go run cmd/server/main.go
```

1. Visitar `http://localhost:3000/` varias veces → verificar que se crean registros en `page_views` (consola PocketBase o admin PocketBase en `:8090`)
2. `http://localhost:3000/admin/reports` → debe mostrar stats (puede ser 0 si no hay datos)
3. `http://localhost:3000/admin/reports/export?type=page_views` → debe descargar CSV
4. `http://localhost:3000/admin/dashboard` → stats del dashboard deben actualizarse (hx-get stats)

Verificar que rutas `/admin/*`, `/frag/*`, `/static/*` no generan page views:
```bash
curl http://localhost:3000/frag/tiendas-grid
# Verificar en PocketBase admin que NO se creó un page_view para /frag/tiendas-grid
```

- [ ] **Step 10: Commit**

```bash
git add internal/auth/collections.go \
        internal/middleware/tracking.go \
        internal/view/pages/admin/reports_page.templ \
        internal/view/pages/admin/reports_page_templ.go \
        internal/handlers/admin/handlers.go \
        cmd/server/main.go
git commit -m "feat: analytics middleware + leads collection + dashboard stats + CSV export"
```

---

## Estado del repositorio al finalizar este task

- Todas las visitas públicas (no `/admin`, no `/frag`, no `/static`) se registran en `page_views`
- Colección `leads` disponible para capturar interés de arrendatarios
- Dashboard stats calcula datos reales de PocketBase
- Informes muestra top pages + totales y permite exportar CSV
- `go build ./...` pasa sin errores
