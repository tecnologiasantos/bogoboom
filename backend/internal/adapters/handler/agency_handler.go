package handler

import (
	"github.com/gofiber/fiber/v3"

	"bogoboom-backend/internal/core/service"
)

// AgencyHandler serves the agency endpoints.
type AgencyHandler struct {
	leads *service.LeadService
	props *service.PropertyService
}

// NewAgencyHandler builds the handler.
func NewAgencyHandler(l *service.LeadService, p *service.PropertyService) *AgencyHandler {
	return &AgencyHandler{leads: l, props: p}
}

// Account handles GET /agency/account (plan, credits, verification).
func (h *AgencyHandler) Account(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	a, err := h.leads.Account(c.Context(), uid)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(a)
}

// Leads handles GET /agency/leads: blurred, best match first.
func (h *AgencyHandler) Leads(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	list, err := h.leads.List(c.Context(), uid)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(list)
}

// Unlock handles POST /agency/leads/:id/unlock.
func (h *AgencyHandler) Unlock(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	id, ok := idParam(c, "id")
	if !ok {
		return notFound(c)
	}
	lead, err := h.leads.Unlock(c.Context(), uid, id, c.IP())
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(lead)
}

// Notifications handles GET /agency/notifications.
func (h *AgencyHandler) Notifications(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	list, err := h.leads.Notifications(c.Context(), uid)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(list)
}

// MarkRead handles POST /agency/notifications/:id/read.
func (h *AgencyHandler) MarkRead(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	id, ok := idParam(c, "id")
	if !ok {
		return notFound(c)
	}
	if err := h.leads.MarkRead(c.Context(), uid, id); err != nil {
		return fail(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ListProperties handles GET /agency/properties.
func (h *AgencyHandler) ListProperties(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	list, err := h.props.ListMine(c.Context(), uid)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(list)
}

// CreateProperty handles POST /agency/properties.
func (h *AgencyHandler) CreateProperty(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	var req service.PropertyInput
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	p, err := h.props.Create(c.Context(), uid, req)
	if err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(p)
}

// UpdateProperty handles PUT /agency/properties/:id.
func (h *AgencyHandler) UpdateProperty(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	id, ok := idParam(c, "id")
	if !ok {
		return notFound(c)
	}
	var req service.PropertyInput
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	p, err := h.props.Update(c.Context(), uid, id, req)
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(p)
}

// DeleteProperty handles DELETE /agency/properties/:id.
func (h *AgencyHandler) DeleteProperty(c fiber.Ctx) error {
	uid, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	id, ok := idParam(c, "id")
	if !ok {
		return notFound(c)
	}
	if err := h.props.Delete(c.Context(), uid, id); err != nil {
		return fail(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
