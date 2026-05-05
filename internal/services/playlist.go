package services

import (
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
)

// Slide represents a single renderable item in a playlist carousel.
// Type determines the content kind; Template refines rendering for content_block slides,
// enabling future layouts without breaking existing ones.
type Slide struct {
	Index    int
	Type     string // "content_block" | "image" | "video"
	Template string // "hero-classic" | "hero-full-video" | "alert-emergencia" | ...
	Title    string // Used for device.current_view tracking and accessibility

	// Media fields (image / video slides)
	MediaURL  string
	StartTime float64 // Video seek position in seconds (applied via #t= URL fragment + JS)

	// Content block fields
	ContentID   string
	Description string
	Category    string
	ImageURL    string
}

// FetchWebHeroSlides returns the ordered slide list for the first web_hero device found,
// along with that device's ID for status tracking.
func FetchWebHeroSlides(pb *pocketbase.PocketBase) (slides []Slide, deviceID string, err error) {
	devices, err := pb.FindRecordsByFilter("devices", "type = 'web_hero'", "", 1, 0)
	if err != nil || len(devices) == 0 {
		return nil, "", fmt.Errorf("no web_hero device found")
	}
	device := devices[0]
	deviceID = device.Id

	playlistID := device.GetString("playlist_id")
	if playlistID == "" {
		return nil, deviceID, fmt.Errorf("web_hero device has no playlist assigned")
	}

	items, err := pb.FindRecordsByFilter("playlist_items",
		"playlist_id = '"+playlistID+"'", "orden", 100, 0)
	if err != nil || len(items) == 0 {
		return nil, deviceID, fmt.Errorf("playlist %s has no items", playlistID)
	}

	for i, item := range items {
		tipo := item.GetString("tipo")
		slide := Slide{Index: i, Type: tipo}

		switch tipo {
		case "content_block":
			cbID := item.GetString("content_block_id")
			if cbID == "" {
				continue
			}
			cb, err := pb.FindRecordById("content_blocks", cbID)
			if err != nil {
				log.Printf("⚠️  playlist: content_block %s not found: %v", cbID, err)
				continue
			}
			slide.Template = cb.GetString("template")
			if slide.Template == "" {
				slide.Template = "hero-classic"
			}
			slide.ContentID = cb.Id
			slide.Title = cb.GetString("title")
			slide.Description = cb.GetString("description")
			slide.Category = cb.GetString("category")
			slide.ImageURL = cb.GetString("image_url")

		case "image":
			mmID := item.GetString("multimedia_id")
			if mmID == "" {
				continue
			}
			mm, err := pb.FindRecordById("multimedia", mmID)
			if err != nil {
				log.Printf("⚠️  playlist: multimedia %s not found: %v", mmID, err)
				continue
			}
			slide.MediaURL = mm.GetString("url_r2")
			slide.Title = mm.GetString("filename")

		case "video":
			mmID := item.GetString("multimedia_id")
			if mmID == "" {
				continue
			}
			mm, err := pb.FindRecordById("multimedia", mmID)
			if err != nil {
				log.Printf("⚠️  playlist: multimedia %s not found: %v", mmID, err)
				continue
			}
			slide.MediaURL = mm.GetString("url_r2")
			slide.StartTime = mm.GetFloat("start_time")
			slide.Title = mm.GetString("filename")

		default:
			continue
		}

		slides = append(slides, slide)
	}

	return slides, deviceID, nil
}

// CalculateCurrentIndex resolves the target slide index from query params.
// Accepts ?slide=N (absolute jump) or ?direction=next|prev&current=N (relative navigation).
func CalculateCurrentIndex(c *fiber.Ctx, total int) int {
	if total <= 0 {
		return 0
	}
	if s := c.QueryInt("slide", -1); s >= 0 && s < total {
		return s
	}
	cur := c.QueryInt("current", 0)
	switch c.Query("direction") {
	case "next":
		return (cur + 1) % total
	case "prev":
		return ((cur-1)%total + total) % total
	}
	return 0
}

// UpdateDeviceStatus records last_seen and current_view for a device.
func UpdateDeviceStatus(pb *pocketbase.PocketBase, deviceID, viewTitle string) {
	if deviceID == "" {
		return
	}
	device, err := pb.FindRecordById("devices", deviceID)
	if err != nil {
		return
	}
	device.Set("last_seen", time.Now().UTC())
	device.Set("current_view", viewTitle)
	if err := pb.Save(device); err != nil {
		log.Printf("⚠️  UpdateDeviceStatus %s: %v", deviceID, err)
	}
}

