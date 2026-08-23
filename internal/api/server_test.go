package api

import (
	"net/http"
	"testing"
)

func TestHealth(t *testing.T) {
	a := newTestAPI(t)
	var body map[string]string
	a.client().mustDo(http.MethodGet, "/api/health", nil, http.StatusOK, &body)
	if body["status"] != "ok" || body["database"] != "ok" {
		t.Fatalf("body = %v", body)
	}
}

func TestUnknownAPIRoute(t *testing.T) {
	a := newTestAPI(t)
	res := a.client().do(http.MethodGet, "/api/does-not-exist", nil)
	if res.Status != http.StatusNotFound {
		t.Fatalf("status %d", res.Status)
	}
}

func TestVersion(t *testing.T) {
	a := newTestAPI(t)
	var body map[string]string
	a.client().mustDo(http.MethodGet, "/api/version", nil, http.StatusOK, &body)
	if body["version"] != "dev" {
		t.Fatalf("version = %q", body["version"])
	}
}
