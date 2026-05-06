package admin

import (
	"cms-plazareal/internal/auth"
	"cms-plazareal/internal/config"
	"cms-plazareal/internal/helpers"
	"cms-plazareal/internal/services/r2"
	"cms-plazareal/internal/view/layout"
	adminView "cms-plazareal/internal/view/pages/admin"
	"context"
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ── LOGIN ──

func LoginPage(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, adminView.Login(""))
	}
}

func LoginSubmit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		email := strings.TrimSpace(c.FormValue("email"))
		password := c.FormValue("password")
		remember := c.FormValue("remember") == "on"

		if email == "" || password == "" {
			return c.Status(fiber.StatusBadRequest).SendString(
				`<div class="toast toast-error">Email y contraseña requeridos</div>`,
			)
		}

		record, err := pb.FindAuthRecordByEmail("users", email)
		if err != nil || record == nil {
			return c.Status(fiber.StatusUnauthorized).SendString(
				`<div class="toast toast-error">Credenciales incorrectas</div>`,
			)
		}

		if !record.ValidatePassword(password) {
			return c.Status(fiber.StatusUnauthorized).SendString(
				`<div class="toast toast-error">Credenciales incorrectas</div>`,
			)
		}

		role := record.GetString("role")
		if role == "" {
			role = "viewer"
		}
		nombre := record.GetString("nombre")
		if nombre == "" {
			nombre = record.GetString("name")
		}

		token, err := auth.GenerateToken(cfg, record.Id, record.GetString("email"), role, nombre)
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error generando sesión</div>`)
		}
		expiry := 24 * time.Hour
		if remember {
			expiry = 72 * time.Hour
		}
		c.Cookie(&fiber.Cookie{
			Name:     "pr_token",
			Value:    token,
			Expires:  time.Now().Add(expiry),
			HTTPOnly: true,
			Secure:   cfg.Env == "production",
			SameSite: "Lax",
			Path:     "/",
		})
		c.Set("HX-Redirect", "/admin/dashboard")
		return c.SendString("")
	}
}

func Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:    "pr_token",
			Value:   "",
			Expires: time.Now().Add(-time.Hour),
			Path:    "/",
		})
		c.Set("HX-Redirect", "/admin/login")
		return c.SendString("")
	}
}

// ── DASHBOARD ──

func Dashboard(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, adminView.Dashboard())
		}
		return helpers.Render(c, layout.Admin("Dashboard", "dashboard", adminView.Dashboard()))
	}
}

// DashboardStats renders stat cards as an HTMX fragment.
//
// Real-data implementation (Task 09):
//   - Counts tiendas publicadas, locales disponibles, reservas pendientes.
//   - Counts page views for "today" (since 00:00) and last 7 days.
//   - Counts leads created in the last 30 days.
//
// Uses FindRecordsByFilter + len() (works across PB versions). Errors are
// silently swallowed so the dashboard always renders.
func DashboardStats(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tiendas, _ := pb.FindRecordsByFilter("tiendas", "status = 'publicado'", "", 1000, 0)
		locales, _ := pb.FindRecordsByFilter("locales_disponibles", "estado = 'disponible'", "", 500, 0)
		reservas, _ := pb.FindRecordsByFilter("reservas", "estado = 'pendiente'", "", 500, 0)

		today := time.Now().Format("2006-01-02")
		weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		monthAgo := time.Now().AddDate(0, -1, 0).Format("2006-01-02")

		visHoy, _ := pb.FindRecordsByFilter("page_views", "created >= '"+today+"'", "", 50000, 0)
		visSem, _ := pb.FindRecordsByFilter("page_views", "created >= '"+weekAgo+"'", "", 100000, 0)
		leads, _ := pb.FindRecordsByFilter("leads", "created >= '"+monthAgo+"'", "", 5000, 0)

		return helpers.Render(c, adminView.DashboardStats(adminView.DashboardStatsData{
			TiendasPublicadas:  len(tiendas),
			LocalesDisponibles: len(locales),
			ReservasPendientes: len(reservas),
			VisitasHoy:         len(visHoy),
			VisitasSemana:      len(visSem),
			NuevosLeads:        len(leads),
		}))
	}
}


func userNombre(r *core.Record) string {
	n := r.GetString("nombre")
	if n == "" {
		n = r.GetString("name")
	}
	return n
}

func currentUID(c *fiber.Ctx) string {
	if v, ok := c.Locals("user_id").(string); ok {
		return v
	}
	if v, ok := c.Locals("userID").(string); ok {
		return v
	}
	return ""
}

func UsersList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter("users", "", "email", 200, 0)
		rows := make([]adminView.UserRow, 0, len(records))
		for _, r := range records {
			rows = append(rows, adminView.UserRow{
				ID:    r.Id,
				Email: r.GetString("email"),
				Name:  userNombre(r),
				Role:  r.GetString("role"),
			})
		}
		data := adminView.UsersPageData{Rows: rows, CurrentUID: currentUID(c)}
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, adminView.UsersPage(data))
		}
		return helpers.Render(c, layout.Admin("Usuarios", "users", adminView.UsersPage(data)))
	}
}

func UserNew(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, adminView.UserForm(adminView.UserFormData{Role: "editor"}))
	}
}

func UserEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("users", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Usuario no encontrado</div>`)
		}
		data := adminView.UserFormData{
			ID:    r.Id,
			Email: r.GetString("email"),
			Name:  userNombre(r),
			Role:  r.GetString("role"),
		}
		return helpers.Render(c, adminView.UserForm(data))
	}
}

func UserCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		email := strings.TrimSpace(c.FormValue("email"))
		password := c.FormValue("password")
		passwordConfirm := c.FormValue("passwordConfirm")
		name := strings.TrimSpace(c.FormValue("name"))
		role := strings.TrimSpace(c.FormValue("role"))

		if email == "" || password == "" || role == "" {
			return c.Status(400).SendString(`<div class="toast toast-error">Email, contraseña y rol son requeridos</div>`)
		}
		if password != passwordConfirm {
			return c.Status(400).SendString(`<div class="toast toast-error">Las contraseñas no coinciden</div>`)
		}

		col, err := pb.FindCollectionByNameOrId("users")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		r := core.NewRecord(col)
		r.Set("email", email)
		r.Set("password", password)
		r.Set("passwordConfirm", passwordConfirm)
		r.Set("nombre", name)
		r.Set("name", name)
		r.Set("role", role)
		r.Set("activo", true)
		r.Set("verified", true)
		if err := pb.Save(r); err != nil {
			return c.Status(400).SendString(`<div class="toast toast-error">Error: ` + template.HTMLEscapeString(err.Error()) + `</div>`)
		}
		c.Set("HX-Redirect", "/admin/users")
		return c.SendString("")
	}
}

func UserUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("users", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Usuario no encontrado</div>`)
		}
		email := strings.TrimSpace(c.FormValue("email"))
		name := strings.TrimSpace(c.FormValue("name"))
		role := strings.TrimSpace(c.FormValue("role"))
		if email != "" {
			r.Set("email", email)
		}
		r.Set("nombre", name)
		r.Set("name", name)
		if role != "" {
			r.Set("role", role)
		}
		if err := pb.Save(r); err != nil {
			return c.Status(400).SendString(`<div class="toast toast-error">Error: ` + template.HTMLEscapeString(err.Error()) + `</div>`)
		}
		return helpers.Render(c, adminView.UserTableRow(adminView.UserRow{
			ID:    r.Id,
			Email: r.GetString("email"),
			Name:  userNombre(r),
			Role:  r.GetString("role"),
		}, currentUID(c)))
	}
}

func UserDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == currentUID(c) {
			return c.Status(400).SendString(`<div class="toast toast-error">No puedes eliminar tu propio usuario</div>`)
		}
		r, err := pb.FindRecordById("users", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		return c.SendString("")
	}
}

// ── TIENDAS CRUD ──

func TiendasList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter("tiendas", "", "nombre", 200, 0)
		rows := make([]adminView.TiendaRow, 0, len(records))
		for _, r := range records {
			gal := r.GetString("gal")
			// Normalize legacy 'norte'/'sur' values to new slugs
			switch gal {
			case "norte":
				gal = "placa-comercial"
			case "sur":
				gal = "torre-flamenco"
			}
			rows = append(rows, adminView.TiendaRow{
				ID:        r.Id,
				Nombre:    r.GetString("nombre"),
				Cat:       r.GetString("cat"),
				Gal:       gal,
				Local:     r.GetString("local"),
				Destacada: r.GetBool("destacada"),
				Status:    r.GetString("status"),
			})
		}
		data := adminView.TiendasPageData{Rows: rows}

		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, adminView.TiendasPage(data))
		}
		return helpers.Render(c, layout.Admin("Tiendas", "tiendas", adminView.TiendasPage(data)))
	}
}

func TiendaForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, adminView.TiendaForm(adminView.TiendaFormData{
			Status: "borrador",
			Gal:    "placa-comercial",
			Cat:    "tiendas",
		}))
	}
}

func TiendaCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		nombre := strings.TrimSpace(c.FormValue("nombre"))
		if nombre == "" {
			return c.SendString(`<div class="toast toast-error">El nombre es requerido</div>`)
		}
		col, err := pb.FindCollectionByNameOrId("tiendas")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		r := core.NewRecord(col)
		setTiendaFields(r, c)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error: ` + template.HTMLEscapeString(err.Error()) + `</div>`)
		}
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Tienda creada
<script>
  document.getElementById('modal-container').innerHTML='';
  htmx.ajax('GET','/admin/tiendas?fragment=rows',{target:'#tiendas-tbody',swap:'innerHTML'});
  setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000);
