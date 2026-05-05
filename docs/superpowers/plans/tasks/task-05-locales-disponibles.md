# Task 05 — Locales Disponibles: colección, admin CRUD y página pública

**Depends on:** Task 01, Task 02
**Estimated complexity:** media — patrón idéntico al módulo tiendas

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Rutas públicas existentes: /, /buscador-tiendas.html, /tienda-individual.html, /noticias.html
Fragmentos HTMX existentes: /fragments/tiendas, /fragments/tiendas-destacadas, /fragments/noticias
Admin CRUD existente: tiendas (referencia de patrón)
```

El mall Plaza Real ofrece locales comerciales en arriendo directamente. Esta sección no existe aún y debe crearse completa.

**Patrón a seguir:** El módulo `tiendas` es el modelo de referencia. La colección, los handlers de admin, el fragmento HTMX y la página pública siguen exactamente el mismo patrón.

Archivo de referencia para el patrón admin: `internal/handlers/fragments/tiendas.go`
Archivo de referencia para el patrón página pública: `web/buscador-tiendas.html`

---

## Objetivo

Crear la sección "Locales Disponibles" completa: colección PocketBase, admin CRUD, fragmento HTMX público con filtros por tipo de uso, y página pública `/locales.html`.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Modificar | `internal/auth/collections.go` — agregar colección `locales_disponibles` |
| Crear | `internal/handlers/fragments/locales.go` |
| Crear | `web/locales.html` |
| Crear | `internal/templates/admin/pages/locales.html` |
| Modificar | `internal/handlers/admin/handlers.go` — agregar CRUD locales |
| Modificar | `cmd/server/main.go` — registrar rutas |
| Modificar | `web/index.html`, `web/buscador-tiendas.html`, `web/noticias.html` — agregar tab nav |

---

## Implementación

- [ ] **Step 1: Agregar colección locales_disponibles en collections.go**

Buscar el final de `ensureCollections`:
```bash
grep -n "return nil" internal/auth/collections.go | tail -3
```

Insertar **antes** del último `return nil`:

```go
// ── LOCALES DISPONIBLES ──
if _, err := app.FindCollectionByNameOrId("locales_disponibles"); err != nil {
    col := core.NewBaseCollection("locales_disponibles")
    col.Fields.Add(
        &core.TextField{Name: "nombre", Required: true},
        &core.TextField{Name: "numero_local"},
        &core.TextField{Name: "piso"},
        &core.NumberField{Name: "superficie_m2"},
        // retail|restaurante|servicios|oficina|otros
        &core.TextField{Name: "tipo_uso"},
        &core.NumberField{Name: "precio_uf"},
        &core.EditorField{Name: "descripcion"},
        &core.TextField{Name: "photos"},
        &core.TextField{Name: "cover"},
        &core.DateField{Name: "disponible_desde"},
        // disponible|reservado|arrendado
        &core.TextField{Name: "status"},
        &core.TextField{Name: "contacto_email"},
        &core.TextField{Name: "contacto_telefono"},
    )
    if err := app.Save(col); err != nil {
        return err
    }
    log.Println("  ✅ Collection 'locales_disponibles' created")
}
```

- [ ] **Step 2: Crear internal/handlers/fragments/locales.go**

```go
package fragments

import (
	"fmt"
	"html/template"
	"strings"

	"cms-plazareal/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
)

