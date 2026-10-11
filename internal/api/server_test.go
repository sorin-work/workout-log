package api

import (
	"net/http"
	"testing"
)

func TestHealth(t *testing.T) {
	a := newTestAPI(t)
	var body map[string]any
	a.client().mustDo(http.MethodGet, "/api/health", nil, http.StatusOK, &body)
	if body["status"] != "ok" || body["database"] != "ok" || body["schemaVersion"].(float64) < 10 {
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

func TestAPIResponsesAreCompressedAndNotCached(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("gz@example.com")
	req, _ := http.NewRequest(http.MethodGet, c.base+"/api/exercises", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	// Disable the transport's transparent decompression to see the header.
	tr := &http.Transport{DisableCompression: true}
	res, err := (&http.Client{Transport: tr, Jar: c.http.Jar}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q", res.Header.Get("Content-Encoding"))
	}
	if res.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q", res.Header.Get("Cache-Control"))
	}
}
