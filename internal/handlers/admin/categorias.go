package admin

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"jcp-gestioninmobiliaria/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ── CATEGORIAS CRUD ──

var catSlugRe = regexp.MustCompile(`[^a-z0-9\-]`)

func slugifyCat(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = catSlugRe.ReplaceAllString(s, "")
	if s == "" {
		s = "categoria"
	}
	return s
}

func CategoriasList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Query("fragment") != "rows" {
			if c.Get("HX-Request") != "true" {
				return c.SendFile("./internal/templates/admin/pages/dashboard.html")
			}
			return c.SendFile("./internal/templates/admin/pages/categorias.html")
		}
		records, err := pb.FindRecordsByFilter("categorias", "", "orden", 200, 0)
		var sb strings.Builder
		if err != nil {
			sb.WriteString(fmt.Sprintf(`<tr><td colspan="6" style="text-align:center;padding:32px;color:#B71C1C">Error: %s</td></tr>`, template.HTMLEscapeString(err.Error())))
		} else if len(records) == 0 {
			sb.WriteString(`<tr class="empty-row"><td colspan="6" style="text-align:center;padding:32px;color:var(--md-outline)">Sin categorías — agrega una con el botón de arriba</td></tr>`)
		} else {
			for _, r := range records {
				activo := r.GetBool("activo")
				badgeClass := "badge-warning"
				stLabel := "Inactiva"
				if activo {
					badgeClass = "badge-success"
					stLabel = "Activa"
				}
				sb.WriteString(fmt.Sprintf(`<tr>
        <td>%d</td>
        <td style="font-size:22px">%s</td>
        <td><strong>%s</strong></td>
        <td><code style="background:var(--md-surface-container-low);padding:2px 6px;border-radius:4px;font-size:12px">%s</code></td>
        <td><span class="badge %s">%s</span></td>
        <td style="white-space:nowrap">
          <button class="topbar-btn topbar-btn-outline" style="padding:4px 10px;font-size:12px"
            hx-get="/admin/categorias/%s/edit" hx-target="#modal-container" hx-swap="innerHTML">Editar</button>
          <button class="topbar-btn" style="padding:4px 10px;font-size:12px;background:#FDECEA;color:#B71C1C;border:none;cursor:pointer"
            hx-delete="/admin/categorias/%s" hx-confirm="¿Eliminar esta categoría?" hx-target="closest tr" hx-swap="outerHTML swap:300ms">Eliminar</button>
        </td></tr>`,
					r.GetInt("orden"),
					template.HTMLEscapeString(r.GetString("icono")),
					template.HTMLEscapeString(r.GetString("nombre")),
					template.HTMLEscapeString(r.GetString("slug")),
					badgeClass, stLabel,
					r.Id, r.Id,
				))
			}
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

func CategoriaForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		html := categoriaFormHTML("", "", "", "", 1, true)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func CategoriaEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("categorias", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrada</div>`)
		}
		html := categoriaFormHTML(
			r.Id,
			r.GetString("slug"),
			r.GetString("nombre"),
			r.GetString("icono"),
			r.GetInt("orden"),
			r.GetBool("activo"),
		)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func CategoriaCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		nombre := strings.TrimSpace(c.FormValue("nombre"))
		if nombre == "" {
			return c.SendString(`<div class="toast toast-error">El nombre es requerido</div>`)
		}
		col, err := pb.FindCollectionByNameOrId("categorias")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		r := core.NewRecord(col)
		setCategoriaFields(r, c)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error: ` + template.HTMLEscapeString(err.Error()) + `</div>`)
		}
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Categoría creada
<script>
  document.getElementById('modal-container').innerHTML='';
  htmx.ajax('GET','/admin/categorias?fragment=rows',{target:'#categorias-tbody',swap:'innerHTML'});
  setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000);
</script></div>`)
	}
}

func CategoriaUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("categorias", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrada</div>`)
		}
		setCategoriaFields(r, c)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error: ` + template.HTMLEscapeString(err.Error()) + `</div>`)
		}
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Categoría actualizada
<script>
  document.getElementById('modal-container').innerHTML='';
  htmx.ajax('GET','/admin/categorias?fragment=rows',{target:'#categorias-tbody',swap:'innerHTML'});
  setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000);
</script></div>`)
	}
}

func CategoriaDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("categorias", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		return c.SendString("")
	}
}

func setCategoriaFields(r *core.Record, c *fiber.Ctx) {
	nombre := strings.TrimSpace(c.FormValue("nombre"))
	slug := strings.TrimSpace(c.FormValue("slug"))
	if slug == "" {
		slug = slugifyCat(nombre)
	} else {
		slug = slugifyCat(slug)
	}
	r.Set("nombre", nombre)
	r.Set("slug", slug)
	r.Set("icono", strings.TrimSpace(c.FormValue("icono")))
	orden, _ := strconv.Atoi(strings.TrimSpace(c.FormValue("orden")))
	if orden < 1 {
		orden = 1
	}
	r.Set("orden", orden)
	r.Set("activo", c.FormValue("activo") == "on")
}

func categoriaFormHTML(id, slug, nombre, icono string, orden int, activo bool) string {
	method := `hx-post="/admin/categorias"`
	if id != "" {
		method = fmt.Sprintf(`hx-put="/admin/categorias/%s"`, id)
	}
	formTitle := map[bool]string{true: "Editar Categoría", false: "Nueva Categoría"}[id != ""]
	chk := ""
	if activo {
		chk = " checked"
	}
	return fmt.Sprintf(`<div class="modal-overlay" onclick="this.remove()">
  <div class="modal-card" onclick="event.stopPropagation()" style="max-width:520px">
    <div class="modal-header">
      <h3>%s</h3>
      <button onclick="document.getElementById('modal-container').innerHTML=''" style="background:none;border:none;cursor:pointer;font-size:20px;color:var(--md-outline)">✕</button>
    </div>
    <form %s hx-target="#toast-area" hx-swap="innerHTML">
      <div class="form-row">
        <div class="form-field" style="grid-column:span 2">
          <label>Nombre *</label>
          <input type="text" name="nombre" value="%s" required class="form-input" placeholder="Tiendas & Moda"/>
        </div>
      </div>
      <div class="form-row">
        <div class="form-field">
          <label>Slug (identificador)</label>
          <input type="text" name="slug" value="%s" class="form-input" placeholder="tiendas"/>
        </div>
        <div class="form-field">
          <label>Ícono (emoji)</label>
          <input type="text" name="icono" value="%s" class="form-input" placeholder="🛍️"/>
        </div>
      </div>
      <div class="form-row">
        <div class="form-field">
          <label>Orden</label>
          <input type="number" name="orden" value="%d" min="1" class="form-input"/>
        </div>
        <div class="form-field">
          <label>Estado</label>
          <label style="display:flex;align-items:center;gap:8px;margin-top:8px;font-size:14px;color:var(--md-on-surface);text-transform:none;letter-spacing:0;font-weight:400">
            <input type="checkbox" name="activo"%s style="width:18px;height:18px"/> Activa
          </label>
        </div>
      </div>
      <div class="modal-actions">
        <button type="button" class="topbar-btn topbar-btn-outline" onclick="document.getElementById('modal-container').innerHTML=''">Cancelar</button>
        <button type="submit" class="topbar-btn topbar-btn-primary">Guardar</button>
      </div>
    </form>
  </div>
</div>`,
		formTitle, method,
		template.HTMLEscapeString(nombre),
		template.HTMLEscapeString(slug),
		template.HTMLEscapeString(icono),
		orden, chk,
	)
}
