# Task 06 — Eventos con Reservas: página pública, formulario HTMX, admin CRUD

**Depends on:** Task 01 (módulo renombrado), Task 02 (colección content_blocks correcta)
**Estimated complexity:** media-alta — nueva colección + página pública + dialog HTMX + admin CRUD

---

## Estado del repositorio al inicio de este task

```
Module: cms-plazareal
Ruta pública /eventos.html: NO existe aún
Colección PocketBase 'reservas': NO existe
Colección 'content_blocks': existe con campo category (filtra "EVENTO")
Fragment /fragments/eventos: existe (muestra en homepage) — se reutiliza
Fragment /fragments/eventos-public: NO existe (es el nuevo que se crea aquí)
```

La app ya tiene un fragment `/fragments/eventos` usado en la homepage. Este task crea:
1. `/fragments/eventos-public` — versión página completa con botón "Reservar"
2. `/fragments/reserva-form` — retorna el HTML del formulario de reserva
3. `POST /api/reservas` — persiste una reserva en PocketBase
4. `web/eventos.html` — la página pública completa
5. CRUD de reservas en el admin

**Ruta en main.go para registros nuevos:**
```go
// Las rutas fragment siguen el grupo: frag := app.Group("/fragments")
// Las rutas api siguen el grupo: api := app.Group("/api")
// Las rutas admin siguen el grupo: adm := app.Group("/admin", middleware.AuthRequired(cfg), middleware.InjectUser())
```

**Patrón de fragment existente (referencia):**
`internal/handlers/fragments/eventos.go` — usa `pb.FindRecordsByFilter("content_blocks", filter, "-date", 6, 0)`

**Patrón de handler API existente (referencia):**
`internal/handlers/api/handlers.go` — package `api`, imports `cms-plazareal/internal/config`

**Patrón de admin handler existente (referencia):**
`internal/handlers/admin/handlers.go` — mismo archivo, agrega funciones al final

---

## Objetivo

Crear la sección de eventos con reservas: página web pública con modal de formulario HTMX, endpoint de reservas y admin para gestionar reservas.

---

## Archivos a tocar

| Acción | Archivo |
|--------|---------|
| Modificar | `internal/auth/collections.go` — nueva colección `reservas` |
| Crear | `web/eventos.html` |
| Crear | `internal/handlers/fragments/eventos_pub.go` |
| Modificar | `internal/handlers/api/handlers.go` — endpoint POST /api/reservas |
| Modificar | `internal/handlers/admin/handlers.go` — CRUD reservas |
| Crear | `internal/templates/admin/pages/reservas.html` |
| Modificar | `cmd/server/main.go` — nuevas rutas |

---

## Implementación

- [ ] **Step 1: Agregar colección reservas en internal/auth/collections.go**

Buscar la función `ensureCollections` y localizar el bloque `return nil` al final:
```bash
grep -n "return nil" internal/auth/collections.go | tail -3
```

Insertar ANTES del último `return nil`:

```go
// ── RESERVAS ──────────────────────────────────────────────
if _, err := app.FindCollectionByNameOrId("reservas"); err != nil {
    col := core.NewBaseCollection("reservas")
    col.Fields.Add(
        &core.TextField{Name: "evento_id", Required: true},
        &core.TextField{Name: "nombre", Required: true},
        &core.TextField{Name: "rut"},
        &core.TextField{Name: "email", Required: true},
        &core.TextField{Name: "telefono"},
        &core.NumberField{Name: "cantidad_personas"},
        &core.TextField{Name: "mensaje"},
        &core.TextField{Name: "status"},
        &core.BoolField{Name: "leido"},
    )
    if err := app.Save(col); err != nil {
        return err
    }
    log.Println("  ✅ Collection 'reservas' created")
}
```

- [ ] **Step 2: Crear internal/handlers/fragments/eventos_pub.go**

```bash
touch internal/handlers/fragments/eventos_pub.go
```

Contenido completo:

