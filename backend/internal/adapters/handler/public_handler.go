package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"bogoboom-backend/internal/core/domain"
	"bogoboom-backend/internal/core/service"
)

// PublicHandler serves the unauthenticated forms and the admin tools.
type PublicHandler struct{ svc *service.PublicService }

// NewPublicHandler builds the handler.
func NewPublicHandler(s *service.PublicService) *PublicHandler { return &PublicHandler{svc: s} }

// honeypot is the hidden "website" field of the public forms: people leave it empty, bots fill it.
type honeypot struct {
	Website string `json:"website"`
}

type inquiryBody struct {
	service.InquiryInput
	honeypot
}

type leadBody struct {
	service.LeadInput
	honeypot
}

// PartnerInquiry handles POST /partner-inquiries.
func (h *PublicHandler) PartnerInquiry(c fiber.Ctx) error {
	var req inquiryBody
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	if req.Website != "" { // a bot: pretend it worked, store nothing
		return c.SendStatus(fiber.StatusAccepted)
	}
	if err := h.svc.CreatePartnerInquiry(c.Context(), req.InquiryInput); err != nil {
		return fail(c, err)
	}
	return c.SendStatus(fiber.StatusAccepted)
}

// Lead handles POST /leads.
func (h *PublicHandler) Lead(c fiber.Ctx) error {
	var req leadBody
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	if req.Website != "" {
		return c.SendStatus(fiber.StatusAccepted)
	}
	if err := h.svc.CreateLead(c.Context(), req.LeadInput); err != nil {
		return fail(c, err)
	}
	return c.SendStatus(fiber.StatusAccepted)
}

// ListInquiries handles GET /admin/partner-inquiries.
func (h *PublicHandler) ListInquiries(c fiber.Ctx) error {
	list, err := h.svc.ListInquiries(c.Context())
	if err != nil {
		return fail(c, err)
	}
	return c.JSON(list)
}

// VerifyAgency handles PUT /admin/agencies/:id/verify with {"verified": true}.
func (h *PublicHandler) VerifyAgency(c fiber.Ctx) error {
	adminID, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	id, ok := idParam(c, "id")
	if !ok {
		return notFound(c)
	}
	var req struct {
		Verified bool `json:"verified"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	if err := h.svc.VerifyAgency(c.Context(), adminID, id, req.Verified, c.IP()); err != nil {
		return fail(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// SetPlan handles PUT /admin/agencies/:id/plan with {"plan":"pro","activeUntil":"2027-01-01T00:00:00Z"}.
func (h *PublicHandler) SetPlan(c fiber.Ctx) error {
	adminID, err := userID(c)
	if err != nil {
		return fail(c, err)
	}
	id, ok := idParam(c, "id")
	if !ok {
		return notFound(c)
	}
	var req struct {
		Plan        string     `json:"plan"`
		ActiveUntil *time.Time `json:"activeUntil"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	if err := h.svc.SetAgencyPlan(c.Context(), adminID, id, domain.Plan(req.Plan), req.ActiveUntil, c.IP()); err != nil {
		return fail(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
