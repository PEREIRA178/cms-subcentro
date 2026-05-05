package middleware

import (
	"strings"

	"cms-plazareal/internal/auth"
	"cms-plazareal/internal/config"

	"github.com/gofiber/fiber/v2"
)

// AuthRequired middleware checks for a valid JWT in cookie or Authorization header
func AuthRequired(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenStr string

		// 1. Check cookie first (browser sessions)
		tokenStr = c.Cookies("pr_token")

		// 2. Fallback to Authorization header (API calls)
		if tokenStr == "" {
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenStr == "" {
			// If HTMX request, return a redirect trigger
			if c.Get("HX-Request") == "true" {
				c.Set("HX-Redirect", "/admin/login")
				return c.Status(fiber.StatusUnauthorized).SendString("")
			}
			return c.Redirect("/admin/login")
		}

		claims, err := auth.ValidateToken(cfg, tokenStr)
		if err != nil {
			// Clear invalid cookie
			c.ClearCookie("pr_token")
			if c.Get("HX-Request") == "true" {
				c.Set("HX-Redirect", "/admin/login")
				return c.Status(fiber.StatusUnauthorized).SendString("")
			}
			return c.Redirect("/admin/login")
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("userID", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("user_role", claims.Role)
		c.Locals("user_nombre", claims.Nombre)

		return c.Next()
	}
}

// RoleRequired middleware checks if user has one of the specified roles
func RoleRequired(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("user_role").(string)
		if !ok || userRole == "" {
			return c.Status(fiber.StatusForbidden).SendString("Acceso denegado")
		}

		for _, r := range roles {
			if userRole == r {
				return c.Next()
			}
		}

		if c.Get("HX-Request") == "true" {
			return c.Status(fiber.StatusForbidden).SendString(
				`<div class="toast toast-error">No tienes permisos para esta acción</div>`,
			)
		}
		return c.Status(fiber.StatusForbidden).SendString("Acceso denegado: rol insuficiente")
	}
}
