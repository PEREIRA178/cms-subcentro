package fragments

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"cms-plazareal/internal/config"
	"cms-plazareal/internal/helpers"
	fragmentsView "cms-plazareal/internal/view/fragments"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
)

// GET /fragments/hero — renders the public landing hero using the
// Liquid Glass design system (page-hero classes from app.css).
// The optional background image is sourced from site_settings.hero_bg_url.
func HeroCarousel(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return helpers.Render(c, fragmentsView.Hero(fragmentsView.HeroData{
			BgURL: helpers.GetSetting(pb, "hero_bg_url"),
		}))
	}
}

// cssClassForCategory maps content_block category to CSS class
func cssClassForCategory(category string) string {
	switch category {
	case "REUNIÓN":
		return "ev-reunion"
	case "EVENTO":
		return "ev-evento"
	case "DEPORTIVO":
		return "ev-deportivo"
	case "ACADÉMICO":
		return "ev-academico"
	default:
		return "ev-info"
	}
}

type contentBlock struct {
	ID       string
	Title    string
	Desc     string
	Category string
	Date     string
	Featured bool
	CSSClass string
	PdfURL   string
	ImageURL string
}

func fetchContentBlocks(pb *pocketbase.PocketBase, filter string, limit int) []contentBlock {
	records, err := pb.FindRecordsByFilter("content_blocks", filter, "-date", limit, 0)
	if err != nil || len(records) == 0 {
		return nil
	}
	result := make([]contentBlock, 0, len(records))
	for _, r := range records {
		category := r.GetString("category")
		dateStr := ""
		if dt := r.GetDateTime("date"); !dt.IsZero() {
			dateStr = dt.Time().Format("2 Jan 2006")
		}
		result = append(result, contentBlock{
			ID:       r.Id,
			Title:    r.GetString("title"),
			Desc:     r.GetString("description"),
			Category: category,
			Date:     dateStr,
			CSSClass: cssClassForCategory(category),
			PdfURL:   r.GetString("pdf_url"),
			ImageURL: r.GetString("image_url"),
		})
	}
	return result
}

// GET /fragments/eventos
func Eventos(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		blocks := fetchContentBlocks(pb,
			"status = 'publicado' && category != 'NOTICIA'", 6)

		if len(blocks) == 0 {
			blocks = []contentBlock{
				{ID: "1", Title: "Cierre temporal del estacionamiento norte", Desc: "Por mantenimiento programado, el estacionamiento norte permanecerá cerrado durante el fin de semana.", Category: "AVISO", Date: "5 may 2026", CSSClass: "ev-info"},
				{ID: "2", Title: "Horarios extendidos por temporada navideña", Desc: "Plaza Real abrirá hasta las 23:00 hrs durante todo el mes de diciembre.", Category: "INFORMACIÓN", Date: "1 dic 2026", CSSClass: "ev-info"},
				{ID: "3", Title: "Nueva tienda inaugurada en Placa Comercial", Desc: "Visita la nueva apertura del primer piso, con promociones exclusivas durante toda la semana.", Category: "EVENTO", Date: "10 may 2026", CSSClass: "ev-info"},
			}
		}

		var sb strings.Builder
		sb.WriteString(`<section class="eventos-section" id="eventos">`)
		sb.WriteString(`<div class="eventos-header-row">`)
		sb.WriteString(`<div><p class="label-primary reveal visible">Comunicados y avisos</p>`)
		sb.WriteString(`<h2 class="headline-l reveal visible" style="margin-bottom:0">Últimos eventos</h2></div>`)
		sb.WriteString(`<a href="comunicados.html" class="eventos-link reveal visible">Ver todos los comunicados →</a>`)
		sb.WriteString(`</div><div class="eventos-grid">`)

		for i, ev := range blocks {
			footerRight := ""
			if ev.PdfURL != "" {
				footerRight = fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener" style="font-size:12px;font-weight:500;color:var(--md-primary);text-decoration:none">Ver comunicado →</a>`,
					template.HTMLEscapeString(ev.PdfURL))
			}
			sb.WriteString(fmt.Sprintf(`
    <div class="evento-card %s reveal visible%s">
      <div class="evento-accent"></div>
      <div class="evento-body">
        <span class="evento-chip">%s</span>
        <h3>%s</h3>
        <p>%s</p>
      </div>
      <div class="evento-footer">
        <span class="evento-fecha"><span class="evento-fecha-dot"></span>%s</span>
        %s
      </div>
    </div>`,
				ev.CSSClass, delayClass(i),
				template.HTMLEscapeString(ev.Category),
				template.HTMLEscapeString(ev.Title),
				template.HTMLEscapeString(ev.Desc),
				ev.Date, footerRight,
			))
		}

		sb.WriteString(`</div></section>`)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// GET /fragments/noticias
