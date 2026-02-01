package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServesWebApp(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>app</html>"), 0o644)
	os.WriteFile(filepath.Join(dir, "assets", "app-123.js"), []byte("console.log(1)"), 0o644)
	srv := NewServer(nil, Options{WebDir: dir})

	get := func(p string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		return rec
	}

	if rec := get("/"); !strings.Contains(rec.Body.String(), "app") {
		t.Fatalf("index: %d %s", rec.Code, rec.Body)
	}
	if rec := get("/workouts/12"); rec.Code != 200 || !strings.Contains(rec.Body.String(), "app") {
		t.Fatalf("client route should fall back to index.html: %d", rec.Code)
	}
	rec := get("/assets/app-123.js")
	if rec.Code != 200 || !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("asset: %d %v", rec.Code, rec.Header())
	}
	if rec := get("/api/nope"); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown api route = %d", rec.Code)
	}
	if rec := get("/../../etc/passwd"); strings.Contains(rec.Body.String(), "root:") {
		t.Fatal("path traversal")
	}
}
