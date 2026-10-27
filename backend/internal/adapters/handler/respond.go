package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"bogoboom-backend/internal/core/domain"
)

// fail answers with the status and message of a domain error. Errors it does not know are returned
// to Fiber, whose GlobalErrorHandler answers 500 with a fixed message and logs the cause.
func fail(c fiber.Ctx, err error) error {
	var ve *domain.ValidationError
	switch {
	case errors.As(err, &ve):
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: true, Message: "validation failed", Details: ve.Fields})
	case errors.Is(err, domain.ErrNotFound):
		return reply(c, fiber.StatusNotFound, err)
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrContactNotAllowed), errors.Is(err, domain.ErrAgencyNotVerified):
		return reply(c, fiber.StatusForbidden, err)
	case errors.Is(err, domain.ErrNotMatching):
		return reply(c, fiber.StatusConflict, err)
	case errors.Is(err, domain.ErrNoCredits):
		return reply(c, fiber.StatusPaymentRequired, err)
	case errors.Is(err, domain.ErrAlreadyRegistered), errors.Is(err, domain.ErrAlreadyExists):
		return reply(c, fiber.StatusConflict, err)
	case errors.Is(err, domain.ErrConsentRequired), errors.Is(err, domain.ErrEmailRequired), errors.Is(err, domain.ErrInvalidRole):
		return reply(c, fiber.StatusUnprocessableEntity, err)
	case errors.Is(err, domain.ErrUserInactive):
		return reply(c, fiber.StatusForbidden, err)
	}
	return err
}

func reply(c fiber.Ctx, status int, err error) error {
	return c.Status(status).JSON(ErrorResponse{Error: true, Message: err.Error()})
}

func badBody(c fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: true, Message: "invalid request body"})
}

// userID returns the internal user ID that UserResolver put in the context.
func userID(c fiber.Ctx) (string, error) {
	id, ok := c.Locals("user_id").(string)
	if !ok || id == "" {
		return "", domain.ErrForbidden
	}
	return id, nil
}
