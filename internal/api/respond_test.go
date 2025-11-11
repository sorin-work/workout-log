package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	type body struct {
		Name string `json:"name"`
	}
	cases := []struct {
		name   string
		input  string
		ok     bool
		substr string
	}{
		{"valid", `{"name":"squat"}`, true, ""},
		{"empty", ``, false, "empty"},
		{"syntax", `{"name":`, false, "Malformed"},
		{"unknown field", `{"nam":"x"}`, false, "Unknown field"},
		{"wrong type", `{"name":5}`, false, "wrong type"},
		{"trailing data", `{"name":"a"}{}`, false, "single JSON object"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.input))
			var dst body
			ok := decodeJSON(rec, req, &dst)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v (body %s)", ok, tc.ok, rec.Body.String())
			}
			if !ok && !strings.Contains(rec.Body.String(), tc.substr) {
				t.Fatalf("body %s does not mention %q", rec.Body.String(), tc.substr)
			}
		})
	}
}
