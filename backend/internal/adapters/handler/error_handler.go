package handler

import (
	"errors"
	"log"
	"net/http"
	"slices"

	"github.com/gofiber/fiber/v3"
)

// InternalErrorMessage is the only text a client gets for an error the server did not map itself.
const InternalErrorMessage = "internal server error"

// unknownRequestMethod stands for a request method that Fiber does not support in the log line.
const unknownRequestMethod = "UNKNOWN"

// GlobalErrorResponse maps an error that reached the app's error handler to the status and the body
// the client gets. The text of a hidden error is never copied into the body.
//
//   - A *fiber.Error with a 4xx code (found with errors.As, so a wrapped one counts) keeps its code
//     and its own message. When the message is empty it gets the standard status text, or "Bad
//     Request" when net/http has no text for the code (499, for one), so it is never empty. The text
//     of a wrapper around it is not used.
//   - A *fiber.Error with a 5xx code keeps its code and gets InternalErrorMessage.
//   - Anything else answers 500 with InternalErrorMessage: a *fiber.Error whose code is outside
//     400-599, a nil *fiber.Error, any other error, and nil.
//
// The body has no details.
func GlobalErrorResponse(err error) (int, ErrorResponse) {
	status := fiber.StatusInternalServerError
	message := InternalErrorMessage

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) && fiberErr != nil {
		switch {
		case fiberErr.Code >= fiber.StatusBadRequest && fiberErr.Code < fiber.StatusInternalServerError:
			status = fiberErr.Code
			message = fiberErr.Message
			if message == "" {
				message = http.StatusText(status)
			}
			if message == "" { // net/http has no text for a code such as 499
				message = http.StatusText(http.StatusBadRequest)
			}
		case fiberErr.Code >= fiber.StatusInternalServerError && fiberErr.Code <= 599:
			status = fiberErr.Code
		}
	}

	return status, ErrorResponse{Error: true, Message: message}
}

// requestMethodForLog returns the request method for the log line. c.Method() panics when Fiber
// has no entry for the method, and the HTTP server hands the error handler such a request when it
// rejects one, before routing and outside the recover middleware, so the panic would stop the
// process. The method is printed as sent only when Fiber supports it. Anything else is
// unknownRequestMethod, so a method chosen by the client never reaches the log.
func requestMethodForLog(c fiber.Ctx) string {
	method := string(c.Request().Header.Method())
	if slices.Contains(c.App().Config().RequestMethods, method) {
		return method
	}
	return unknownRequestMethod
}

// GlobalErrorHandler is the fiber.ErrorHandler of the app. It answers with GlobalErrorResponse. It
// receives an error that a handler returned instead of answering, a recovered panic, and the
// failures of the HTTP server (a body over the limit, for one). It never panics, whatever request
// the server hands it.
//
// Every hidden error (status 500 and above) writes exactly one line to the standard log, which is
// the only place its cause appears:
//
//	unhandled error: GET "/api/v1/path" -> 500: "cause"
//
// The path comes from c.Path(), so it has no query string. The path and the cause are quoted, so a
// line break in either cannot start a second line. The cause is cut after 1024 characters, so a long
// driver message or provider page cannot make a huge line. The method is the one the client sent
// when Fiber supports it, and UNKNOWN otherwise. The handler reads no request body and no header
// field. An error answered with a 4xx status is not logged.
func GlobalErrorHandler(c fiber.Ctx, err error) error {
	status, body := GlobalErrorResponse(err)
	if status >= fiber.StatusInternalServerError {
		log.Printf("unhandled error: %s %q -> %d: %.1024q", requestMethodForLog(c), c.Path(), status, err)
	}
	return c.Status(status).JSON(body)
}
