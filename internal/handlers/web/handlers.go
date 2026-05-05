package web

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"cms-plazareal/internal/config"
	"cms-plazareal/internal/helpers"
	"cms-plazareal/internal/view/pages/public"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
)

type noticiaData struct {
	Title    string
	Date     string
	Category string
	ImgHTML  template.HTML
	BodyHTML template.HTML
}

// IndexHandler renders the public home page using the templ Index() component.
func IndexHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Index())
	}
}

// TiendasPageHandler renders the public Tiendas index page.
func TiendasPageHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Tiendas())
	}
}

// NoticiasPageHandler renders the public Noticias page.
func NoticiasPageHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Noticias())
	}
}

// ComunicadosPageHandler renders the public Comunicados page.
func ComunicadosPageHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Comunicados())
	}
}

// LocalesPageHandler renders the public Locales Disponibles page with
// server-side data (only locales with estado='disponible').
func LocalesPageHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter(
			"locales_disponibles",
			"estado = 'disponible'",
			"galeria,numero,nombre",
			100, 0,
		)
		items := make([]public.LocalPublico, 0, len(records))
		for _, r := range records {
			items = append(items, public.LocalPublico{
				ID:          r.Id,
				Nombre:      r.GetString("nombre"),
				Galeria:     r.GetString("galeria"),
				Numero:      r.GetString("numero"),
				Piso:        r.GetString("piso"),
				Descripcion: r.GetString("descripcion"),
				PrecioRef:   r.GetString("precio_ref"),
				ImagenURL:   r.GetString("imagen_url"),
				M2:          r.GetFloat("m2"),
			})
		}
		return helpers.Render(c, public.LocalesPublicPage(public.LocalesPublicData{Locales: items}))
	}
}

// PromocionesPageHandler renders the public Eventos y Promociones page.
func PromocionesPageHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, public.Promociones())
	}
}

// TiendaDetailHandler renders a single tienda detail page server-side.
func TiendaDetailHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		slug := sanitizeSlug(c.Params("slug"))
		if slug == "" {
			return c.Redirect("/buscador-tiendas", fiber.StatusFound)
		}

		records, err := pb.FindRecordsByFilter(
			"tiendas",
			"(slug = '"+slug+"' || id = '"+slug+"') && status = 'publicado'",
			"",
			1, 0,
		)
		if err != nil || len(records) == 0 {
			return c.Status(fiber.StatusNotFound).SendString("Tienda no encontrada")
		}
		r := records[0]

		data := public.TiendaDetailData{
			ID:         r.Id,
			Nombre:     r.GetString("nombre"),
			Cat:        r.GetString("cat"),
			Gal:        r.GetString("gal"),
			Local:      r.GetString("local"),
			Logo:       r.GetString("logo"),
			Tags:       splitCSV(r.GetString("tags")),
			Desc:       r.GetString("desc"),
			About:      r.GetString("about"),
			About2:     r.GetString("about2"),
			Pay:        r.GetString("pay"),
			Photos:     splitCSV(r.GetString("photos")),
			Whatsapp:   r.GetString("whatsapp"),
			Telefono:   r.GetString("telefono"),
			HorarioLV:  r.GetString("horario_lv"),
			HorarioSab: r.GetString("horario_sab"),
			HorarioDom: r.GetString("horario_dom"),
			HeroBg:     r.GetString("hero_bg"),
		}

		isOpen, statusLabel := computeOpenStatus(
			data.HorarioLV, data.HorarioSab, data.HorarioDom,
			r.GetString("status_horario"),
		)
		data.IsOpen = isOpen
		data.StatusLabel = statusLabel

		// Similar tiendas
		similarSlugs := splitCSV(r.GetString("similar"))
		if len(similarSlugs) > 0 {
			parts := make([]string, 0, len(similarSlugs))
			for _, s := range similarSlugs {
				parts = append(parts, "'"+sanitizeSlug(s)+"'")
			}
			simRecs, simErr := pb.FindRecordsByFilter(
				"tiendas",
				"slug ?| ["+strings.Join(parts, ",")+"] && status = 'publicado'",
				"nombre", 4, 0,
			)
			if simErr == nil {
				for _, sr := range simRecs {
					data.Similar = append(data.Similar, public.SimilarTienda{
						Slug:   sr.GetString("slug"),
						Nombre: sr.GetString("nombre"),
						Logo:   sr.GetString("logo"),
						Cat:    sr.GetString("cat"),
					})
				}
			}
		}

		return helpers.Render(c, public.TiendaDetail(data))
	}
}

// PageHandler serves static sub-pages (kept for any callers still using it).
func PageHandler(cfg *config.Config, page string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile(fmt.Sprintf("./web/%s.html", page))
	}
}

// DeviceDisplay serves the kiosk mode display for a horizontal/screen device
func DeviceDisplay(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Params("code")
		if code == "" {
			return c.Status(404).SendString("Código de dispositivo inválido")
		}
		return c.SendFile("./internal/templates/devices/display.html")
	}
}

// TotemDisplay serves the vertical totem kiosk for a totem device
func TotemDisplay(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := c.Params("code")
		if code == "" {
			return c.Status(404).SendString("Código de dispositivo inválido")
		}
		return c.SendFile("./internal/templates/devices/totem.html")
	}
}

