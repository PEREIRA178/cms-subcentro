package admin

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"jcp-gestioninmobiliaria/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ── CAROUSEL SLIDES CRUD ──

func CarouselList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Query("fragment") != "rows" {
			if c.Get("HX-Request") != "true" {
				return c.SendFile("./internal/templates/admin/pages/dashboard.html")
			}
			return c.SendFile("./internal/templates/admin/pages/carousel.html")
		}
		records, err := pb.FindRecordsByFilter("carousel_slides", "", "orden", 100, 0)
		var sb strings.Builder
		if err != nil {
			sb.WriteString(fmt.Sprintf(`<tr><td colspan="6" style="text-align:center;padding:32px;color:#B71C1C">Error: %s</td></tr>`, template.HTMLEscapeString(err.Error())))
		} else if len(records) == 0 {
			sb.WriteString(`<tr class="empty-row"><td colspan="6" style="text-align:center;padding:32px;color:var(--md-outline)">Sin banners — agrega uno con el botón de arriba</td></tr>`)
		} else {
			for _, r := range records {
				activo := r.GetBool("activo")
				badgeClass := "badge-warning"
				stLabel := "Inactivo"
				if activo {
					badgeClass = "badge-success"
					stLabel = "Activo"
				}
				thumb := r.GetString("image_url")
				thumbImg := ""
				if thumb != "" {
					thumbImg = fmt.Sprintf(`<img src="%s" alt="" style="width:80px;height:44px;object-fit:cover;border-radius:6px;border:1px solid var(--md-outline-variant)"/>`, template.HTMLEscapeString(thumb))
				}
				sb.WriteString(fmt.Sprintf(`<tr>
        <td>%d</td>
        <td>%s</td>
        <td>%s</td>
        <td style="max-width:280px"><div style="font-weight:500">%s</div><div style="font-size:12px;color:var(--md-outline);overflow:hidden;text-overflow:ellipsis;white-space:nowrap">%s</div></td>
        <td><span class="badge %s">%s</span></td>
        <td style="white-space:nowrap">
          <button class="topbar-btn topbar-btn-outline" style="padding:4px 10px;font-size:12px"
            hx-get="/admin/carousel/%s/edit" hx-target="#modal-container" hx-swap="innerHTML">Editar</button>
          <button class="topbar-btn" style="padding:4px 10px;font-size:12px;background:#FDECEA;color:#B71C1C;border:none;cursor:pointer"
            hx-delete="/admin/carousel/%s" hx-confirm="¿Eliminar este banner?" hx-target="closest tr" hx-swap="outerHTML swap:300ms">Eliminar</button>
        </td></tr>`,
					r.GetInt("orden"),
					thumbImg,
					template.HTMLEscapeString(r.GetString("title")),
					template.HTMLEscapeString(r.GetString("subtitle")),
					template.HTMLEscapeString(r.GetString("link_url")),
					badgeClass, stLabel,
					r.Id, r.Id,
				))
			}
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

func CarouselForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		html := carouselFormHTML("", "", "", "", "", 1, true)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func CarouselEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("carousel_slides", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		html := carouselFormHTML(
			r.Id,
			r.GetString("title"),
			r.GetString("subtitle"),
			r.GetString("image_url"),
			r.GetString("link_url"),
			r.GetInt("orden"),
			r.GetBool("activo"),
		)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

func CarouselCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		img := strings.TrimSpace(c.FormValue("image_url"))
		if img == "" {
			return c.SendString(`<div class="toast toast-error">La URL de imagen es requerida</div>`)
		}
		col, err := pb.FindCollectionByNameOrId("carousel_slides")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		r := core.NewRecord(col)
		setCarouselFields(r, c)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error: ` + template.HTMLEscapeString(err.Error()) + `</div>`)
		}
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Banner creado
<script>
  document.getElementById('modal-container').innerHTML='';
  htmx.ajax('GET','/admin/carousel?fragment=rows',{target:'#carousel-tbody',swap:'innerHTML'});
  setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000);
</script></div>`)
	}
}

func CarouselUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("carousel_slides", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		setCarouselFields(r, c)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error: ` + template.HTMLEscapeString(err.Error()) + `</div>`)
		}
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Banner actualizado
<script>
  document.getElementById('modal-container').innerHTML='';
  htmx.ajax('GET','/admin/carousel?fragment=rows',{target:'#carousel-tbody',swap:'innerHTML'});
  setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000);
</script></div>`)
	}
}

func CarouselDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("carousel_slides", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		return c.SendString("")
	}
}

func setCarouselFields(r *core.Record, c *fiber.Ctx) {
	r.Set("title", strings.TrimSpace(c.FormValue("title")))
	r.Set("subtitle", strings.TrimSpace(c.FormValue("subtitle")))
	r.Set("image_url", strings.TrimSpace(c.FormValue("image_url")))
	r.Set("link_url", strings.TrimSpace(c.FormValue("link_url")))
	orden, _ := strconv.Atoi(strings.TrimSpace(c.FormValue("orden")))
	if orden < 1 {
		orden = 1
	}
	r.Set("orden", orden)
	r.Set("activo", c.FormValue("activo") == "on")
}

func carouselFormHTML(id, title, subtitle, imageURL, linkURL string, orden int, activo bool) string {
	method := `hx-post="/admin/carousel"`
	if id != "" {
		method = fmt.Sprintf(`hx-put="/admin/carousel/%s"`, id)
	}
	formTitle := map[bool]string{true: "Editar Banner", false: "Nuevo Banner"}[id != ""]
	chk := ""
	if activo {
		chk = " checked"
	}
	return fmt.Sprintf(`<div class="modal-overlay" onclick="this.remove()">
  <div class="modal-card" onclick="event.stopPropagation()" style="max-width:600px">
    <div class="modal-header">
      <h3>%s</h3>
      <button onclick="document.getElementById('modal-container').innerHTML=''" style="background:none;border:none;cursor:pointer;font-size:20px;color:var(--md-outline)">✕</button>
    </div>
    <form %s hx-target="#toast-area" hx-swap="innerHTML">
      <div class="form-field">
        <label>URL de imagen *</label>
        <input type="url" name="image_url" value="%s" required class="form-input" placeholder="https://..."/>
      </div>
      <div class="form-field">
        <label>Título</label>
        <input type="text" name="title" value="%s" class="form-input" placeholder="Bienvenido a Plaza Real"/>
      </div>
      <div class="form-field">
        <label>Subtítulo</label>
        <input type="text" name="subtitle" value="%s" class="form-input" placeholder="Más de 100 tiendas..."/>
      </div>
      <div class="form-field">
        <label>Hipervínculo (al hacer clic)</label>
        <input type="text" name="link_url" value="%s" class="form-input" placeholder="/buscador-tiendas.html o https://..."/>
      </div>
      <div class="form-row">
        <div class="form-field">
          <label>Orden</label>
          <input type="number" name="orden" value="%d" min="1" class="form-input"/>
        </div>
        <div class="form-field">
          <label>Estado</label>
          <label style="display:flex;align-items:center;gap:8px;margin-top:8px;font-size:14px;color:var(--md-on-surface);text-transform:none;letter-spacing:0;font-weight:400">
            <input type="checkbox" name="activo"%s style="width:18px;height:18px"/> Activo (visible en el sitio)
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
		template.HTMLEscapeString(imageURL),
		template.HTMLEscapeString(title),
		template.HTMLEscapeString(subtitle),
		template.HTMLEscapeString(linkURL),
		orden, chk,
	)
}
