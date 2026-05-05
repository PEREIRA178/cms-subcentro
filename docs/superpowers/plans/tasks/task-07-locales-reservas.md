# Task 07 — Locales Disponibles + Reservas: colecciones, admin, páginas públicas

**Depends on:** Task 05 (admin shell), Task 02 (design system)
**Estimated complexity:** alta — dos colecciones nuevas, admin + público + lógica de disponibilidad

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Layout admin ya existe: internal/view/layout/admin.templ
Render helper ya existe: internal/helpers/render.go
Colecciones existentes: tiendas, content_blocks, devices, playlists, page_views
Colecciones a crear: locales_disponibles, reservas
```

---

## Objetivo

Crear las colecciones `locales_disponibles` y `reservas` en PocketBase. CRUD admin para ambas. Página pública de locales disponibles (para potenciales arrendatarios). Sistema de reservas para eventos de tipo PROMOCION con cupo limitado.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Modificar | `internal/auth/collections.go` — agregar colecciones |
| Crear | `internal/view/pages/admin/locales_page.templ` |
| Crear | `internal/view/pages/admin/local_form.templ` |
| Crear | `internal/view/pages/admin/reservas_page.templ` |
| Crear | `internal/view/pages/public/locales_page.templ` |
| Crear | `internal/view/fragments/locales_cards.templ` |
| Modificar | `internal/handlers/admin/handlers.go` — agregar handlers locales + reservas |
| Modificar | `internal/handlers/web/handlers.go` — página pública locales |
| Modificar | `cmd/server/main.go` — registrar rutas |

---

## Implementación

- [ ] **Step 1: Definir colecciones en internal/auth/collections.go**

Buscar dónde se definen las colecciones existentes:
```bash
grep -n "func.*Collection\|CreateCollection\|FindOrCreate\|collectionDef\|initCollections" internal/auth/collections.go | head -20
```

Agregar la definición de `locales_disponibles`. Buscar un ejemplo de cómo se define una colección en el archivo (observar la función y estructura usada), luego agregar:

```go
// locales_disponibles — espacios comerciales en arriendo
// Campos:
//   nombre      string   "Local 101 — Placa Comercial"
//   galeria     select   ["placa-comercial", "torre-flamenco"]
//   numero      string   "101"
//   piso        string   "1"
//   m2          number   superficie en m²
//   descripcion string
//   precio_ref  string   "UF X / mes" (referencial, no contractual)
//   estado      select   ["disponible", "en-negociacion", "arrendado"]
//   imagen_url  string
//   contacto_email string
//   contacto_tel   string
```

El patrón de definición seguirá el mismo que otros collections en el archivo. Ejemplo de estructura esperada (ajustar según el patrón real encontrado):

```go
&migrate.CreateCollection{
    Collection: &core.Collection{
        Name: "locales_disponibles",
        Type: core.CollectionTypeBase,
        Fields: core.FieldsList{
            &core.TextField{Name: "nombre", Required: true},
            &core.SelectField{Name: "galeria", Values: []string{"placa-comercial", "torre-flamenco"}},
            &core.TextField{Name: "numero"},
            &core.TextField{Name: "piso"},
            &core.NumberField{Name: "m2"},
            &core.TextField{Name: "descripcion"},
            &core.TextField{Name: "precio_ref"},
            &core.SelectField{Name: "estado", Values: []string{"disponible", "en-negociacion", "arrendado"}, Required: true},
            &core.TextField{Name: "imagen_url"},
            &core.TextField{Name: "contacto_email"},
            &core.TextField{Name: "contacto_tel"},
        },
    },
},
```

Agregar también `reservas`:

```go
// reservas — reservas para eventos PROMOCION con cupo limitado
// Campos:
//   content_block_id  relation  → content_blocks (ID del PROMOCION de tipo evento)
//   nombre            string    nombre del visitante
//   email             string
//   telefono          string
//   cantidad          number    número de personas
//   estado            select    ["pendiente", "confirmada", "cancelada"]
//   notas             string    (campo libre para el admin)
```

```go
&migrate.CreateCollection{
    Collection: &core.Collection{
        Name: "reservas",
        Type: core.CollectionTypeBase,
        Fields: core.FieldsList{
            &core.RelationField{Name: "content_block_id", CollectionId: "content_blocks"},
            &core.TextField{Name: "nombre", Required: true},
            &core.EmailField{Name: "email", Required: true},
            &core.TextField{Name: "telefono"},
            &core.NumberField{Name: "cantidad"},
            &core.SelectField{Name: "estado", Values: []string{"pendiente", "confirmada", "cancelada"}},
            &core.TextField{Name: "notas"},
        },
    },
},
```

**Nota:** Verificar los tipos exactos usados en el archivo (ej: `core.TextField` vs `core.StringField`, `core.NumberField` vs `core.FloatField`). Ajustar según los imports y tipos reales.

- [ ] **Step 2: Crear internal/view/pages/admin/locales_page.templ**

```templ
package admin