```go
package fragments

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"cms-plazareal/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
)

// EventosPublic retorna la grilla completa de eventos próximos con botón de reserva.
// GET /fragments/eventos-public
func EventosPublic(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		now := time.Now().Format("2006-01-02 15:04:05")
		filter := fmt.Sprintf("status = 'publicado' && (category = 'EVENTO' || category = 'INFORMACIÓN') && date >= '%s'", now)

		records, err := pb.FindRecordsByFilter("content_blocks", filter, "date", 20, 0)
		if err != nil {
			return c.Status(500).SendString(`<p class="error-msg">Error cargando eventos.</p>`)
		}

		if len(records) == 0 {
			return c.SendString(`<div class="empty-state"><span class="material-symbols-outlined">event_busy</span><p>No hay eventos próximos.</p></div>`)
		}

		var sb strings.Builder
		sb.WriteString(`<div class="eventos-grid">`)
		for _, r := range records {
			id := r.Id
			title := template.HTMLEscapeString(r.GetString("title"))
			desc := template.HTMLEscapeString(r.GetString("description"))
			cat := template.HTMLEscapeString(r.GetString("category"))
			imgURL := r.GetString("image_url")
			dateStr := ""
			if dt := r.GetDateTime("date"); !dt.IsZero() {
				dateStr = dt.Time().Format("02/01/2006 15:04")
			}

			imgHTML := `<div class="evento-cover-placeholder"><span class="material-symbols-outlined">event</span></div>`
			if imgURL != "" {
				imgHTML = fmt.Sprintf(`<img src="%s" alt="%s" class="evento-cover"/>`,
					template.HTMLEscapeString(imgURL), title)
			}

			sb.WriteString(fmt.Sprintf(`
<div class="evento-card">
  %s
  <div class="evento-info">
    <span class="evento-cat">%s</span>
    <h3 class="evento-title">%s</h3>
    <p class="evento-date"><span class="material-symbols-outlined">calendar_today</span> %s</p>
    <p class="evento-desc">%s</p>
    <button class="btn-reservar"
      hx-get="/fragments/reserva-form?evento_id=%s&titulo=%s"
      hx-target="#reserva-modal-inner"
      hx-swap="innerHTML"
      onclick="document.getElementById('reserva-modal').showModal()">
      <span class="material-symbols-outlined">event_seat</span> Reservar lugar
    </button>
  </div>
</div>`, imgHTML, cat, title, dateStr, desc, id, template.URLQueryEscaper(r.GetString("title"))))
		}
		sb.WriteString(`</div>`)

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// ReservaForm retorna el formulario HTML para reservar un evento específico.
// GET /fragments/reserva-form?evento_id=XXX&titulo=YYY
func ReservaForm(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		eventoID := template.HTMLEscapeString(c.Query("evento_id"))
		titulo := template.HTMLEscapeString(c.Query("titulo"))

		html := fmt.Sprintf(`
<form hx-post="/api/reservas" hx-target="#reserva-feedback" hx-swap="innerHTML">
  <input type="hidden" name="evento_id" value="%s"/>
  <h3 class="modal-title">Reservar lugar</h3>
  <p class="modal-subtitle">%s</p>
  <div class="form-group">
    <label>Nombre completo *</label>
    <input type="text" name="nombre" required placeholder="Tu nombre"/>
  </div>
  <div class="form-group">
    <label>Email *</label>
    <input type="email" name="email" required placeholder="tu@email.com"/>
  </div>
  <div class="form-group">
    <label>Teléfono</label>
    <input type="tel" name="telefono" placeholder="+56 9 1234 5678"/>
  </div>
  <div class="form-group">
    <label>Cantidad de personas</label>
    <input type="number" name="cantidad_personas" min="1" max="10" value="1"/>
  </div>
  <div class="form-group">
    <label>Mensaje (opcional)</label>
    <textarea name="mensaje" rows="3" placeholder="¿Alguna consulta?"></textarea>
  </div>
  <div id="reserva-feedback"></div>
  <div class="form-actions">
    <button type="submit" class="btn-primary">
      <span class="material-symbols-outlined">check_circle</span> Confirmar reserva
    </button>
    <button type="button" class="btn-outline"
      onclick="document.getElementById('reserva-modal').close()">
      Cancelar
    </button>
  </div>
