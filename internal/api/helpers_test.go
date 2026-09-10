package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sorin-work/workout-log/internal/store"
	"github.com/sorin-work/workout-log/internal/testdb"
)

// testAPI is a running server backed by an isolated database.
type testAPI struct {
	t     *testing.T
	srv   *httptest.Server
	store *store.Store
}

func newTestAPI(t *testing.T) *testAPI {
	t.Helper()
	st := store.New(testdb.New(t))
	srv := httptest.NewServer(NewServer(st, Options{}))
	t.Cleanup(srv.Close)
	return &testAPI{t: t, srv: srv, store: st}
}

// client is an HTTP client with its own cookie jar, i.e. one browser.
type client struct {
	t    *testing.T
	base string
	http *http.Client
}

func (a *testAPI) client() *client {
	jar, _ := cookiejar.New(nil)
	return &client{t: a.t, base: a.srv.URL, http: &http.Client{Jar: jar}}
}

// signup creates a user and returns a client logged in as that user.
func (a *testAPI) signup(email string) *client {
	a.t.Helper()
	c := a.client()
	c.mustDo(http.MethodPost, "/api/auth/signup", map[string]any{
		"email": email, "password": "password123", "displayName": "Test",
	}, http.StatusCreated, nil)
	return c
}

type response struct {
	Status int
	Body   []byte
}

func (r response) decode(t *testing.T, dst any) {
	t.Helper()
	if err := json.Unmarshal(r.Body, dst); err != nil {
		t.Fatalf("decode %s: %v", r.Body, err)
	}
}

// upload posts a raw body with a content type, e.g. a CSV file.
func (c *client) upload(path, contentType, body string) response {
	c.t.Helper()
	res, err := c.http.Post(c.base+path, contentType, strings.NewReader(body))
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return response{Status: res.StatusCode, Body: b}
}

func (c *client) do(method, path string, body any) response {
	c.t.Helper()
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.base+path, rd)
	if err != nil {
		c.t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return response{Status: res.StatusCode, Body: b}
}

// mustDo performs a request, checks the status and decodes the body into dst.
func (c *client) mustDo(method, path string, body any, wantStatus int, dst any) response {
	c.t.Helper()
	res := c.do(method, path, body)
	if res.Status != wantStatus {
		c.t.Fatalf("%s %s: status %d, want %d: %s", method, path, res.Status, wantStatus, res.Body)
	}
	if dst != nil {
		res.decode(c.t, dst)
	}
	return res
}

func errorFields(t *testing.T, res response) map[string]string {
	t.Helper()
	var e apiError
	res.decode(t, &e)
	return e.Fields
}
