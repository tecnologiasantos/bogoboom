package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// NewRateLimiter creates a new rate limiter middleware keyed by the client
// address, c.IP(). The key is trustworthy only when the app was built with
// ConfigureClientIP, which limits forwarding headers to trusted proxies.
func NewRateLimiter(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   true,
				"message": "too many requests, please try again later",
			})
		},
	})
}
