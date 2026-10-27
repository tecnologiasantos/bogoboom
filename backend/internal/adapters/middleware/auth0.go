package middleware

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// errInvalidAudience is returned when a token was not issued for this API.
var errInvalidAudience = errors.New("invalid token audience")

// Sentinel errors for a token the client got wrong. The text of each one is the exact message the
// client receives, so tokenRejection answers with sentinel.Error(). validateToken wraps a sentinel
// together with the underlying detail when there is one (fmt.Errorf("%w: %w", sentinel, detail)): the
// detail reaches the log and never the response body.
var (
	errTokenFormat          = errors.New("invalid token format")
	errTokenHeaderEncoding  = errors.New("failed to decode token header")
	errTokenHeaderJSON      = errors.New("failed to parse token header")
	errTokenPayloadEncoding = errors.New("failed to decode token payload")
	errTokenClaimsJSON      = errors.New("failed to parse token claims")
	errTokenExpired         = errors.New("token has expired")
	errInvalidIssuer        = errors.New("invalid token issuer")
	errSigningKeyNotFound   = errors.New("signing key not found")
	errInvalidSignature     = errors.New("signature verification failed")
)

// errSigningKeysUnavailable means the token cannot be checked because the signing keys cannot be
// fetched or used. It is a failure on the server's side, not a problem with the client's token.
var errSigningKeysUnavailable = errors.New("signing keys unavailable")

const (
	// invalidTokenMessage answers 401 for a validation error that no sentinel names.
	invalidTokenMessage = "invalid token"
	// authUnavailableMessage answers 503 when the signing keys are unavailable.
	authUnavailableMessage = "authentication is temporarily unavailable, try again later"
	// jwksFetchTimeout bounds one fetch of the signing key set.
	jwksFetchTimeout = 10 * time.Second
)

// clientTokenErrors are the sentinels for problems the client caused, in the order tokenRejection
// looks for them.
var clientTokenErrors = []error{
	errTokenFormat,
	errTokenHeaderEncoding,
	errTokenHeaderJSON,
	errTokenPayloadEncoding,
	errTokenClaimsJSON,
	errTokenExpired,
	errInvalidIssuer,
	errInvalidAudience,
	errSigningKeyNotFound,
	errInvalidSignature,
}

// tokenRejection maps an error from validateToken to the HTTP status and the fixed message of the
// response. The message never comes from the error's own text, so a wrapped detail (a transport
// error, a library message) cannot reach the client.
//
// Unavailable signing keys are tested first. A key that cannot be decoded is found while the signature
// is being verified, so its error can arrive wrapped inside a signature error. It must still answer
// 503, because the key material comes from the key set and not from the token.
func tokenRejection(err error) (int, string) {
	if errors.Is(err, errSigningKeysUnavailable) {
		return fiber.StatusServiceUnavailable, authUnavailableMessage
	}
	for _, sentinel := range clientTokenErrors {
		if errors.Is(err, sentinel) {
			return fiber.StatusUnauthorized, sentinel.Error()
		}
	}
	return fiber.StatusUnauthorized, invalidTokenMessage
}

// Auth0Config holds Auth0 configuration
type Auth0Config struct {
	Domain   string
	Audience string
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	Sub   string          `json:"sub"`
	Email string          `json:"email"`
	Iss   string          `json:"iss"`
	Aud   json.RawMessage `json:"aud"`
	Exp   int64           `json:"exp"`
	Iat   int64           `json:"iat"`
	Scope string          `json:"scope"`
	Azp   string          `json:"azp"`
}

// ContextKey is the type for context keys
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "userID"
	// UserEmailKey is the context key for user email
	UserEmailKey ContextKey = "userEmail"
	// UserClaimsKey is the context key for full claims
	UserClaimsKey ContextKey = "userClaims"
)

// JWKS represents the JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

// Auth0Middleware creates JWT validation middleware for Auth0
type Auth0Middleware struct {
	config Auth0Config

	// cacheMu guards the key cache: the fields below, up to fetching.
	cacheMu sync.Mutex
	// jwksCache is the last key set fetched successfully. It is used until cacheExp, never after.
	jwksCache *JWKS
	cacheExp  time.Time
	// lastFetch is when the last fetch attempt ended, and lastFetchErr its error (nil after a success).
	lastFetch    time.Time
	lastFetchErr error
	// fetching is the fetch in flight, or nil.
	fetching *keySetFetch
	// cacheNow is the key cache's clock. Tests replace it; nil means time.Now. Token expiry is checked
	// against time.Now.
	cacheNow func() time.Time

	// client fetches the key set. Tests replace it; a nil client falls back to the default one.
	client *http.Client
}

// NewAuth0Middleware creates a new Auth0 middleware instance
func NewAuth0Middleware(config Auth0Config) *Auth0Middleware {
	return &Auth0Middleware{
		config:   config,
		cacheNow: time.Now,
		client:   &http.Client{Timeout: jwksFetchTimeout},
	}
}

