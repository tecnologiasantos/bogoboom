package config

import (
	"fmt"
	"net"
	"strings"
)

// ParseTrustedProxies parses a comma-separated list of IP addresses and CIDR
// ranges of reverse proxies allowed to report the client address. Entries are
// returned in canonical form (the form Fiber matches against), in input order.
// Empty or blank input means no proxy is trusted and returns an empty list.
// Hostnames, malformed entries and ranges that trust every address (prefix
// length 0) are rejected.
func ParseTrustedProxies(raw string) ([]string, error) {
	var proxies []string
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		if strings.Contains(entry, "/") {
			_, ipNet, err := net.ParseCIDR(entry)
			if err != nil {
				return nil, fmt.Errorf("TRUSTED_PROXIES: %q is not an IP address or CIDR range", entry)
			}
			// Check the canonical form: an IPv4-mapped range such as ::ffff:0.0.0.0/96
			// canonicalises to 0.0.0.0/0 and must be rejected like any trust-all range.
			canonical := ipNet.String()
			if _, canonicalNet, err := net.ParseCIDR(canonical); err == nil {
				if ones, _ := canonicalNet.Mask.Size(); ones == 0 {
					return nil, fmt.Errorf("TRUSTED_PROXIES: %q trusts every address; list the proxy addresses instead", entry)
				}
			}
			proxies = append(proxies, canonical)
			continue
		}

		ip := net.ParseIP(entry)
		if ip == nil {
			return nil, fmt.Errorf("TRUSTED_PROXIES: %q is not an IP address or CIDR range", entry)
		}
		proxies = append(proxies, ip.String())
	}
	return proxies, nil
}

// LoadTrustedProxies reads and validates the TRUSTED_PROXIES environment
// variable. Unset or empty means no proxy is trusted.
func LoadTrustedProxies() ([]string, error) {
	return ParseTrustedProxies(getEnv("TRUSTED_PROXIES", ""))
}