// BuildHeroHTML is the central dispatcher: selects the correct renderer
// based on slide.Type and slide.Template, enabling clean extensibility.
func BuildHeroHTML(slide Slide, currentIndex, totalSlides int) string {
	switch slide.Type {
	case "image":
		return buildImageSlideHTML(slide, currentIndex, totalSlides)
	case "video":
		return buildVideoSlideHTML(slide, currentIndex, totalSlides)
	default: // "content_block" and any future types default to content block rendering
		return buildContentBlockSlideHTML(slide, currentIndex, totalSlides)
	}
}

// FallbackHeroHTML renders the static hero used when no playlist is configured.
func FallbackHeroHTML() string {
	return `<section class="hero-wrap" id="hero">
  <div class="hero-carousel">
    <div class="hero-slide active">
      <div class="hero-slide-bg slide-grad-1"></div>
      <div class="hero-slide-content">
        <div class="hero-text">
          <p class="hero-eyebrow">Copiapó, Atacama</p>
          <h1 class="hero-title">Plaza Real<br/><em>Copiapó</em></h1>
          <p class="hero-desc">Tu centro comercial en el corazón de Copiapó. Más de 100 tiendas, restaurantes y entretenimiento. Visítanos todos los días.</p>
          <div class="hero-actions">
            <a href="tiendas.html" class="btn-filled">Ver tiendas</a>
            <a href="comunicados.html" class="btn-tonal-white">Noticias</a>
          </div>
        </div>
        <div class="hero-float-card">
          <div class="hero-stat-row">
            <div><div class="hero-stat-num">100+</div><div class="hero-stat-label">tiendas<br/>y servicios</div></div>
            <div><div class="hero-stat-num">7</div><div class="hero-stat-label">días<br/>a la semana</div></div>
          </div>
          <div class="hero-lema-float"><em>"Tu centro comercial en Copiapó"</em></div>
        </div>
      </div>
    </div>
  </div>
  <div class="hero-controls">
    <button class="hero-arrow prev" aria-label="Anterior" disabled>‹</button>
    <div class="hero-dots"><span class="hero-dot active"></span></div>
    <button class="hero-arrow next" aria-label="Siguiente" disabled>›</button>
  </div>
</section>`
}

// ── Internal renderers ─────────────────────────────────────────────────────────

// buildCarouselControls produces the prev/next arrows and dot indicators.
// All navigation is pure HTMX: hx-get replaces the entire #hero section.
func buildCarouselControls(currentIndex, totalSlides int) string {
	if totalSlides <= 1 {
		return `<div class="hero-controls">
    <button class="hero-arrow prev" aria-label="Anterior" disabled>‹</button>
    <div class="hero-dots"><span class="hero-dot active"></span></div>
    <button class="hero-arrow next" aria-label="Siguiente" disabled>›</button>
  </div>`
	}

	var sb strings.Builder
	sb.WriteString(`<div class="hero-controls">`)
	sb.WriteString(fmt.Sprintf(
		`<button class="hero-arrow prev" aria-label="Anterior"`+
			` hx-get="/fragments/hero?direction=prev&current=%d"`+
			` hx-target="#hero" hx-swap="outerHTML">‹</button>`,
		currentIndex))
	sb.WriteString(`<div class="hero-dots">`)
	for i := 0; i < totalSlides; i++ {
		cls := "hero-dot"
		if i == currentIndex {
			cls += " active"
		}
		sb.WriteString(fmt.Sprintf(
			`<button class="%s" aria-label="Slide %d"`+
				` hx-get="/fragments/hero?slide=%d"`+
				` hx-target="#hero" hx-swap="outerHTML"></button>`,
			cls, i+1, i))
	}
	sb.WriteString(`</div>`)
	sb.WriteString(fmt.Sprintf(
		`<button class="hero-arrow next" aria-label="Siguiente"`+
			` hx-get="/fragments/hero?direction=next&current=%d"`+
			` hx-target="#hero" hx-swap="outerHTML">›</button>`,
		currentIndex))
	sb.WriteString(`</div>`)
	return sb.String()
}