import "cms-plazareal/internal/view/layout"

type LocalRow struct {
	ID        string
	Nombre    string
	Galeria   string
	Numero    string
	M2        float64
	Estado    string
}

type LocalesPageData struct {
	Rows []LocalRow
}

templ LocalesPage(d LocalesPageData) {
	@layout.Admin("Locales Disponibles", "locales", localesPageBody(d))
}

templ localesPageBody(d LocalesPageData) {
	<div class="admin-page">
		<div class="page-header">
			<h1 class="page-title">Locales Disponibles</h1>
			<button
				class="btn btn-primary"
				hx-get="/admin/locales/new"
				hx-target="#modal-container"
				hx-swap="innerHTML"
			>
				<span class="material-symbols-outlined">add</span>
				Nuevo Local
			</button>
		</div>
		<div class="table-card">
			<table class="admin-table">
				<thead>
					<tr>
						<th>Nombre</th>
						<th>Galería</th>
						<th>N°</th>
						<th>m²</th>
						<th>Estado</th>
						<th></th>
					</tr>
				</thead>
				<tbody id="locales-tbody">
					for _, row := range d.Rows {
						@LocalTableRow(row)
					}
				</tbody>
			</table>
		</div>
		<div id="modal-container"></div>
	</div>
}

templ LocalTableRow(r LocalRow) {
	<tr id={ "local-row-" + r.ID }>
		<td class="font-medium">{ r.Nombre }</td>
		<td>{ r.Galeria }</td>
		<td>{ r.Numero }</td>
		<td>{ fmt.Sprintf("%.0f", r.M2) }</td>
		<td>
			<span class={
				"badge",
				templ.KV("badge-success", r.Estado == "disponible"),
				templ.KV("badge-warning", r.Estado == "en-negociacion"),
				templ.KV("badge-muted", r.Estado == "arrendado"),
			}>{ r.Estado }</span>
		</td>
		<td class="row-actions">
			<button class="btn-icon"
				hx-get={ "/admin/locales/" + r.ID + "/edit" }
				hx-target="#modal-container"
				hx-swap="innerHTML"
				title="Editar">
				<span class="material-symbols-outlined">edit</span>
			</button>
			<button class="btn-icon btn-danger"
				hx-delete={ "/admin/locales/" + r.ID }
				hx-target={ "#local-row-" + r.ID }
				hx-swap="outerHTML"
				hx-confirm="¿Eliminar este local?"
				title="Eliminar">
				<span class="material-symbols-outlined">delete</span>
			</button>
		</td>
	</tr>
}
```

- [ ] **Step 3: Crear internal/view/pages/admin/local_form.templ**

```templ
package admin

import "cms-plazareal/internal/view/components"

type LocalFormData struct {
	ID            string
	Nombre        string
	Galeria       string
	Numero        string
	Piso          string
	M2            string
	Descripcion   string
	PrecioRef     string
	Estado        string
	ImagenURL     string
	ContactoEmail string
	ContactoTel   string
	ErrorMsg      string
}

