package api

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// rateLimiter allows a fixed number of requests per key in each window.
// It is in memory, which is fine for a single instance.
type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	now     func() time.Time
	buckets map[string]*bucket
}

type bucket struct {
	start time.Time
	count int
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, window: window, now: time.Now, buckets: map[string]*bucket{}}
}

// allow records a request for key and reports whether it is within the limit,
// and if not, how long until the window resets.
func (l *rateLimiter) allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	b := l.buckets[key]
	if b == nil || now.Sub(b.start) >= l.window {
		if len(l.buckets) > 10000 {
			l.sweep(now)
		}
		b = &bucket{start: now}
		l.buckets[key] = b
	}
	b.count++
	if b.count > l.limit {
		return false, b.start.Add(l.window).Sub(now)
	}
	return true, 0
}

func (l *rateLimiter) sweep(now time.Time) {
	for k, b := range l.buckets {
		if now.Sub(b.start) >= l.window {
			delete(l.buckets, k)
		}
	}
}

// middleware limits requests per client IP.
func (l *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		if ok, retry := l.allow(ip); !ok {
			w.Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())+1))
			writeError(w, http.StatusTooManyRequests, "rate_limited", "Too many attempts. Try again in a minute.")
			return
		}
		next.ServeHTTP(w, r)
	})
}