// LocalesPage devuelve la grilla de locales disponibles filtrada por tipo_uso.
func LocalesPage(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tipoUso := c.Query("tipo_uso")
		filter := "status = 'disponible'"
		if tipoUso != "" && tipoUso != "todos" {
			filter += fmt.Sprintf(" && tipo_uso = '%s'",
				strings.ReplaceAll(tipoUso, "'", ""))
		}

		records, err := pb.FindRecordsByFilter("locales_disponibles", filter, "-created", 50, 0)
		if err != nil {
			return c.Status(500).SendString(`<p class="error-msg">Error cargando locales.</p>`)
		}

		if len(records) == 0 {
			return c.SendString(`
<div class="empty-state">
  <span class="material-symbols-outlined" style="font-size:48px;color:var(--md-outline)">store_off</span>
  <p>No hay locales disponibles en este momento.</p>
  <p style="font-size:14px;color:var(--md-outline)">Consulta próximamente o contáctanos directamente.</p>
</div>`)
		}

		var sb strings.Builder
		sb.WriteString(`<div class="locales-grid">`)
		for _, r := range records {
			nombre := template.HTMLEscapeString(r.GetString("nombre"))
			local := template.HTMLEscapeString(r.GetString("numero_local"))
			piso := template.HTMLEscapeString(r.GetString("piso"))
			tipoStr := template.HTMLEscapeString(r.GetString("tipo_uso"))
			sup := r.GetFloat("superficie_m2")
			precio := r.GetFloat("precio_uf")
			cover := r.GetString("cover")
			email := template.HTMLEscapeString(r.GetString("contacto_email"))
			if email == "" {
				email = "arriendo@plazareal.cl"
			}

			imgHTML := `<div class="local-cover-placeholder"><span class="material-symbols-outlined">store</span></div>`
			if cover != "" {
				imgHTML = fmt.Sprintf(`<img src="%s" alt="%s" class="local-cover" loading="lazy"/>`,
					template.HTMLEscapeString(cover), nombre)
			}

			precioStr := ""
			if precio > 0 {
				precioStr = fmt.Sprintf(`<span class="local-precio">UF %.0f/mes</span>`, precio)
			}

			supStr := ""
			if sup > 0 {
				supStr = fmt.Sprintf(`<span class="local-sup">%g m²</span>`, sup)
			}

			ubicacion := ""
			if local != "" || piso != "" {
				ubicacion = fmt.Sprintf(`<p class="local-ubicacion">Local %s · Piso %s</p>`, local, piso)
			}

			sb.WriteString(fmt.Sprintf(`
<div class="local-card">
  %s
  <div class="local-info">
    <h3 class="local-nombre">%s</h3>
    <div class="local-meta">
      <span class="local-tipo">%s</span>
      %s%s
    </div>
    %s
    <a href="mailto:%s?subject=Consulta%%20local%%20%s"
       class="btn-consultar"
       hx-post="/api/leads"
       hx-vals='{"tipo":"consulta_local","source":"%s"}'
       hx-swap="none">
      Consultar disponibilidad
    </a>
  </div>
</div>`, imgHTML, nombre, tipoStr, supStr, precioStr,
				ubicacion, email,
				template.URLQueryEscaper(r.GetString("nombre")),
				template.HTMLEscapeString(r.GetString("nombre"))))
		}
		sb.WriteString(`</div>`)

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}
```

- [ ] **Step 3: Crear web/locales.html**

Leer `web/buscador-tiendas.html` como referencia de estructura (navegación, hero, filtros, grid HTMX). Crear `web/locales.html` con estas diferencias clave:

1. `<title>` → `Locales Disponibles — Plaza Real`
2. El héroe: título "Locales en Arriendo", descripción sobre espacios disponibles
3. Los filtros de categoría → tipos de uso:
```html
<div class="filter-bar" id="locales-filters">
  <button class="filter-chip active"
    hx-get="/fragments/locales"
    hx-target="#locales-results"
    hx-swap="innerHTML"
    onclick="setActive(this)">Todos</button>
  <button class="filter-chip"
    hx-get="/fragments/locales?tipo_uso=retail"
    hx-target="#locales-results"
    hx-swap="innerHTML"
    onclick="setActive(this)">Retail</button>
  <button class="filter-chip"
    hx-get="/fragments/locales?tipo_uso=restaurante"
    hx-target="#locales-results"
    hx-swap="innerHTML"
    onclick="setActive(this)">Gastronómico</button>
  <button class="filter-chip"
    hx-get="/fragments/locales?tipo_uso=servicios"
    hx-target="#locales-results"
    hx-swap="innerHTML"
    onclick="setActive(this)">Servicios</button>
  <button class="filter-chip"
    hx-get="/fragments/locales?tipo_uso=oficina"
    hx-target="#locales-results"
    hx-swap="innerHTML"
    onclick="setActive(this)">Oficinas</button>
</div>

<div id="locales-results"
     hx-get="/fragments/locales"
     hx-trigger="load"
     hx-swap="innerHTML">
  <div class="loading-state">
    <div class="spinner"></div>
    <p>Cargando locales disponibles...</p>
  </div>
</div>
```

4. La navegación debe incluir el tab de Locales marcado como activo.

- [ ] **Step 4: Agregar admin handlers en handlers/admin/handlers.go**

Agregar al final del archivo (el patrón es idéntico al de tiendas):

```go
// ── LOCALES DISPONIBLES CRUD ──

