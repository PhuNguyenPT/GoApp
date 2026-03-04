package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRegisterPageHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/register", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %v", ct)
	}
}

func TestValidationMessage(t *testing.T) {
	tests := []struct {
		name     string
		form     url.Values
		wantBody string
	}{
		{
			name:     "invalid email",
			form:     url.Values{"name": {"Test"}, "email": {"not-an-email"}, "password": {"password123"}},
			wantBody: "valid email address",
		},
		{
			name:     "missing name",
			form:     url.Values{"email": {"test@example.com"}, "password": {"password123"}},
			wantBody: "required",
		},
		{
			name:     "name too long",
			form:     url.Values{"name": {strings.Repeat("a", 101)}, "email": {"test@example.com"}, "password": {"password123"}},
			wantBody: "at most 100 characters",
		},
		{
			name:     "email too long",
			form:     url.Values{"name": {"Test"}, "email": {strings.Repeat("a", 246) + "@test.com"}, "password": {"password123"}},
			wantBody: "at most 254 characters",
		},
		{
			name:     "password too long",
			form:     url.Values{"name": {"Test"}, "email": {"test@example.com"}, "password": {strings.Repeat("a", 73)}},
			wantBody: "at most 72 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.form.Encode()))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)
			if !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got: %s", tt.wantBody, rr.Body.String())
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Test User")
		form.Set("email", "newuser@example.com")
		form.Set("password", "password123")

		req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}

		cookies := rr.Result().Cookies()
		found := false
		for _, c := range cookies {
			if c.Name == "session_token" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected session_token cookie to be set")
		}

		if !strings.Contains(rr.Body.String(), "/dashboard") {
			t.Errorf("expected dashboard link in success response")
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Test User")
		// missing email and password
		req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "required") {
			t.Errorf("expected validation error message")
		}
	})

	t.Run("password too short", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Test User")
		form.Set("email", "test@example.com")
		form.Set("password", "short")

		req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "at least 8 characters") {
			t.Errorf("expected password length error message")
		}
	})
	t.Run("name too long", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", strings.Repeat("a", 101))
		form.Set("email", "test@example.com")
		form.Set("password", "password123")

		req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "at most 100 characters") {
			t.Errorf("expected name max length error message")
		}
	})

	t.Run("password too long", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Test User")
		form.Set("email", "test@example.com")
		form.Set("password", strings.Repeat("a", 73))

		req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "at most 72 characters") {
			t.Errorf("expected password max length error message")
		}
	})
}

func TestLoginPageHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/login", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %v", ct)
	}
}

func TestLoginHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		form := url.Values{}
		form.Set("email", "test@example.com")
		form.Set("password", "password123")

		req, err := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}

		// Check session cookie is set
		cookies := rr.Result().Cookies()
		var found bool
		for _, c := range cookies {
			if c.Name == "session_token" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected session_token cookie to be set")
		}

		// Check success response renders dashboard link
		if !strings.Contains(rr.Body.String(), "/dashboard") {
			t.Errorf("expected dashboard link in success response")
		}
	})

	t.Run("missing fields", func(t *testing.T) {
		form := url.Values{}
		form.Set("email", "test@example.com")

		req, err := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "Invalid email or password") {
			t.Errorf("expected error message in body")
		}

		if rr.Header().Get("HX-Redirect") != "" {
			t.Errorf("unexpected redirect on failed login")
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		form := url.Values{}
		form.Set("email", "not-an-email")
		form.Set("password", "password123")

		req, err := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "Invalid email or password") {
			t.Errorf("expected error message")
		}
	})
}

func TestLogoutHandler(t *testing.T) {
	t.Run("clears cookie and redirects", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/logout", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusFound {
			t.Errorf("expected redirect 302, got %v", rr.Code)
		}
		if rr.Header().Get("Location") != "/login" {
			t.Errorf("expected redirect to /login, got %v", rr.Header().Get("Location"))
		}
	})
}

func TestRevokeSessionHandler(t *testing.T) {
	t.Run("unauthenticated redirects to login", func(t *testing.T) {
		token, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("failed to generate uuid: %v", err)
		}
		req, err := http.NewRequest(http.MethodDelete, "/dashboard/session/"+token.String(), nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusFound {
			t.Errorf("expected redirect 302, got %v", rr.Code)
		}
		if rr.Header().Get("Location") != "/login" {
			t.Errorf("expected redirect to /login, got %v", rr.Header().Get("Location"))
		}
	})

	t.Run("invalid session_id returns 400", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "/dashboard/session/not-a-valid-uuid", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %v", rr.Code)
		}
	})

	t.Run("valid session_id returns 200", func(t *testing.T) {
		token, err := uuid.NewV7()
		if err != nil {
			t.Fatalf("failed to generate uuid: %v", err)
		}
		req, err := http.NewRequest(http.MethodDelete, "/dashboard/session/"+token.String(), nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
	})
}
