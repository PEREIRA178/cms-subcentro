# Task 05 — Admin Shell + Core Sections: login, dashboard, tiendas → templ

**Depends on:** Task 02 (design system), Task 03 (R2 upload widget)
**Estimated complexity:** alta — múltiples páginas + migrar el handler más grande del proyecto

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
layout/admin.templ: existe y compila (Task 02)
internal/templates/admin/pages/*.html: aún existen, servidos con c.SendFile
Patrón actual admin:
  - HX-Request == true → SendFile(fragment.html)
  - otherwise → SendFile(dashboard.html) [el shell]
Nuevo patrón:
  - HX-Request == true → helpers.Render(c, AdminPage_Content(data))
  - otherwise → helpers.Render(c, layout.Admin(title, page, AdminPage_Content(data)))
```

---

## Objetivo

Migrar el admin de archivos `.html` + `strings.Builder` a templ components usando `layout.Admin`. Cubrir: login, dashboard con stats, tiendas CRUD. El resto de secciones (content, locales, reservas, users, reports) se migran en tasks 06-09.

---

## Archivos a crear/modificar

| Acción | Archivo |
|--------|---------|
| Crear | `internal/view/pages/admin/login.templ` |
| Crear | `internal/view/pages/admin/dashboard.templ` |
| Crear | `internal/view/pages/admin/tiendas_page.templ` |
| Crear | `internal/view/pages/admin/tienda_form.templ` |
| Modificar | `internal/handlers/admin/handlers.go` — reemplazar SendFile + strings.Builder |

---

## Implementación

- [ ] **Step 1: Crear internal/view/pages/admin/login.templ**

```bash
mkdir -p internal/view/pages/admin
```

Contenido de `internal/view/pages/admin/login.templ`:

```templ
package admin

templ Login(errorMsg string) {
	<!DOCTYPE html>
	<html lang="es">
	<head>
		<meta charset="UTF-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
		<title>Login — Plaza Real Admin</title>
		<link rel="icon" type="image/svg+xml" href="/static/favicon.svg"/>
		<link rel="preconnect" href="https://fonts.googleapis.com"/>
		<link href="https://fonts.googleapis.com/css2?family=Montserrat:ital,wght@0,900;1,900&family=Geist:wght@300;400;500;600;700&display=swap" rel="stylesheet"/>
		<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200"/>
		<link rel="stylesheet" href="/static/css/admin.css"/>
		<script src="https://unpkg.com/htmx.org@1.9.12"></script>
	</head>
	<body>
		<div class="login-bg">
			<div class="login-card">
				<div style="text-align:center;margin-bottom:32px">
					<img src="https://plazareal.cl/wp-content/uploads/2025/12/logo-mall-plaza-real-2024@3x.png" alt="Plaza Real" style="height:48px;filter:brightness(0) invert(1);margin:0 auto"/>
					<p style="color:var(--text-muted);font-size:.85rem;margin-top:12px">Panel de administración</p>
				</div>
				<form hx-post="/admin/login" hx-target="#login-feedback" hx-swap="innerHTML">
					<div class="form-group">
						<label>Email</label>
						<input type="email" name="email" required placeholder="admin@plazareal.cl" autofocus/>
					</div>
					<div class="form-group" style="margin-bottom:20px">
						<label>Contraseña</label>
						<input type="password" name="password" required placeholder="••••••••"/>
					</div>
					<label style="display:flex;align-items:center;gap:8px;font-size:.82rem;color:var(--text-muted);margin-bottom:20px;cursor:pointer">
						<input type="checkbox" name="remember"/> Recordar sesión (72 h)
					</label>
					<div id="login-feedback" style="margin-bottom:12px">
						if errorMsg != "" {
							<div class="toast toast-error">{ errorMsg }</div>
						}
					</div>
					<button type="submit" class="topbar-btn topbar-btn-primary" style="width:100%;justify-content:center;padding:12px">
						Iniciar sesión
					</button>
				</form>
			</div>
		</div>
	</body>
	</html>
}
```

- [ ] **Step 2: Crear internal/view/pages/admin/dashboard.templ**

Contenido de `internal/view/pages/admin/dashboard.templ`:

```templ
package admin

import "fmt"

type DashboardStatsData struct {
	TiendasPublicadas  int
	LocalesDisponibles int
	ReservasPendientes int
	VisitasHoy         int
	VisitasSemana      int
	NuevosLeads        int
}

templ Dashboard() {
	<div>
		<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px">
			<div>
				<h2 style="font-size:1.1rem;font-weight:800;margin-bottom:4px">Resumen</h2>
				<p style="font-size:.8rem;color:var(--text-muted)">Indicadores en tiempo real</p>
			</div>
			<a href="/index.html" target="_blank" class="topbar-btn topbar-btn-outline">
				<span class="material-symbols-outlined" style="font-size:14px">open_in_new</span> Ver sitio
			</a>
		</div>
		<div id="dashboard-stats"
			hx-get="/admin/dashboard/stats"
			hx-trigger="load"
			hx-swap="innerHTML">
			<div class="stats-grid">
				for i := 0; i < 6; i++ {
					<div class="stat-card" style="opacity:.4">
						<div class="stat-card-icon" style="background:var(--surface-3)"></div>
						<div class="stat-card-value">—</div>
						<div class="stat-card-label">Cargando...</div>
					</div>
				}
			</div>
		</div>
		<div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(200px,1fr));gap:16px;margin-top:24px">
			<a href="/admin/tiendas" class="card" style="padding:20px;display:flex;align-items:center;gap:12px">
				<span class="material-symbols-outlined" style="font-size:22px;color:var(--red)">storefront</span>
				<div>
					<p style="font-size:.84rem;font-weight:700">Tiendas</p>
					<p style="font-size:.74rem;color:var(--text-muted)">Gestionar directorio</p>
				</div>
			</a>
			<a href="/admin/locales" class="card" style="padding:20px;display:flex;align-items:center;gap:12px">
				<span class="material-symbols-outlined" style="font-size:22px;color:var(--blue)">real_estate_agent</span>
				<div>
					<p style="font-size:.84rem;font-weight:700">Locales</p>
					<p style="font-size:.74rem;color:var(--text-muted)">Disponibles en arriendo</p>
				</div>
			</a>
			<a href="/admin/content?cat=PROMOCION" class="card" style="padding:20px;display:flex;align-items:center;gap:12px">
				<span class="material-symbols-outlined" style="font-size:22px;color:var(--lime)">celebration</span>
				<div>
					<p style="font-size:.84rem;font-weight:700">Promociones</p>
					<p style="font-size:.74rem;color:var(--text-muted)">Eventos y ofertas</p>
				</div>
			</a>
			<a href="/admin/reservas" class="card" style="padding:20px;display:flex;align-items:center;gap:12px">
				<span class="material-symbols-outlined" style="font-size:22px;color:var(--text-mid)">event_seat</span>
				<div>
					<p style="font-size:.84rem;font-weight:700">Reservas</p>
					<p style="font-size:.74rem;color:var(--text-muted)">Ver pendientes</p>
				</div>
			</a>
		</div>
	</div>
}

templ DashboardStats(d DashboardStatsData) {
	<div class="stats-grid">
		<div class="stat-card">
			<div class="stat-card-icon" style="background:rgba(215,16,85,.1)">
				<span class="material-symbols-outlined" style="color:var(--red)">storefront</span>
			</div>
			<div class="stat-card-value">{ fmt.Sprintf("%d", d.TiendasPublicadas) }</div>
			<div class="stat-card-label">Tiendas publicadas</div>
		</div>
		<div class="stat-card">
			<div class="stat-card-icon" style="background:rgba(6,160,224,.1)">
				<span class="material-symbols-outlined" style="color:var(--blue)">real_estate_agent</span>
			</div>
			<div class="stat-card-value">{ fmt.Sprintf("%d", d.LocalesDisponibles) }</div>
			<div class="stat-card-label">Locales disponibles</div>
		</div>
		<div class="stat-card">
			<div class="stat-card-icon" style="background:rgba(245,158,11,.1)">
				<span class="material-symbols-outlined" style="color:#f59e0b">event_seat</span>
			</div>
			<div class="stat-card-value">{ fmt.Sprintf("%d", d.ReservasPendientes) }</div>
			<div class="stat-card-label">Reservas pendientes</div>
		</div>
		<div class="stat-card">
			<div class="stat-card-icon" style="background:rgba(45,179,122,.1)">
				<span class="material-symbols-outlined" style="color:#2db37a">visibility</span>
			</div>
			<div class="stat-card-value">{ fmt.Sprintf("%d", d.VisitasHoy) }</div>
			<div class="stat-card-label">Visitas hoy</div>
		</div>
		<div class="stat-card">
			<div class="stat-card-icon" style="background:rgba(45,179,122,.07)">
				<span class="material-symbols-outlined" style="color:#2db37a">bar_chart</span>
			</div>
			<div class="stat-card-value">{ fmt.Sprintf("%d", d.VisitasSemana) }</div>
			<div class="stat-card-label">Visitas esta semana</div>
		</div>
		<div class="stat-card">
			<div class="stat-card-icon" style="background:rgba(172,198,13,.1)">
				<span class="material-symbols-outlined" style="color:var(--lime)">group_add</span>
			</div>
			<div class="stat-card-value">{ fmt.Sprintf("%d", d.NuevosLeads) }</div>
			<div class="stat-card-label">Nuevos leads (7 días)</div>
		</div>
	</div>
}
```

- [ ] **Step 3: Crear internal/view/pages/admin/tiendas_page.templ**

Contenido de `internal/view/pages/admin/tiendas_page.templ`:

```templ
package admin

import (
	"fmt"
	"strings"
)

type TiendaRow struct {
	ID        string
	Nombre    string
	Cat       string
	Gal       string
	Local     string
	Destacada bool
	Status    string
}

type TiendasPageData struct {
	Rows []TiendaRow
}

templ TiendasPage(d TiendasPageData) {
	<div>
		<div class="card" style="margin-bottom:20px;padding:16px 20px">
			<div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
				<input id="t-search" type="text" placeholder="Buscar por nombre..."
					style="flex:1;min-width:200px;padding:8px 12px;border-radius:var(--r-md);border:1px solid var(--border);background:var(--surface-2);color:var(--text);font-family:var(--font-body);font-size:.84rem;outline:none"
					oninput="filterTiendaRows()"/>
				<select id="t-cat" onchange="filterTiendaRows()"
					style="padding:8px 12px;border-radius:var(--r-md);border:1px solid var(--border);background:var(--surface-2);color:var(--text);font-size:.84rem;outline:none">
					<option value="">Todas las categorías</option>
					<option value="tiendas">Moda &amp; Tiendas</option>
					<option value="restaurantes">Restaurantes</option>
					<option value="farmacias">Farmacias</option>
					<option value="salud">Salud &amp; Belleza</option>
					<option value="tecnologia">Tecnología</option>
					<option value="servicios">Servicios</option>
				</select>
				<select id="t-gal" onchange="filterTiendaRows()"
					style="padding:8px 12px;border-radius:var(--r-md);border:1px solid var(--border);background:var(--surface-2);color:var(--text);font-size:.84rem;outline:none">
					<option value="">Todas las galerías</option>
					<option value="placa-comercial">Placa Comercial</option>
					<option value="torre-flamenco">Torre Flamenco</option>
				</select>
				<select id="t-status" onchange="filterTiendaRows()"
					style="padding:8px 12px;border-radius:var(--r-md);border:1px solid var(--border);background:var(--surface-2);color:var(--text);font-size:.84rem;outline:none">
					<option value="">Todos los estados</option>
					<option value="publicado">Publicado</option>
					<option value="borrador">Borrador</option>
				</select>
				<button class="topbar-btn topbar-btn-primary"
					hx-get="/admin/tiendas/new" hx-target="#modal-container" hx-swap="innerHTML">
					<span class="material-symbols-outlined" style="font-size:14px">add</span> Nueva tienda
				</button>
			</div>
		</div>
		<div class="card">
			<div class="card-header">
				<h2 class="card-title">Tiendas</h2>
				<span id="t-count" style="font-size:.76rem;color:var(--text-muted)">{ fmt.Sprintf("%d registros", len(d.Rows)) }</span>
			</div>
			<div style="overflow-x:auto">
				<table>
					<thead>
						<tr>
							<th>Nombre</th><th>Categoría</th><th>Galería</th><th>Local</th><th>Estado</th><th>Acciones</th>
						</tr>
					</thead>
					<tbody id="tiendas-tbody">
						if len(d.Rows) == 0 {
							<tr><td colspan="6" class="empty-state-cell">Sin tiendas — agrega una con el botón de arriba</td></tr>
						}
						for _, t := range d.Rows {
							@TiendaTableRow(t)
						}
					</tbody>
				</table>
			</div>
		</div>
		<script>
		function filterTiendaRows() {
			var q = document.getElementById('t-search').value.toLowerCase();
			var cat = document.getElementById('t-cat').value.toLowerCase();
			var gal = document.getElementById('t-gal').value.toLowerCase();
			var status = document.getElementById('t-status').value.toLowerCase();
			var visible = 0;
			document.querySelectorAll('#tiendas-tbody tr[data-name]').forEach(function(tr) {
				var show = tr.dataset.name.includes(q) &&
					(cat === '' || tr.dataset.cat === cat) &&
					(gal === '' || tr.dataset.gal === gal) &&
					(status === '' || tr.dataset.status === status);
				tr.style.display = show ? '' : 'none';
				if (show) visible++;
			});
			var el = document.getElementById('t-count');
			if (el) el.textContent = visible + ' registros';
		}
		</script>
	</div>
}

templ TiendaTableRow(t TiendaRow) {
	<tr data-name={ strings.ToLower(t.Nombre) } data-cat={ t.Cat } data-gal={ t.Gal } data-status={ t.Status }>
		<td style="font-weight:600">
			if t.Destacada {
				<span class="material-symbols-outlined" style="font-size:14px;color:var(--lime);vertical-align:-2px">star</span>
			}
			{ t.Nombre }
		</td>
		<td>{ t.Cat }</td>
		<td>{ galLabel(t.Gal) }</td>
		<td>{ t.Local }</td>
		<td><span class={ "badge badge-" + t.Status }>{ t.Status }</span></td>
		<td style="white-space:nowrap">
			<button class="btn-icon" title="Editar"
				hx-get={ "/admin/tiendas/" + t.ID + "/edit" }
				hx-target="#modal-container" hx-swap="innerHTML">
				<span class="material-symbols-outlined">edit</span>
			</button>
			<button class="btn-icon btn-danger" title="Eliminar"
				hx-delete={ "/admin/tiendas/" + t.ID }
				hx-confirm={ "¿Eliminar tienda " + t.Nombre + "?" }
				hx-target="closest tr" hx-swap="outerHTML swap:0.3s">
				<span class="material-symbols-outlined">delete</span>
			</button>
		</td>
	</tr>
}

func galLabel(gal string) string {
	switch gal {
	case "placa-comercial":
		return "Placa Comercial"
	case "torre-flamenco":
		return "Torre Flamenco"
	}
	return gal
}
```

- [ ] **Step 4: Crear internal/view/pages/admin/tienda_form.templ**

Contenido de `internal/view/pages/admin/tienda_form.templ`:

```templ
package admin

import "cms-plazareal/internal/view/components"

type TiendaFormData struct {
	ID         string
	Nombre     string
	Slug       string
	Cat        string
	Gal        string
	Local      string
	Logo       string
	Tags       string
	Desc       string
	About      string
	About2     string
	Pay        string
	Photos     string
	Similar    string
	Whatsapp   string
	Telefono   string
	Rating     string
	HorarioLV  string
	HorarioSab string
	HorarioDom string
	Status     string
	Destacada  bool
}

templ TiendaForm(d TiendaFormData) {
	@components.UploadFieldScript()
	<div class="modal-header">
		if d.ID == "" {
			<h3 class="modal-title">Nueva tienda</h3>
		} else {
			<h3 class="modal-title">Editar: { d.Nombre }</h3>
		}
		<button class="btn-icon"
			hx-get="/admin/empty"
			hx-target="#modal-container"
			hx-swap="innerHTML"
			title="Cerrar">
			<span class="material-symbols-outlined">close</span>
		</button>
	</div>
	<div class="modal-body" style="max-height:70vh;overflow-y:auto">
		if d.ID == "" {
			<form hx-post="/admin/tiendas" hx-target="#toast-area" hx-swap="innerHTML">
				@tiendaFormFields(d)
				<div class="modal-footer">
					<button type="submit" class="topbar-btn topbar-btn-primary">Crear tienda</button>
				</div>
			</form>
		} else {
			<form hx-put={ "/admin/tiendas/" + d.ID } hx-target="#toast-area" hx-swap="innerHTML">
				@tiendaFormFields(d)
				<div class="modal-footer">
					<button type="submit" class="topbar-btn topbar-btn-primary">Guardar cambios</button>
				</div>
			</form>
		}
	</div>
	<div id="toast-area" style="padding:0 24px 16px"></div>
}

templ tiendaFormFields(d TiendaFormData) {
	<div style="display:grid;grid-template-columns:1fr 1fr;gap:0 16px">
		<div class="form-group" style="grid-column:span 2">
			<label>Nombre *</label>
			<input type="text" name="nombre" value={ d.Nombre } required/>
		</div>
		<div class="form-group">
			<label>Slug *</label>
			<input type="text" name="slug" value={ d.Slug } required placeholder="nombre-tienda"/>
		</div>
		<div class="form-group">
			<label>Local (ej: Local 10)</label>
			<input type="text" name="local" value={ d.Local }/>
		</div>
		<div class="form-group">
			<label>Categoría</label>
			<select name="cat">
				<option value="tiendas" selected?={ d.Cat == "tiendas" }>Moda &amp; Tiendas</option>
				<option value="restaurantes" selected?={ d.Cat == "restaurantes" }>Restaurantes</option>
				<option value="farmacias" selected?={ d.Cat == "farmacias" }>Farmacias</option>
				<option value="salud" selected?={ d.Cat == "salud" }>Salud &amp; Belleza</option>
				<option value="tecnologia" selected?={ d.Cat == "tecnologia" }>Tecnología</option>
				<option value="servicios" selected?={ d.Cat == "servicios" }>Servicios</option>
			</select>
		</div>
		<div class="form-group">
			<label>Galería</label>
			<select name="gal">
				<option value="placa-comercial" selected?={ d.Gal == "placa-comercial" }>Placa Comercial</option>
				<option value="torre-flamenco" selected?={ d.Gal == "torre-flamenco" }>Torre Flamenco</option>
			</select>
		</div>
	</div>
	@components.UploadField("logo", d.Logo, "Logo")
	<div class="form-group" style="margin-top:12px">
		<label>Descripción corta</label>
		<textarea name="desc" rows="2">{ d.Desc }</textarea>
	</div>
	<div class="form-group">
		<label>Sobre la tienda</label>
		<textarea name="about" rows="3">{ d.About }</textarea>
	</div>
	<div style="display:grid;grid-template-columns:1fr 1fr;gap:0 16px">
		<div class="form-group">
			<label>WhatsApp (ej: 56912345678)</label>
			<input type="text" name="whatsapp" value={ d.Whatsapp }/>
		</div>
		<div class="form-group">
			<label>Teléfono</label>
			<input type="text" name="telefono" value={ d.Telefono }/>
		</div>
		<div class="form-group">
			<label>Horario Lun–Vie</label>
			<input type="text" name="horario_lv" value={ d.HorarioLV } placeholder="9:00 – 21:00"/>
		</div>
		<div class="form-group">
			<label>Horario Sábado</label>
			<input type="text" name="horario_sab" value={ d.HorarioSab }/>
		</div>
	</div>
	<div class="form-group">
		<label>Tags (separados por coma)</label>
		<input type="text" name="tags" value={ d.Tags } placeholder="Moda, Dama, Casual"/>
	</div>
	<div class="form-group">
		<label>Medios de pago</label>
		<input type="text" name="pay" value={ d.Pay } placeholder="Efectivo · Tarjetas"/>
	</div>
	<div style="display:grid;grid-template-columns:1fr 1fr;gap:0 16px">
		<div class="form-group">
			<label>Estado</label>
			<select name="status">
				<option value="publicado" selected?={ d.Status == "publicado" }>Publicado</option>
				<option value="borrador" selected?={ d.Status == "borrador" || d.Status == "" }>Borrador</option>
			</select>
		</div>
		<div class="form-group">
			<label style="display:flex;align-items:center;gap:8px;cursor:pointer;margin-top:20px">
				<input type="checkbox" name="destacada" value="on" checked?={ d.Destacada }/>
				Tienda destacada
			</label>
		</div>
	</div>
}
```

- [ ] **Step 5: Agregar ruta GET /admin/empty en cmd/server/main.go**

Esta ruta limpia el modal-container de forma segura vía HTMX (alternativa segura a innerHTML=''):

```go
// En el grupo adm:
adm.Get("/empty", func(c *fiber.Ctx) error {
    return c.SendString("")
})
```

- [ ] **Step 6: Actualizar handlers.go — Login, Dashboard, Tiendas**

Agregar imports al inicio de `internal/handlers/admin/handlers.go`:
```go
import (
    // ... existing imports ...
    "cms-plazareal/internal/helpers"
    adminView "cms-plazareal/internal/view/pages/admin"
    "cms-plazareal/internal/view/layout"
)
```

Reemplazar `LoginPage`:
```go
func LoginPage(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, adminView.Login(""))
	}
}
```

Reemplazar `Dashboard`:
```go
func Dashboard(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, adminView.Dashboard())
		}
		return helpers.Render(c, layout.Admin("Dashboard", "dashboard", adminView.Dashboard()))
	}
}
```

Reemplazar `DashboardStats`:
```go
func DashboardStats(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		publishedTiendas, _ := pb.FindRecordsByFilter("tiendas", "status = 'publicado'", "", 500, 0)
		localesDisp, _ := pb.FindRecordsByFilter("locales_disponibles", "status = 'disponible'", "", 200, 0)
		reservasPend, _ := pb.FindRecordsByFilter("reservas", "status = 'pendiente'", "", 200, 0)

		hoyStr := time.Now().Format("2006-01-02")
		visitasHoy, _ := pb.FindRecordsByFilter("page_views", "created >= '"+hoyStr+"'", "", 500, 0)
		semanaStr := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		visitasSemana, _ := pb.FindRecordsByFilter("page_views", "created >= '"+semanaStr+"'", "", 2000, 0)
		leadsRecs, _ := pb.FindRecordsByFilter("leads", "created >= '"+semanaStr+"'", "", 500, 0)

		d := adminView.DashboardStatsData{
			TiendasPublicadas:  len(publishedTiendas),
			LocalesDisponibles: len(localesDisp),
			ReservasPendientes: len(reservasPend),
			VisitasHoy:         len(visitasHoy),
			VisitasSemana:      len(visitasSemana),
			NuevosLeads:        len(leadsRecs),
		}
		return helpers.Render(c, adminView.DashboardStats(d))
	}
}
```

Reemplazar `TiendasList`:
```go
func TiendasList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter("tiendas", "", "nombre", 500, 0)
		var rows []adminView.TiendaRow
		for _, r := range records {
			rows = append(rows, adminView.TiendaRow{
				ID: r.Id, Nombre: r.GetString("nombre"),
				Cat: r.GetString("cat"), Gal: r.GetString("gal"),
				Local: r.GetString("local"), Destacada: r.GetBool("destacada"),
				Status: r.GetString("status"),
			})
		}
		content := adminView.TiendasPage(adminView.TiendasPageData{Rows: rows})
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, content)
		}
		return helpers.Render(c, layout.Admin("Tiendas", "tiendas", content))
	}
}
```

Reemplazar `TiendaNew` (o crear si no existe):
```go
func TiendaNew(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, adminView.TiendaForm(adminView.TiendaFormData{
			Status: "borrador", Gal: "placa-comercial",
		}))
	}
}
```

Reemplazar `TiendaEdit`:
```go
func TiendaEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("tiendas", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("Tienda no encontrada")
		}
		return helpers.Render(c, adminView.TiendaForm(adminView.TiendaFormData{
			ID: r.Id, Nombre: r.GetString("nombre"), Slug: r.GetString("slug"),
			Cat: r.GetString("cat"), Gal: r.GetString("gal"), Local: r.GetString("local"),
			Logo: r.GetString("logo"), Tags: r.GetString("tags"), Desc: r.GetString("desc"),
			About: r.GetString("about"), About2: r.GetString("about2"), Pay: r.GetString("pay"),
			Photos: r.GetString("photos"), Similar: r.GetString("similar"),
			Whatsapp: r.GetString("whatsapp"), Telefono: r.GetString("telefono"),
			Rating: r.GetString("rating"), HorarioLV: r.GetString("horario_lv"),
			HorarioSab: r.GetString("horario_sab"), HorarioDom: r.GetString("horario_dom"),
			Status: r.GetString("status"), Destacada: r.GetBool("destacada"),
		}))
	}
}
```

Los handlers `TiendaCreate`, `TiendaUpdate`, `TiendaDelete` permanecen igual (reciben form, persisten, devuelven toast string).

- [ ] **Step 7: templ generate + compilar**

```bash
make generate
go build ./...
```

Errores comunes:
- `undefined: fmt` → verificar `import "fmt"` al inicio del archivo .templ
- `undefined: strings` → verificar `import "strings"` al inicio del archivo .templ
- Import cycle → verificar que view/pages/admin no importa handlers

- [ ] **Step 8: Test manual**

```bash
go run cmd/server/main.go
```

1. `http://localhost:3000/admin/login` → card de login glass dark, formulario funcional
2. Login correcto → redirect dashboard
3. Dashboard → sidebar con links, stat cards en loading → carga stats
4. `/admin/tiendas` → tabla con filtros, botón "Nueva tienda"
5. Click "Nueva tienda" → modal con form y upload widget
6. Click X en modal → limpia modal vía HTMX a /admin/empty

- [ ] **Step 9: Commit**

```bash
git add internal/view/pages/admin/ internal/handlers/admin/handlers.go cmd/server/main.go
git commit -m "feat: admin login, dashboard, tiendas in templ — Liquid Glass dark design"
```

---

## Estado del repositorio al finalizar este task

- Login, Dashboard y Tiendas admin en templ con design Liquid Glass dark
- `layout.Admin` compuesto correctamente en cada handler
- `DashboardStats` carga datos reales de PocketBase
- `TiendaForm` con `UploadField` para logos
- Cierre de modal seguro vía `GET /admin/empty`
- `go build ./...` pasa sin errores

---

## Amendments (agregados después del diseño inicial)

### A1: Campo hero_bg en TiendaFormData y TiendaForm

Agregar `HeroBg string` al struct `TiendaFormData`.

En `TiendaForm`, agregar después del `UploadField` de logo:

```templ
@components.UploadField("hero_bg", d.HeroBg, "Imagen de fondo hero (opcional)")
```

El handler `TiendaCreate` y `TiendaUpdate` deben leer y guardar `c.FormValue("hero_bg")`.
El handler `TiendaEdit` debe cargar `r.GetString("hero_bg")` en el form.

La colección `tiendas` debe tener el campo en `collections.go`:
```go
&core.TextField{Name: "hero_bg"},
```

### A2: Campo status_horario en TiendaForm

Agregar `StatusHorario string` al struct `TiendaFormData`.

En `TiendaForm`, dentro de la sección de horarios, agregar:

```templ
<div class="form-group">
    <label class="form-label">Estado de apertura</label>
    <select name="status_horario" class="form-select">
        <option value="normal" selected?={ d.StatusHorario == "normal" || d.StatusHorario == "" }>Calculado por horario</option>
        <option value="solo-reserva" selected?={ d.StatusHorario == "solo-reserva" }>Solo con reserva</option>
        <option value="cerrado-temporal" selected?={ d.StatusHorario == "cerrado-temporal" }>Cerrado temporalmente</option>
    </select>
    <span class="form-hint">Sobreescribe el cálculo automático de abierto/cerrado.</span>
</div>
```

Los handlers Create/Update guardan `c.FormValue("status_horario")`.
El campo en `collections.go`:
```go
&core.SelectField{Name: "status_horario", Values: []string{"normal", "solo-reserva", "cerrado-temporal"}},
```
