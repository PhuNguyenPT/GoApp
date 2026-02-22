package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestAuthHeaderHandler(t *testing.T) {
	t.Run("unauthenticated returns sign in fragment", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/auth-header", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "/register") {
			t.Errorf("expected register link in unauthenticated fragment")
		}
		if !strings.Contains(rr.Body.String(), "Get Started") {
			t.Errorf("expected Get Started button in unauthenticated fragment")
		}
	})

	t.Run("authenticated returns user fragment", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/auth-header", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		// mockDB returns "Test User" → first letter "T"
		if !strings.Contains(rr.Body.String(), "T") {
			t.Errorf("expected user initial in authenticated fragment")
		}
		if !strings.Contains(rr.Body.String(), "/logout") {
			t.Errorf("expected logout link in authenticated fragment")
		}
		if !strings.Contains(rr.Body.String(), "/dashboard") {
			t.Errorf("expected dashboard link in authenticated fragment")
		}
	})
}

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

		if rr.Header().Get("HX-Redirect") != "/dashboard" {
			t.Errorf("expected HX-Redirect to /dashboard, got %v", rr.Header().Get("HX-Redirect"))
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

		if got := rr.Header().Get("HX-Redirect"); got != "/dashboard" {
			t.Errorf("expected HX-Redirect to /dashboard, got %q", got)
		}

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