</script></div>`)
	}
}

func TiendaEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("tiendas", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		gal := r.GetString("gal")
		switch gal {
		case "norte":
			gal = "placa-comercial"
		case "sur":
			gal = "torre-flamenco"
		}
		data := adminView.TiendaFormData{
			ID:            r.Id,
			Nombre:        r.GetString("nombre"),
			Slug:          r.GetString("slug"),
			Cat:           r.GetString("cat"),
			Gal:           gal,
			Local:         r.GetString("local"),
			Logo:          r.GetString("logo"),
			HeroBg:        r.GetString("hero_bg"),
			Tags:          r.GetString("tags"),
			Desc:          r.GetString("desc"),
			About:         r.GetString("about"),
			About2:        r.GetString("about2"),
			Pay:           r.GetString("pay"),
			Photos:        r.GetString("photos"),
			Similar:       r.GetString("similar"),
			Whatsapp:      r.GetString("whatsapp"),
			Telefono:      r.GetString("telefono"),
			HorarioLV:     r.GetString("horario_lv"),
			HorarioSab:    r.GetString("horario_sab"),
			HorarioDom:    r.GetString("horario_dom"),
			StatusHorario: r.GetString("status_horario"),
			Status:        r.GetString("status"),
			Destacada:     r.GetBool("destacada"),
		}
		return helpers.Render(c, adminView.TiendaForm(data))
	}
}

func TiendaUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("tiendas", id)
		if err != nil {
			return c.SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		setTiendaFields(r, c)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error actualizando</div>`)
		}
		return c.SendString(`<div class="toast toast-success" id="toast-area">✅ Tienda actualizada
<script>
  document.getElementById('modal-container').innerHTML='';
  htmx.ajax('GET','/admin/tiendas?fragment=rows',{target:'#tiendas-tbody',swap:'innerHTML'});
  setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000);
</script></div>`)
	}
}

func TiendaDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("tiendas", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		return c.SendString("")
	}
}