</form>`, eventoID, titulo)

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}
```

- [ ] **Step 3: Agregar CreateReserva en internal/handlers/api/handlers.go**

Verificar los imports actuales del archivo:
```bash
grep -n '"github.com/pocketbase/pocketbase/core"' internal/handlers/api/handlers.go
```

Si el import `core` no está, agregarlo al bloque import. Luego agregar al final del archivo:

```go
// CreateReserva persiste una reserva de evento en PocketBase.
// POST /api/reservas
func CreateReserva(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		eventoID := strings.TrimSpace(c.FormValue("evento_id"))
		nombre := strings.TrimSpace(c.FormValue("nombre"))
		email := strings.TrimSpace(c.FormValue("email"))

		if eventoID == "" || nombre == "" || email == "" {
			return c.Status(400).SendString(`<p class="form-error">Nombre y email son obligatorios.</p>`)
		}

		col, err := pb.FindCollectionByNameOrId("reservas")
		if err != nil {
			return c.Status(500).SendString(`<p class="form-error">Error interno.</p>`)
		}

		r := core.NewRecord(col)
		r.Set("evento_id", eventoID)
		r.Set("nombre", nombre)
		r.Set("email", email)
		r.Set("telefono", strings.TrimSpace(c.FormValue("telefono")))
		r.Set("cantidad_personas", c.FormValue("cantidad_personas"))
		r.Set("mensaje", strings.TrimSpace(c.FormValue("mensaje")))
		r.Set("status", "pendiente")
		r.Set("leido", false)

		if err := pb.Save(r); err != nil {
			return c.Status(500).SendString(`<p class="form-error">No se pudo guardar la reserva. Intenta nuevamente.</p>`)
		}

		return c.SendString(`<div class="form-success"><span class="material-symbols-outlined">check_circle</span> ¡Reserva recibida! Te contactaremos a la brevedad.</div>`)
	}
}
```

**Nota:** Si `core` no estaba en los imports, el bloque import de `handlers.go` debe quedar:
```go
import (
    "strings"

    "cms-plazareal/internal/config"

    "github.com/gofiber/fiber/v2"
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)
```

- [ ] **Step 4: Crear web/eventos.html**

Crear basándose en la estructura de `web/noticias.html` (mismo stack: Tailwind CDN, HTMX, Montserrat/Geist, design tokens Plaza Real). Diferencias clave:
- `<title>Eventos — Plaza Real Copiapó`
- Hero: texto "Eventos & Actividades del Mall"
- Sin barra de búsqueda/chips — solo la grilla de eventos
- Agregar el `<dialog>` para el modal de reserva
- El contenido se carga via HTMX on load

Contenido completo del archivo:

```html
<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Eventos — Plaza Real Copiapó</title>
  <meta name="description" content="Eventos, actividades y novedades de Plaza Real Copiapó. Reserva tu lugar en línea.">
  <link rel="icon" type="image/png" href="https://plazareal.cl/wp-content/uploads/2025/12/logo-mall-plaza-real-2024@3x.png">
  <link rel="apple-touch-icon" href="https://plazareal.cl/wp-content/uploads/2025/12/logo-mall-plaza-real-2024@3x.png">
  <script src="https://unpkg.com/htmx.org@1.9.12"></script>
  <script src="https://cdn.tailwindcss.com"></script>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link href="https://fonts.googleapis.com/css2?family=Montserrat:ital,wght@0,300;0,400;0,500;0,600;0,700;0,800;0,900;1,700;1,900&family=Geist:wght@300;400;500;600;700&display=swap" rel="stylesheet">
  <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200"/>
  <style>
    :root{
      --pr-red:#d71055;--pr-blue:#06a0e0;--pr-lime:#acc60d;
      --red:var(--pr-red);--red-d:#a80940;--blue:var(--pr-blue);--lime:var(--pr-lime);
      --dark:#0E0E0E;--dark2:#181818;--mid:#555;
      --surface:#F5F4F1;--surface2:#EBEBEA;
      --white:#FFF;--text:#0E0E0E;--muted:#6B6B6B;--border:#E0DFDC;
    }
    *{box-sizing:border-box;margin:0;padding:0}
    html{scroll-behavior:smooth}
    body{font-family:'Geist',system-ui,sans-serif;background:var(--surface);color:var(--text);overflow-x:hidden;-webkit-font-smoothing:antialiased}

    /* NAVBAR */
    .navbar{position:fixed;top:0;left:0;right:0;z-index:200;background:rgba(14,14,14,.97);backdrop-filter:blur(20px);border-bottom:1px solid rgba(255,255,255,.06);padding:0 32px;height:68px;display:flex;align-items:center;justify-content:space-between}
    .logo{display:flex;align-items:center;text-decoration:none}
    .logo img{height:42px;width:auto;display:block;filter:brightness(0)invert(1)}
    .nav-links{display:flex;gap:36px;align-items:center}
    .nav-links a{color:rgba(255,255,255,.65);text-decoration:none;font-size:.87rem;font-weight:500;transition:color .18s}
    .nav-links a:hover,.nav-links a.active{color:#fff}
    .nav-cta{background:var(--red)!important;color:#fff!important;padding:9px 22px;border-radius:100px;font-weight:700!important;font-size:.87rem!important}
    .nav-cta:hover{opacity:.88}
    .burger{display:none;background:none;border:none;cursor:pointer;flex-direction:column;gap:5px;padding:4px}
    .burger span{display:block;width:24px;height:2px;background:#fff;border-radius:2px}
    .mob-menu{display:none;position:fixed;top:68px;left:0;right:0;background:var(--dark2);z-index:199;padding:16px 24px 24px;flex-direction:column;gap:2px;border-bottom:1px solid rgba(255,255,255,.08)}
    .mob-menu a{color:rgba(255,255,255,.78);text-decoration:none;font-size:1rem;font-weight:500;padding:12px 0;border-bottom:1px solid rgba(255,255,255,.06);display:block}
    .mob-menu.open{display:flex}

    /* HERO */
    .page-hdr{background:linear-gradient(rgba(14,14,14,.75),rgba(14,14,14,.88)),url('https://plazareal.cl/wp-content/uploads/2025/12/plaza-real-hall-pantallas.jpeg');background-size:cover;background-position:center;background-attachment:fixed;padding:120px 24px 64px;position:relative;overflow:hidden}
    .page-hdr-inner{max-width:1200px;margin:0 auto;position:relative;z-index:1}
    .page-eye{font-size:.7rem;font-weight:700;letter-spacing:2.5px;text-transform:uppercase;color:var(--red);margin-bottom:8px}
    .page-h1{font-family:'Montserrat',sans-serif;font-size:clamp(2rem,5vw,3.4rem);font-weight:900;color:#fff;letter-spacing:-1px;margin-bottom:10px;line-height:1.05}
    .page-h1 em{font-style:italic;color:var(--blue)}
    .page-sub{color:rgba(255,255,255,.65);font-size:.97rem;max-width:620px;line-height:1.65;font-weight:300}

    /* SECTION */
    .container{max-width:1200px;margin:0 auto;padding:0 24px}
    .section{padding:56px 0 96px}

    /* EVENTOS GRID */
    .eventos-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(320px,1fr));gap:28px}
    .evento-card{background:var(--white);border-radius:16px;overflow:hidden;border:1.5px solid var(--border);transition:box-shadow .2s,transform .2s}
    .evento-card:hover{box-shadow:0 8px 32px rgba(0,0,0,.10);transform:translateY(-3px)}
    .evento-cover{width:100%;height:200px;object-fit:cover}
    .evento-cover-placeholder{width:100%;height:200px;background:linear-gradient(135deg,var(--surface2),var(--border));display:flex;align-items:center;justify-content:center}
    .evento-cover-placeholder .material-symbols-outlined{font-size:48px;color:var(--muted)}
    .evento-info{padding:20px 24px 24px}
    .evento-cat{font-size:.68rem;font-weight:700;letter-spacing:1.5px;text-transform:uppercase;color:var(--red);display:block;margin-bottom:8px}
    .evento-title{font-family:'Montserrat',sans-serif;font-size:1.15rem;font-weight:800;color:var(--text);margin-bottom:8px;line-height:1.25}
    .evento-date{display:flex;align-items:center;gap:6px;font-size:.82rem;color:var(--muted);margin-bottom:10px}
    .evento-date .material-symbols-outlined{font-size:16px}
    .evento-desc{font-size:.88rem;color:var(--mid);line-height:1.55;margin-bottom:16px}
    .btn-reservar{display:inline-flex;align-items:center;gap:8px;background:var(--red);color:#fff;border:none;padding:11px 22px;border-radius:100px;font-family:'Geist',sans-serif;font-size:.88rem;font-weight:700;cursor:pointer;transition:opacity .18s}
    .btn-reservar:hover{opacity:.86}
    .btn-reservar .material-symbols-outlined{font-size:18px}

    /* EMPTY STATE */
    .empty-state{display:flex;flex-direction:column;align-items:center;gap:12px;padding:80px 24px;color:var(--muted)}
    .empty-state .material-symbols-outlined{font-size:56px;color:var(--border)}
    .empty-state p{font-size:1rem}

    /* MODAL */
    dialog#reserva-modal{border:none;border-radius:20px;padding:0;width:min(540px,95vw);box-shadow:0 24px 80px rgba(0,0,0,.28)}
    dialog#reserva-modal::backdrop{background:rgba(0,0,0,.55);backdrop-filter:blur(4px)}
    dialog#reserva-modal>div{padding:32px}
    .modal-title{font-family:'Montserrat',sans-serif;font-size:1.3rem;font-weight:800;margin-bottom:4px}
    .modal-subtitle{font-size:.9rem;color:var(--muted);margin-bottom:20px}
    .form-group{margin-bottom:16px;display:flex;flex-direction:column;gap:6px}
    .form-group label{font-size:.83rem;font-weight:600;color:var(--mid)}
    .form-group input,.form-group textarea{padding:10px 14px;border:1.5px solid var(--border);border-radius:10px;font-family:'Geist',sans-serif;font-size:.93rem;outline:none;transition:border-color .18s;width:100%}
    .form-group input:focus,.form-group textarea:focus{border-color:var(--red)}
    .form-group textarea{resize:vertical;min-height:80px}
    .form-actions{display:flex;gap:12px;margin-top:20px;flex-wrap:wrap}
    .btn-primary{display:inline-flex;align-items:center;gap:8px;background:var(--red);color:#fff;border:none;padding:12px 24px;border-radius:100px;font-family:'Geist',sans-serif;font-size:.9rem;font-weight:700;cursor:pointer;transition:opacity .18s}
    .btn-primary:hover{opacity:.86}
    .btn-outline{background:transparent;color:var(--mid);border:1.5px solid var(--border);padding:12px 24px;border-radius:100px;font-family:'Geist',sans-serif;font-size:.9rem;font-weight:600;cursor:pointer;transition:all .18s}
    .btn-outline:hover{border-color:var(--mid)}
    .form-error{color:#c0392b;font-size:.85rem;padding:10px 14px;background:#fdecea;border-radius:8px;margin-top:8px}
    .form-success{display:flex;align-items:center;gap:8px;color:#1a7f52;font-size:.9rem;font-weight:600;padding:12px 16px;background:#e8f8f0;border-radius:10px;margin-top:8px}
    .form-success .material-symbols-outlined{font-size:20px}

    /* FOOTER */
    footer{background:var(--dark);color:rgba(255,255,255,.5);padding:32px 24px;text-align:center;font-size:.82rem}

    /* RESPONSIVE */
    @media(max-width:768px){
      .nav-links{display:none}
      .burger{display:flex}
      .eventos-grid{grid-template-columns:1fr}
    }
  </style>
</head>
<body>

<!-- NAVBAR -->
<nav class="navbar">
  <a href="/index.html" class="logo">
    <img src="https://plazareal.cl/wp-content/uploads/2025/12/logo-mall-plaza-real-2024@3x.png" alt="Plaza Real"/>
  </a>
  <div class="nav-links">
    <a href="/index.html">Inicio</a>
    <a href="/buscador-tiendas.html">Tiendas</a>
    <a href="/locales.html">Locales Disponibles</a>
    <a href="/eventos.html" class="active">Eventos</a>
    <a href="/noticias.html">Noticias</a>
    <a href="/admin" class="nav-cta">Admin</a>
  </div>
  <button class="burger" onclick="document.querySelector('.mob-menu').classList.toggle('open')" aria-label="Menú">
    <span></span><span></span><span></span>
  </button>
</nav>
<div class="mob-menu">
  <a href="/index.html">Inicio</a>
  <a href="/buscador-tiendas.html">Tiendas</a>
  <a href="/locales.html">Locales Disponibles</a>
  <a href="/eventos.html">Eventos</a>
  <a href="/noticias.html">Noticias</a>
  <a href="/admin">Admin</a>
</div>

<!-- HERO -->
<div class="page-hdr">
  <div class="page-hdr-inner">
    <div class="page-eye">Plaza Real · Copiapó</div>
    <h1 class="page-h1">Eventos &amp; <em>Actividades</em></h1>
    <p class="page-sub">Descubre lo que viene en el mall y reserva tu lugar en línea.</p>
  </div>
</div>

<!-- CONTENIDO -->
<div class="section">
  <div class="container">
    <div id="eventos-results"
         hx-get="/fragments/eventos-public"
         hx-trigger="load"
         hx-swap="innerHTML">
      <div style="display:flex;align-items:center;gap:10px;color:var(--muted);padding:40px 0">
        <span class="material-symbols-outlined" style="animation:spin 1s linear infinite">progress_activity</span>
        Cargando eventos...
      </div>
    </div>
  </div>
</div>

<!-- MODAL DE RESERVA -->
<dialog id="reserva-modal">
  <div id="reserva-modal-inner">
    <!-- El formulario se inyecta aquí vía HTMX al hacer click en "Reservar lugar" -->
  </div>
</dialog>

<!-- FOOTER -->
<footer>
  <p>&copy; 2026 Plaza Real Copiapó. Todos los derechos reservados.</p>
</footer>

<style>
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
<script>
  // Cierra el modal al hacer clic fuera del contenido
  document.getElementById('reserva-modal').addEventListener('click', function(e) {
    if (e.target === this) this.close();
  });
</script>
</body>
</html>
```

