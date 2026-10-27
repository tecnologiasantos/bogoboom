package middleware

import (
	"github.com/gofiber/fiber/v3"

	"bogoboom-backend/internal/core/domain"
)

// RolesAllowed lets a request through only when the resolved user has one of the roles. It must run
// after UserResolver: without a registered user the answer is 403, so an unregistered token holder
// can reach nothing but /me and /register.
func RolesAllowed(allowed ...domain.UserRole) fiber.Handler {
	return func(c fiber.Ctx) error {
		role, ok := c.Locals("role").(domain.UserRole)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": true, "message": "register your account first"})
		}
		for _, a := range allowed {
			if role == a {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": true, "message": "forbidden - insufficient permissions"})
	}
}