func TiendaToggleStatus(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("tiendas", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		newStatus := "publicado"
		if r.GetString("status") == "publicado" {
			newStatus = "borrador"
		}
		r.Set("status", newStatus)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error</div>`)
		}
		label := map[string]string{"publicado": "✅ Publicada", "borrador": "📝 Borrador"}[newStatus]
		return c.SendString(fmt.Sprintf(`<div class="toast toast-success" id="toast-area">%s
<script>
  htmx.ajax('GET','/admin/tiendas?fragment=rows',{target:'#tiendas-tbody',swap:'innerHTML'});
  setTimeout(function(){var t=document.getElementById('toast-area');if(t)t.innerHTML=''},2000);
</script></div>`, label))
	}
}

func setTiendaFields(r *core.Record, c *fiber.Ctx) {
	r.Set("nombre", strings.TrimSpace(c.FormValue("nombre")))
	r.Set("slug", strings.TrimSpace(c.FormValue("slug")))
	r.Set("cat", c.FormValue("cat"))
	r.Set("gal", c.FormValue("gal"))
	r.Set("local", c.FormValue("local"))
	r.Set("logo", c.FormValue("logo"))
	r.Set("hero_bg", c.FormValue("hero_bg"))
	r.Set("tags", c.FormValue("tags"))
	r.Set("desc", c.FormValue("desc"))
	r.Set("about", c.FormValue("about"))
	r.Set("about2", c.FormValue("about2"))
	r.Set("pay", c.FormValue("pay"))
	r.Set("photos", c.FormValue("photos"))
	r.Set("similar", c.FormValue("similar"))
	r.Set("whatsapp", c.FormValue("whatsapp"))
	r.Set("telefono", c.FormValue("telefono"))
	r.Set("rating", c.FormValue("rating"))
	r.Set("horario_lv", c.FormValue("horario_lv"))
	r.Set("horario_sab", c.FormValue("horario_sab"))
	r.Set("horario_dom", c.FormValue("horario_dom"))
	r.Set("status_horario", c.FormValue("status_horario"))
	r.Set("status", c.FormValue("status"))
	r.Set("destacada", c.FormValue("destacada") == "on")
}

func tiendaFormHTML(id, nombre, slug, cat, gal, local, logo, tags, desc, about, about2, pay, photos, similar, whatsapp, telefono, rating, horarioLV, horarioSab, horarioDom, status string, destacada bool) string {
	method := `hx-post="/admin/tiendas"`
	if id != "" {
		method = fmt.Sprintf(`hx-put="/admin/tiendas/%s"`, id)
	}
	title := map[bool]string{true: "Editar Tienda", false: "Nueva Tienda"}[id != ""]

	chk := func(b bool) string {
		if b {
			return " checked"
		}
		return ""
	}

	cats := []struct{ v, l string }{
		{"tiendas", "Tiendas & Moda"}, {"restaurantes", "Restaurantes & Café"},
		{"farmacias", "Farmacias"}, {"salud", "Salud & Belleza"},
		{"tecnologia", "Tecnología"}, {"servicios", "Servicios"},
	}
	var catOpts strings.Builder
	for _, o := range cats {
		s := ""
		if o.v == cat {
			s = " selected"
		}
		catOpts.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, o.v, s, o.l))
	}

	gals := []struct{ v, l string }{{"norte", "Placa Comercial"}, {"sur", "Torre Flamenco"}}
	var galOpts strings.Builder
	for _, o := range gals {
		s := ""
		if o.v == gal {
			s = " selected"
		}
		galOpts.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, o.v, s, o.l))
	}

	statuses := []struct{ v, l string }{{"borrador", "Borrador"}, {"publicado", "Publicado"}}
	var statusOpts strings.Builder
	for _, st := range statuses {
		s := ""
		if st.v == status {
			s = " selected"
		}
		statusOpts.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, st.v, s, st.l))
	}

	return fmt.Sprintf(`<div class="modal-overlay" onclick="this.remove()">
  <div class="modal-card" onclick="event.stopPropagation()" style="max-width:700px;max-height:90vh;overflow-y:auto">
    <div class="modal-header">
      <h3>%s</h3>
      <button onclick="document.getElementById('modal-container').innerHTML=''" style="background:none;border:none;cursor:pointer;font-size:20px;color:var(--md-outline)">✕</button>
    </div>
    <form %s hx-target="#toast-area" hx-swap="innerHTML">
      <div class="form-row">
        <div class="form-field" style="grid-column:span 2">
          <label>Nombre *</label>
          <input type="text" name="nombre" value="%s" required class="form-input" placeholder="Nombre de la tienda"/>
        </div>
      </div>
      <div class="form-row">
        <div class="form-field"><label>Slug (URL)</label><input type="text" name="slug" value="%s" class="form-input" placeholder="nombre-tienda"/></div>
        <div class="form-field"><label>Local (ej: Local 8)</label><input type="text" name="local" value="%s" class="form-input" placeholder="Local 8"/></div>
      </div>
      <div class="form-row">
        <div class="form-field"><label>Categoría</label><select name="cat" class="form-input">%s</select></div>
        <div class="form-field"><label>Nivel</label><select name="gal" class="form-input">%s</select></div>
      </div>
      <div class="form-field"><label>URL Logo</label><input type="url" name="logo" value="%s" class="form-input" placeholder="https://..."/></div>
      <div class="form-field"><label>Tags (separados por coma)</label><input type="text" name="tags" value="%s" class="form-input" placeholder="Café,Frappuccino,WiFi"/></div>
      <div class="form-field"><label>Descripción corta (hero)</label><input type="text" name="desc" value="%s" class="form-input" placeholder="Descripción breve visible en la página de la tienda"/></div>
      <div class="form-field"><label>Sobre la tienda (párrafo 1)</label><textarea name="about" class="form-input" rows="3">%s</textarea></div>
      <div class="form-field"><label>Sobre la tienda (párrafo 2)</label><textarea name="about2" class="form-input" rows="2">%s</textarea></div>
      <div class="form-field"><label>Medios de pago</label><input type="text" name="pay" value="%s" class="form-input" placeholder="Efectivo · Tarjetas · Débito"/></div>
      <div class="form-field"><label>Fotos galería (URLs separadas por coma, mín 4)</label><textarea name="photos" class="form-input" rows="3" placeholder="https://img1.jpg,https://img2.jpg,...">%s</textarea></div>
      <div class="form-field"><label>Tiendas similares (slugs separados por coma)</label><input type="text" name="similar" value="%s" class="form-input" placeholder="starbucks,oakberry,krispy-kreme"/></div>
      <div class="form-row">
        <div class="form-field"><label>WhatsApp</label><input type="text" name="whatsapp" value="%s" class="form-input" placeholder="56912345678"/></div>
        <div class="form-field"><label>Teléfono</label><input type="text" name="telefono" value="%s" class="form-input" placeholder="+56 2 1234 5678"/></div>
      </div>
      <div class="form-row">
        <div class="form-field"><label>Rating</label><input type="text" name="rating" value="%s" class="form-input" placeholder="4.7"/></div>
        <div class="form-field"><label>Estado</label><select name="status" class="form-input">%s</select></div>
      </div>
      <div class="form-row">
        <div class="form-field"><label>Horario Lun–Vie</label><input type="text" name="horario_lv" value="%s" class="form-input" placeholder="9:00 – 21:00"/></div>
        <div class="form-field"><label>Horario Sábado</label><input type="text" name="horario_sab" value="%s" class="form-input" placeholder="10:00 – 20:00"/></div>
      </div>
      <div class="form-row">
        <div class="form-field"><label>Horario Domingo</label><input type="text" name="horario_dom" value="%s" class="form-input" placeholder="Cerrado"/></div>
        <div class="form-field" style="display:flex;align-items:center;padding-top:22px">
          <label style="display:flex;align-items:center;gap:8px;font-size:14px;cursor:pointer">
            <input type="checkbox" name="destacada"%s style="width:16px;height:16px;accent-color:var(--md-primary)"/> Destacada
          </label>
        </div>
      </div>
      <div class="modal-actions">
        <button type="button" onclick="document.getElementById('modal-container').innerHTML=''" class="topbar-btn topbar-btn-outline">Cancelar</button>
        <button type="submit" class="topbar-btn topbar-btn-primary">Guardar</button>
      </div>
    </form>
  </div>
</div>`,
		title, method,
		template.HTMLEscapeString(nombre),
		template.HTMLEscapeString(slug),
		template.HTMLEscapeString(local),
		catOpts.String(), galOpts.String(),
		template.HTMLEscapeString(logo),
		template.HTMLEscapeString(tags),
		template.HTMLEscapeString(desc),
		template.HTMLEscapeString(about),
		template.HTMLEscapeString(about2),
		template.HTMLEscapeString(pay),
		template.HTMLEscapeString(photos),
		template.HTMLEscapeString(similar),
		template.HTMLEscapeString(whatsapp),
		template.HTMLEscapeString(telefono),
		template.HTMLEscapeString(rating),
		statusOpts.String(),
		template.HTMLEscapeString(horarioLV),
		template.HTMLEscapeString(horarioSab),
		template.HTMLEscapeString(horarioDom),
		chk(destacada),
	)
}

// sel returns " selected" if val == target
func sel(val, target string) string {
	if val == target {
		return " selected"
	}
	return ""
}

// ── R2 UPLOAD ──

// UploadFile recibe multipart/form-data con campo "file",
// lo sube a R2 y devuelve {"url": "https://..."} como JSON.
// POST /admin/upload
func UploadFile(cfg *config.Config, r2Client *r2.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fh, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "no file provided"})
		}

		const maxSize = 10 << 20
		if fh.Size > maxSize {
			return c.Status(400).JSON(fiber.Map{"error": "file too large (max 10 MB)"})
		}

		f, err := fh.Open()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "cannot open file"})
		}
		defer f.Close()

		key := fmt.Sprintf("uploads/%s/%d-%s",
			time.Now().Format("2006-01"),
			time.Now().UnixMilli(),
			sanitizeFilename(fh.Filename),
		)

		contentType := fh.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		url, err := r2Client.Upload(context.Background(), key, f, contentType)
		if err != nil {
			log.Printf("R2 upload error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "upload failed"})
		}

		return c.JSON(fiber.Map{"url": url})
	}
}

func sanitizeFilename(name string) string {
	name = strings.ToLower(name)
	r := strings.NewReplacer(" ", "-", "(", "", ")", "", "[", "", "]", "")
	return r.Replace(name)
}

// ── CONTENT BLOCKS (NOTICIA / COMUNICADO / PROMOCION) ──────────────────────────

// contentTitleByCategory returns the human-readable title for a category code.
func contentTitleByCategory(category string) string {
	switch category {
	case "NOTICIA":
		return "Noticias"
	case "COMUNICADO":
		return "Comunicados"
	case "PROMOCION":
		return "Promociones"
	}
	return category
}

// contentActivePage returns the sidebar active key for a category code.
func contentActivePage(category string) string {
	switch category {
	case "NOTICIA":
		return "noticias"
	case "COMUNICADO":
		return "comunicados"
	case "PROMOCION":
		return "promociones"
	}
	return "noticias"
}

// contentDateForRow renders a friendly date for the table row.
func contentDateForRow(r *core.Record) string {
	if dt := r.GetDateTime("published_at"); !dt.IsZero() {
		return dt.Time().Format("2 Jan 2006 15:04")
	}
	if dt := r.GetDateTime("date"); !dt.IsZero() {
		return dt.Time().Format("2 Jan 2006")
	}
	return ""
}

// contentDateForInput renders the datetime-local value (YYYY-MM-DDTHH:MM).
func contentDateForInput(r *core.Record, field string) string {
	if dt := r.GetDateTime(field); !dt.IsZero() {
		return dt.Time().Format("2006-01-02T15:04")
	}
	return ""
}

// ContentList renders the admin list for a single category.
func ContentList(cfg *config.Config, pb *pocketbase.PocketBase, category string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter := fmt.Sprintf("category = '%s'", category)
		records, _ := pb.FindRecordsByFilter("content_blocks", filter, "-published_at,-date,-created", 200, 0)
		rows := make([]adminView.ContentRow, 0, len(records))
		for _, r := range records {
			rows = append(rows, adminView.ContentRow{
				ID:          r.Id,
				Title:       r.GetString("title"),
				Category:    r.GetString("category"),
				Status:      r.GetString("status"),
				Featured:    r.GetBool("featured"),
				PublishedAt: contentDateForRow(r),
			})
		}
		data := adminView.ContentPageData{
			Category: category,
			Title:    contentTitleByCategory(category),
			Rows:     rows,
		}
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, adminView.ContentPage(data))
		}
		return helpers.Render(c, layout.Admin(data.Title, contentActivePage(category), adminView.ContentPage(data)))
	}
}

// ContentNew returns the empty form (modal).
func ContentNew(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		category := strings.ToUpper(strings.TrimSpace(c.Query("cat", "NOTICIA")))
		switch category {
		case "NOTICIA", "COMUNICADO", "PROMOCION":
		default:
			category = "NOTICIA"
		}
		return helpers.Render(c, adminView.ContentForm(adminView.ContentFormData{
			Category: category,
			Status:   "borrador",
		}))
	}
}

// ContentEdit returns the populated form (modal).
func ContentEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("content_blocks", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		data := adminView.ContentFormData{
			ID:          r.Id,
			Category:    r.GetString("category"),
			Title:       r.GetString("title"),
			Description: r.GetString("description"),
			Body:        r.GetString("body"),
			ImageURL:    r.GetString("image_url"),
			Status:      r.GetString("status"),
			Featured:    r.GetBool("featured"),
			PublishedAt: contentDateForInput(r, "published_at"),
			ExpiresAt:   contentDateForInput(r, "expires_at"),
		}
		return helpers.Render(c, adminView.ContentForm(data))
	}
}

// setContentFields applies form values to a record.
func setContentFields(r *core.Record, c *fiber.Ctx) {
	r.Set("title", strings.TrimSpace(c.FormValue("title")))
	r.Set("description", c.FormValue("description"))
	r.Set("body", c.FormValue("body"))
	r.Set("image_url", c.FormValue("image_url"))
	category := strings.ToUpper(strings.TrimSpace(c.FormValue("category")))
	switch category {
	case "NOTICIA", "COMUNICADO", "PROMOCION":
		r.Set("category", category)
	}
	status := c.FormValue("status")
	if status == "" {
		status = "borrador"
	}
	r.Set("status", status)
	r.Set("featured", c.FormValue("featured") == "true" || c.FormValue("featured") == "on")
	if v := strings.TrimSpace(c.FormValue("published_at")); v != "" {
		// HTML datetime-local: 2006-01-02T15:04
		if t, err := time.Parse("2006-01-02T15:04", v); err == nil {
			r.Set("published_at", t)
		}
	}
	if v := strings.TrimSpace(c.FormValue("expires_at")); v != "" {
		if t, err := time.Parse("2006-01-02T15:04", v); err == nil {
			r.Set("expires_at", t)
		}
	}
}

// contentSaveResponse mirrors the toast pattern used by tiendas/multimedia
// handlers: closes the modal, refreshes the listing in #content, and
// auto-clears the toast after 2s.
func contentSaveResponse(c *fiber.Ctx, msg, listURL string) error {
	return c.SendString(fmt.Sprintf(`<div class="toast toast-success" id="toast-area">%s
<script>
  var m=document.getElementById('modal-container'); if(m){m.replaceChildren();}
  htmx.ajax('GET','%s',{target:'#content',swap:'innerHTML'});
  setTimeout(function(){var t=document.getElementById('toast-area');if(t){t.replaceChildren();}},2000);
</script></div>`, template.HTMLEscapeString(msg), template.HTMLEscapeString(listURL)))
}

// ContentCreate handles POST /admin/content.
func ContentCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		title := strings.TrimSpace(c.FormValue("title"))
		if title == "" {
			return c.SendString(`<div class="toast toast-error">El título es requerido</div>`)
		}
		col, err := pb.FindCollectionByNameOrId("content_blocks")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		r := core.NewRecord(col)
		setContentFields(r, c)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error: ` + template.HTMLEscapeString(err.Error()) + `</div>`)
		}
		return contentSaveResponse(c, "✅ Registro creado", "/admin/content?cat="+r.GetString("category"))
	}
}