templ LocalForm(d LocalFormData) {
	<div class="modal-overlay" hx-get="/admin/empty" hx-trigger="click[target===this]" hx-target="#modal-container" hx-swap="innerHTML">
		<div class="modal-card modal-card-lg" @click.stop="">
			<div class="modal-header">
				<h2 class="modal-title">
					if d.ID == "" { Nuevo Local } else { Editar Local }
				</h2>
				<button class="btn-icon modal-close" hx-get="/admin/empty" hx-target="#modal-container" hx-swap="innerHTML">
					<span class="material-symbols-outlined">close</span>
				</button>
			</div>
			if d.ErrorMsg != "" {
				<div class="alert alert-danger">{ d.ErrorMsg }</div>
			}
			<form
				if d.ID == "" { hx-post="/admin/locales" } else { hx-put={ "/admin/locales/" + d.ID } }
				hx-target={ "#local-row-" + d.ID }
				hx-swap="outerHTML"
				enctype="multipart/form-data"
				class="modal-form"
			>
				<div class="form-row">
					<div class="form-group form-group-grow">
						<label class="form-label">Nombre / Descripción del local *</label>
						<input type="text" name="nombre" class="form-input" value={ d.Nombre } required placeholder="Local 101 — Placa Comercial"/>
					</div>
				</div>
				<div class="form-row">
					<div class="form-group">
						<label class="form-label">Galería</label>
						<select name="galeria" class="form-select">
							<option value="">— Seleccionar —</option>
							<option value="placa-comercial" selected?={ d.Galeria == "placa-comercial" }>Placa Comercial</option>
							<option value="torre-flamenco" selected?={ d.Galeria == "torre-flamenco" }>Torre Flamenco</option>
						</select>
					</div>
					<div class="form-group">
						<label class="form-label">N° Local</label>
						<input type="text" name="numero" class="form-input" value={ d.Numero } placeholder="101"/>
					</div>
					<div class="form-group">
						<label class="form-label">Piso</label>
						<input type="text" name="piso" class="form-input" value={ d.Piso } placeholder="1"/>
					</div>
					<div class="form-group">
						<label class="form-label">Superficie (m²)</label>
						<input type="number" name="m2" class="form-input" value={ d.M2 } placeholder="0" step="0.1"/>
					</div>
				</div>
				<div class="form-group">
					<label class="form-label">Descripción</label>
					<textarea name="descripcion" class="form-textarea" rows="3" placeholder="Características del local, orientación, acceso...">{ d.Descripcion }</textarea>
				</div>
				<div class="form-row">
					<div class="form-group">
						<label class="form-label">Precio referencial</label>
						<input type="text" name="precio_ref" class="form-input" value={ d.PrecioRef } placeholder="UF 30 / mes"/>
					</div>
					<div class="form-group">
						<label class="form-label">Estado *</label>
						<select name="estado" class="form-select" required>
							<option value="disponible" selected?={ d.Estado == "disponible" || d.Estado == "" }>Disponible</option>
							<option value="en-negociacion" selected?={ d.Estado == "en-negociacion" }>En negociación</option>
							<option value="arrendado" selected?={ d.Estado == "arrendado" }>Arrendado</option>
						</select>
					</div>
				</div>
				@components.UploadField("imagen_url", d.ImagenURL, "Imagen del local")
				<div class="form-row">
					<div class="form-group">
						<label class="form-label">Email de contacto</label>
						<input type="email" name="contacto_email" class="form-input" value={ d.ContactoEmail }/>
					</div>
					<div class="form-group">
						<label class="form-label">Teléfono de contacto</label>
						<input type="text" name="contacto_tel" class="form-input" value={ d.ContactoTel }/>
					</div>
				</div>
				<div class="modal-footer">
					<button type="button" class="btn btn-ghost" hx-get="/admin/empty" hx-target="#modal-container" hx-swap="innerHTML">Cancelar</button>
					<button type="submit" class="btn btn-primary">Guardar</button>
				</div>
			</form>
			@components.UploadFieldScript()
		</div>
	</div>
}
```

- [ ] **Step 4: Crear internal/view/pages/admin/reservas_page.templ**

```templ
package admin

import "cms-plazareal/internal/view/layout"

type ReservaRow struct {
	ID              string
	Nombre          string
	Email           string
	Telefono        string
	Cantidad        int
	Estado          string
	ContentBlockTitle string
	CreatedAt       string
}

type ReservasPageData struct {
	Rows []ReservaRow
}

templ ReservasPage(d ReservasPageData) {
	@layout.Admin("Reservas", "reservas", reservasPageBody(d))
}