- [ ] **Step 5: Agregar admin handlers de reservas en internal/handlers/admin/handlers.go**

Agregar al final del archivo:

```go
func ReservasList(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter("reservas", "status != ''", "-created", 200, 0)
		type row struct {
			ID, EventoID, Nombre, Email, Telefono, Personas, Status string
		}
		var rows []row
		for _, r := range records {
			rows = append(rows, row{
				ID:       r.Id,
				EventoID: r.GetString("evento_id"),
				Nombre:   r.GetString("nombre"),
				Email:    r.GetString("email"),
				Telefono: r.GetString("telefono"),
				Personas: fmt.Sprintf("%d", r.GetInt("cantidad_personas")),
				Status:   r.GetString("status"),
			})
		}
		tmpl, err := template.ParseFiles("./internal/templates/admin/pages/reservas.html")
		if err != nil {
			return c.Status(500).SendString("Template error: " + err.Error())
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.Execute(c, map[string]any{"Reservas": rows})
	}
}

func ReservaConfirm(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("reservas", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Reserva no encontrada</div>`)
		}
		r.Set("status", "confirmada")
		r.Set("leido", true)
		if err := pb.Save(r); err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error al confirmar</div>`)
		}
		c.Set("HX-Redirect", "/admin/reservas")
		return c.SendStatus(200)
	}
}

func ReservaCancel(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("reservas", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Reserva no encontrada</div>`)
		}
		r.Set("status", "cancelada")
		if err := pb.Save(r); err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error al cancelar</div>`)
		}
		c.Set("HX-Redirect", "/admin/reservas")
		return c.SendStatus(200)
	}
}

func ReservaDelete(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r, err := pb.FindRecordById("reservas", c.Params("id"))
		if err != nil {
			return c.Status(404).SendString(`<div class="toast toast-error">Reserva no encontrada</div>`)
		}
		if err := pb.Delete(r); err != nil {
			return c.Status(500).SendString(`<div class="toast toast-error">Error al eliminar</div>`)
		}
		return c.SendStatus(200)
	}
}
```

