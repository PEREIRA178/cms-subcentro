package web

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"cms-plazareal/internal/config"

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

// IndexHandler serves the main index.html (with HTMX fragment placeholders)
func IndexHandler(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendFile("./web/index.html")
	}
}

// PageHandler serves static sub-pages
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

