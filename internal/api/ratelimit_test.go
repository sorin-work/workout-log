package api

import (
	"net/http"
	"testing"
	"time"
)

func TestRateLimiterWindow(t *testing.T) {
	now := time.Date(2026, 3, 3, 20, 0, 0, 0, time.UTC)
	l := newRateLimiter(2, time.Minute)
	l.now = func() time.Time { return now }
	for i := 0; i < 2; i++ {
		if ok, _ := l.allow("1.2.3.4"); !ok {
			t.Fatalf("request %d blocked", i+1)
		}
	}
	ok, retry := l.allow("1.2.3.4")
	if ok || retry != time.Minute {
		t.Fatalf("third request: ok=%v retry=%v", ok, retry)
	}
	if ok, _ := l.allow("5.6.7.8"); !ok {
		t.Fatal("other clients are not affected")
	}
	now = now.Add(time.Minute)
	if ok, _ := l.allow("1.2.3.4"); !ok {
		t.Fatal("window did not reset")
	}
}

func TestLoginIsRateLimited(t *testing.T) {
	a := newTestAPI(t)
	c := a.client()
	body := map[string]any{"email": "x@example.com", "password": "password123"}
	for i := 0; i < loginAttemptsPerMinute; i++ {
		c.mustDo(http.MethodPost, "/api/auth/login", body, http.StatusUnauthorized, nil)
	}
	res := c.do(http.MethodPost, "/api/auth/login", body)
	if res.Status != http.StatusTooManyRequests {
		t.Fatalf("status %d", res.Status)
	}
}