**Verificar que `fmt` y `template` están en los imports de handlers.go:**
```bash
grep -n '"fmt"\|"html/template"' internal/handlers/admin/handlers.go
```
Si alguno falta, agregarlo al bloque `import (`.

- [ ] **Step 6: Crear internal/templates/admin/pages/reservas.html**

```html
<header class="topbar">
  <h1 class="topbar-title">Reservas</h1>
  <div class="topbar-actions">
    <a href="/eventos.html" target="_blank" class="topbar-btn topbar-btn-outline">
      <span class="material-symbols-outlined" style="font-size:16px">open_in_new</span> Ver eventos
    </a>
  </div>
</header>

<div class="content">
  <div class="card">
    <div class="card-header">
      <h2 class="card-title">Listado de reservas</h2>
    </div>
    <div style="overflow-x:auto">
      <table>
        <thead>
          <tr>
            <th>Nombre</th>
            <th>Email</th>
            <th>Teléfono</th>
            <th>Evento ID</th>
            <th>Personas</th>
            <th>Estado</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {{range .Reservas}}
          <tr>
            <td>{{.Nombre}}</td>
            <td>{{.Email}}</td>
            <td>{{.Telefono}}</td>
            <td style="font-size:11px;color:var(--md-outline)">{{.EventoID}}</td>
            <td>{{.Personas}}</td>
            <td>
              <span class="badge badge-{{.Status}}">{{.Status}}</span>
            </td>
            <td>
              {{if eq .Status "pendiente"}}
              <button class="btn-icon"
                hx-post="/admin/reservas/{{.ID}}/confirm"
                hx-confirm="¿Confirmar reserva de {{.Nombre}}?"
                title="Confirmar">
                <span class="material-symbols-outlined">check_circle</span>
              </button>
              <button class="btn-icon"
                hx-post="/admin/reservas/{{.ID}}/cancel"
                hx-confirm="¿Cancelar reserva de {{.Nombre}}?"
                title="Cancelar">
                <span class="material-symbols-outlined">cancel</span>
              </button>
              {{end}}
              <button class="btn-icon btn-danger"
                hx-delete="/admin/reservas/{{.ID}}"
                hx-confirm="¿Eliminar reserva de {{.Nombre}}?"
                hx-target="closest tr"
                hx-swap="outerHTML swap:0.3s"
                title="Eliminar">
                <span class="material-symbols-outlined">delete</span>
              </button>
            </td>
          </tr>
          {{else}}
          <tr><td colspan="7" class="empty-state-cell">No hay reservas aún.</td></tr>
          {{end}}
        </tbody>
      </table>
    </div>
  </div>
</div>
```

