# Task 06 — Content Blocks: NOTICIA / COMUNICADO / PROMOCION (admin + público)

**Depends on:** Task 05 (admin shell), Task 02 (design system)
**Estimated complexity:** media — 3 categorías, un formulario compartido, fragmentos públicos

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Layout admin ya existe: internal/view/layout/admin.templ
Render helper ya existe: internal/helpers/render.go
content_blocks collection: categorías ya corregidas a NOTICIA/COMUNICADO/PROMOCION (Task 01)
Fragmentos legacy: internal/handlers/fragments/*.go usan strings.Builder → a migrar a templ
```

---

## Objetivo

Crear las páginas y fragmentos templ para content_blocks. Admin CRUD completo (lista + form). Fragmentos públicos para homepage y páginas de archivo. Handlers HTMX-aware.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Crear | `internal/view/pages/admin/content_page.templ` |
| Crear | `internal/view/pages/admin/content_form.templ` |
| Crear | `internal/view/fragments/noticias_cards.templ` |
| Crear | `internal/view/fragments/comunicados_cards.templ` |
| Crear | `internal/view/fragments/promos_cards.templ` |
| Modificar | `internal/handlers/admin/handlers.go` — agregar handlers content_blocks |
| Modificar | `internal/handlers/fragments/handlers.go` — migrar a templ |
| Modificar | `cmd/server/main.go` — registrar rutas admin y fragmentos |

---

## Implementación

- [ ] **Step 1: Revisar colección content_blocks en PocketBase**

```bash
grep -n "content_blocks\|ContentBlock\|content_block" internal/auth/collections.go | head -30
```

Identificar los campos disponibles. Los campos esperados son:
- `title` string
- `description` string (resumen)
- `body` string (HTML completo, opcional)
- `category` select: NOTICIA | COMUNICADO | PROMOCION
- `image_url` string (URL de imagen, cargada vía R2)
- `status` string: "borrador" | "publicado"
- `featured` bool
- `published_at` datetime (para PROMOCION: fecha de inicio)
- `expires_at` datetime (para PROMOCION: fecha de fin, opcional)

Si algún campo no existe, agregar en `collections.go` y verificar compilación.

- [ ] **Step 2: Crear internal/view/pages/admin/content_page.templ**

```templ
package admin

import "cms-plazareal/internal/view/layout"

type ContentRow struct {
	ID        string
	Title     string
	Category  string
	Status    string
	Featured  bool
	PublishedAt string
}

type ContentPageData struct {
	Category string // "NOTICIA" | "COMUNICADO" | "PROMOCION"
	Title    string // "Noticias" | "Comunicados" | "Promociones"
	Rows     []ContentRow
}

templ ContentPage(d ContentPageData) {
	@layout.Admin(d.Title, strings.ToLower(d.Category), contentPageBody(d))
}

templ contentPageBody(d ContentPageData) {
	<div class="admin-page">
		<div class="page-header">
			<h1 class="page-title">{ d.Title }</h1>
			<button
				class="btn btn-primary"
				hx-get={ "/admin/content/new?cat=" + d.Category }
				hx-target="#modal-container"
				hx-swap="innerHTML"
			>
				<span class="material-symbols-outlined">add</span>
				Nuevo
			</button>
		</div>
		<div class="table-card">
			<table class="admin-table">
				<thead>
					<tr>
						<th>Título</th>
						<th>Estado</th>
						<th>Destacado</th>
						<th>Publicado</th>
						<th></th>
					</tr>
				</thead>
				<tbody id="content-tbody">
					for _, row := range d.Rows {
						@ContentTableRow(row)
					}
				</tbody>
			</table>
		</div>
		<div id="modal-container"></div>
	</div>
}

templ ContentTableRow(r ContentRow) {
	<tr id={ "content-row-" + r.ID }>
		<td class="font-medium">{ r.Title }</td>
		<td>
			<span class={ "badge", templ.KV("badge-success", r.Status == "publicado"), templ.KV("badge-muted", r.Status == "borrador") }>
				{ r.Status }
			</span>
		</td>
		<td>
			if r.Featured {
				<span class="material-symbols-outlined text-yellow" style="font-size:18px">star</span>
			}
		</td>
		<td class="text-muted text-sm">{ r.PublishedAt }</td>
		<td class="row-actions">
			<button
				class="btn-icon"
				hx-get={ "/admin/content/" + r.ID + "/edit" }
				hx-target="#modal-container"
				hx-swap="innerHTML"
				title="Editar"
			>
				<span class="material-symbols-outlined">edit</span>
			</button>
			<button
				class="btn-icon btn-danger"
				hx-delete={ "/admin/content/" + r.ID }
				hx-target={ "#content-row-" + r.ID }
				hx-swap="outerHTML"
				hx-confirm="¿Eliminar este contenido?"
				title="Eliminar"
			>
				<span class="material-symbols-outlined">delete</span>
			</button>
		</td>
	</tr>
}
```

**Nota:** `strings.ToLower` requiere importar `"strings"` en el paquete — en templ se importa normalmente con `import "strings"` dentro del bloque de paquete.

- [ ] **Step 3: Crear internal/view/pages/admin/content_form.templ**

```templ
package admin

import "cms-plazareal/internal/view/components"

type ContentFormData struct {
	ID          string
	Category    string // "NOTICIA" | "COMUNICADO" | "PROMOCION"
	Title       string
	Description string
	Body        string
	ImageURL    string
	Status      string
	Featured    bool
	PublishedAt string
	ExpiresAt   string
	ErrorMsg    string
}

templ ContentForm(d ContentFormData) {
	<div class="modal-overlay" hx-get="/admin/empty" hx-trigger="click[target===this]" hx-target="#modal-container" hx-swap="innerHTML">
		<div class="modal-card" @click.stop="">
			<div class="modal-header">
				<h2 class="modal-title">
					if d.ID == "" {
						Nuevo { d.Category }
					} else {
						Editar { d.Category }
					}
				</h2>
				<button
					class="btn-icon modal-close"
					hx-get="/admin/empty"
					hx-target="#modal-container"
					hx-swap="innerHTML"
				>
					<span class="material-symbols-outlined">close</span>
				</button>
			</div>
			if d.ErrorMsg != "" {
				<div class="alert alert-danger">{ d.ErrorMsg }</div>
			}
			<form
				if d.ID == "" {
					hx-post="/admin/content"
				} else {
					hx-put={ "/admin/content/" + d.ID }
				}
				hx-target={ "#content-row-" + d.ID }
				hx-swap="outerHTML"
				enctype="multipart/form-data"
				class="modal-form"
			>
				<input type="hidden" name="category" value={ d.Category }/>
				<div class="form-group">
					<label class="form-label">Título *</label>
					<input type="text" name="title" class="form-input" value={ d.Title } required/>
				</div>
				<div class="form-group">
					<label class="form-label">Descripción / Resumen</label>
					<textarea name="description" class="form-textarea" rows="3">{ d.Description }</textarea>
				</div>
				<div class="form-group">
					<label class="form-label">Contenido completo (HTML)</label>
					<textarea name="body" class="form-textarea" rows="6">{ d.Body }</textarea>
				</div>
				@components.UploadField("image_url", d.ImageURL, "Imagen")
				<div class="form-row">
					<div class="form-group">
						<label class="form-label">Estado</label>
						<select name="status" class="form-select">
							<option value="borrador" selected?={ d.Status == "borrador" || d.Status == "" }>Borrador</option>
							<option value="publicado" selected?={ d.Status == "publicado" }>Publicado</option>
						</select>
					</div>
					<div class="form-group">
						<label class="form-label">
							<input type="checkbox" name="featured" value="true" checked?={ d.Featured }/>
							Destacado
						</label>
					</div>
				</div>
				<div class="form-row">
					<div class="form-group">
						<label class="form-label">Publicado desde</label>
						<input type="datetime-local" name="published_at" class="form-input" value={ d.PublishedAt }/>
					</div>
					if d.Category == "PROMOCION" {
						<div class="form-group">
							<label class="form-label">Expira</label>
							<input type="datetime-local" name="expires_at" class="form-input" value={ d.ExpiresAt }/>
						</div>
					}
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

**Nota HTMX sobre respuesta de create:** Cuando se crea un nuevo registro (POST), el servidor debe devolver el `ContentTableRow` del nuevo registro con `HX-Retarget: #content-tbody` y `HX-Reswap: afterbegin` para insertar la fila al inicio. También cerrar el modal con `HX-Trigger: closeModal`. Agregar event listener `htmx:afterRequest` para limpiar modal cuando `closeModal` dispara — o usar OOB swap para limpiar `#modal-container`.

La forma más simple: el handler POST devuelve `HX-Redirect: /admin/noticias` para recargar la página completa. Simple y confiable.

- [ ] **Step 4: Crear fragmentos públicos**

**`internal/view/fragments/noticias_cards.templ`:**

```templ
package fragments

type NoticiaCard struct {
	ID          string
	Title       string
	Description string
	ImageURL    string
	PublishedAt string
	Slug        string
}

templ NoticiasCards(items []NoticiaCard) {
	for _, n := range items {
		<article class="content-card glass-card" data-id={ n.ID }>
			if n.ImageURL != "" {
				<img src={ n.ImageURL } alt={ n.Title } class="content-card-img" loading="lazy"/>
			}
			<div class="content-card-body">
				<time class="content-card-date text-muted text-sm">{ n.PublishedAt }</time>
				<h3 class="content-card-title">{ n.Title }</h3>
				<p class="content-card-desc text-muted">{ n.Description }</p>
				<a href={ templ.URL("/noticias/" + n.ID) } class="btn btn-ghost btn-sm">Leer más</a>
			</div>
		</article>
	}
}
```

**`internal/view/fragments/comunicados_cards.templ`:**

```templ
package fragments

type ComunicadoCard struct {
	ID          string
	Title       string
	Description string
	PublishedAt string
}

templ ComunicadosCards(items []ComunicadoCard) {
	for _, c := range items {
		<article class="comunicado-card glass-card" data-id={ c.ID }>
			<div class="comunicado-card-body">
				<time class="text-muted text-sm">{ c.PublishedAt }</time>
				<h3 class="comunicado-title">{ c.Title }</h3>
				<p class="text-muted">{ c.Description }</p>
			</div>
		</article>
	}
}
```

**`internal/view/fragments/promos_cards.templ`:**

```templ
package fragments

type PromoCard struct {
	ID          string
	Title       string
	Description string
	ImageURL    string
	PublishedAt string
	ExpiresAt   string
	Featured    bool
}

templ PromosCards(items []PromoCard) {
	for _, p := range items {
		<article class={ "promo-card glass-card", templ.KV("promo-featured", p.Featured) } data-id={ p.ID }>
			if p.ImageURL != "" {
				<img src={ p.ImageURL } alt={ p.Title } class="promo-img" loading="lazy"/>
			}
			<div class="promo-body">
				<h3 class="promo-title">{ p.Title }</h3>
				<p class="text-muted">{ p.Description }</p>
				if p.ExpiresAt != "" {
					<p class="promo-expires text-sm">
						<span class="material-symbols-outlined" style="font-size:14px">schedule</span>
						Hasta { p.ExpiresAt }
					</p>
				}
			</div>
		</article>
	}
}
```

- [ ] **Step 5: Agregar handlers admin en internal/handlers/admin/handlers.go**

Agregar al final del archivo:

```go
// ContentList retorna la lista de content_blocks filtrada por categoría.
// GET /admin/noticias | /admin/comunicados | /admin/promociones
func ContentList(cfg *config.Config, pb *pocketbase.PocketBase, category string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, err := pb.FindRecordsByFilter(
			"content_blocks",
			"category = '"+category+"'",
			"-created",
			100, 0,
		)

		title := map[string]string{
			"NOTICIA":    "Noticias",
			"COMUNICADO": "Comunicados",
			"PROMOCION":  "Promociones",
		}[category]

		rows := make([]adminView.ContentRow, 0, len(records))
		if err == nil {
			for _, r := range records {
				rows = append(rows, adminView.ContentRow{
					ID:          r.Id,
					Title:       r.GetString("title"),
					Category:    r.GetString("category"),
					Status:      r.GetString("status"),
					Featured:    r.GetBool("featured"),
					PublishedAt: r.GetString("published_at"),
				})
			}
		}

		content := adminView.ContentPage(adminView.ContentPageData{
			Category: category,
			Title:    title,
			Rows:     rows,
		})

		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, content)
		}
		return helpers.Render(c, layout.Admin(title, strings.ToLower(category), content))
	}
}

// ContentNew sirve el formulario vacío en el modal.
// GET /admin/content/new?cat=NOTICIA
func ContentNew(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cat := c.Query("cat", "NOTICIA")
		return helpers.Render(c, adminView.ContentForm(adminView.ContentFormData{Category: cat}))
	}
}

// ContentEdit sirve el formulario precargado con datos del registro.
// GET /admin/content/:id/edit
func ContentEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("content_blocks", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("not found")
		}
		return helpers.Render(c, adminView.ContentForm(adminView.ContentFormData{
			ID:          r.Id,
			Category:    r.GetString("category"),
			Title:       r.GetString("title"),
			Description: r.GetString("description"),
			Body:        r.GetString("body"),
			ImageURL:    r.GetString("image_url"),
			Status:      r.GetString("status"),
			Featured:    r.GetBool("featured"),
			PublishedAt: r.GetString("published_at"),
			ExpiresAt:   r.GetString("expires_at"),
		}))
	}
}

// ContentCreate procesa el formulario de creación.
// POST /admin/content
func ContentCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		category := c.FormValue("category")
		collection, err := pb.FindCollectionByNameOrId("content_blocks")
		if err != nil {
			return c.Status(500).SendString("collection error")
		}
		record := core.NewRecord(collection)
		record.Set("title", c.FormValue("title"))
		record.Set("description", c.FormValue("description"))
		record.Set("body", c.FormValue("body"))
		record.Set("image_url", c.FormValue("image_url"))
		record.Set("category", category)
		record.Set("status", c.FormValue("status"))
		record.Set("featured", c.FormValue("featured") == "true")
		record.Set("published_at", c.FormValue("published_at"))
		record.Set("expires_at", c.FormValue("expires_at"))
		if err := pb.Save(record); err != nil {
			return helpers.Render(c, adminView.ContentForm(adminView.ContentFormData{
				Category: category, ErrorMsg: err.Error(),
			}))
		}
		redirectPath := map[string]string{
			"NOTICIA":    "/admin/noticias",
			"COMUNICADO": "/admin/comunicados",
			"PROMOCION":  "/admin/promociones",
		}[category]
		c.Set("HX-Redirect", redirectPath)
		return c.SendStatus(204)
	}
}

// ContentUpdate procesa la edición.
// PUT /admin/content/:id
func ContentUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("content_blocks", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString("not found")
		}
		r.Set("title", c.FormValue("title"))
		r.Set("description", c.FormValue("description"))
		r.Set("body", c.FormValue("body"))
		r.Set("image_url", c.FormValue("image_url"))
		r.Set("status", c.FormValue("status"))
		r.Set("featured", c.FormValue("featured") == "true")
		r.Set("published_at", c.FormValue("published_at"))
		r.Set("expires_at", c.FormValue("expires_at"))
		if err := pb.Save(r); err != nil {
			return helpers.Render(c, adminView.ContentForm(adminView.ContentFormData{
				ID: r.Id, ErrorMsg: err.Error(),
			}))
		}
		row := adminView.ContentRow{
			ID:          r.Id,
			Title:       r.GetString("title"),
			Category:    r.GetString("category"),
			Status:      r.GetString("status"),
			Featured:    r.GetBool("featured"),
			PublishedAt: r.GetString("published_at"),
		}
		c.Set("HX-Trigger", "closeModal")
		return helpers.Render(c, adminView.ContentTableRow(row))
	}
}

// ContentDelete elimina el registro.
// DELETE /admin/content/:id
func ContentDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("content_blocks", c.Params("id"))
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

**Verificar imports necesarios en handlers.go:**
```bash
grep -n '"strings"\|"cms-plazareal/internal/view/layout"\|adminView\|fragmentsView' internal/handlers/admin/handlers.go | head -10
```

Agregar imports que falten:
```go
import (
    "strings"
    "cms-plazareal/internal/view/layout"
    adminView "cms-plazareal/internal/view/pages/admin"
    "pocketbase.io/pocketbase/core"
)
```

- [ ] **Step 6: Migrar fragment handlers a templ**

Revisar handlers de fragmentos existentes:
```bash
grep -n "func \|noticias\|comunicados\|promos\|content_blocks" internal/handlers/fragments/handlers.go | head -30
```

Reemplazar la lógica de `strings.Builder` por llamadas a los nuevos componentes templ. Estructura de cada handler de fragmento:

```go
// En internal/handlers/fragments/handlers.go:

import (
    "cms-plazareal/internal/helpers"
    fragmentsView "cms-plazareal/internal/view/fragments"
)

func NoticiasCards(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
    return func(c *fiber.Ctx) error {
        limit := 6
        records, _ := pb.FindRecordsByFilter(
            "content_blocks",
            "category = 'NOTICIA' && status = 'publicado'",
            "-published_at",
            limit, 0,
        )
        items := make([]fragmentsView.NoticiaCard, 0, len(records))
        for _, r := range records {
            items = append(items, fragmentsView.NoticiaCard{
                ID:          r.Id,
                Title:       r.GetString("title"),
                Description: r.GetString("description"),
                ImageURL:    r.GetString("image_url"),
                PublishedAt: r.GetString("published_at"),
            })
        }
        return helpers.Render(c, fragmentsView.NoticiasCards(items))
    }
}

// Similar para ComunicadosCards y PromosCards
```

- [ ] **Step 7: Registrar rutas en cmd/server/main.go**

Buscar el grupo admin:
```bash
grep -n "adm\.\|admin\." cmd/server/main.go | head -20
```

Agregar dentro del grupo `adm`:
```go
// Content blocks
adm.Get("/noticias", middleware.RoleRequired("superadmin","director","admin","editor"), admin.ContentList(cfg, pb, "NOTICIA"))
adm.Get("/comunicados", middleware.RoleRequired("superadmin","director","admin","editor"), admin.ContentList(cfg, pb, "COMUNICADO"))
adm.Get("/promociones", middleware.RoleRequired("superadmin","director","admin","editor"), admin.ContentList(cfg, pb, "PROMOCION"))
adm.Get("/content/new", middleware.RoleRequired("superadmin","director","admin","editor"), admin.ContentNew(cfg))
adm.Get("/content/:id/edit", middleware.RoleRequired("superadmin","director","admin","editor"), admin.ContentEdit(cfg, pb))
adm.Post("/content", middleware.RoleRequired("superadmin","director","admin","editor"), admin.ContentCreate(cfg, pb))
adm.Put("/content/:id", middleware.RoleRequired("superadmin","director","admin","editor"), admin.ContentUpdate(cfg, pb))
adm.Delete("/content/:id", middleware.RoleRequired("superadmin","director","admin","editor"), admin.ContentDelete(cfg, pb))
```

Buscar las rutas de fragmentos existentes:
```bash
grep -n "noticias\|comunicados\|promos" cmd/server/main.go | head -10
```

Actualizar/agregar:
```go
frag.Get("/noticias-cards", fragments.NoticiasCards(cfg, pb))
frag.Get("/comunicados-cards", fragments.ComunicadosCards(cfg, pb))
frag.Get("/promos-cards", fragments.PromosCards(cfg, pb))
```

- [ ] **Step 8: Ejecutar templ generate y compilar**

```bash
make generate
go build ./...
```

Si hay errores de `undefined`:
- `adminView.ContentPage` → verificar que `content_page_templ.go` se generó en `internal/view/pages/admin/`
- `fragmentsView.NoticiasCards` → verificar `noticias_cards_templ.go` en `internal/view/fragments/`

- [ ] **Step 9: Test manual**

```bash
go run cmd/server/main.go
```

Abrir en browser:
1. `http://localhost:3000/admin/noticias` — debe mostrar tabla vacía
2. Click "Nuevo" — debe abrir modal con formulario
3. Rellenar título + descripción, click Guardar → debe redirigir a `/admin/noticias` con la fila nueva
4. Click editar → modal con datos, editar y guardar → fila actualizada sin reload completo
5. Click eliminar → fila desaparece sin reload

Fragmentos públicos (si las rutas públicas ya existen):
```bash
curl http://localhost:3000/frag/noticias-cards
# Espera: HTML con article.content-card (vacío si no hay datos)
```

- [ ] **Step 10: Commit**

```bash
git add internal/view/pages/admin/content_page.templ \
        internal/view/pages/admin/content_page_templ.go \
        internal/view/pages/admin/content_form.templ \
        internal/view/pages/admin/content_form_templ.go \
        internal/view/fragments/noticias_cards.templ \
        internal/view/fragments/noticias_cards_templ.go \
        internal/view/fragments/comunicados_cards.templ \
        internal/view/fragments/comunicados_cards_templ.go \
        internal/view/fragments/promos_cards.templ \
        internal/view/fragments/promos_cards_templ.go \
        internal/handlers/admin/handlers.go \
        internal/handlers/fragments/handlers.go \
        cmd/server/main.go
git commit -m "feat: add content_blocks admin CRUD + public fragments (NOTICIA/COMUNICADO/PROMOCION)"
```

---

## Estado del repositorio al finalizar este task

- Admin CRUD completo para NOTICIA, COMUNICADO, PROMOCION en templ
- Fragmentos públicos `noticias-cards`, `comunicados-cards`, `promos-cards` migrados a templ
- `go build ./...` pasa sin errores
- Rutas registradas: `/admin/noticias`, `/admin/comunicados`, `/admin/promociones`