// ContentUpdate handles PUT /admin/content/:id.
func ContentUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("content_blocks", id)
		if err != nil {
			return c.SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		setContentFields(r, c)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error actualizando</div>`)
		}
		return contentSaveResponse(c, "✅ Registro actualizado", "/admin/content?cat="+r.GetString("category"))
	}
}

// ContentDelete handles DELETE /admin/content/:id.
func ContentDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("content_blocks", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		return c.SendString("")
	}
}

// ── LOCALES DISPONIBLES (CRUD) ─────────────────────────────────────────────

// localesSaveResponse closes the modal and refreshes the listing.
func localesSaveResponse(c *fiber.Ctx, msg string) error {
	return c.SendString(fmt.Sprintf(`<div class="toast toast-success" id="toast-area">%s
<script>
  var m=document.getElementById('modal-container'); if(m){m.replaceChildren();}
  htmx.ajax('GET','/admin/locales',{target:'#content',swap:'innerHTML'});
  setTimeout(function(){var t=document.getElementById('toast-area');if(t){t.replaceChildren();}},2000);
</script></div>`, template.HTMLEscapeString(msg)))
}

// LocalesList renders the admin list of locales_disponibles.
func LocalesList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter("locales_disponibles", "", "galeria,numero,nombre", 200, 0)
		rows := make([]adminView.LocalRow, 0, len(records))
		for _, r := range records {
			rows = append(rows, adminView.LocalRow{
				ID:      r.Id,
				Nombre:  r.GetString("nombre"),
				Galeria: r.GetString("galeria"),
				Numero:  r.GetString("numero"),
				Estado:  r.GetString("estado"),
				M2:      r.GetFloat("m2"),
			})
		}
		data := adminView.LocalesPageData{Rows: rows}
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, adminView.LocalesPage(data))
		}
		return helpers.Render(c, layout.Admin("Locales Disponibles", "locales", adminView.LocalesPage(data)))
	}
}

// LocalNew returns the empty local form (modal).
func LocalNew(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, adminView.LocalForm(adminView.LocalFormData{
			Estado: "disponible",
		}))
	}
}

// LocalEdit returns the populated local form (modal).
func LocalEdit(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("locales_disponibles", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		m2Str := ""
		if v := r.GetFloat("m2"); v > 0 {
			m2Str = fmt.Sprintf("%g", v)
		}
		data := adminView.LocalFormData{
			ID:            r.Id,
			Nombre:        r.GetString("nombre"),
			Galeria:       r.GetString("galeria"),
			Numero:        r.GetString("numero"),
			Piso:          r.GetString("piso"),
			M2:            m2Str,
			Descripcion:   r.GetString("descripcion"),
			PrecioRef:     r.GetString("precio_ref"),
			Estado:        r.GetString("estado"),
			ImagenURL:     r.GetString("imagen_url"),
			ContactoEmail: r.GetString("contacto_email"),
			ContactoTel:   r.GetString("contacto_tel"),
		}
		return helpers.Render(c, adminView.LocalForm(data))
	}
}

// setLocalFields applies form values to a local record.
func setLocalFields(r *core.Record, c *fiber.Ctx) {
	r.Set("nombre", strings.TrimSpace(c.FormValue("nombre")))
	r.Set("galeria", c.FormValue("galeria"))
	r.Set("numero", c.FormValue("numero"))
	r.Set("piso", c.FormValue("piso"))
	if v := strings.TrimSpace(c.FormValue("m2")); v != "" {
		var f float64
		if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
			r.Set("m2", f)
		}
	} else {
		r.Set("m2", 0)
	}
	r.Set("descripcion", c.FormValue("descripcion"))
	r.Set("precio_ref", c.FormValue("precio_ref"))
	estado := c.FormValue("estado")
	if estado == "" {
		estado = "disponible"
	}
	r.Set("estado", estado)
	r.Set("imagen_url", c.FormValue("imagen_url"))
	r.Set("contacto_email", c.FormValue("contacto_email"))
	r.Set("contacto_tel", c.FormValue("contacto_tel"))
}

// LocalCreate handles POST /admin/locales.
func LocalCreate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		nombre := strings.TrimSpace(c.FormValue("nombre"))
		if nombre == "" {
			return c.SendString(`<div class="toast toast-error">El nombre es requerido</div>`)
		}
		col, err := pb.FindCollectionByNameOrId("locales_disponibles")
		if err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error de BD</div>`)
		}
		r := core.NewRecord(col)
		setLocalFields(r, c)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error: ` + template.HTMLEscapeString(err.Error()) + `</div>`)
		}
		return localesSaveResponse(c, "✅ Local creado")
	}
}

// LocalUpdate handles PUT /admin/locales/:id.
func LocalUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("locales_disponibles", id)
		if err != nil {
			return c.SendString(`<div class="toast toast-error">No encontrado</div>`)
		}
		setLocalFields(r, c)
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error actualizando</div>`)
		}
		return localesSaveResponse(c, "✅ Local actualizado")
	}
}

