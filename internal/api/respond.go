package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sorin-work/workout-log/internal/store"
)

// maxBodyBytes caps request bodies; workouts with many sets stay far below it.
const maxBodyBytes = 1 << 20

// apiError is the JSON error body returned by every endpoint.
type apiError struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, apiError{Error: code, Message: message})
}

func writeValidation(w http.ResponseWriter, fields map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, apiError{
		Error:   "validation_failed",
		Message: "Some fields are invalid.",
		Fields:  fields,
	})
}

func writeInternal(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("request failed", "method", r.Method, "path", r.URL.Path,
		"request_id", middleware.GetReqID(r.Context()), "err", err)
	writeError(w, http.StatusInternalServerError, "internal", "Something went wrong.")
}

// writeInternalLogOnly logs an error that happened after the response
// started streaming.
func writeInternalLogOnly(r *http.Request, err error) {
	slog.Error("request failed mid-stream", "method", r.Method, "path", r.URL.Path, "err", err)
}

func writeNotFound(w http.ResponseWriter, what string) {
	writeError(w, http.StatusNotFound, "not_found", what+" not found.")
}

// decodeJSON reads a JSON body into dst, rejecting unknown fields and
// trailing data. On failure it writes a 400 response and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", describeDecodeError(err))
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "bad_request", "Request body must contain a single JSON object.")
		return false
	}
	return true
}

func describeDecodeError(err error) string {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var maxErr *http.MaxBytesError
	switch {
	case errors.Is(err, io.EOF):
		return "Request body is empty."
	case errors.As(err, &syntaxErr):
		return fmt.Sprintf("Malformed JSON at position %d.", syntaxErr.Offset)
	case errors.As(err, &typeErr):
		return fmt.Sprintf("Field %q has the wrong type.", typeErr.Field)
	case errors.As(err, &maxErr):
		return "Request body is too large."
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return "Unknown field " + strings.TrimPrefix(err.Error(), "json: unknown field ") + "."
	default:
		return "Malformed JSON."
	}
}

// pathID parses a positive integer URL parameter. On failure it writes a 404
// and returns false.
func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusNotFound, "not_found", "Not found.")
		return 0, false
	}
	return id, true
}

// writeStoreError turns a store error into a response: 404 for ErrNotFound,
// 500 for anything else.
func writeStoreError(w http.ResponseWriter, r *http.Request, err error, what string) {
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, what)
		return
	}
	writeInternal(w, r, err)
}

// textLength counts characters, not bytes, so limits shown to users as
// "at most N characters" hold for accented and non-Latin text too.
func textLength(s string) int {
	return utf8.RuneCountInString(s)
}
