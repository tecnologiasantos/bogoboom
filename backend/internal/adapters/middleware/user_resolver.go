package middleware

import (
	"github.com/gofiber/fiber/v3"

	"bogoboom-backend/internal/core/domain"
)

// UserResolver maps the Auth0 subject of a valid token to the internal user. A token whose subject
// has no registration passes without locals: handlers then treat the caller as unregistered. The
// match is by subject only, never by e-mail, so a stranger with the same address gets no account.
type UserResolver struct{ users domain.UserRepository }

// NewUserResolver builds the middleware.
func NewUserResolver(users domain.UserRepository) *UserResolver { return &UserResolver{users: users} }

// Resolve returns the Fiber handler.
func (m *UserResolver) Resolve() fiber.Handler {
	return func(c fiber.Ctx) error {
		claims := GetUserClaims(c)
		if claims == nil || claims.Sub == "" {
			return c.Next()
		}
		user, err := m.users.GetByAuth0Sub(c.Context(), claims.Sub)
		if err != nil {
			return err // 500 through the global handler; never continue unauthenticated by accident
		}
		if user == nil {
			return c.Next()
		}
		if !user.IsActive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": true, "message": domain.ErrUserInactive.Error()})
		}
		c.Locals("user_id", user.ID)
		c.Locals("role", user.Role)
		return c.Next()
	}
}
