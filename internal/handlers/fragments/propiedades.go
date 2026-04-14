package fragments

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"jcp-gestioninmobiliaria/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
)

// propiedad is a flattened listing used by the public fragment/detail handlers.
type propiedad struct {
	ID             string
	Titulo         string
	Slug           string
	Descripcion    string
	Operacion      string // VENTA | ARRIENDO
	Tipo           string // CASA | DEPARTAMENTO | ...
	Direccion      string
	Comuna         string
	Region         string
	PrecioUF       float64
	PrecioCLP      float64
	Dormitorios    int
	Banos          int
	Estac          int
	SupUtil        float64
	SupTotal       float64
	Ano            int
	Estado         string
	Amenidades     []string
	Destacada      bool
	Oportunidad    bool
	CoverImage     string
	Gallery        []string
	TourURL        string
	Lat, Lng       float64
	FechaPublicado string
	Whatsapp       string
}

// formatCLP renders a price in CLP with thousands separators (Chilean: 1.234.567).
func formatCLP(v float64) string {
	if v <= 0 {
		return ""
	}
	n := int64(v)
	s := fmt.Sprintf("%d", n)
	// insert dots as thousand separators
	out := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, c)
	}
	return "$ " + string(out)
}

func formatUF(v float64) string {
	if v <= 0 {
		return ""
	}
	if v == float64(int64(v)) {
		return fmt.Sprintf("UF %d", int64(v))
	}
	return fmt.Sprintf("UF %.1f", v)
}

func priceLabel(p propiedad) string {
	uf := formatUF(p.PrecioUF)
	clp := formatCLP(p.PrecioCLP)
	switch {
	case uf != "" && clp != "":
		return uf + `<span class="prop-price-clp"> · ` + clp + `</span>`
	case uf != "":
		return uf
	default:
		return clp
	}
}

func fetchPropiedades(pb *pocketbase.PocketBase, filter, sort string, limit, offset int) []propiedad {
	if sort == "" {
		sort = "-destacada,-publicado_en"
	}
	records, err := pb.FindRecordsByFilter("propiedades", filter, sort, limit, offset)
	if err != nil || len(records) == 0 {
		return nil
	}
	out := make([]propiedad, 0, len(records))
	for _, r := range records {
		amen := splitCSV(r.GetString("amenidades"))
		gal := splitCSV(r.GetString("gallery"))
		date := ""
		if dt := r.GetDateTime("publicado_en"); !dt.IsZero() {
			date = dt.Time().Format("2 Jan 2006")
		}
		out = append(out, propiedad{
			ID:             r.Id,
			Titulo:         r.GetString("titulo"),
			Slug:           r.GetString("slug"),
			Descripcion:    r.GetString("descripcion"),
			Operacion:      r.GetString("operacion"),
			Tipo:           r.GetString("tipo"),
			Direccion:      r.GetString("direccion"),
			Comuna:         r.GetString("comuna"),
			Region:         r.GetString("region"),
			PrecioUF:       r.GetFloat("precio_uf"),
			PrecioCLP:      r.GetFloat("precio_clp"),
			Dormitorios:    r.GetInt("dormitorios"),
			Banos:          r.GetInt("banos"),
			Estac:          r.GetInt("estacionamientos"),
			SupUtil:        r.GetFloat("superficie_util"),
			SupTotal:       r.GetFloat("superficie_total"),
			Ano:            r.GetInt("ano_construccion"),
			Estado:         r.GetString("estado_propiedad"),
			Amenidades:     amen,
			Destacada:      r.GetBool("destacada"),
			Oportunidad:    r.GetBool("oportunidad"),
			CoverImage:     r.GetString("cover_image"),
			Gallery:        gal,
			TourURL:        r.GetString("tour_url"),
			Lat:            r.GetFloat("lat"),
			Lng:            r.GetFloat("lng"),
			FechaPublicado: date,
			Whatsapp:       r.GetString("contacto_whatsapp"),
		})
	}
	return out
}

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