templ reservasPageBody(d ReservasPageData) {
	<div class="admin-page">
		<div class="page-header">
			<h1 class="page-title">Reservas</h1>
		</div>
		<div class="table-card">
			<table class="admin-table">
				<thead>
					<tr>
						<th>Evento</th>
						<th>Nombre</th>
						<th>Email</th>
						<th>Personas</th>
						<th>Estado</th>
						<th>Fecha</th>
						<th></th>
					</tr>
				</thead>
				<tbody id="reservas-tbody">
					for _, row := range d.Rows {
						@ReservaTableRow(row)
					}
				</tbody>
			</table>
		</div>
	</div>
}

templ ReservaTableRow(r ReservaRow) {
	<tr id={ "reserva-row-" + r.ID }>
		<td class="text-sm text-muted">{ r.ContentBlockTitle }</td>
		<td class="font-medium">{ r.Nombre }</td>
		<td>{ r.Email }</td>
		<td>{ fmt.Sprintf("%d", r.Cantidad) }</td>
		<td>
			<select
				class="form-select form-select-sm"
				hx-put={ "/admin/reservas/" + r.ID + "/estado" }
				hx-target={ "#reserva-row-" + r.ID }
				hx-swap="outerHTML"
				name="estado"
			>
				<option value="pendiente" selected?={ r.Estado == "pendiente" }>Pendiente</option>
				<option value="confirmada" selected?={ r.Estado == "confirmada" }>Confirmada</option>
				<option value="cancelada" selected?={ r.Estado == "cancelada" }>Cancelada</option>
			</select>
		</td>
		<td class="text-muted text-sm">{ r.CreatedAt }</td>
		<td class="row-actions">
			<button class="btn-icon btn-danger"
				hx-delete={ "/admin/reservas/" + r.ID }
				hx-target={ "#reserva-row-" + r.ID }
				hx-swap="outerHTML"
				hx-confirm="¿Eliminar reserva?"
				title="Eliminar">
				<span class="material-symbols-outlined">delete</span>
			</button>
		</td>
	</tr>
}
```

- [ ] **Step 5: Crear internal/view/pages/public/locales_page.templ**

```templ
package public

import "cms-plazareal/internal/view/layout"

type LocalPublico struct {
	ID          string
	Nombre      string
	Galeria     string
	Numero      string
	Piso        string
	M2          float64
	Descripcion string
	PrecioRef   string
	ImagenURL   string
}

type LocalesPublicData struct {
	Locales []LocalPublico
}

templ LocalesPublicPage(d LocalesPublicData) {
	@layout.Base("Locales Disponibles — Plaza Real", "locales", localesPublicBody(d))
}

templ localesPublicBody(d LocalesPublicData) {
	<section class="page-hero">
		<div class="container">
			<h1 class="hero-title">Locales Disponibles</h1>
			<p class="hero-subtitle">Instala tu negocio en Plaza Real — el centro comercial de Copiapó.</p>
		</div>
	</section>
	<section class="section container">
		if len(d.Locales) == 0 {
			<div class="empty-state glass-card">
				<span class="material-symbols-outlined empty-icon">store_off</span>
				<p>Por el momento no hay locales disponibles. Contáctanos para más información.</p>
			</div>
		} else {
			<div class="locales-grid">
				for _, loc := range d.Locales {
					@LocalCard(loc)
				}
			</div>
		}
		<div class="cta-card glass-card" style="margin-top:3rem; text-align:center;">
			<h2>¿Interesado en arrendar?</h2>
			<p class="text-muted">Contáctanos directamente y te asesoraremos.</p>
			<a href="https://wa.me/56XXXXXXXXX" target="_blank" rel="noopener" class="btn btn-primary">
				<span class="material-symbols-outlined">chat</span>
				Consultar por WhatsApp
			</a>
		</div>
	</section>
}

templ LocalCard(loc LocalPublico) {
	<article class="local-card glass-card">
		if loc.ImagenURL != "" {
			<img src={ loc.ImagenURL } alt={ loc.Nombre } class="local-card-img" loading="lazy"/>
		}
		<div class="local-card-body">
			<div class="local-card-meta text-sm text-muted">
				{ loc.Galeria } · Piso { loc.Piso } · { fmt.Sprintf("%.0f m²", loc.M2) }
			</div>
			<h3 class="local-card-title">{ loc.Nombre }</h3>
			if loc.Descripcion != "" {
				<p class="text-muted local-card-desc">{ loc.Descripcion }</p>
			}
			if loc.PrecioRef != "" {
				<p class="local-card-price">{ loc.PrecioRef }</p>
			}
		</div>
	</article>
}
```

- [ ] **Step 6: Crear internal/view/fragments/locales_cards.templ**

```templ
package fragments