// RSSFeed generates an RSS feed from published events and news
func RSSFeed(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Fetch from PocketBase events + news_articles where status = 'publicado'
		now := time.Now().Format(time.RFC1123Z)

		rss := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Plaza Real — Noticias y Comunicados</title>
    <link>%s</link>
    <description>Noticias, comunicados y promociones de Plaza Real Copiapó</description>
    <language>es-cl</language>
    <lastBuildDate>%s</lastBuildDate>
    <atom:link href="%s/rss.xml" rel="self" type="application/rss+xml"/>
    <!-- Items from PocketBase -->
    <item>
      <title>Plaza Real Copiapó</title>
      <link>%s/comunicados.html</link>
      <description>Noticias y actividades del centro comercial.</description>
      <pubDate>%s</pubDate>
      <guid>%s/events/1</guid>
    </item>
  </channel>
</rss>`, cfg.BaseURL, now, cfg.BaseURL, cfg.BaseURL, now, cfg.BaseURL)

		c.Set("Content-Type", "application/rss+xml; charset=utf-8")
		return c.SendString(rss)
	}
}

// WhatsAppWebhook handles inbound WhatsApp messages
func WhatsAppWebhook(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendString("<Response></Response>")
	}
}

// NoticiaHandler renders a single news article using the full site layout template.
func NoticiaHandler(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		r, err := pb.FindRecordById("content_blocks", id)
		if err != nil || r.GetString("status") != "publicado" {
			return c.Redirect("/", fiber.StatusFound)
		}
		cat := r.GetString("category")
		if cat != "NOTICIA" && cat != "AVISO" && cat != "PUBLICIDAD" {
			return c.Redirect("/", fiber.StatusFound)
		}
		catLabel := map[string]string{
			"NOTICIA":    "Noticia",
			"AVISO":      "Aviso",
			"PUBLICIDAD": "Publicidad",
		}[cat]

		title := r.GetString("title")
		desc := r.GetString("description")
		body := r.GetString("body")
		imageURL := r.GetString("image_url")

		dateStr := ""
		if dt := r.GetDateTime("date"); !dt.IsZero() {
			dateStr = dt.Time().Format("2 de January de 2006")
		}

		var imgHTML template.HTML
		if imageURL != "" {
			imgHTML = template.HTML(fmt.Sprintf(
				`<div style="width:100%%;aspect-ratio:16/6;border-radius:18px;margin-bottom:40px;overflow:hidden"><img src="%s" style="width:100%%;height:100%%;object-fit:cover" alt="%s"/></div>`,
				template.HTMLEscapeString(imageURL), template.HTMLEscapeString(title)))
		} else {
			imgHTML = `<div style="width:100%;aspect-ratio:16/6;background:linear-gradient(135deg,#d60d5222,#00a0e322);border-radius:18px;margin-bottom:40px;display:flex;align-items:center;justify-content:center;font-size:3.5rem">📰</div>`
		}

		src := body
		if src == "" {
			src = desc
		}
		var bodyParts []string
		for _, p := range strings.Split(strings.TrimSpace(src), "\n\n") {
			p = strings.TrimSpace(p)
			if p != "" {
				bodyParts = append(bodyParts, "<p>"+template.HTMLEscapeString(p)+"</p>")
			}
		}
		bodyHTML := template.HTML(strings.Join(bodyParts, "\n"))

		tmpl, err2 := template.ParseFiles("./internal/templates/web/noticia.html")
		if err2 != nil {
			return c.Status(500).SendString("Template error")
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return tmpl.ExecuteTemplate(c, "noticia.html", noticiaData{
			Title:    title,
			Date:     dateStr,
			Category: catLabel,
			ImgHTML:  imgHTML,
			BodyHTML: bodyHTML,
		})
	}
}

// ── helpers ────────────────────────────────────────────────────────────────

// splitCSV trims and returns non-empty comma-separated tokens.
func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// sanitizeSlug strips characters that would break a PocketBase filter expression.
// Keeps alphanumerics, dashes and underscores; everything else is dropped.
func sanitizeSlug(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_':
			b.WriteRune(r)
		}
	}
	return b.String()
}

// computeOpenStatus returns whether the store is currently open and a label.
// Uses the day-specific schedule string ("HH:MM - HH:MM") and an optional
// override stored in `status_horario` ("solo-reserva" | "cerrado-temporal" | "normal").
func computeOpenStatus(lv, sab, dom, statusOverride string) (bool, string) {
	if statusOverride == "cerrado-temporal" {
		return false, "Cerrado temporalmente"
	}
	if statusOverride == "solo-reserva" {
		return false, "Solo con reserva"
	}
	now := time.Now()
	var horario string
	switch now.Weekday() {
	case time.Saturday:
		horario = sab
	case time.Sunday:
		horario = dom
	default:
		horario = lv
	}
	if horario == "" {
		return false, "Sin horario"
	}
	// Normalize separators: support "HH:MM - HH:MM" and "HH:MM – HH:MM" (en dash).
	norm := strings.ReplaceAll(horario, "–", "-")
	parts := strings.SplitN(norm, " - ", 2)
	if len(parts) != 2 {
		return false, "Sin horario"
	}
	parseHM := func(s string) (int, int, bool) {
		var h, m int
		_, err := fmt.Sscanf(strings.TrimSpace(s), "%d:%02d", &h, &m)
		return h, m, err == nil
	}
	oh, om, ok1 := parseHM(parts[0])
	ch, cm, ok2 := parseHM(parts[1])
	if !ok1 || !ok2 {
		return false, "Sin horario"
	}
	nowMin := now.Hour()*60 + now.Minute()
	openMin := oh*60 + om
	closeMin := ch*60 + cm
	if nowMin >= openMin && nowMin < closeMin {
		return true, "Abierto ahora"
	}
	return false, "Cerrado"
}