// LocalDelete handles DELETE /admin/locales/:id.
func LocalDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("locales_disponibles", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		return c.SendString("")
	}
}

// ── RESERVAS (admin: list + change estado + delete) ────────────────────────

// ReservasList renders the admin list of reservas.
func ReservasList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter("reservas", "", "-created", 200, 0)
		rows := make([]adminView.ReservaRow, 0, len(records))
		for _, r := range records {
			cbTitle := ""
			if id := r.GetString("content_block_id"); id != "" {
				if cb, err := pb.FindRecordById("content_blocks", id); err == nil {
					cbTitle = cb.GetString("title")
				}
			}
			created := ""
			if dt := r.GetDateTime("created"); !dt.IsZero() {
				created = dt.Time().Format("2 Jan 2006 15:04")
			}
			rows = append(rows, adminView.ReservaRow{
				ID:                r.Id,
				Nombre:            r.GetString("nombre"),
				Email:             r.GetString("email"),
				Telefono:          r.GetString("telefono"),
				Estado:            r.GetString("estado"),
				ContentBlockTitle: cbTitle,
				CreatedAt:         created,
				Cantidad:          int(r.GetFloat("cantidad")),
			})
		}
		data := adminView.ReservasPageData{Rows: rows}
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, adminView.ReservasPage(data))
		}
		return helpers.Render(c, layout.Admin("Reservas", "reservas", adminView.ReservasPage(data)))
	}
}