func LocalesList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter("locales_disponibles", "", "-created", 200, 0)
		type row struct {
			ID, Nombre, Piso, Local, TipoUso, Status, Superficie, PrecioUF string
		}
		var rows []row
		for _, r := range records {
			rows = append(rows, row{
				ID:         r.Id,
				Nombre:     r.GetString("nombre"),
				Piso:       r.GetString("piso"),
				Local:      r.GetString("numero_local"),
				TipoUso:    r.GetString("tipo_uso"),
				Status:     r.GetString("status"),
				Superficie: fmt.Sprintf("%.0f m²", r.GetFloat("superficie_m2")),
				PrecioUF:   fmt.Sprintf("UF %.0f", r.GetFloat("precio_uf")),
			})
		}
		tmpl, err := template.ParseFiles("./internal/templates/admin/pages/locales.html")
		if err != nil {
			return c.Status(500).SendString("Template error: " + err.Error())
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.Execute(c, map[string]any{"Locales": rows})
	}
}

func LocalForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tmpl, _ := template.ParseFiles("./internal/templates/admin/pages/locales.html")
		c.Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.ExecuteTemplate(c, "form", map[string]any{"Local": nil})
	}
}

func LocalCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		col, err := pb.FindCollectionByNameOrId("locales_disponibles")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error interno</div>`)
		}
		r := core.NewRecord(col)
		r.Set("nombre", c.FormValue("nombre"))
		r.Set("numero_local", c.FormValue("numero_local"))
		r.Set("piso", c.FormValue("piso"))
		r.Set("tipo_uso", c.FormValue("tipo_uso"))
		r.Set("superficie_m2", c.FormValue("superficie_m2"))
		r.Set("precio_uf", c.FormValue("precio_uf"))
		r.Set("descripcion", c.FormValue("descripcion"))
		r.Set("photos", c.FormValue("photos"))
		r.Set("cover", c.FormValue("cover"))
		r.Set("contacto_email", c.FormValue("contacto_email"))
		r.Set("contacto_telefono", c.FormValue("contacto_telefono"))
		r.Set("status", "disponible")
		if err := pb.Save(r); err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error creando local</div>`)
		}
		c.Set("HX-Redirect", "/admin/locales")
		return c.SendString(`<div class="toast toast-success">Local creado</div>`)
	}
}

func LocalEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("locales_disponibles", c.Params("id"))
		if err != nil {
			return c.Redirect("/admin/locales")
		}
		tmpl, _ := template.ParseFiles("./internal/templates/admin/pages/locales.html")
		c.Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.ExecuteTemplate(c, "form", map[string]any{"Local": r})
	}
}

func LocalUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("locales_disponibles", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Local no encontrado</div>`)
		}
		r.Set("nombre", c.FormValue("nombre"))
		r.Set("numero_local", c.FormValue("numero_local"))
		r.Set("piso", c.FormValue("piso"))
		r.Set("tipo_uso", c.FormValue("tipo_uso"))
		r.Set("superficie_m2", c.FormValue("superficie_m2"))
		r.Set("precio_uf", c.FormValue("precio_uf"))
		r.Set("descripcion", c.FormValue("descripcion"))
		r.Set("photos", c.FormValue("photos"))
		r.Set("cover", c.FormValue("cover"))
		r.Set("contacto_email", c.FormValue("contacto_email"))
		r.Set("contacto_telefono", c.FormValue("contacto_telefono"))
		if err := pb.Save(r); err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error actualizando</div>`)
		}
		c.Set("HX-Redirect", "/admin/locales")
		return c.SendString(`<div class="toast toast-success">Local actualizado</div>`)
	}
}

func LocalDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("locales_disponibles", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Local no encontrado</div>`)
		}
		pb.Delete(r)
		return c.SendStatus(200)
	}
}

func LocalToggleStatus(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("locales_disponibles", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Local no encontrado</div>`)
		}
		next := "disponible"
		if r.GetString("status") == "disponible" {
			next = "arrendado"
		}
		r.Set("status", next)
		pb.Save(r)
		c.Set("HX-Redirect", "/admin/locales")
		return c.SendStatus(200)
	}
}
```

- [ ] **Step 5: Crear internal/templates/admin/pages/locales.html**

Copiar `internal/templates/admin/pages/tiendas.html` como base con:
```bash
cp internal/templates/admin/pages/tiendas.html internal/templates/admin/pages/locales.html
```

Modificar `locales.html`:
- Título de la página: "Locales Disponibles"
- Columnas de la tabla: Nombre, Tipo de Uso, Piso, Local N°, Superficie, Precio UF, Status, Acciones
- Formulario de creación/edición: campos `nombre`, `numero_local`, `piso`, `tipo_uso` (select), `superficie_m2`, `precio_uf`, `descripcion`, `cover`, `contacto_email`, `contacto_telefono`
- El iterador del template: `{{range .Locales}}` en lugar de `{{range .Tiendas}}`

- [ ] **Step 6: Registrar rutas en cmd/server/main.go**

Buscar el bloque de rutas de tiendas como referencia:
```bash
grep -n "tiendas\|locales" cmd/server/main.go | head -20
```

Agregar después del bloque de tiendas:

```go
// Locales disponibles (public)
app.Get("/locales.html", web.PageHandler(cfg, "locales"))
frag.Get("/locales", fragments.LocalesPage(cfg, pb))

// Admin — Locales
adm.Get("/locales", admin.LocalesList(cfg, pb))
adm.Get("/locales/new", admin.LocalForm(cfg))
adm.Post("/locales", admin.LocalCreate(cfg, pb))
adm.Get("/locales/:id/edit", admin.LocalEdit(cfg, pb))
adm.Put("/locales/:id", admin.LocalUpdate(cfg, pb))
adm.Delete("/locales/:id", admin.LocalDelete(cfg, pb))
adm.Post("/locales/:id/publish", admin.LocalToggleStatus(cfg, pb))
```

- [ ] **Step 7: Agregar tab "Locales" al sidebar del admin**

En `internal/templates/admin/pages/dashboard.html`, buscar el bloque de navegación "Plaza Real":
```bash
grep -n "Plaza Real\|storefront\|Tiendas" internal/templates/admin/pages/dashboard.html | head -10
```

Agregar después del link de Tiendas:
```html
<a href="/admin/locales" class="sidebar-link"
   hx-get="/admin/locales"
   hx-target="#main-content"
   hx-push-url="true">
  <span class="material-symbols-outlined">store_mall_directory</span> Locales Disponibles
</a>
```

- [ ] **Step 8: Agregar tab "Locales" en navegación pública**

En `web/index.html`, `web/buscador-tiendas.html`, `web/noticias.html` — buscar la barra de navegación:
```bash
grep -n "nav\|buscador-tiendas\|noticias" web/index.html | head -10
```

Agregar el link a `/locales.html` en la navegación de cada página. La navegación completa del sitio debe quedar:
- Inicio (`/index.html`)
- Tiendas (`/buscador-tiendas.html`)
- Locales Disponibles (`/locales.html`) ← nuevo
- Eventos (`/eventos.html`) ← se agrega en Task 06
- Noticias (`/noticias.html`)

- [ ] **Step 9: Compilar**

```bash
go build ./...
# Esperado: sin errores
```

- [ ] **Step 10: Test manual**

```bash
go run cmd/server/main.go
```

1. Abrir `http://localhost:3000/locales.html` → debe cargar con empty state (sin locales aún)
2. Abrir `http://localhost:3000/admin/locales` → debe mostrar tabla vacía con botón "Nuevo"
3. Crear un local desde el admin → debe aparecer en la tabla
4. Volver a `/locales.html` → debe aparecer el local creado
5. Probar filtros: Retail, Gastronómico, etc. → solo deben aparecer los del tipo seleccionado

- [ ] **Step 11: Commit**

```bash
git add internal/auth/collections.go \
        internal/handlers/fragments/locales.go \
        internal/handlers/admin/handlers.go \
        internal/templates/admin/pages/locales.html \
        web/locales.html \
        web/index.html web/buscador-tiendas.html web/noticias.html \
        internal/templates/admin/pages/dashboard.html \
        cmd/server/main.go
git commit -m "feat: add locales_disponibles — collection, admin CRUD, public page with HTMX filters"
```

---

## Estado del repositorio al finalizar este task

- Colección `locales_disponibles` existe en PocketBase
- `/locales.html` muestra locales disponibles con filtros por tipo de uso
- Admin `/admin/locales` permite crear/editar/eliminar locales
- Tab "Locales Disponibles" aparece en la navegación de todas las páginas públicas
- `go build ./...` pasa sin errores
