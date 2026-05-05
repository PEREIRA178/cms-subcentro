package fragments

import (
	"cms-plazareal/internal/config"
	"cms-plazareal/internal/helpers"
	fragmentsView "cms-plazareal/internal/view/fragments"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// formatPublishedAt returns a localized "2 Jan 2006" string for the record,
// preferring published_at, then date, then created.
func formatPublishedAt(r *core.Record) string {
	if dt := r.GetDateTime("published_at"); !dt.IsZero() {
		return dt.Time().Format("2 Jan 2006")
	}
	if dt := r.GetDateTime("date"); !dt.IsZero() {
		return dt.Time().Format("2 Jan 2006")
	}
	if dt := r.GetDateTime("created"); !dt.IsZero() {
		return dt.Time().Format("2 Jan 2006")
	}
	return ""
}

func formatExpiresAt(r *core.Record) string {
	if dt := r.GetDateTime("expires_at"); !dt.IsZero() {
		return dt.Time().Format("2 Jan 2006")
	}
	return ""
}

// NoticiasCards — GET /fragments/noticias and /fragments/noticias-page.
// Renders a templ-based card grid for category=NOTICIA, status=publicado.
func NoticiasCards(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter(
			"content_blocks",
			"category = 'NOTICIA' && status = 'publicado'",
			"-published_at,-date,-created",
			24, 0,
		)
		items := make([]fragmentsView.NoticiaCard, 0, len(records))
		for _, r := range records {
			items = append(items, fragmentsView.NoticiaCard{
				ID:          r.Id,
				Title:       r.GetString("title"),
				Description: r.GetString("description"),
				ImageURL:    r.GetString("image_url"),
				PublishedAt: formatPublishedAt(r),
			})
		}
		return helpers.Render(c, fragmentsView.NoticiasCards(items))
	}
}

// ComunicadosCards — GET /fragments/comunicados-cards and /fragments/comunicados-page.
func ComunicadosCards(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter(
			"content_blocks",
			"category = 'COMUNICADO' && status = 'publicado'",
			"-published_at,-date,-created",
			24, 0,
		)
		items := make([]fragmentsView.ComunicadoCard, 0, len(records))
		for _, r := range records {
			items = append(items, fragmentsView.ComunicadoCard{
				ID:          r.Id,
				Title:       r.GetString("title"),
				Description: r.GetString("description"),
				PublishedAt: formatPublishedAt(r),
			})
		}
		return helpers.Render(c, fragmentsView.ComunicadosCards(items))
	}
}

// PromosCards — GET /fragments/promos and /fragments/promos-page.
func PromosCards(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		records, _ := pb.FindRecordsByFilter(
			"content_blocks",
			"category = 'PROMOCION' && status = 'publicado'",
			"-featured,-published_at,-date,-created",
			24, 0,
		)
		items := make([]fragmentsView.PromoCard, 0, len(records))
		for _, r := range records {
			items = append(items, fragmentsView.PromoCard{
				ID:          r.Id,
				Title:       r.GetString("title"),
				Description: r.GetString("description"),
				ImageURL:    r.GetString("image_url"),
				PublishedAt: formatPublishedAt(r),
				ExpiresAt:   formatExpiresAt(r),
				Featured:    r.GetBool("featured"),
			})
		}
		return helpers.Render(c, fragmentsView.PromosCards(items))
	}
}
