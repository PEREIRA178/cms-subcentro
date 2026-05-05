package fragments

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"cms-plazareal/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
)

// tienda is a flattened store record for public fragments.
type tienda struct {
	ID         string
	Nombre     string
	Slug       string
	Cat        string
	Gal        string
	Local      string
	Logo       string
	Tags       []string
	Desc       string
	About      string
	About2     string
	Pay        string
	Photos     []string
	Similar    []string
	Whatsapp   string
	Telefono   string
	Rating     string
	HorarioLV  string
	HorarioSab string
	HorarioDom string
	Destacada  bool
}

func catLabel(cat string) string {
	switch cat {
	case "tiendas":
		return "Tiendas & Moda"
	case "restaurantes":
		return "Restaurantes & Café"
	case "farmacias":
		return "Farmacias"
	case "salud":
		return "Salud & Belleza"
	case "tecnologia":
		return "Tecnología"
	case "servicios":
		return "Servicios"
	}
	return strings.Title(cat)
}

func catEmoji(cat string) string {
	switch cat {
	case "tiendas":
		return "👗"
	case "restaurantes":
		return "🍽️"
	case "farmacias":
		return "💊"
	case "salud":
		return "✨"
	case "tecnologia":
		return "📱"
	case "servicios":
		return "🛎️"
	}
	return "🏬"
}

func fetchTiendas(pb *pocketbase.PocketBase, filter, sort string, limit, offset int) []tienda {
	if sort == "" {
		sort = "-destacada,nombre"
	}
	records, err := pb.FindRecordsByFilter("tiendas", filter, sort, limit, offset)
	if err != nil || len(records) == 0 {
		return nil
	}
	out := make([]tienda, 0, len(records))
	for _, r := range records {
		slug := r.GetString("slug")
		if slug == "" {
			slug = r.Id
		}
		t := tienda{
			ID:         r.Id,
			Nombre:     r.GetString("nombre"),
			Slug:       slug,
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
			Similar:    splitCSV(r.GetString("similar")),
			Whatsapp:   r.GetString("whatsapp"),
			Telefono:   r.GetString("telefono"),
			Rating:     r.GetString("rating"),
			HorarioLV:  r.GetString("horario_lv"),
			HorarioSab: r.GetString("horario_sab"),
			HorarioDom: r.GetString("horario_dom"),
			Destacada:  r.GetBool("destacada"),
		}
		if t.Rating == "" {
			t.Rating = "4.5"
		}
		out = append(out, t)
	}
	return out
}

// renderTiendaCard returns the HTML card for the buscador grid.
func renderTiendaCard(t tienda, i int) string {
	delay := (i % 24) * 35
	logoSrc := t.Logo
	if logoSrc == "" {
		logoSrc = fmt.Sprintf("https://picsum.photos/seed/%s/320/148", t.Slug)
	}
	galBadge := "Placa Comercial"
	if t.Gal == "sur" {
		galBadge = "Torre Flamenco"
	}
	featured := ""
	if t.Destacada {
		featured = `<span class="c-badge"><span class="material-symbols-outlined">grade</span>Destacado</span>`
	}
	tagsStr := ""
	if len(t.Tags) > 0 {
		tagsStr = strings.Join(t.Tags, " · ")
	}
	catDisp := catLabel(t.Cat)

	return fmt.Sprintf(`<a href="tienda-individual.html?tienda=%s" class="s-card" role="listitem" aria-label="Ver %s" style="animation-delay:%dms">
  <div class="c-top">
    <img src="%s" alt="Logo %s" loading="lazy" onerror="this.src='https://picsum.photos/seed/%s/320/148'">
    <span class="c-gal">%s</span>%s
  </div>
  <div class="c-body">
    <div class="c-name">%s</div>
    <div class="c-cat">%s</div>
    <div class="c-foot"><span class="c-tag">%s</span><span class="c-arr">→</span></div>
  </div>
</a>`,
		template.HTMLEscapeString(t.Slug),
		template.HTMLEscapeString(t.Nombre),
		delay,
		template.HTMLEscapeString(logoSrc),
		template.HTMLEscapeString(t.Nombre),
		template.HTMLEscapeString(t.Slug),
		galBadge, featured,
		template.HTMLEscapeString(t.Nombre),
		template.HTMLEscapeString(tagsStr),
		template.HTMLEscapeString(catDisp),
	)
}

