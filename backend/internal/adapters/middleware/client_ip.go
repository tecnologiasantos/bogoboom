package middleware

import "github.com/gofiber/fiber/v3"

// ClientIPHeader is the only header a trusted proxy may use to report the
// client address. Both proxies we run (Traefik and nginx) overwrite it with
// the address they saw. X-Forwarded-For is deliberately never used: it is a
// chain that proxies append to, so its leading entries are not written by the
// proxy.
const ClientIPHeader = "X-Real-IP"

// ConfigureClientIP makes c.IP() resolve the client address from the
// connection, or from ClientIPHeader when the connection comes from one of
// trustedProxies (already validated IP addresses or CIDR ranges). With an
// empty list no peer is trusted and c.IP() is the connection address.
//
// TrustProxy stays true even with an empty list: in Fiber v3.0.0-rc.3 a
// ProxyHeader combined with TrustProxy == false trusts every peer.
func ConfigureClientIP(cfg *fiber.Config, trustedProxies []string) {
	cfg.TrustProxy = true
	cfg.ProxyHeader = ClientIPHeader
	cfg.EnableIPValidation = true
	cfg.TrustProxyConfig = fiber.TrustProxyConfig{
		Proxies: append([]string(nil), trustedProxies...),
	}
}
