package handler

import (
	"github.com/gofiber/fiber/v3"

	"bogoboom-backend/internal/adapters/middleware"
	"bogoboom-backend/internal/core/domain"
	"bogoboom-backend/internal/core/service"
)

// AccountHandler serves /me and /register/*.
type AccountHandler struct {
	reg   *service.RegistrationService
	users domain.UserRepository
}

// NewAccountHandler builds the handler.
func NewAccountHandler(reg *service.RegistrationService, users domain.UserRepository) *AccountHandler {
	return &AccountHandler{reg: reg, users: users}
}

// MeResponse says who the token belongs to and whether that identity is registered.
type MeResponse struct {
	IsRegistered bool            `json:"isRegistered"`
	Email        string          `json:"email"`
	ID           string          `json:"id,omitempty"`
	Role         domain.UserRole `json:"role,omitempty"`
	Locale       string          `json:"locale,omitempty"`
}

// Me handles GET /me. A valid token without a registration is not an error: the app then sends the
// person to the registration form.
func (h *AccountHandler) Me(c fiber.Ctx) error {
	claims := middleware.GetUserClaims(c)
	if claims == nil {
		return reply(c, fiber.StatusUnauthorized, domain.ErrForbidden)
	}
	id, err := userID(c)
	if err != nil {
		return c.JSON(MeResponse{IsRegistered: false, Email: claims.Email})
	}
	u, err := h.users.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	if u == nil {
		return c.JSON(MeResponse{IsRegistered: false, Email: claims.Email})
	}
	return c.JSON(MeResponse{IsRegistered: true, Email: u.Email, ID: u.ID, Role: u.Role, Locale: u.Locale})
}

func identity(c fiber.Ctx) service.Identity {
	claims := middleware.GetUserClaims(c)
	if claims == nil {
		return service.Identity{}
	}
	return service.Identity{Sub: claims.Sub, Email: claims.Email}
}

func (h *AccountHandler) created(c fiber.Ctx, u *domain.User, err error) error {
	if err != nil {
		return fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(MeResponse{IsRegistered: true, Email: u.Email, ID: u.ID, Role: u.Role, Locale: u.Locale})
}

// RegisterNewcomer handles POST /register/newcomer.
func (h *AccountHandler) RegisterNewcomer(c fiber.Ctx) error {
	var req service.NewcomerRegistration
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	u, err := h.reg.RegisterNewcomer(c.Context(), identity(c), req)
	return h.created(c, u, err)
}

// RegisterHR handles POST /register/hr.
func (h *AccountHandler) RegisterHR(c fiber.Ctx) error {
	var req service.HRRegistration
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	u, err := h.reg.RegisterHR(c.Context(), identity(c), req)
	return h.created(c, u, err)
}

// RegisterAgency handles POST /register/agency.
func (h *AccountHandler) RegisterAgency(c fiber.Ctx) error {
	var req service.AgencyRegistration
	if err := c.Bind().JSON(&req); err != nil {
		return badBody(c)
	}
	u, err := h.reg.RegisterAgency(c.Context(), identity(c), req)
	return h.created(c, u, err)
}
