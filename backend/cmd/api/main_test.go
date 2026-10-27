package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"bogoboom-backend/internal/adapters/handler"
	"bogoboom-backend/internal/config"
)

func testApp(t *testing.T) *fiber.App {
	t.Helper()
	origins, err := handler.NewOriginPolicy("https://bogoboom.test")
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Env: "development", Server: config.ServerConfig{FrontendURL: "https://bogoboom.test", ReadTimeout: 5, WriteTimeout: 5},
		Auth0: config.Auth0Config{Domain: "bogoboom-test.auth0.com", Audience: "https://api.bogoboom.test"}}
	return buildApp(cfg, nil, nil, origins)
}

// doRaw returns the transport error too: the server may refuse a request before routing it.
func doRaw(app *fiber.App, method, path, body string, headers map[string]string) (*http.Response, error) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return app.Test(req)
}

func do(t *testing.T, app *fiber.App, method, path, body string, headers map[string]string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

// publicRoutes are the only routes reachable without a token. Adding a route here is a security
// decision: it must be reviewed.
var publicRoutes = map[string]bool{
	"GET /health":                    true,
	"GET /api/v1":                    true,
	"GET /api/v1/":                   true,
	"POST /api/v1/partner-inquiries": true,
	"POST /api/v1/leads":             true,
}

func TestEveryOtherRouteRequiresAToken(t *testing.T) {
	app := testApp(t)
	checked := 0
	for _, r := range app.GetRoutes(true) {
		if r.Method == "HEAD" || r.Method == "OPTIONS" || r.Method == "CONNECT" || r.Method == "TRACE" || r.Method == "PATCH" {
			continue
		}
		key := r.Method + " " + r.Path
		if publicRoutes[key] {
			continue
		}
		path := strings.NewReplacer(":id", "00000000-0000-0000-0000-000000000001").Replace(r.Path)
		resp := do(t, app, r.Method, path, "", nil)
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("%s answered %d without a token, want 401", key, resp.StatusCode)
		}
		checked++
	}
	if checked < 20 {
		t.Fatalf("only %d protected routes found: the route table changed unexpectedly", checked)
	}
}

func TestRejectsGarbageTokens(t *testing.T) {
	app := testApp(t)
	for _, h := range []string{"Bearer", "Bearer x", "Bearer a.b.c", "Basic abc", "bearer " + strings.Repeat("a", 5000)} {
		resp, err := doRaw(app, "GET", "/api/v1/me", "", map[string]string{"Authorization": h})
		if err != nil {
			continue // refused by the HTTP server (oversized header): never reaches a handler
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("Authorization %.20q: status %d", h, resp.StatusCode)
		}
	}
	// A token in the URL is never accepted.
	if resp := do(t, app, "GET", "/api/v1/me?access_token=x.y.z", "", nil); resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("query-string token: status %d", resp.StatusCode)
	}
}

func TestHealthAndSecurityHeaders(t *testing.T) {
	resp := do(t, testApp(t), "GET", "/health", "", nil)
	if resp.StatusCode != 200 {
		t.Fatalf("health: %d", resp.StatusCode)
	}
	for k, want := range map[string]string{"X-Content-Type-Options": "nosniff", "X-Frame-Options": "DENY", "Cache-Control": "no-store"} {
		if got := resp.Header.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if resp.Header.Get("Content-Security-Policy") == "" || resp.Header.Get("Strict-Transport-Security") == "" {
		t.Error("CSP and HSTS must be set")
	}
}

func TestCORSAllowsOnlyTheFrontend(t *testing.T) {
	app := testApp(t)
	ok := do(t, app, "OPTIONS", "/api/v1/me", "", map[string]string{
		"Origin": "https://bogoboom.test", "Access-Control-Request-Method": "GET", "Access-Control-Request-Headers": "authorization"})
	if got := ok.Header.Get("Access-Control-Allow-Origin"); got != "https://bogoboom.test" {
		t.Errorf("frontend origin: %q", got)
	}
	bad := do(t, app, "OPTIONS", "/api/v1/me", "", map[string]string{
		"Origin": "https://evil.example", "Access-Control-Request-Method": "GET"})
	if got := bad.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("foreign origin must get no CORS grant, got %q", got)
	}
}

func TestPublicFormsValidateBeforeTouchingTheDatabase(t *testing.T) {
	app := testApp(t) // nil pool: any query would panic, so these paths must stop earlier
	resp := do(t, app, "POST", "/api/v1/partner-inquiries", `{"fullName":"","email":"bad"}`, nil)
	if resp.StatusCode != 400 {
		t.Errorf("invalid inquiry: %d", resp.StatusCode)
	}
	// The honeypot field makes the answer look successful and stores nothing.
	resp = do(t, app, "POST", "/api/v1/partner-inquiries", `{"fullName":"Bot","email":"b@x.co","collaborationType":"hr","acceptPrivacy":true,"website":"http://spam"}`, nil)
	if resp.StatusCode != 202 {
		t.Errorf("honeypot: %d", resp.StatusCode)
	}
	resp = do(t, app, "POST", "/api/v1/leads", `{"name":"Bot","email":"b@x.co","acceptPrivacy":true,"website":"x"}`, nil)
	if resp.StatusCode != 202 {
		t.Errorf("lead honeypot: %d", resp.StatusCode)
	}
	if resp := do(t, app, "POST", "/api/v1/leads", `not json`, nil); resp.StatusCode != 400 {
		t.Errorf("bad json: %d", resp.StatusCode)
	}
}

func TestBodyLimit(t *testing.T) {
	big := `{"name":"` + strings.Repeat("a", 70*1024) + `"}`
	resp, err := doRaw(testApp(t), "POST", "/api/v1/leads", big, nil)
	if err != nil {
		return // refused by the body limit before reaching a handler
	}
	if resp.StatusCode != fiber.StatusRequestEntityTooLarge {
		t.Errorf("oversized body: %d", resp.StatusCode)
	}
}