// TiendasPage serves the HTMX fragment for buscador-tiendas.html card grid.
func TiendasPage(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		catFilter := c.Query("cat")
		galFilter := c.Query("gal")
		q := strings.TrimSpace(c.Query("q"))
		pageNum := 1
		if p := c.QueryInt("page", 1); p > 0 {
			pageNum = p
		}
		const pageSize = 24

		pbFilter := "status = 'publicado'"
		if catFilter != "" && catFilter != "all" {
			pbFilter += fmt.Sprintf(" && cat = '%s'", escapeFilter(catFilter))
		}
		if galFilter != "" && galFilter != "all" {
			pbFilter += fmt.Sprintf(" && gal = '%s'", escapeFilter(galFilter))
		}
		if q != "" {
			esc := escapeFilter(q)
			pbFilter += fmt.Sprintf(" && (nombre ~ '%s' || tags ~ '%s' || cat ~ '%s')", esc, esc, esc)
		}

		offset := (pageNum - 1) * pageSize
		stores := fetchTiendas(pb, pbFilter, "", pageSize+1, offset)

		hasMore := len(stores) > pageSize
		if hasMore {
			stores = stores[:pageSize]
		}

		if len(stores) == 0 {
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.SendString(`<div id="empty" style="display:block;text-align:center;padding:72px 24px;max-width:360px;margin:0 auto"><div style="font-size:3.5rem;margin-bottom:16px;opacity:.45">🔍</div><h3 style="font-family:'Montserrat',sans-serif;font-size:1.3rem;font-weight:900;margin-bottom:6px">Sin resultados</h3><p style="font-size:.87rem;color:#6B6B6B">Intenta con otro término o quita los filtros.</p></div>`)
		}

		var sb strings.Builder
		for i, t := range stores {
			sb.WriteString(renderTiendaCard(t, offset+i))
		}

		if hasMore {
			nextPage := pageNum + 1
			sb.WriteString(fmt.Sprintf(`<div id="load-more-trigger" style="grid-column:1/-1;text-align:center;padding:16px 0">
  <button id="lmBtn"
    hx-get="/fragments/tiendas?cat=%s&gal=%s&q=%s&page=%d"
    hx-target="#grid"
    hx-swap="beforeend"
    hx-on::after-request="document.getElementById('load-more-trigger')?.remove()"
    style="background:#0E0E0E;color:#fff;font-family:'Geist',sans-serif;font-weight:700;font-size:.95rem;padding:15px 44px;border-radius:100px;border:none;cursor:pointer">
    ⬇ Cargar más
  </button>
</div>`,
				template.URLQueryEscaper(catFilter),
				template.URLQueryEscaper(galFilter),
				template.URLQueryEscaper(q),
				nextPage,
			))
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// renderIndexCard renders an s-card for the home page featured section.
func renderIndexCard(t tienda, i int) string {
	logoSrc := t.Logo
	if logoSrc == "" {
		logoSrc = fmt.Sprintf("https://picsum.photos/seed/%s/320/160", t.Slug)
	}
	galBadge := "Placa Comercial"
	if t.Gal == "sur" {
		galBadge = "Torre Flamenco"
	}
	featured := ""
	if t.Destacada {
		featured = `<span class="s-badge"><span class="material-symbols-outlined">grade</span>Destacado</span>`
	}
	tag := catLabel(t.Cat)
	return fmt.Sprintf(`<a href="tienda-individual.html?tienda=%s" class="s-card">
  <div class="s-card-top">
    <img src="%s" alt="Logo %s" loading="lazy" onerror="this.src='https://picsum.photos/seed/%s/320/160'">
    <span class="s-gallery">%s</span>%s
  </div>
  <div class="s-card-body">
    <div class="s-card-name">%s</div>
    <div class="s-card-cat">%s</div>
    <div class="s-card-foot"><span class="s-card-tag">%s</span><span class="s-card-arr">→</span></div>
  </div>
</a>`,
		template.HTMLEscapeString(t.Slug),
		template.HTMLEscapeString(logoSrc),
		template.HTMLEscapeString(t.Nombre),
		template.HTMLEscapeString(t.Slug),
		galBadge, featured,
		template.HTMLEscapeString(t.Nombre),
		template.HTMLEscapeString(t.Desc),
		tag,
	)
}

// TiendasDestacadas serves up to 6 featured store cards for the home page.
// Only stores with destacada = true are returned — never falls back to
// non-featured stores, per product requirement.
func TiendasDestacadas(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		all := fetchTiendas(pb, "status = 'publicado'", "-destacada,nombre", 200, 0)
		var sb strings.Builder
		count := 0
		for _, t := range all {
			if !t.Destacada {
				continue
			}
			sb.WriteString(renderIndexCard(t, count))
			count++
			if count >= 6 {
				break
			}
		}
		if sb.Len() == 0 {
			sb.WriteString(`<p style="grid-column:1/-1;text-align:center;color:#6B6B6B">No hay tiendas destacadas aún.</p>`)
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// TiendasMarquee returns the inline HTML for the home marquee ticker.
// Lists every published store name twice (duplicated for the CSS loop).
func TiendasMarquee(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		stores := fetchTiendas(pb, "status = 'publicado'", "nombre", 500, 0)
		var sb strings.Builder
		writeRun := func() {
			for _, t := range stores {
				sb.WriteString(`<span class="mq-item">`)
				sb.WriteString(template.HTMLEscapeString(t.Nombre))
				sb.WriteString(`<span class="mq-dot"></span></span>`)
			}
		}
		if len(stores) > 0 {
			writeRun()
			writeRun()
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}

// TiendaDetail returns the full store content for tienda-individual.html via HTMX.
func TiendaDetail(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Params("key")

		pbFilter := fmt.Sprintf("(slug = '%s' || id = '%s') && status = 'publicado'", escapeFilter(key), escapeFilter(key))
		stores := fetchTiendas(pb, pbFilter, "", 1, 0)

		if len(stores) == 0 {
			c.Set("Content-Type", "text/html; charset=utf-8")
			return c.SendString(`<div style="text-align:center;padding:80px 24px"><h2 style="font-family:'Montserrat',sans-serif">Tienda no encontrada</h2><p><a href="buscador-tiendas.html">Ver todas las tiendas</a></p></div>`)
		}

		t := stores[0]

		// Build gallery HTML (up to 4 photos)
		photos := t.Photos
		for len(photos) < 4 {
			photos = append(photos, fmt.Sprintf("https://images.unsplash.com/photo-1441986300917-64674bd600d8?w=400&q=80"))
		}
		photosJSON, _ := json.Marshal(photos)
		photosAttr := template.HTMLEscapeString(string(photosJSON))

		// Build tags HTML
		var tagsSB strings.Builder
		for _, tag := range t.Tags {
			tagsSB.WriteString(fmt.Sprintf(`<span class="s-tag">%s</span>`, template.HTMLEscapeString(tag)))
		}
		tagsSB.WriteString(`<span class="s-open"><span class="live-dot"></span>Abierto ahora</span>`)

		// Similar stores (fetch from DB)
		var simHTML strings.Builder
		if len(t.Similar) > 0 {
			slugList := ""
			for i, s := range t.Similar {
				if i > 0 {
					slugList += ","
				}
				slugList += fmt.Sprintf("'%s'", escapeFilter(s))
			}
			simStores := fetchTiendas(pb, fmt.Sprintf("slug ?| [%s] && status = 'publicado'", slugList), "nombre", 4, 0)
			// Fallback: fetch any published stores (exclude current)
			if len(simStores) == 0 {
				simStores = fetchTiendas(pb, fmt.Sprintf("status = 'publicado' && slug != '%s'", escapeFilter(t.Slug)), "nombre", 4, 0)
			}
			for _, sim := range simStores {
				simLogo := sim.Logo
				if simLogo == "" {
					simLogo = fmt.Sprintf("https://picsum.photos/seed/%s/46/46", sim.Slug)
				}
				simTags := strings.Join(sim.Tags, " · ")
				simHTML.WriteString(fmt.Sprintf(`<a href="tienda-individual.html?tienda=%s" class="sim-card">
  <div class="sim-logo"><img src="%s" alt="%s" onerror="this.src='https://picsum.photos/seed/%s/46/46'"></div>
  <div><div class="sim-name">%s</div><div class="sim-sub">%s</div></div>
  <span class="sim-arr">→</span>
</a>`,
					template.HTMLEscapeString(sim.Slug),
					template.HTMLEscapeString(simLogo),
					template.HTMLEscapeString(sim.Nombre),
					template.HTMLEscapeString(sim.Slug),
					template.HTMLEscapeString(sim.Nombre),
					template.HTMLEscapeString(simTags),
				))
			}
		}

		// Build schedule table
		days := []struct{ name, field string }{
			{"Lunes", "lv"}, {"Martes", "lv"}, {"Miércoles", "lv"},
			{"Jueves", "lv"}, {"Viernes", "lv"}, {"Sábado", "sab"}, {"Domingo", "dom"},
		}
		horarios := map[string]string{
			"lv":  t.HorarioLV,
			"sab": t.HorarioSab,
			"dom": t.HorarioDom,
		}
		var schSB strings.Builder
		for _, d := range days {
			hrs := horarios[d.field]
			if hrs == "" {
				hrs = "Consultar"
			}
			schSB.WriteString(fmt.Sprintf(`<div class="sch-row"><span class="sch-day">%s</span><span class="sch-hrs">%s</span></div>`,
				template.HTMLEscapeString(d.name),
				template.HTMLEscapeString(hrs),
			))
		}

		galLabel := "Placa Comercial"
		if t.Gal == "sur" {
			galLabel = "Torre Flamenco"
		}
		localInfo := fmt.Sprintf("%s · %s", galLabel, t.Local)
		telLink := fmt.Sprintf("tel:%s", strings.ReplaceAll(t.Telefono, " ", ""))

		logoSrc := t.Logo
		if logoSrc == "" {
			logoSrc = fmt.Sprintf("https://picsum.photos/seed/%s/148/148", t.Slug)
		}

		aboutTitle := fmt.Sprintf(`<span class="material-symbols-outlined" style="vertical-align:middle;font-size:1.1em;margin-right:6px">auto_awesome</span>Sobre %s`, t.Nombre)

		html := fmt.Sprintf(`
<div class="bc">
  <nav class="bc-inner">
    <a href="index.html">Inicio</a><span class="sep">›</span>
    <a href="buscador-tiendas.html">Tiendas</a><span class="sep">›</span>
    <span class="cur">%s</span>
  </nav>
</div>
<section class="s-hero">
  <div class="s-hero-inner">
    <div class="logo-zone">
      <img id="storeLogoImg" src="%s" alt="Logo %s" onerror="this.src='https://picsum.photos/seed/%s/148/148';this.style.objectFit='cover';this.style.padding='0'">
    </div>
    <div class="s-hero-info">
      <div class="s-cat-badge">%s %s</div>
      <h1 class="s-hero-name">%s</h1>
      <p class="s-hero-desc">%s</p>
      <div class="s-tags">%s</div>
    </div>
  </div>
</section>
<div class="main">
  <div>
    <div class="card fi" style="padding:20px">
      <h2><span class="material-symbols-outlined" style="vertical-align:middle;font-size:1.1em;margin-right:6px">photo_library</span>Galería</h2>
      <div class="gal-grid" data-photos="%s">
        <div class="gal-item" onclick="lbOpen(JSON.parse(this.parentNode.dataset.photos),0)"><img src="%s" alt="Foto 1"><div class="gal-over"></div></div>
        <div class="gal-item" onclick="lbOpen(JSON.parse(this.parentNode.dataset.photos),1)"><img src="%s" alt="Foto 2"><div class="gal-over"></div></div>
        <div class="gal-item" onclick="lbOpen(JSON.parse(this.parentNode.dataset.photos),2)"><img src="%s" alt="Foto 3"><div class="gal-over"></div></div>
        <div class="gal-item" onclick="lbOpen(JSON.parse(this.parentNode.dataset.photos),3)"><img src="%s" alt="Más"><div class="gal-over"><span class="gal-more">Ver todas</span></div></div>
      </div>
    </div>
    <div class="card fi">
      <h2>%s</h2>
      <p style="font-size:.93rem;line-height:1.75;color:var(--muted);font-weight:300">%s</p>
      %s
    </div>
    <div class="card fi">
      <h2><span class="material-symbols-outlined" style="vertical-align:middle;font-size:1.1em;margin-right:6px">schedule</span>Horarios</h2>
      <div>%s</div>
    </div>
    <div class="card fi" style="padding:20px">
      <h2><span class="material-symbols-outlined" style="vertical-align:middle;font-size:1.1em;margin-right:6px">location_on</span>Ubicación en Plaza Real</h2>
      <div class="map-mini">
        <iframe src="https://www.google.com/maps?q=Plaza+Real+Copiap%%C3%%B3&output=embed" allowfullscreen loading="lazy" referrerpolicy="no-referrer-when-downgrade" title="Ubicación"></iframe>
      </div>
      <div style="margin-top:13px;display:flex;align-items:center;gap:9px;flex-wrap:wrap">
        <span style="background:var(--surface2);border-radius:9px;padding:7px 13px;font-size:.83rem;font-weight:600;display:inline-flex;align-items:center;gap:5px"><span class="material-symbols-outlined" style="font-size:15px">pin_drop</span>%s</span>
        <span style="background:var(--surface2);border-radius:9px;padding:7px 13px;font-size:.83rem;font-weight:600;color:var(--muted);display:inline-flex;align-items:center;gap:5px"><span class="material-symbols-outlined" style="font-size:15px">apartment</span>Centro de Copiapó</span>
      </div>
    </div>
  </div>
  <div>
    <div class="card fi">
      <div class="rating-strip">
        <div><div class="r-big">%s</div></div>
        <div><div class="stars">★★★★★</div><div class="r-count">Valoraciones Google</div></div>
      </div>
      <div class="i-row"><div class="i-icon r"><span class="material-symbols-outlined">location_on</span></div><div class="i-txt"><h4>Dirección</h4><p>Colipí 484, Copiapó, Atacama</p><a href="https://maps.google.com/?q=Plaza+Real+Copiap%%C3%%B3" target="_blank" rel="noopener">Ver en Google Maps →</a></div></div>
      <div class="i-row"><div class="i-icon y"><span class="material-symbols-outlined">storefront</span></div><div class="i-txt"><h4>Nivel &amp; Local</h4><p>%s</p></div></div>
      <div class="i-row"><div class="i-icon g"><span class="material-symbols-outlined">schedule</span></div><div class="i-txt"><h4>Horario Lun–Vie</h4><p>%s</p></div></div>
      <div class="i-row"><div class="i-icon b"><span class="material-symbols-outlined">payments</span></div><div class="i-txt"><h4>Medios de pago</h4><p>%s</p></div></div>
    </div>
    <div class="act-card fi">
      <h2>Contactar &amp; Llegar</h2>
      <a href="%s" class="a-btn red"><span class="ai"><span class="material-symbols-outlined">call</span></span><span class="al">Llamar al local</span><span class="ar">→</span></a>
      <a href="https://maps.google.com/?q=Plaza+Real+Copiap%%C3%%B3" target="_blank" rel="noopener" class="a-btn"><span class="ai"><span class="material-symbols-outlined">directions</span></span><span class="al">Cómo llegar</span><span class="ar">→</span></a>
      <a href="buscador-tiendas.html" class="a-btn"><span class="ai"><span class="material-symbols-outlined">search</span></span><span class="al">Ver todas las tiendas</span><span class="ar">→</span></a>
    </div>
    <div class="card fi">
      <h2><span class="material-symbols-outlined" style="vertical-align:middle;font-size:1.1em;margin-right:6px">storefront</span>También te puede gustar</h2>
      <div class="sim-grid">%s</div>
    </div>
  </div>
</div>`,
			// bc
			template.HTMLEscapeString(t.Nombre),
			// hero
			template.HTMLEscapeString(logoSrc),
			template.HTMLEscapeString(t.Nombre),
			template.HTMLEscapeString(t.Slug),
			catEmoji(t.Cat),
			template.HTMLEscapeString(catLabel(t.Cat)),
			template.HTMLEscapeString(t.Nombre),
			template.HTMLEscapeString(t.Desc),
			tagsSB.String(),
			// gallery
			photosAttr,
			template.HTMLEscapeString(photos[0]),
			template.HTMLEscapeString(photos[1]),
			template.HTMLEscapeString(photos[2]),
			template.HTMLEscapeString(photos[3]),
			// about
			template.HTMLEscapeString(aboutTitle),
			template.HTMLEscapeString(t.About),
			func() string {
				if t.About2 == "" {
					return ""
				}
				return fmt.Sprintf(`<p style="font-size:.93rem;line-height:1.75;color:var(--muted);font-weight:300;margin-top:12px">%s</p>`, template.HTMLEscapeString(t.About2))
			}(),
			// schedule
			schSB.String(),
			// location badge
			template.HTMLEscapeString(localInfo),
			// rating card
			template.HTMLEscapeString(t.Rating),
			template.HTMLEscapeString(localInfo),
			template.HTMLEscapeString(t.HorarioLV),
			template.HTMLEscapeString(t.Pay),
			// action buttons
			telLink,
			// similar
			simHTML.String(),
		)

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	}
}

// ── shared helpers ───────────────────────────────────────────────────────────

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

// escapeFilter strips characters that would break a PocketBase filter expression.
func escapeFilter(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "'", ""), "\\", "")
}