- [ ] **Step 7: Agregar link Reservas al sidebar del admin**

Buscar el sidebar en el layout del admin:
```bash
grep -rn "sidebar-link\|event\|Eventos" internal/templates/admin/ | grep -v ".html:$" | head -20
```

Localizar el bloque del sidebar (normalmente en `internal/templates/admin/layout.html` o similar). Agregar después del link de Noticias/Events:

```html
<a href="/admin/reservas" class="sidebar-link {{if eq .ActivePage "reservas"}}active{{end}}">
  <span class="material-symbols-outlined">event_seat</span>
  Reservas
</a>
```

- [ ] **Step 8: Actualizar navegación en web/index.html y web/noticias.html**

Verificar que la nav incluye todos los tabs. Buscar el nav en cada página:
```bash
grep -n "nav-links\|buscador-tiendas\|locales\|eventos\|noticias" web/index.html | head -10
grep -n "nav-links\|buscador-tiendas\|locales\|eventos\|noticias" web/noticias.html | head -10
```

La nav debe incluir (en ese orden):
```html
<a href="/index.html">Inicio</a>
<a href="/buscador-tiendas.html">Tiendas</a>
<a href="/locales.html">Locales Disponibles</a>
<a href="/eventos.html">Eventos</a>
<a href="/noticias.html">Noticias</a>
<a href="/admin" class="nav-cta">Admin</a>
```