type LocalCard struct {
	ID        string
	Nombre    string
	Galeria   string
	M2        float64
	PrecioRef string
	ImagenURL string
}

templ LocalesCards(items []LocalCard) {
	for _, loc := range items {
		<article class="local-preview-card glass-card">
			if loc.ImagenURL != "" {
				<img src={ loc.ImagenURL } alt={ loc.Nombre } class="local-preview-img" loading="lazy"/>
			}
			<div class="local-preview-body">
				<span class="text-sm text-muted">{ loc.Galeria } · { fmt.Sprintf("%.0f m²", loc.M2) }</span>
				<h4>{ loc.Nombre }</h4>
				if loc.PrecioRef != "" {
					<span class="local-price-badge">{ loc.PrecioRef }</span>
				}
			</div>
		</article>
	}
}
```

- [ ] **Step 7: Agregar handlers en internal/handlers/admin/handlers.go**

Agregar al final:

```go
// LocalesList — GET /admin/locales
func LocalesList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter("locales_disponibles", "estado != 'arrendado'", "galeria,numero", 200, 0)
		rows := make([]adminView.LocalRow, 0, len(records))
		for _, r := range records {
			rows = append(rows, adminView.LocalRow{
				ID:      r.Id,
				Nombre:  r.GetString("nombre"),
				Galeria: r.GetString("galeria"),
				Numero:  r.GetString("numero"),
				M2:      r.GetFloat("m2"),
				Estado:  r.GetString("estado"),
			})
		}
		content := adminView.LocalesPage(adminView.LocalesPageData{Rows: rows})
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, content)
		}
		return helpers.Render(c, layout.Admin("Locales Disponibles", "locales", content))
	}
}

// LocalNew — GET /admin/locales/new
func LocalNew(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, adminView.LocalForm(adminView.LocalFormData{}))
	}
}

// LocalEdit — GET /admin/locales/:id/edit
func LocalEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("locales_disponibles", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("not found")
		}
		return helpers.Render(c, adminView.LocalForm(adminView.LocalFormData{
			ID:            r.Id,
			Nombre:        r.GetString("nombre"),
			Galeria:       r.GetString("galeria"),
			Numero:        r.GetString("numero"),
			Piso:          r.GetString("piso"),
			M2:            fmt.Sprintf("%.1f", r.GetFloat("m2")),
			Descripcion:   r.GetString("descripcion"),
			PrecioRef:     r.GetString("precio_ref"),
			Estado:        r.GetString("estado"),
			ImagenURL:     r.GetString("imagen_url"),
			ContactoEmail: r.GetString("contacto_email"),
			ContactoTel:   r.GetString("contacto_tel"),
		}))
	}
}

// LocalCreate — POST /admin/locales
func LocalCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		col, err := pb.FindCollectionByNameOrId("locales_disponibles")
		if err != nil {
			return c.Status(500).SendString("collection error")
		}
		r := core.NewRecord(col)
		r.Set("nombre", c.FormValue("nombre"))
		r.Set("galeria", c.FormValue("galeria"))
		r.Set("numero", c.FormValue("numero"))
		r.Set("piso", c.FormValue("piso"))
		r.Set("m2", c.FormValue("m2"))
		r.Set("descripcion", c.FormValue("descripcion"))
		r.Set("precio_ref", c.FormValue("precio_ref"))
		r.Set("estado", c.FormValue("estado"))
		r.Set("imagen_url", c.FormValue("imagen_url"))
		r.Set("contacto_email", c.FormValue("contacto_email"))
		r.Set("contacto_tel", c.FormValue("contacto_tel"))
		if err := pb.Save(r); err != nil {
			return helpers.Render(c, adminView.LocalForm(adminView.LocalFormData{ErrorMsg: err.Error()}))
		}
		c.Set("HX-Redirect", "/admin/locales")
		return c.SendStatus(204)
	}
}