// ReservaUpdateEstado handles POST /admin/reservas/:id/estado.
func ReservaUpdateEstado(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("reservas", id)
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">No encontrada</div>`)
		}
		estado := c.FormValue("estado")
		switch estado {
		case "pendiente", "confirmada", "cancelada":
			r.Set("estado", estado)
		default:
			return c.SendString(`<div class="toast toast-error">Estado inválido</div>`)
		}
		if err := pb.Save(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error actualizando</div>`)
		}
		return c.SendString("")
	}
}

// ReservaDelete handles DELETE /admin/reservas/:id.
func ReservaDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("reservas", id)
		if err != nil {
			return c.Status(404).SendString("")
		}
		if err := pb.Delete(r); err != nil {
			return c.SendString(`<div class="toast toast-error">Error eliminando</div>`)
		}
		return c.SendString("")
	}
}

// ── REPORTS / ANALYTICS ──

// ReportsPageHandler renders the analytics overview page (Task 09).
//
// Aggregates real data from page_views and leads:
//   - Visits in last 7 / 30 days
//   - Top pages in the last 30 days (by raw count, sorted desc, top 20)
//   - Total leads + leads created in the last 30 days
func ReportsPageHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		monthAgo := time.Now().AddDate(0, -1, 0).Format("2006-01-02")

		visSem, _ := pb.FindRecordsByFilter("page_views", "created >= '"+weekAgo+"'", "", 100000, 0)
		visMes, _ := pb.FindRecordsByFilter("page_views", "created >= '"+monthAgo+"'", "", 500000, 0)

		// Aggregate top pages (last 30 days) in-memory.
		counts := make(map[string]int, len(visMes))
		for _, r := range visMes {
			path := r.GetString("path")
			if path == "" {
				continue
			}
			counts[path]++
		}
		top := make([]adminView.TopPage, 0, len(counts))
		for p, n := range counts {
			top = append(top, adminView.TopPage{Path: p, Count: n})
		}
		sort.Slice(top, func(i, j int) bool { return top[i].Count > top[j].Count })
		if len(top) > 20 {
			top = top[:20]
		}

		allLeads, _ := pb.FindRecordsByFilter("leads", "", "", 100000, 0)
		newLeads, _ := pb.FindRecordsByFilter("leads", "created >= '"+monthAgo+"'", "", 5000, 0)

		data := adminView.ReportsPageData{
			TotalVisitasSemana: len(visSem),
			TotalVisitasMes:    len(visMes),
			TopPages:           top,
			TotalLeads:         len(allLeads),
			LeadsNuevos:        len(newLeads),
		}

		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, adminView.ReportsPage(data))
		}
		return helpers.Render(c, layout.Admin("Informes", "reports", adminView.ReportsPage(data)))
	}
}

// ReportsExport streams CSV downloads for either page_views or leads.
//
// Query: ?type=page_views | ?type=leads (default: page_views)
// Sets Content-Type: text/csv and a Content-Disposition attachment with a
// timestamped filename. Caps export at 100k rows for safety.
func ReportsExport(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		t := c.Query("type", "page_views")
		stamp := time.Now().Format("20060102-150405")

		c.Set("Content-Type", "text/csv; charset=utf-8")
		c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-%s.csv"`, t, stamp))

		var sb strings.Builder
		w := csv.NewWriter(&sb)

		switch t {
		case "leads":
			_ = w.Write([]string{"id", "created", "nombre", "email", "telefono", "local_id", "estado", "mensaje"})
			records, _ := pb.FindRecordsByFilter("leads", "", "-created", 100000, 0)
			for _, r := range records {
				created := ""
				if dt := r.GetDateTime("created"); !dt.IsZero() {
					created = dt.Time().Format(time.RFC3339)
				}
				_ = w.Write([]string{
					r.Id,
					created,
					r.GetString("nombre"),
					r.GetString("email"),
					r.GetString("telefono"),
					r.GetString("local_id"),
					r.GetString("estado"),
					r.GetString("mensaje"),
				})
			}
		default: // page_views
			_ = w.Write([]string{"id", "created", "path", "referrer", "user_agent", "ip"})
			records, _ := pb.FindRecordsByFilter("page_views", "", "-created", 100000, 0)
			for _, r := range records {
				created := ""
				if dt := r.GetDateTime("created"); !dt.IsZero() {
					created = dt.Time().Format(time.RFC3339)
				}
				_ = w.Write([]string{
					r.Id,
					created,
					r.GetString("path"),
					r.GetString("referrer"),
					r.GetString("user_agent"),
					r.GetString("ip"),
				})
			}
		}
		w.Flush()
		return c.SendString(sb.String())
	}
}

