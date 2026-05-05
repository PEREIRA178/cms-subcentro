package middleware

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"cms-plazareal/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// TrackPageView records public-page views to the page_views collection.
//
// The middleware MUST never block or fail the request. Errors from PocketBase
// are silently swallowed. Tracking is skipped for /admin, /frag, /fragments,
// /static, /api/, and any response with status >= 400.
//
// IPs are hashed with SHA-256 (first 8 bytes hex) so we can count unique
// visitors without storing raw IP addresses.
func TrackPageView(cfg *config.Config, pb *pocketbase.PocketBase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		path := c.Path()
		if strings.HasPrefix(path, "/admin") ||
			strings.HasPrefix(path, "/frag") ||
			strings.HasPrefix(path, "/fragments") ||
			strings.HasPrefix(path, "/static") ||
			strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/ws") ||
			c.Response().StatusCode() >= 400 {
			return err
		}

		col, colErr := pb.FindCollectionByNameOrId("page_views")
		if colErr != nil || col == nil {
			return err
		}

		record := core.NewRecord(col)
		record.Set("path", path)
		record.Set("referrer", c.Get("Referer"))
		record.Set("user_agent", c.Get("User-Agent"))
		h := sha256.Sum256([]byte(c.IP()))
		record.Set("ip", fmt.Sprintf("%x", h[:8]))
		_ = pb.Save(record)

		return err
	}
}