// PropiedadesDestacadas — /fragments/propiedades-destacadas — home featured grid.
func PropiedadesDestacadas(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		items := fetchPropiedades(pb,
			"status = 'publicado' && destacada = true", "-publicado_en", 6, 0)
		if len(items) == 0 {
			items = fetchPropiedades(pb, "status = 'publicado'", "-publicado_en", 6, 0)
		}

		var sb strings.Builder
		sb.WriteString(`<section class="prop-featured-section" id="destacadas"><div class="container">`)
		sb.WriteString(`<div class="prop-featured-head">`)
		sb.WriteString(`<div><p class="label-primary reveal visible">Propiedades destacadas</p>`)
		sb.WriteString(`<h2 class="headline-l reveal visible" style="margin-bottom:0">Casas y departamentos seleccionados para ti</h2></div>`)
		sb.WriteString(`<a href="/propiedades.html" class="prop-featured-link reveal visible">Ver todas las propiedades →</a>`)
		sb.WriteString(`</div>`)
		sb.WriteString(`<div class="prop-grid">`)

		for i, p := range items {
			sb.WriteString(renderPropCard(p, i))
		}

		sb.WriteString(`</div></div></section>`)
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// PropiedadesPage — /fragments/propiedades-page — full searchable listings.
// Supports filters: operacion, tipo, comuna, dormitorios, precio_min, precio_max, q
func PropiedadesPage(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		operacion := strings.ToUpper(strings.TrimSpace(c.Query("operacion", "")))
		tipo := strings.ToUpper(strings.TrimSpace(c.Query("tipo", "")))
		comuna := strings.TrimSpace(c.Query("comuna", ""))
		q := strings.TrimSpace(c.Query("q", ""))
		dormQ := c.QueryInt("dormitorios", 0)
		precioMin := c.QueryFloat("precio_min", 0)
		precioMax := c.QueryFloat("precio_max", 0)
		sortParam := c.Query("sort", "recientes")
		page := c.QueryInt("page", 1)
		if page < 1 {
			page = 1
		}
		const pageSize = 9

		pbFilter := "status = 'publicado'"
		if operacion == "VENTA" || operacion == "ARRIENDO" {
			pbFilter += " && operacion = '" + operacion + "'"
		}
		if tipo != "" {
			pbFilter += " && tipo = '" + tipo + "'"
		}
		if comuna != "" {
			pbFilter += fmt.Sprintf(" && comuna ~ '%s'", escapeFilter(comuna))
		}
		if dormQ > 0 {
			pbFilter += fmt.Sprintf(" && dormitorios >= %d", dormQ)
		}
		if precioMin > 0 {
			pbFilter += fmt.Sprintf(" && precio_uf >= %f", precioMin)
		}
		if precioMax > 0 {
			pbFilter += fmt.Sprintf(" && precio_uf <= %f", precioMax)
		}
		if q != "" {
			safeQ := escapeFilter(q)
			pbFilter += fmt.Sprintf(" && (titulo ~ '%s' || descripcion ~ '%s' || comuna ~ '%s' || direccion ~ '%s')",
				safeQ, safeQ, safeQ, safeQ)
		}

		sortClause := "-destacada,-publicado_en"
		switch sortParam {
		case "precio_asc":
			sortClause = "precio_uf"
		case "precio_desc":
			sortClause = "-precio_uf"
		case "superficie":
			sortClause = "-superficie_util"
		}

		offset := (page - 1) * pageSize
		items := fetchPropiedades(pb, pbFilter, sortClause, pageSize+1, offset)
		hasMore := len(items) > pageSize
		if hasMore {
			items = items[:pageSize]
		}

		var sb strings.Builder

		if page == 1 {
			sb.WriteString(`<div class="prop-grid" id="prop-grid">`)
		}

		if len(items) == 0 && page == 1 {
			msg := "No se encontraron propiedades con estos filtros."
			sb.WriteString(fmt.Sprintf(
				`<div style="grid-column:1/-1;text-align:center;padding:64px 24px;color:var(--md-outline)">
           <p style="font-size:15px">%s</p>
           <p style="font-size:13px;margin-top:8px">Intenta ampliar el rango de búsqueda.</p>
         </div>`, msg))
		} else {
			for i, p := range items {
				sb.WriteString(renderPropCard(p, i))
			}
		}

		nextURL := func(p int) string {
			v := url.Values{}
			v.Set("page", fmt.Sprintf("%d", p))
			if operacion != "" {
				v.Set("operacion", operacion)
			}
			if tipo != "" {
				v.Set("tipo", tipo)
			}
			if comuna != "" {
				v.Set("comuna", comuna)
			}
			if dormQ > 0 {
				v.Set("dormitorios", fmt.Sprintf("%d", dormQ))
			}
			if precioMin > 0 {
				v.Set("precio_min", fmt.Sprintf("%g", precioMin))
			}
			if precioMax > 0 {
				v.Set("precio_max", fmt.Sprintf("%g", precioMax))
			}
			if sortParam != "" {
				v.Set("sort", sortParam)
			}
			if q != "" {
				v.Set("q", q)
			}
			return "/fragments/propiedades-page?" + v.Encode()
		}

		if page == 1 {
			sb.WriteString(`</div>`)
			if hasMore {
				sb.WriteString(fmt.Sprintf(`<div id="prop-load-more" style="text-align:center;padding:32px 0 8px">
  <button class="prop-chip" style="padding:12px 28px;font-size:13px"
          hx-get="%s" hx-target="#prop-grid" hx-swap="beforeend">
    Cargar más propiedades
  </button>
</div>`, nextURL(page+1)))
			} else {
				sb.WriteString(`<div id="prop-load-more"></div>`)
			}
		} else {
			if hasMore {
				sb.WriteString(fmt.Sprintf(`<div id="prop-load-more" hx-swap-oob="true" style="text-align:center;padding:32px 0 8px">
  <button class="prop-chip" style="padding:12px 28px;font-size:13px"
          hx-get="%s" hx-target="#prop-grid" hx-swap="beforeend">
    Cargar más propiedades
  </button>
</div>`, nextURL(page+1)))
			} else {
				sb.WriteString(`<div id="prop-load-more" hx-swap-oob="true"></div>`)
			}
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

func escapeFilter(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "'", ""), "\\", "")
}

// renderPropCard emits a single realtor-style card.
func renderPropCard(p propiedad, i int) string {
	imgHTML := ""
	if p.CoverImage != "" {
		imgHTML = fmt.Sprintf(`<img src="%s" alt="%s" loading="lazy"/>`,
			template.HTMLEscapeString(p.CoverImage),
			template.HTMLEscapeString(p.Titulo))
	} else {
		imgHTML = `<div class="prop-img-placeholder">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M3 12L12 3l9 9v9a1 1 0 0 1-1 1h-5v-7h-6v7H4a1 1 0 0 1-1-1v-9z"/>
      </svg>
    </div>`
	}

	badges := ""
	if p.Destacada {
		badges += `<span class="prop-badge prop-badge-featured">Destacada</span>`
	}
	if p.Oportunidad {
		badges += `<span class="prop-badge prop-badge-deal">Oportunidad</span>`
	}
	opBadge := ""
	switch p.Operacion {
	case "VENTA":
		opBadge = `<span class="prop-badge prop-badge-venta">En Venta</span>`
	case "ARRIENDO":
		opBadge = `<span class="prop-badge prop-badge-arriendo">En Arriendo</span>`
	}

	// Feature icons: dormitorios, banos, superficie
	feats := `<div class="prop-feats">`
	if p.Dormitorios > 0 {
		feats += fmt.Sprintf(`<span class="prop-feat"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 4v16M22 4v16M2 12h20M5 8h4M15 8h4"/></svg>%d dorm</span>`, p.Dormitorios)
	}
	if p.Banos > 0 {
		feats += fmt.Sprintf(`<span class="prop-feat"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 12V5a2 2 0 0 1 2-2h2v3M4 12h16v4a4 4 0 0 1-4 4H8a4 4 0 0 1-4-4zM8 20l-1 2M16 20l1 2"/></svg>%d baño%s</span>`,
			p.Banos, plural(p.Banos))
	}
	if p.SupUtil > 0 {
		feats += fmt.Sprintf(`<span class="prop-feat"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/></svg>%g m² útiles</span>`, p.SupUtil)
	} else if p.SupTotal > 0 {
		feats += fmt.Sprintf(`<span class="prop-feat"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/></svg>%g m² terreno</span>`, p.SupTotal)
	}
	if p.Estac > 0 {
		feats += fmt.Sprintf(`<span class="prop-feat"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M10 8h3a2 2 0 0 1 0 4h-3v4"/></svg>%d estac.</span>`, p.Estac)
	}
	feats += `</div>`

	location := template.HTMLEscapeString(strings.TrimSpace(strings.Trim(
		fmt.Sprintf("%s · %s", p.Comuna, p.Region), " ·")))
	if p.Comuna == "" && p.Region == "" {
		location = template.HTMLEscapeString(p.Direccion)
	}

	delay := ""
	if i > 0 {
		delay = fmt.Sprintf(" reveal-delay-%d", i%4)
	}

	return fmt.Sprintf(`
<article class="prop-card reveal visible%s" data-id="%s">
  <a href="/propiedades/%s" class="prop-card-link" aria-label="Ver %s">
    <div class="prop-media">
      %s
      <div class="prop-badges">%s%s</div>
      <button class="prop-fav" type="button" aria-label="Guardar propiedad" onclick="event.preventDefault();event.stopPropagation();this.classList.toggle('is-active')">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78L12 21.23l8.84-8.84a5.5 5.5 0 0 0 0-7.78z"/></svg>
      </button>
    </div>
    <div class="prop-body">
      <div class="prop-price">%s</div>
      %s
      <h3 class="prop-title">%s</h3>
      <p class="prop-loc"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>%s</p>
    </div>
  </a>
</article>`,
		delay,
		template.HTMLEscapeString(p.ID),
		template.HTMLEscapeString(func() string {
			if p.Slug != "" {
				return p.Slug
			}
			return p.ID
		}()),
		template.HTMLEscapeString(p.Titulo),
		imgHTML,
		badges, opBadge,
		priceLabel(p),
		feats,
		template.HTMLEscapeString(p.Titulo),
		location,
	)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