// LocalUpdate — PUT /admin/locales/:id
func LocalUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("locales_disponibles", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("not found")
		}
		r.Set("nombre", c.FormValue("nombre"))
		r.Set("galeria", c.FormValue("galeria"))
		r.Set("numero", c.FormValue("numero"))
		r.Set("piso", c.FormValue("piso"))
		r.Set("m2", c.FormValue("m2"))
		r.Set("descripcion", c.FormValue("descripcion"))
		r.Set("precio_ref", c.FormValue("precio_ref"))
		r.Set("estado", c.FormValue("estado"))
		r.Set("imagen_url", c.FormValue("imagen_url"))
		r.Set("contacto_email", c.FormValue("contacto_email"))
		r.Set("contacto_tel", c.FormValue("contacto_tel"))
		if err := pb.Save(r); err != nil {
			return helpers.Render(c, adminView.LocalForm(adminView.LocalFormData{ID: r.Id, ErrorMsg: err.Error()}))
		}
		row := adminView.LocalRow{
			ID:      r.Id,
			Nombre:  r.GetString("nombre"),
			Galeria: r.GetString("galeria"),
			Numero:  r.GetString("numero"),
			M2:      r.GetFloat("m2"),
			Estado:  r.GetString("estado"),
		}
		c.Set("HX-Trigger", "closeModal")
		return helpers.Render(c, adminView.LocalTableRow(row))
	}
}

// LocalDelete — DELETE /admin/locales/:id
func LocalDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("locales_disponibles", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("not found")
		}
		if err := pb.Delete(r); err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.SendStatus(200)
	}
}

// ReservasList — GET /admin/reservas
func ReservasList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter("reservas", "1=1", "-created", 200, 0)
		rows := make([]adminView.ReservaRow, 0, len(records))
		for _, r := range records {
			cbTitle := ""
			if cb, err := pb.FindRecordById("content_blocks", r.GetString("content_block_id")); err == nil {
				cbTitle = cb.GetString("title")
			}
			rows = append(rows, adminView.ReservaRow{
				ID:                r.Id,
				Nombre:            r.GetString("nombre"),
				Email:             r.GetString("email"),
				Telefono:          r.GetString("telefono"),
				Cantidad:          int(r.GetFloat("cantidad")),
				Estado:            r.GetString("estado"),
				ContentBlockTitle: cbTitle,
				CreatedAt:         r.GetString("created"),
			})
		}
		content := adminView.ReservasPage(adminView.ReservasPageData{Rows: rows})
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, content)
		}
		return helpers.Render(c, layout.Admin("Reservas", "reservas", content))
	}
}

// ReservaUpdateEstado — PUT /admin/reservas/:id/estado
func ReservaUpdateEstado(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("reservas", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("not found")
		}
		r.Set("estado", c.FormValue("estado"))
		if err := pb.Save(r); err != nil {
			return c.Status(500).SendString(err.Error())
		}
		cbTitle := ""
		if cb, err := pb.FindRecordById("content_blocks", r.GetString("content_block_id")); err == nil {
			cbTitle = cb.GetString("title")
		}
		row := adminView.ReservaRow{
			ID:                r.Id,
			Nombre:            r.GetString("nombre"),
			Email:             r.GetString("email"),
			Telefono:          r.GetString("telefono"),
			Cantidad:          int(r.GetFloat("cantidad")),
			Estado:            r.GetString("estado"),
			ContentBlockTitle: cbTitle,
			CreatedAt:         r.GetString("created"),
		}
		return helpers.Render(c, adminView.ReservaTableRow(row))
	}
}