// buildContentBlockSlideHTML dispatches to the right layout based on slide.Template.
// Add new cases here as new templates are created — existing templates are never broken.
func buildContentBlockSlideHTML(slide Slide, currentIndex, totalSlides int) string {
	switch slide.Template {
	case "hero-classic":
		return buildHeroClassicHTML(currentIndex, totalSlides)
	default:
		// Unknown templates fall back to hero-classic to avoid blank screens.
		return buildHeroClassicHTML(currentIndex, totalSlides)
	}
}

// buildHeroClassicHTML renders the Plaza Real Copiapó landing hero exactly
// as it appeared before dynamic playlists were introduced.
func buildHeroClassicHTML(currentIndex, totalSlides int) string {
	controls := buildCarouselControls(currentIndex, totalSlides)
	return fmt.Sprintf(`<section class="hero-wrap" id="hero">
  <div class="hero-carousel">
    <div class="hero-slide active">
      <div class="hero-slide-bg slide-grad-1"></div>
      <div class="hero-slide-content">
        <div class="hero-text">
          <p class="hero-eyebrow">Copiapó, Atacama</p>
          <h1 class="hero-title">Plaza Real<br/><em>Copiapó</em></h1>
          <p class="hero-desc">Tu centro comercial en el corazón de Copiapó. Más de 100 tiendas, restaurantes y entretenimiento. Visítanos todos los días.</p>
          <div class="hero-actions">
            <a href="tiendas.html" class="btn-filled">Ver tiendas</a>
            <a href="comunicados.html" class="btn-tonal-white">Noticias</a>
          </div>
        </div>
        <div class="hero-float-card">
          <div class="hero-stat-row">
            <div><div class="hero-stat-num">100+</div><div class="hero-stat-label">tiendas<br/>y servicios</div></div>
            <div><div class="hero-stat-num">7</div><div class="hero-stat-label">días<br/>a la semana</div></div>
          </div>
          <div class="hero-lema-float"><em>"Tu centro comercial en Copiapó"</em></div>
        </div>
      </div>
    </div>
  </div>
  %s
</section>`, controls)
}

// buildImageSlideHTML renders a full-bleed image slide.
// NOTE: .hero-slide is already position:absolute;inset:0 via CSS — no inline override needed.
func buildImageSlideHTML(slide Slide, currentIndex, totalSlides int) string {
	controls := buildCarouselControls(currentIndex, totalSlides)
	return fmt.Sprintf(`<section class="hero-wrap" id="hero">
  <div class="hero-carousel">
    <div class="hero-slide active">
      <img src="%s" alt="%s"
           style="position:absolute;inset:0;width:100%%;height:100%%;object-fit:cover" />
    </div>
  </div>
  %s
</section>`,
		template.HTMLEscapeString(slide.MediaURL),
		template.HTMLEscapeString(slide.Title),
		controls)
}

// buildVideoSlideHTML renders a full-bleed, auto-playing muted video slide.
// The start time is applied via both the #t= Media Fragment URI and an inline script
// for maximum browser compatibility. The script is minimal — HTMX executes inline
// scripts found in swapped responses by default.
func buildVideoSlideHTML(slide Slide, currentIndex, totalSlides int) string {
	controls := buildCarouselControls(currentIndex, totalSlides)

	videoURL := slide.MediaURL
	startScript := ""
	if slide.StartTime > 0 {
		// Media Fragment URI: appended only if no fragment already present.
		if !strings.Contains(videoURL, "#") {
			videoURL = fmt.Sprintf("%s#t=%.2f", videoURL, slide.StartTime)
		}
		// Inline JS fallback for browsers that ignore #t= on <source> tags.
		startScript = fmt.Sprintf(
			"\n  <script>(function(){var v=document.querySelector('#hero video');if(v){v.currentTime=%.2f;}})()</script>",
			slide.StartTime)
	}

	return fmt.Sprintf(`<section class="hero-wrap" id="hero">
  <div class="hero-carousel">
    <div class="hero-slide active">
      <video autoplay muted playsinline loop
             style="position:absolute;inset:0;width:100%%;height:100%%;object-fit:cover">
        <source src="%s" type="video/mp4">
      </video>
    </div>
  </div>
  %s%s
</section>`,
		template.HTMLEscapeString(videoURL),
		controls,
		startScript)
}
