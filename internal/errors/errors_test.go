package errors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPErrorBasics(t *testing.T) {
	e := NotFound("missing")
	if e.Error() != "missing" {
		t.Fatalf("expected message 'missing', got %q", e.Error())
	}
	if e.HTTPCode() != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, e.HTTPCode())
	}

	// Is should compare by status code
	other := NotFound("nope")
	if !e.Is(other) {
		t.Fatalf("expected Is to be true for same status code")
	}

	// WriteHTTPError writes the error and status code
	rr := httptest.NewRecorder()
	WriteHTTPError(rr, e, "default")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected recorder code %d, got %d", http.StatusNotFound, rr.Code)
	}

	// Non-HTTPCoder should cause default internal server error
	rr2 := httptest.NewRecorder()
	WriteHTTPError(rr2, HttpErr("oops"), "fallback")
	if rr2.Code != http.StatusInternalServerError {
		t.Fatalf("expected fallback code %d, got %d", http.StatusInternalServerError, rr2.Code)
	}
}

// HttpErr implements error but not HTTPCoder
type HttpErr string

func (h HttpErr) Error() string { return string(h) }
