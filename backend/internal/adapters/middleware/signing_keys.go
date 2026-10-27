package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	// jwksCacheTTL is how long a fetched key set is used. An expired set is never used.
	jwksCacheTTL = time.Hour
	// jwksFailureTTL is how long a failed fetch is remembered while there is no current key set. Until it
	// has passed, a token that needs a fetch answers 503 at once instead of starting another one. With a
	// current set, a failed refetch for an unknown key id is remembered for jwksKidRefetchInterval.
	jwksFailureTTL = 5 * time.Second
	// jwksKidRefetchInterval is how often a token whose key id the current set lacks may make the
	// middleware fetch the set again. Auth0 adds the next key to the set before it signs with it.
	jwksKidRefetchInterval = time.Minute
)

// keySetFetch is one fetch of the key set. A request that needs the key set while the fetch is in flight
// waits on done, unless it is answered from a failure that is already remembered. keys and err are set
// before done is closed.
type keySetFetch struct {
	done chan struct{}
	keys *JWKS
	err  error
}

// cacheTime is the key cache's clock: cacheNow, or time.Now when it is nil.
func (m *Auth0Middleware) cacheTime() time.Time {
	if m.cacheNow != nil {
		return m.cacheNow()
	}
	return time.Now()
}

// key returns a copy of the key whose id is kid, or nil when the set has none.
func (s *JWKS) key(kid string) *JWK {
	for i := range s.Keys {
		if s.Keys[i].Kid == kid {
			k := s.Keys[i]
			return &k
		}
	}
	return nil
}

// signingKey returns the key that kid names. A key set that was fetched without error and lacks kid
// answers errSigningKeyNotFound. When the key set cannot be had, the error wraps
// errSigningKeysUnavailable. That includes a current set whose refetch for an unknown kid failed.
//
// One fetch at a time: a request that needs the key set while a fetch is in flight waits for that
// fetch instead of starting its own, and no fetch runs under the cache lock, so requests whose key
// is cached are never held up by one. After a failed fetch, requests answer the remembered failure at
// once, and do not wait for a retry that is in flight. A new fetch waits jwksFailureTTL when there is no
// current set, and jwksKidRefetchInterval when an unknown kid asks for a refetch of a current one.
func (m *Auth0Middleware) signingKey(ctx context.Context, kid string) (*JWK, error) {
	m.cacheMu.Lock()
	now := m.cacheTime()
	fresh := m.jwksCache != nil && now.Before(m.cacheExp)
	if fresh {
		if key := m.jwksCache.key(kid); key != nil {
			m.cacheMu.Unlock()
			return key, nil
		}
	}

	// The key is not in a current key set, so it takes a fetch. How long after the last attempt a new
	// one may start: an unknown key id with a current set waits jwksKidRefetchInterval, a missing or
	// expired set waits jwksFailureTTL after a failure and nothing after a success.
	gap := time.Duration(0)
	switch {
	case fresh:
		gap = jwksKidRefetchInterval
	case m.lastFetchErr != nil:
		gap = jwksFailureTTL
	}

	f := m.fetching
	switch {
	case f != nil && m.lastFetchErr != nil:
		err := m.lastFetchErr
		m.cacheMu.Unlock()
		return nil, rememberedFetchFailure(err)
	case f != nil:
		m.cacheMu.Unlock()
		select {
		case <-f.done:
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %w", errSigningKeysUnavailable, ctx.Err())
		}
	case now.Before(m.lastFetch.Add(gap)):
		err := m.lastFetchErr
		m.cacheMu.Unlock()
		if err != nil {
			return nil, rememberedFetchFailure(err)
		}
		return nil, errSigningKeyNotFound
	default:
		f = &keySetFetch{done: make(chan struct{})}
		m.fetching = f
		m.cacheMu.Unlock()
		m.runFetch(ctx, f)
	}

	if f.err != nil {
		return nil, fmt.Errorf("%w: %w", errSigningKeysUnavailable, f.err)
	}
	if key := f.keys.key(kid); key != nil {
		return key, nil
	}
	return nil, errSigningKeyNotFound
}

// rememberedFetchFailure is the error for a request answered from a failure that is remembered.
func rememberedFetchFailure(err error) error {
	return fmt.Errorf("%w: previous fetch failed, not retried yet: %w", errSigningKeysUnavailable, err)
}

// runFetch fetches the key set for f and records the result. The fetch does not stop when the
// request that started it goes away, because other requests wait for it; the client's timeout bounds
// it. The deferred block runs even if the fetch panics, so no request is left waiting.
func (m *Auth0Middleware) runFetch(ctx context.Context, f *keySetFetch) {
	f.err = errors.New("key set fetch did not finish")
	defer func() {
		m.cacheMu.Lock()
		now := m.cacheTime()
		m.lastFetch = now
		m.lastFetchErr = f.err
		if f.err == nil {
			m.jwksCache = f.keys
			m.cacheExp = now.Add(jwksCacheTTL)
		}
		m.fetching = nil
		m.cacheMu.Unlock()
		close(f.done)
	}()
	f.keys, f.err = m.fetchJWKS(context.WithoutCancel(ctx))
}

// fetchJWKS fetches the tenant's key set. It does not touch the cache.
func (m *Auth0Middleware) fetchJWKS(ctx context.Context) (*JWKS, error) {
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", m.config.Domain)

	req, err := http.NewRequestWithContext(ctx, "GET", jwksURL, nil)
	if err != nil {
		return nil, err
	}

	client := m.client
	if client == nil {
		client = &http.Client{Timeout: jwksFetchTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS request failed with status: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, err
	}
	// A key set without a key cannot check any token, so it is as unusable as one that could not be
	// fetched: it counts as a failed fetch and is never cached.
	if len(jwks.Keys) == 0 {
		return nil, errors.New("key set has no keys")
	}
	return &jwks, nil
}
