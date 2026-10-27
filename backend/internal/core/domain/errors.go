package domain

import "errors"

// Errors the services return. Handlers map them to HTTP statuses; the text is safe to show.
var (
	ErrNotFound          = errors.New("not found")
	ErrForbidden         = errors.New("forbidden")
	ErrUserInactive      = errors.New("account is inactive")
	ErrAlreadyRegistered = errors.New("this account is already registered")
	ErrEmailRequired     = errors.New("the access token has no e-mail: add the e-mail claim in Auth0")
	ErrInvalidRole       = errors.New("invalid account type")
	ErrConsentRequired   = errors.New("you must accept the terms and the privacy policy")
	ErrNoCredits         = errors.New("no free contacts left this month: upgrade to Pro")
	ErrContactNotAllowed = errors.New("this person did not allow agencies to contact them")
	ErrNotMatching       = errors.New("this lead does not match any of your listings")
	ErrAgencyNotVerified = errors.New("the agency must be verified before it can contact leads")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidTransition = errors.New("invalid stage change")
)

// ValidationError lists one message per invalid field.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "validation failed" }

// NewValidationError starts an empty error; use Add and then Err.
func NewValidationError() *ValidationError { return &ValidationError{Fields: map[string]string{}} }

// Add records a problem for a field (the first message for a field wins).
func (e *ValidationError) Add(field, msg string) {
	if _, ok := e.Fields[field]; !ok {
		e.Fields[field] = msg
	}
}

// Err returns the error when something was added, and nil otherwise.
func (e *ValidationError) Err() error {
	if len(e.Fields) == 0 {
		return nil
	}
	return e
}
