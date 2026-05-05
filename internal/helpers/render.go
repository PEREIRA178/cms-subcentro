package helpers

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
)

// Render writes a templ component to the Fiber response.
// Use this in every handler instead of c.SendFile or strings.Builder.
func Render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return component.Render(c.UserContext(), c.Response().BodyWriter())
}