Hacer el mismo cambio en el mob-menu (versión móvil) de cada archivo. Archivos a revisar: `web/index.html`, `web/noticias.html`, `web/buscador-tiendas.html`, `web/tienda-individual.html`.

- [ ] **Step 9: Registrar rutas en cmd/server/main.go**

Buscar dónde están los grupos de rutas:
```bash
grep -n 'frag :=\|api :=\|adm :=' cmd/server/main.go
```

Agregar las nuevas rutas en cada grupo:

```go
// Fragmentos públicos de eventos y reservas
frag.Get("/eventos-public", fragments.EventosPublic(cfg, pb))
frag.Get("/reserva-form", fragments.ReservaForm(cfg))

// API de reservas
api.Post("/reservas", apiHandlers.CreateReserva(cfg, pb))

// Admin de reservas
adm.Get("/reservas", middleware.RoleRequired("superadmin", "director", "admin"), admin.ReservasList(cfg, pb))
adm.Post("/reservas/:id/confirm", middleware.RoleRequired("superadmin", "director", "admin"), admin.ReservaConfirm(cfg, pb))
adm.Post("/reservas/:id/cancel", middleware.RoleRequired("superadmin", "director", "admin"), admin.ReservaCancel(cfg, pb))
adm.Delete("/reservas/:id", middleware.RoleRequired("superadmin", "director"), admin.ReservaDelete(cfg, pb))
```

También agregar la ruta pública de la página:
```go
app.Get("/eventos.html", func(c *fiber.Ctx) error {
    return c.SendFile("./web/eventos.html")
})
```

Verificar cómo sirve noticias.html actualmente para usar el mismo patrón:
```bash
grep -n "noticias.html\|SendFile\|web.PageHandler" cmd/server/main.go | head -5
```

- [ ] **Step 10: Compilar**

```bash
go build ./...
# Esperado: sin errores
```

Si hay errores de imports no resueltos, revisar:
- `fragments.EventosPublic` — verificar que el package es `fragments`
- `apiHandlers.CreateReserva` — verificar el alias del package api en main.go:
  ```bash
  grep "apiHandlers\|handlers/api" cmd/server/main.go | head -3
  ```

- [ ] **Step 11: Test manual**

```bash
go run cmd/server/main.go
```

1. Abrir `http://localhost:3000/eventos.html` — debe cargar la grilla (o empty state si no hay eventos)
2. Si hay eventos: hacer click en "Reservar lugar" → debe abrir el `<dialog>` con el formulario
3. Completar nombre + email → Submit → debe mostrar mensaje de éxito verde
4. Verificar en `/admin/reservas` que aparece la reserva con status "pendiente"
5. Click "Confirmar" → status debe cambiar a "confirmada"

- [ ] **Step 12: Commit**

```bash
git add internal/auth/collections.go \
        web/eventos.html \
        internal/handlers/fragments/eventos_pub.go \
        internal/handlers/api/handlers.go \
        internal/handlers/admin/handlers.go \
        internal/templates/admin/pages/reservas.html \
        cmd/server/main.go \
        web/index.html web/noticias.html web/buscador-tiendas.html web/tienda-individual.html
git commit -m "feat: add eventos con reservas — public page, HTMX dialog form, admin CRUD"
```

---

## Estado del repositorio al finalizar este task

- Colección `reservas` existe en PocketBase
- `web/eventos.html` accesible en `/eventos.html`
- Formulario de reserva abre como `<dialog>` nativo vía HTMX
- `POST /api/reservas` persiste reservas en PocketBase
- `/admin/reservas` lista, confirma, cancela y elimina reservas
- Navegación pública incluye todos los tabs (Tiendas, Locales, Eventos, Noticias)
- `go build ./...` pasa sin errores
