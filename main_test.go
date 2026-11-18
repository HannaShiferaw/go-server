package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func resetStore() { submissions = &store{} }

func TestHelloHandler(t *testing.T) {
	t.Run("GET returns Hello", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		rec := httptest.NewRecorder()
		helloHandler(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if rec.Body.String() != "Hello!" {
			t.Errorf("body = %q, want %q", rec.Body.String(), "Hello!")
		}
	})

	t.Run("POST returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/hello", nil)
		rec := httptest.NewRecorder()
		helloHandler(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestFormHandler(t *testing.T) {
	t.Run("valid POST stores and echoes escaped values", func(t *testing.T) {
		resetStore()
		form := url.Values{"name": {"Hanna"}, "address": {"Addis Ababa"}}
		req := httptest.NewRequest(http.MethodPost, "/form", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		formHandler(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if !strings.Contains(rec.Body.String(), "Hanna") || !strings.Contains(rec.Body.String(), "Addis Ababa") {
			t.Errorf("body does not contain the submitted values: %s", rec.Body.String())
		}
		if got := submissions.all(); len(got) != 1 {
			t.Errorf("stored submissions = %d, want 1", len(got))
		}
	})

	t.Run("HTML in input is escaped", func(t *testing.T) {
		resetStore()
		form := url.Values{"name": {"<script>alert(1)</script>"}, "address": {"x"}}
		req := httptest.NewRequest(http.MethodPost, "/form", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		formHandler(rec, req)
		if strings.Contains(rec.Body.String(), "<script>") {
			t.Errorf("body contains unescaped HTML: %s", rec.Body.String())
		}
	})

	t.Run("GET returns 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/form", nil)
		rec := httptest.NewRecorder()
		formHandler(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("empty name returns 400", func(t *testing.T) {
		form := url.Values{"name": {""}, "address": {"Addis Ababa"}}
		req := httptest.NewRequest(http.MethodPost, "/form", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		formHandler(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("empty address returns 400", func(t *testing.T) {
		form := url.Values{"name": {"Hanna"}, "address": {"  "}}
		req := httptest.NewRequest(http.MethodPost, "/form", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		formHandler(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestSubmissionsHandler(t *testing.T) {
	resetStore()
	submissions.add(Submission{Name: "Hanna", Address: "Addis Ababa"})
	req := httptest.NewRequest(http.MethodGet, "/submissions", nil)
	rec := httptest.NewRecorder()
	submissionsHandler(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Hanna") {
		t.Errorf("body does not list the submission: %s", rec.Body.String())
	}
}

func TestPort(t *testing.T) {
	t.Setenv("PORT", "")
	if got := port(); got != "8080" {
		t.Errorf("port() with no env = %q, want 8080", got)
	}
	t.Setenv("PORT", "9090")
	if got := port(); got != "9090" {
		t.Errorf("port() with PORT=9090 = %q, want 9090", got)
	}
}
