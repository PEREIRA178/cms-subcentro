package fragments

import (
	"fmt"
	"html/template"
	"strings"

	"jcp-gestioninmobiliaria/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
)

// HomeCarousel returns the HTML slides for the home page banner carousel.
// Loads active slides ordered by 'orden' from the carousel_slides collection.
func HomeCarousel(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, err := pb.FindRecordsByFilter("carousel_slides", "activo = true", "orden", 20, 0)
		var sb strings.Builder
		if err != nil || len(records) == 0 {
			// Fallback defaults if nothing in DB
			sb.WriteString(`<div class="carousel-slide"><img src="https://images.unsplash.com/photo-1555396273-367ea4eb4db5?w=1400&q=85" alt="Banner 1"><div class="carousel-caption"><h3>Bienvenido a Plaza Real</h3><p>Más de 100 tiendas en el centro de Copiapó</p></div></div>`)
		} else {
			for i, r := range records {
				img := r.GetString("image_url")
				if img == "" {
					continue
				}
				title := r.GetString("title")
				subtitle := r.GetString("subtitle")
				link := r.GetString("link_url")
				caption := ""
				if title != "" || subtitle != "" {
					caption = fmt.Sprintf(`<div class="carousel-caption"><h3>%s</h3><p>%s</p></div>`,
						template.HTMLEscapeString(title),
						template.HTMLEscapeString(subtitle))
				}
				slideInner := fmt.Sprintf(`<img src="%s" alt="Banner %d">%s`,
					template.HTMLEscapeString(img), i+1, caption)
				if link != "" {
					sb.WriteString(fmt.Sprintf(`<div class="carousel-slide"><a href="%s" style="display:block;width:100%%;height:100%%;text-decoration:none;color:inherit">%s</a></div>`,
						template.HTMLEscapeString(link), slideInner))
				} else {
					sb.WriteString(fmt.Sprintf(`<div class="carousel-slide">%s</div>`, slideInner))
				}
			}
		}
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(sb.String())
	}
}
