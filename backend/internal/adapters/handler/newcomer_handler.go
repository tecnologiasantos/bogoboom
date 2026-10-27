package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"bogoboom-backend/internal/core/service"
)

// NewcomerHandler serves the newcomer and HR endpoints.
type NewcomerHandler struct {
	newcomers *service.NewcomerService
	props     *service.PropertyService
}

// NewNewcomerHandler builds the handler.
func NewNewcomerHandler(n *service.NewcomerService, p *service.PropertyService) *NewcomerHandler {
	return &NewcomerHandler{newcomers: n, props: p}
}

// Me handles GET /newcomers/me.
func (h *NewcomerHandler) Me(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	n, err := h.newcomers.Me(c.Context(), uid)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(n)
}

// Update handles PUT /newcomers/me.
func (h *NewcomerHandler) Update(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	var req service.NewcomerUpdate
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	n, err := h.newcomers.Update(c.Context(), uid, c.IP(), req)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(n)
}

// SearchProperties handles GET /properties?city=&maxPrice=.
func (h *NewcomerHandler) SearchProperties(c fiber.Ctx) error {
	var max *float64
	if raw := c.Query("maxPrice"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil || v < 0 {
			return badBody(c)
		}
		max = &v
	}
	list, err := h.props.Search(c.Context(), c.Query("city"), max)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(list)
}

// Favorites handles GET /favorites.
func (h *NewcomerHandler) Favorites(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	list, err := h.props.Favorites(c.Context(), uid)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(list)
}

// AddFavorite handles PUT /favorites/:id.
func (h *NewcomerHandler) AddFavorite(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	id, ok := idParam(c, "id")
	if !ok {
		return notFound(c)
	}
	if err := h.props.Favorite(c.Context(), uid, id); err != nil {
		return fail(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// RemoveFavorite handles DELETE /favorites/:id.
func (h *NewcomerHandler) RemoveFavorite(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	id, ok := idParam(c, "id")
	if !ok {
		return notFound(c)
	}
	if err := h.props.Unfavorite(c.Context(), uid, id); err != nil {
		return fail(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Invite handles POST /hr/newcomers.
func (h *NewcomerHandler) Invite(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	var req service.Invite
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	n, err := h.newcomers.Invite(c.Context(), uid, c.IP(), req)
	if err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(n)
}

// ListForHR handles GET /hr/newcomers.
func (h *NewcomerHandler) ListForHR(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	list, err := h.newcomers.ListForHR(c.Context(), uid)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(list)
}
