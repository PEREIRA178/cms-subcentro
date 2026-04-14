package api

import (
	"strings"

	"jcp-gestioninmobiliaria/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
)

// sanitize removes characters that could cause PocketBase filter injection.
func sanitize(s string) string {
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\\", "")
	return s
}

// DevicePlaylist returns the playlist and items assigned to a device (by URL code).
// GET /api/devices/:code/playlist
func DevicePlaylist(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		code := sanitize(c.Params("code"))
		if code == "" {
			return c.Status(400).JSON(fiber.Map{"error": "missing device code"})
		}

		devices, err := pb.FindRecordsByFilter("devices", "code = '"+code+"'", "", 1, 0)
		if err != nil || len(devices) == 0 {
			return c.Status(404).JSON(fiber.Map{"error": "device not found"})
		}
		device := devices[0]

		resp := fiber.Map{
			"device": fiber.Map{
				"id":   device.Id,
				"name": device.GetString("name"),
				"type": device.GetString("type"),
				"code": device.GetString("code"),
			},
			"playlist": nil,
			"items":    []fiber.Map{},
		}

		playlistID := device.GetString("playlist_id")
		if playlistID == "" {
			return c.JSON(resp)
		}

		playlist, err := pb.FindRecordById("playlists", playlistID)
		if err != nil {
			return c.JSON(resp)
		}
		resp["playlist"] = fiber.Map{
			"id":   playlist.Id,
			"name": playlist.GetString("name"),
		}

		piRecords, err := pb.FindRecordsByFilter("playlist_items",
			"playlist_id = '"+playlistID+"'", "orden", 100, 0)
		if err != nil {
			return c.JSON(resp)
		}

		items := make([]fiber.Map, 0, len(piRecords))
		for _, pi := range piRecords {
			tipo := pi.GetString("tipo")
			duracion := pi.GetFloat("duracion_segundos")
			if duracion <= 0 {
				duracion = 10
			}

			item := fiber.Map{
				"id":               pi.Id,
				"tipo":             tipo,
				"duration_seconds": duracion,
			}

			switch tipo {
			case "content_block":
				cbID := pi.GetString("content_block_id")
				if cbID == "" {
					continue
				}
				cb, err := pb.FindRecordById("content_blocks", cbID)
				if err != nil {
					continue
				}
				item["name"] = cb.GetString("title")
				item["title"] = cb.GetString("title")
				item["description"] = cb.GetString("description")
				item["category"] = cb.GetString("category")
				item["urgency"] = cb.GetBool("urgency")
				item["image_url"] = cb.GetString("image_url")
				item["url"] = cb.GetString("image_url")

			case "image":
				mmID := pi.GetString("multimedia_id")
				if mmID == "" {
					continue
				}
				mm, err := pb.FindRecordById("multimedia", mmID)
				if err != nil {
					continue
				}
				item["name"] = mm.GetString("filename")
				item["url"] = mm.GetString("url_r2")
				item["start_time"] = float64(0)

			case "video":
				mmID := pi.GetString("multimedia_id")
				if mmID == "" {
					continue
				}
				mm, err := pb.FindRecordById("multimedia", mmID)
				if err != nil {
					continue
				}
				item["name"] = mm.GetString("filename")
				item["url"] = mm.GetString("url_r2")
				item["start_time"] = mm.GetFloat("start_time")

			default:
				continue
			}

			items = append(items, item)
		}

		resp["items"] = items
		return c.JSON(resp)
	}
}

// UpcomingEvents returns the next N upcoming events for display panels.
// GET /api/events/upcoming?limit=3
func UpcomingEvents(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit := c.QueryInt("limit", 3)
		if limit < 1 || limit > 10 {
			limit = 3
		}

		records, err := pb.FindRecordsByFilter("content_blocks",
			"status = 'publicado' && category != 'NOTICIA'",
			"-urgency,date", limit, 0)

		events := make([]fiber.Map, 0)
		if err == nil {
			for _, r := range records {
				dateStr := ""
				if dt := r.GetDateTime("date"); !dt.IsZero() {
					dateStr = dt.Time().Format("02 Jan 2006")
				}
				events = append(events, fiber.Map{
					"id":       r.Id,
					"title":    r.GetString("title"),
					"category": r.GetString("category"),
					"date":     dateStr,
					"urgency":  r.GetBool("urgency"),
				})
			}
		}

		return c.JSON(events)
	}
}
