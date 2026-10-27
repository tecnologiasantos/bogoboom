package middleware

import "github.com/gofiber/fiber/v3"

// SecurityHeaders sets conservative headers on every API response. The API serves JSON only, so the
// CSP forbids everything and responses are never cached (they can hold personal data).
func SecurityHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Set("Cache-Control", "no-store")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		return c.Next()
	}
}