func Noticias(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		blocks := fetchContentBlocks(pb,
			"status = 'publicado' && (category = 'NOTICIA' || category = 'AVISO' || category = 'PUBLICIDAD')", 3)

		type noticia struct{ ID, Title, Excerpt, Category, Date, ImageURL string }
		var noticias []noticia

		for _, b := range blocks {
			noticias = append(noticias, noticia{
				ID: b.ID, Title: b.Title, Excerpt: b.Desc,
				Category: b.Category, Date: b.Date, ImageURL: b.ImageURL,
			})
		}

		var sb strings.Builder
		if len(noticias) == 0 {
			sb.WriteString(`<div style="grid-column:1/-1;text-align:center;padding:48px 0;color:#6B6B6B">Aún no hay comunicados publicados.</div>`)
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.SendString(sb.String())
		}

		for i, n := range noticias {
			thumbInner := `<div style="width:100%;height:100%;background:linear-gradient(135deg,#d60d5222,#00a0e322);display:flex;align-items:center;justify-content:center;font-size:2rem">📰</div>`
			if n.ImageURL != "" {
				thumbInner = fmt.Sprintf(`<img src="%s" alt="%s" loading="lazy" style="width:100%%;height:100%%;object-fit:cover;display:block">`,
					template.HTMLEscapeString(n.ImageURL), template.HTMLEscapeString(n.Title))
			}
			leerHref := "noticias.html"
			if n.ID != "" {
				leerHref = "/noticias/" + n.ID
			}
			extraClass := ""
			if i == 0 {
				extraClass = " n-big"
			}
			sb.WriteString(fmt.Sprintf(`<a href="%s" class="n-card%s">
  <div class="n-thumb">%s<span class="n-cat">%s</span></div>
  <div class="n-body">
    <h3 class="n-title">%s</h3>
    <p class="n-excerpt">%s</p>
  </div>
  <div class="n-foot"><span class="n-date">%s</span><span class="n-arr">Leer más →</span></div>
</a>`,
				template.HTMLEscapeString(leerHref),
				extraClass,
				thumbInner,
				template.HTMLEscapeString(n.Category),
				template.HTMLEscapeString(n.Title),
				template.HTMLEscapeString(n.Excerpt),
				template.HTMLEscapeString(n.Date),
			))
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// GET /fragments/comunicados — full grid for /comunicados.html
func Comunicados(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter := c.Query("filter", "todos")
		q := strings.TrimSpace(c.Query("q", ""))
		page := c.QueryInt("page", 1)
		if page < 1 {
			page = 1
		}
		const pageSize = 12

		pbFilter := "status = 'publicado' && category != 'NOTICIA'"
		switch filter {
		case "urgente":
			pbFilter += " && category = 'EMERGENCIA'"
		case "reunion":
			pbFilter += " && category = 'REUNIÓN'"
		case "info":
			pbFilter += " && category = 'INFORMACIÓN'"
		case "academico":
			pbFilter += " && category = 'ACADÉMICO'"
		case "evento":
			pbFilter += " && category = 'EVENTO'"
		case "deportivo":
			pbFilter += " && category = 'DEPORTIVO'"
		}
		if q != "" {
			safeQ := strings.ReplaceAll(strings.ReplaceAll(q, "'", ""), "\\", "")
			pbFilter += fmt.Sprintf(" && (title ~ '%s' || description ~ '%s')", safeQ, safeQ)
		}

		offset := (page - 1) * pageSize
		records, err := pb.FindRecordsByFilter("content_blocks", pbFilter, "-date", pageSize+1, offset)

		var blocks []contentBlock
		if err == nil {
			for _, r := range records {
				cat := r.GetString("category")
				dateStr := ""
				if dt := r.GetDateTime("date"); !dt.IsZero() {
					dateStr = dt.Time().Format("2 Jan 2006")
				}
				blocks = append(blocks, contentBlock{
					ID:       r.Id,
					Title:    r.GetString("title"),
					Desc:     r.GetString("description"),
					Category: cat,
					Date:     dateStr,
					CSSClass: cssClassForCategory(cat),
					PdfURL:   r.GetString("pdf_url"),
					ImageURL: r.GetString("image_url"),
				})
			}
		}

		hasMore := len(blocks) > pageSize
		if hasMore {
			blocks = blocks[:pageSize]
		}

		// Empty state for specific filter (not fallback)
		if len(blocks) == 0 && filter != "todos" {
			c.Set("Content-Type", "text/html; charset=utf-8")
			if page == 1 {
				return c.SendString(`<div class="comunicados-grid" id="comunicados-grid"><div style="grid-column:1/-1;text-align:center;padding:56px 24px;color:var(--md-outline)"><p style="font-size:15px">No hay comunicados en esta categoría.</p></div></div><div id="load-more-wrap"></div>`)
			}
			return c.SendString(`<div id="load-more-wrap" hx-swap-oob="true"></div>`)
		}

		// Empty state for search
		if len(blocks) == 0 && q != "" {
			c.Set("Content-Type", "text/html; charset=utf-8")
			if page == 1 {
				return c.SendString(`<div class="comunicados-grid" id="comunicados-grid"><div style="grid-column:1/-1;text-align:center;padding:56px 24px;color:var(--md-outline)"><p style="font-size:15px">Sin resultados para esa búsqueda.</p></div></div><div id="load-more-wrap"></div>`)
			}
			return c.SendString(`<div id="load-more-wrap" hx-swap-oob="true"></div>`)
		}

		// Fallback only for "todos" filter with no search
		if len(blocks) == 0 {
			blocks = hardcodedComunicados()
			hasMore = false
		}

		// Build next-page URL
		nextURL := func(p int) string {
			v := url.Values{}
			v.Set("filter", filter)
			v.Set("page", fmt.Sprintf("%d", p))
			if q != "" {
				v.Set("q", q)
			}
			return "/fragments/comunicados?" + v.Encode()
		}

		loadMoreBtn := func(nextPage int) string {
			return fmt.Sprintf(`<div id="load-more-wrap" style="text-align:center;padding:24px 0 8px">
  <button class="filtro-chip" style="padding:12px 28px;font-size:13px"
          hx-get="%s"
          hx-target="#comunicados-grid"
          hx-swap="beforeend">
    Cargar más
  </button>
</div>`, nextURL(nextPage))
		}

		var sb strings.Builder

		if page == 1 {
			sb.WriteString(`<div class="comunicados-grid" id="comunicados-grid">`)
		}

		for i, b := range blocks {
			_, chip, accentClass := categoryToTipo(b.Category)
			verMasBtn := ""
			if b.PdfURL != "" {
				verMasBtn = fmt.Sprintf(`<a href="%s" class="comunicado-btn" target="_blank" rel="noopener">Ver comunicado →</a>`,
					template.HTMLEscapeString(b.PdfURL))
			}
			sb.WriteString(fmt.Sprintf(`
      <div class="comunicado-card %s reveal visible%s">
        <div class="comunicado-accent"></div>
        <div class="comunicado-body">
          <span class="comunicado-chip">%s</span>
          <h3>%s</h3>
          <p>%s</p>
        </div>
        <div class="comunicado-footer">
          <span class="comunicado-fecha">%s</span>
          %s
        </div>
      </div>`,
				accentClass, delayClass(i%4),
				chip,
				template.HTMLEscapeString(b.Title),
				template.HTMLEscapeString(b.Desc),
				b.Date,
				verMasBtn,
			))
		}

		if page == 1 {
			sb.WriteString(`</div>`)
			if hasMore {
				sb.WriteString(loadMoreBtn(page + 1))
			} else {
				sb.WriteString(`<div id="load-more-wrap"></div>`)
			}
		} else {
			// OOB update for load-more-wrap
			if hasMore {
				sb.WriteString(fmt.Sprintf(`<div id="load-more-wrap" hx-swap-oob="true" style="text-align:center;padding:24px 0 8px">
  <button class="filtro-chip" style="padding:12px 28px;font-size:13px"
          hx-get="%s"
          hx-target="#comunicados-grid"
          hx-swap="beforeend">
    Cargar más
  </button>
</div>`, nextURL(page+1)))
			} else {
				sb.WriteString(`<div id="load-more-wrap" hx-swap-oob="true"></div>`)
			}
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

func categoryToTipo(category string) (tipo, chip, accentClass string) {
	switch category {
	case "REUNIÓN":
		return "reunion", "Reunión", "tipo-reunion"
	case "ACADÉMICO":
		return "academico", "Académico", "tipo-academico"
	case "EVENTO":
		return "evento", "Evento", "tipo-evento"
	case "DEPORTIVO":
		return "deportivo", "Deportivo", "tipo-deportivo"
	case "NOTICIA":
		return "info", "Noticia", "tipo-info"
	default:
		return "info", "Información", "tipo-info"
	}
}

func hardcodedComunicados() []contentBlock {
	return []contentBlock{
		{Title: "Horarios extendidos por temporada navideña", Desc: "Plaza Real abrirá sus puertas hasta las 23:00 hrs durante todo el mes de diciembre. Disfruta de nuestras tiendas, restaurantes y entretenimiento con horarios extendidos para acompañarte en estas fiestas.", Category: "INFORMACIÓN", Date: "1 de diciembre, 2026"},
		{Title: "Cierre temporal del estacionamiento norte por mantenimiento", Desc: "Informamos a nuestros visitantes que el estacionamiento norte permanecerá cerrado el sábado 9 y domingo 10 de mayo por trabajos de mantenimiento programado. Te invitamos a utilizar el estacionamiento sur durante esos días.", Category: "AVISO", Date: "5 de mayo, 2026"},
		{Title: "Nueva tienda inaugurada en Placa Comercial", Desc: "Te invitamos a conocer la más reciente incorporación a Plaza Real. Disfruta de promociones exclusivas durante toda la semana de apertura y descubre nuestra renovada oferta comercial.", Category: "EVENTO", Date: "10 de mayo, 2026"},
	}
}

// Blog — removed. Returns empty to avoid breaking existing route.
func Blog(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString("")
	}
}

// GET /fragments/noticias-page — full searchable listing of noticias (category=NOTICIA).
// Used by web/noticias.html via HTMX.
func NoticiasPage(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		q := strings.TrimSpace(c.Query("q", ""))
		catFilter := strings.ToUpper(strings.TrimSpace(c.Query("cat", "")))
		page := c.QueryInt("page", 1)
		if page < 1 {
			page = 1
		}
		const pageSize = 9

		pbFilter := "status = 'publicado' && (category = 'NOTICIA' || category = 'AVISO' || category = 'PUBLICIDAD')"
		if catFilter == "NOTICIA" || catFilter == "AVISO" || catFilter == "PUBLICIDAD" {
			pbFilter = fmt.Sprintf("status = 'publicado' && category = '%s'", catFilter)
		}
		if q != "" {
			safeQ := strings.ReplaceAll(strings.ReplaceAll(q, "'", ""), "\\", "")
			pbFilter += fmt.Sprintf(" && (title ~ '%s' || description ~ '%s')", safeQ, safeQ)
		}

		offset := (page - 1) * pageSize
		records, err := pb.FindRecordsByFilter("content_blocks", pbFilter, "-date", pageSize+1, offset)

		type noticia struct {
			ID       string
			Title    string
			Excerpt  string
			Category string
			Date     string
			ImageURL string
		}
		var noticias []noticia

		hasMore := false
		if err == nil && len(records) > 0 {
			hasMore = len(records) > pageSize
			if hasMore {
				records = records[:pageSize]
			}
			for _, r := range records {
				dateStr := ""
				if dt := r.GetDateTime("date"); !dt.IsZero() {
					dateStr = dt.Time().Format("2 January 2006")
				}
				noticias = append(noticias, noticia{
					ID:       r.Id,
					Title:    r.GetString("title"),
					Excerpt:  r.GetString("description"),
					Category: r.GetString("category"),
					Date:     dateStr,
					ImageURL: r.GetString("image_url"),
				})
			}
		}

		var sb strings.Builder

		if page == 1 {
			sb.WriteString(`<div class="noticias-page-grid" id="noticias-page-grid">`)
		}

		if len(noticias) == 0 && page == 1 {
			msg := "No hay comunicados publicados."
			if q != "" {
				msg = "Sin resultados para esa búsqueda."
			}
			sb.WriteString(fmt.Sprintf(`<div style="grid-column:1/-1;text-align:center;padding:64px 24px;color:#6B6B6B"><p style="font-size:15px">%s</p></div>`, msg))
		} else {
			for _, n := range noticias {
				thumbInner := `<div style="width:100%;height:100%;background:linear-gradient(135deg,#d60d5222,#00a0e322);display:flex;align-items:center;justify-content:center;font-size:2.2rem">📰</div>`
				if n.ImageURL != "" {
					thumbInner = fmt.Sprintf(`<img src="%s" alt="%s" loading="lazy" style="width:100%%;height:100%%;object-fit:cover;display:block">`,
						template.HTMLEscapeString(n.ImageURL), template.HTMLEscapeString(n.Title))
				}
				leerHref := "#"
				if n.ID != "" {
					leerHref = "/noticias/" + n.ID
				}
				sb.WriteString(fmt.Sprintf(`<a href="%s" class="n-card">
  <div class="n-thumb">%s<span class="n-cat">%s</span></div>
  <div class="n-body">
    <h3 class="n-title">%s</h3>
    <p class="n-excerpt">%s</p>
  </div>
  <div class="n-foot"><span class="n-date">%s</span><span class="n-arr">Leer más →</span></div>
</a>`,
					template.HTMLEscapeString(leerHref),
					thumbInner,
					template.HTMLEscapeString(n.Category),
					template.HTMLEscapeString(n.Title),
					template.HTMLEscapeString(n.Excerpt),
					template.HTMLEscapeString(n.Date),
				))
			}
		}

		buildNext := func() string {
			nextURL := fmt.Sprintf("/fragments/noticias-page?page=%d", page+1)
			if q != "" {
				nextURL += "&q=" + url.QueryEscape(q)
			}
			if catFilter != "" {
				nextURL += "&cat=" + url.QueryEscape(catFilter)
			}
			return nextURL
		}

		if page == 1 {
			sb.WriteString(`</div>`)
			if hasMore {
				sb.WriteString(fmt.Sprintf(`<div id="np-load-more" style="text-align:center;padding:32px 0 8px">
  <button class="filtro-chip" style="padding:12px 28px;font-size:13px"
          hx-get="%s" hx-target="#noticias-page-grid" hx-swap="beforeend">
    Cargar más
  </button>
</div>`, buildNext()))
			} else {
				sb.WriteString(`<div id="np-load-more"></div>`)
			}
		} else {
			if hasMore {
				sb.WriteString(fmt.Sprintf(`<div id="np-load-more" hx-swap-oob="true" style="text-align:center;padding:32px 0 8px">
  <button class="filtro-chip" style="padding:12px 28px;font-size:13px"
          hx-get="%s" hx-target="#noticias-page-grid" hx-swap="beforeend">
    Cargar más
  </button>
</div>`, buildNext()))
			} else {
				sb.WriteString(`<div id="np-load-more" hx-swap-oob="true"></div>`)
			}
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

func delayClass(i int) string {
	if i == 0 {
		return ""
	}
	return fmt.Sprintf(" reveal-delay-%d", i)
}
