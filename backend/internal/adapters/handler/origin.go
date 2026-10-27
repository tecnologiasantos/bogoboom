package handler

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
)

// OriginPolicy decides which web origins may use the API from a browser: only these may open the
// chat socket, and CORS (buildApp in cmd/api) grants only these. Origins are compared
// after normalisation (lowercase scheme and host, default port dropped); there is no
// wildcard or suffix matching.
type OriginPolicy struct {
	allowed map[string]struct{}
}

// NewOriginPolicy builds a policy from configured origins such as the value of
// FRONTEND_URL. Each must be an http or https URL with a host and no user info; a
// trailing slash is ignored. An error says which rule a value broke and its position, never
// the value itself: it goes to the startup log, and the value may carry user info.
//
// A host with non-ASCII letters must be configured in punycode (xn--...), the form a browser sends
// in Origin. The policy does not convert between the two, so a Unicode value never matches a
// browser's origin: it fails closed.
func NewOriginPolicy(origins ...string) (*OriginPolicy, error) {
	if len(origins) == 0 {
		return nil, errors.New("no allowed origin configured")
	}
	allowed := make(map[string]struct{}, len(origins))
	for i, raw := range origins {
		key, err := normaliseOrigin(raw, true)
		if err != nil {
			return nil, fmt.Errorf("origin %d: %w", i+1, err)
		}
		allowed[key] = struct{}{}
	}
	return &OriginPolicy{allowed: allowed}, nil
}

// Allows reports whether the Origin header value is acceptable. An empty value (header
// missing, i.e. not a browser) is accepted; authentication applies to those clients. CORS never
// asks about an empty value, because a request without Origin is outside CORS.
func (p *OriginPolicy) Allows(origin string) bool {
	if origin == "" {
		return true
	}
	key, err := normaliseOrigin(origin, false)
	if err != nil {
		return false
	}
	_, ok := p.allowed[key]
	return ok
}

// CheckOrigin is the origin check for the websocket upgrader and for the early check
// that runs before a ticket is redeemed.
func (p *OriginPolicy) CheckOrigin(ctx *fasthttp.RequestCtx) bool {
	return p.Allows(string(ctx.Request.Header.Peek("Origin")))
}

// normaliseOrigin returns scheme://host[:port]. A configured value may carry an empty
// path or "/"; a value received from a client must be a bare origin. The errors never
// contain the value (the error of url.Parse repeats the whole URL, and part of it can be a
// secret).
func normaliseOrigin(raw string, configured bool) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("must not be empty")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("not a valid URL")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", errors.New("scheme must be http or https")
	}
	if u.Host == "" || u.Hostname() == "" {
		return "", errors.New("missing host")
	}
	if u.User != nil {
		return "", errors.New("must not contain user info")
	}
	if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" {
		return "", errors.New("must not have a query or fragment")
	}
	if u.Path != "" && !(configured && u.Path == "/") {
		return "", errors.New("must not have a path")
	}
	host := strings.ToLower(u.Hostname())
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	port := u.Port()
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		port = ""
	}
	if port != "" {
		host += ":" + port
	}
	return scheme + "://" + host, nil
}