// ReservaDelete — DELETE /admin/reservas/:id
func ReservaDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("reservas", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("not found")
		}
		if err := pb.Delete(r); err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.SendStatus(200)
	}
}
```

- [ ] **Step 8: Agregar handler público en internal/handlers/web/handlers.go**

```go
// LocalesPublicPage — GET /locales-disponibles
func LocalesPublicPage(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter(
			"locales_disponibles",
			"estado = 'disponible'",
			"galeria,numero",
			100, 0,
		)
		items := make([]publicView.LocalPublico, 0, len(records))
		for _, r := range records {
			items = append(items, publicView.LocalPublico{
				ID:          r.Id,
				Nombre:      r.GetString("nombre"),
				Galeria:     r.GetString("galeria"),
				Numero:      r.GetString("numero"),
				Piso:        r.GetString("piso"),
				M2:          r.GetFloat("m2"),
				Descripcion: r.GetString("descripcion"),
				PrecioRef:   r.GetString("precio_ref"),
				ImagenURL:   r.GetString("imagen_url"),
			})
		}
		return helpers.Render(c, publicView.LocalesPublicPage(publicView.LocalesPublicData{Locales: items}))
	}
}
```

- [ ] **Step 9: Agregar handler fragmento en internal/handlers/fragments/handlers.go**

```go
// LocalesCards — GET /frag/locales-cards
func LocalesCardsHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter(
			"locales_disponibles",
			"estado = 'disponible'",
			"-created",
			4, 0,
		)
		items := make([]fragmentsView.LocalCard, 0, len(records))
		for _, r := range records {
			items = append(items, fragmentsView.LocalCard{
				ID:        r.Id,
				Nombre:    r.GetString("nombre"),
				Galeria:   r.GetString("galeria"),
				M2:        r.GetFloat("m2"),
				PrecioRef: r.GetString("precio_ref"),
				ImagenURL: r.GetString("imagen_url"),
			})
		}
		return helpers.Render(c, fragmentsView.LocalesCards(items))
	}
}
```

- [ ] **Step 10: Registrar rutas en cmd/server/main.go**

```go
// Admin — locales
adm.Get("/locales", middleware.RoleRequired(...), admin.LocalesList(cfg, pb))
adm.Get("/locales/new", middleware.RoleRequired(...), admin.LocalNew(cfg))
adm.Get("/locales/:id/edit", middleware.RoleRequired(...), admin.LocalEdit(cfg, pb))
adm.Post("/locales", middleware.RoleRequired(...), admin.LocalCreate(cfg, pb))
adm.Put("/locales/:id", middleware.RoleRequired(...), admin.LocalUpdate(cfg, pb))
adm.Delete("/locales/:id", middleware.RoleRequired(...), admin.LocalDelete(cfg, pb))

// Admin — reservas
adm.Get("/reservas", middleware.RoleRequired(...), admin.ReservasList(cfg, pb))
adm.Put("/reservas/:id/estado", middleware.RoleRequired(...), admin.ReservaUpdateEstado(cfg, pb))
adm.Delete("/reservas/:id", middleware.RoleRequired(...), admin.ReservaDelete(cfg, pb))

// Público
app.Get("/locales-disponibles", web.LocalesPublicPage(cfg, pb))

// Fragmento
frag.Get("/locales-cards", fragments.LocalesCardsHandler(cfg, pb))
```

- [ ] **Step 11: Ejecutar templ generate y compilar**

```bash
make generate
go build ./...
```

- [ ] **Step 12: Test manual**

```bash
go run cmd/server/main.go
```

Verificar:
1. `http://localhost:3000/admin/locales` — tabla de locales
2. Crear local → formulario modal, guardar → HX-Redirect a /admin/locales
3. Editar local → actualiza fila en tabla via HTMX
4. `http://localhost:3000/locales-disponibles` — página pública vacía o con locales

- [ ] **Step 13: Commit**

```bash
git add internal/auth/collections.go \
        internal/view/pages/admin/locales_page.templ \
        internal/view/pages/admin/locales_page_templ.go \
        internal/view/pages/admin/local_form.templ \
        internal/view/pages/admin/local_form_templ.go \
        internal/view/pages/admin/reservas_page.templ \
        internal/view/pages/admin/reservas_page_templ.go \
        internal/view/pages/public/locales_page.templ \
        internal/view/pages/public/locales_page_templ.go \
        internal/view/fragments/locales_cards.templ \
        internal/view/fragments/locales_cards_templ.go \
        internal/handlers/admin/handlers.go \
        internal/handlers/web/handlers.go \
        internal/handlers/fragments/handlers.go \
        cmd/server/main.go
git commit -m "feat: add locales_disponibles + reservas collections, admin CRUD, public page"
```

---

## Estado del repositorio al finalizar este task

- Colecciones `locales_disponibles` y `reservas` definidas en PocketBase
- Admin CRUD completo para locales y gestión de estado de reservas
- Página pública `/locales-disponibles` muestra locales con `estado=disponible`
- Fragmento `/frag/locales-cards` para homepage
- `go build ./...` pasa sin errores