// Protect returns a Fiber middleware that validates JWT tokens
func (m *Auth0Middleware) Protect() fiber.Handler {
	return func(c fiber.Ctx) error {
		// The token is read from the "Authorization: Bearer" header only. Credentials in a
		// URL end up in logs and history, so no query parameter is ever accepted.
		authHeader := c.Get("Authorization")
		var token string

		if authHeader != "" {
			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			}
		}

		if token == "" {
			log.Println("[Auth0] Unauthorized: Missing token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "missing authorization",
			})
		}

		// Validate token
		claims, err := m.validateToken(c.Context(), token)
		if err != nil {
			// The response carries a fixed message. The error, with its detail, only goes to the
			// log, and a token that cannot be checked because the signing keys are unavailable is
			// logged apart from one the client got wrong. The middleware answers both itself: an
			// error returned to Fiber would reach the global error handler and be called unhandled.
			status, message := tokenRejection(err)
			if status == fiber.StatusServiceUnavailable {
				log.Printf("[Auth0] Service unavailable: %.1024q", err)
			} else {
				log.Printf("[Auth0] Unauthorized: Invalid token: %v", err)
			}
			return c.Status(status).JSON(fiber.Map{
				"error":   true,
				"message": message,
			})
		}

		// Store claims in context
		c.Locals(string(UserClaimsKey), claims)
		c.Locals(string(UserEmailKey), claims.Email)

		// Parse user ID from sub claim (Auth0 format: auth0|<id>)
		if claims.Sub != "" {
			parts := strings.Split(claims.Sub, "|")
			if len(parts) == 2 {
				c.Locals(string(UserIDKey), parts[1])
			} else {
				c.Locals(string(UserIDKey), claims.Sub)
			}
		}

		return c.Next()
	}
}

// validateToken validates the JWT token against Auth0
func (m *Auth0Middleware) validateToken(ctx context.Context, tokenString string) (*JWTClaims, error) {
	// Split token parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errTokenFormat
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errTokenHeaderEncoding
	}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, errTokenHeaderJSON
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errTokenPayloadEncoding
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errTokenClaimsJSON
	}

	// Validate expiration
	if time.Now().Unix() > claims.Exp {
		return nil, errTokenExpired
	}

	// Validate issuer
	expectedIssuer := fmt.Sprintf("https://%s/", m.config.Domain)
	if claims.Iss != expectedIssuer {
		return nil, errInvalidIssuer
	}

	// Validate audience: only tokens issued for this API are accepted.
	if !audienceAllowed(claims.Aud, m.config.Audience) {
		log.Printf("[Auth0] Audience rejected: expected %q, token aud %q", m.config.Audience, string(claims.Aud))
		return nil, errInvalidAudience
	}

	// The key the token names: errSigningKeyNotFound, or an error wrapping errSigningKeysUnavailable.
	key, err := m.signingKey(ctx, header.Kid)
	if err != nil {
		return nil, err
	}

	// Verify signature (simplified - in production use a proper JWT library)
	if err := m.verifySignature(parts, key); err != nil {
		// A key that cannot be decoded is a failure of the key set, not of the token.
		if errors.Is(err, errSigningKeysUnavailable) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %w", errInvalidSignature, err)
	}

	return &claims, nil
}

// audienceAllowed reports whether the token's aud claim (a string or a list of
// strings) contains the expected audience. The match is exact; an empty
// expected audience, or a missing or malformed claim, never matches.
func audienceAllowed(raw json.RawMessage, expected string) bool {
	if expected == "" {
		return false
	}
	// Pointers so a JSON null entry is detectable: it makes the claim malformed.
	var entries []*string
	if err := json.Unmarshal(raw, &entries); err != nil {
		var single string
		if err := json.Unmarshal(raw, &single); err != nil {
			return false
		}
		entries = []*string{&single}
	}
	found := false
	for _, aud := range entries {
		if aud == nil {
			return false
		}
		if *aud == expected {
			found = true
		}
	}
	return found
}

// verifySignature verifies the JWT signature using RSA with SHA-256
func (m *Auth0Middleware) verifySignature(parts []string, key *JWK) error {
	// Decode n and e from JWK
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return fmt.Errorf("%w: failed to decode key modulus", errSigningKeysUnavailable)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return fmt.Errorf("%w: failed to decode key exponent", errSigningKeysUnavailable)
	}

	// Create RSA public key
	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	pubKey := &rsa.PublicKey{
		N: n,
		E: e,
	}

	// Decode signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return errors.New("failed to decode signature")
	}

	// Create the message to verify (header.payload)
	message := parts[0] + "." + parts[1]
	hash := sha256.Sum256([]byte(message))

	// Verify signature using RSA-PKCS1-v1.5 with SHA-256
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// GetUserID extracts user ID from Fiber context
func GetUserID(c fiber.Ctx) (uuid.UUID, error) {
	userIDStr := c.Locals(string(UserIDKey))
	if userIDStr == nil {
		return uuid.Nil, errors.New("user ID not found in context")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID format")
	}

	return userID, nil
}

// GetUserEmail extracts user email from Fiber context
func GetUserEmail(c fiber.Ctx) string {
	email := c.Locals(string(UserEmailKey))
	if email == nil {
		return ""
	}
	return email.(string)
}

// GetUserClaims extracts full claims from Fiber context
func GetUserClaims(c fiber.Ctx) *JWTClaims {
	claims := c.Locals(string(UserClaimsKey))
	if claims == nil {
		return nil
	}
	return claims.(*JWTClaims)
}