// ── SITE SETTINGS ─────────────────────────────────────────────────────────────

// SettingsPageHandler renders the global site settings admin page.
// Loads the current values for hero_bg_url and search_bg_url so the upload
// widgets can show existing previews.
func SettingsPageHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := adminView.SettingsPageData{
			HeroBgURL:   helpers.GetSetting(pb, "hero_bg_url"),
			SearchBgURL: helpers.GetSetting(pb, "search_bg_url"),
		}
		content := adminView.SettingsPage(data)
		if c.Get("HX-Request") == "true" {
			return helpers.Render(c, content)
		}
		return helpers.Render(c, layout.Admin("Ajustes", "settings", content))
	}
}

// SettingsUpdate processes the admin settings form submission. Saves
// hero_bg_url and search_bg_url via helpers.SetSetting (creates the records on
// first save). Returns the form HTML with a success or error toast.
func SettingsUpdate(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		heroBg := strings.TrimSpace(c.FormValue("hero_bg_url"))
		searchBg := strings.TrimSpace(c.FormValue("search_bg_url"))
		var errMsg string
		if err := helpers.SetSetting(pb, "hero_bg_url", heroBg); err != nil {
			errMsg = "Error guardando hero_bg_url: " + err.Error()
		}
		if err := helpers.SetSetting(pb, "search_bg_url", searchBg); err != nil {
			errMsg = "Error guardando search_bg_url: " + err.Error()
		}
		success := ""
		if errMsg == "" {
			success = "Ajustes guardados correctamente."
		}
		return helpers.Render(c, adminView.SettingsPage(adminView.SettingsPageData{
			HeroBgURL:   heroBg,
			SearchBgURL: searchBg,
			SuccessMsg:  success,
			ErrorMsg:    errMsg,
		}))
	}
}
