package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadinessEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/readiness", nil)
	w := httptest.NewRecorder()

	readinessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	expected := "OK"
	if w.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, w.Body.String())
	}
}