package handler

import (
	"regexp"

	"github.com/gofiber/fiber/v3"

	"bogoboom-backend/internal/core/domain"
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// idParam returns the :id route parameter when it is a UUID. Anything else answers 404 at once, so
// a malformed ID never reaches a `::uuid` cast in SQL.
func idParam(c fiber.Ctx, name string) (string, bool) {
	id := c.Params(name)
	return id, uuidRe.MatchString(id)
}

func notFound(c fiber.Ctx) error {
	return reply(c, fiber.StatusNotFound, domain.ErrNotFound)
}
