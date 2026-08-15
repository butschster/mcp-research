package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// rateLimiter is a fixed-window counter keyed by an arbitrary string.
//
// It is deliberately small. The share prefix is the only unauthenticated read
// surface in the product, and what protects a link from being guessed is its
// 256 bits of randomness, not this — no rate a limiter can impose makes 2^256
// reachable, and no rate it fails to impose makes it easier. What this is for
// is the one place guessing is cheap: the password on a protected link, and the
// noise a public URL attracts once it has been passed around.
type rateLimiter struct {
	mu      sync.Mutex
	windows map[string]*rateWindow
	limit   int
	window  time.Duration
	// swept bounds the map. Buckets are dropped wholesale when it grows, which
	// costs a caller nothing but a fresh window.
	maxKeys int
}

type rateWindow struct {
	count   int
	resetAt time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		windows: make(map[string]*rateWindow),
		limit:   limit,
		window:  window,
		maxKeys: 16384,
	}
}

// allow records one hit against a key and reports whether it fits in the
// window.
func (l *rateLimiter) allow(key string) bool {
	if key == "" {
		return true
	}
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.windows) >= l.maxKeys {
		// Expired windows first. Dropping the map wholesale was a way in: an
		// attacker who invented sixteen thousand share tokens pushed the map
		// over the limit, wiped it, and got a fresh password budget against the
		// one link they actually cared about. Only if sweeping frees nothing
		// does anything live get dropped.
		for k, v := range l.windows {
			if now.After(v.resetAt) {
				delete(l.windows, k)
			}
		}
		if len(l.windows) >= l.maxKeys {
			l.windows = make(map[string]*rateWindow)
		}
	}

	w, ok := l.windows[key]
	if !ok || now.After(w.resetAt) {
		l.windows[key] = &rateWindow{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	w.count++
	return w.count <= l.limit
}

// limitByIP wraps a handler with a per-caller rate limit.
func (l *rateLimiter) limitByIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, map[string]any{"error": "too many requests"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP is who to charge a request to.
//
// Only the hop this server can vouch for. `X-Real-IP` is what the bundled nginx
// config sets from `$remote_addr`, and it is overwritten rather than appended,
// so a client cannot choose its own value.
//
// `X-Forwarded-For` is read from the **right**, not the left. The proxy builds
// it with `$proxy_add_x_forwarded_for`, which appends the real address to
// whatever the client sent — so the leftmost element is attacker-supplied and
// the rightmost is the proxy's own observation. Taking the left one, as this
// did, meant a caller could mint a fresh bucket per request by sending
// `X-Forwarded-For: 1.2.3.<random>` and the limiter counted nothing at all.
func clientIP(r *http.Request) string {
	if real := strings.TrimSpace(r.Header.Get("X-Real-IP")); real != "" {
		return real
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.LastIndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[i+1:])
		}
		return strings.TrimSpace(xff)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
