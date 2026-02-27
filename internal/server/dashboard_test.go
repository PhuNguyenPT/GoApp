package server

import (
	"GoApp/internal/database"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"user@example.com", "u***@example.com"},
		{"a@example.com", "***@example.com"},   // single char local part
		{"@example.com", "***@example.com"},    // empty local part
		{"ab@example.com", "a***@example.com"}, // two char local part
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := maskEmail(tt.input)
			if got != tt.want {
				t.Errorf("maskEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDashboardPageHandler(t *testing.T) {
	t.Run("unauthenticated redirects to login", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/dashboard", nil)
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

	t.Run("authenticated shows dashboard", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/dashboard", nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
			t.Errorf("expected HTML content type, got %v", ct)
		}
		// mockDB returns "test@example.com" → masked to "t***@example.com"
		if !strings.Contains(rr.Body.String(), "t***@example.com") {
			t.Errorf("expected masked email in dashboard body")
		}
		if !strings.Contains(rr.Body.String(), "Mozilla/5.0 Test Browser") {
			t.Errorf("expected session user agent in dashboard body")
		}
		if !strings.Contains(rr.Body.String(), "Active Sessions") {
			t.Errorf("expected active sessions count in dashboard body")
		}
	})
}

func TestUpdateUserNameHandler(t *testing.T) {
	t.Run("unauthenticated redirects to login", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "New Name")

		req, err := http.NewRequest(http.MethodPut, "/dashboard/name", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusFound {
			t.Errorf("expected redirect 302, got %v", rr.Code)
		}
		if rr.Header().Get("Location") != "/login" {
			t.Errorf("expect redirect to /login, got %v", rr.Header().Get("Location"))
		}
	})
	t.Run("missing name", func(t *testing.T) {
		form := url.Values{}

		req, err := http.NewRequest(http.MethodPut, "/dashboard/name", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "required") {
			t.Errorf("expected validation error in body")
		}
	})
	t.Run("success", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "New Name")

		req, err := http.NewRequest(http.MethodPut, "/dashboard/name", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "successfully") {
			t.Errorf("expected success message in body")
		}
		if !strings.Contains(rr.Body.String(), "Welcome, New Name!") {
			t.Errorf("expected updated welcome heading in body")
		}
		if !strings.Contains(rr.Body.String(), "hx-swap-oob") {
			t.Errorf("expected oob swap attributes in body")
		}
	})
}

type mockDBWithPassword struct {
	mockDB
	passwordHash string
}

func (m *mockDBWithPassword) GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error) {
	return database.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: m.passwordHash,
	}, nil
}
func TestUpdateUserPasswordHandler(t *testing.T) {
	t.Run("unauthenticated redirects to login", func(t *testing.T) {
		form := url.Values{}
		form.Set("current_password", "password123")
		form.Set("new_password", "newpassword123")
		form.Set("confirm_password", "newpassword123")

		req, err := http.NewRequest(http.MethodPut, "/dashboard/password", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusFound {
			t.Errorf("expected redirect 302, got %v", rr.Code)
		}
		if rr.Header().Get("Location") != "/login" {
			t.Errorf("expected redirect to /login, got %v", rr.Header().Get("Location"))
		}
	})
	t.Run("missing fields", func(t *testing.T) {
		form := url.Values{}
		form.Set("current_password", "password123")
		// missing new_password and confirm_password

		req, err := http.NewRequest(http.MethodPut, "/dashboard/password", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "required") {
			t.Errorf("expected required error in body")
		}
	})
	t.Run("passwords do not match", func(t *testing.T) {
		form := url.Values{}
		form.Set("current_password", "password123")
		form.Set("new_password", "newpassword123")
		form.Set("confirm_password", "differentpassword")

		req, err := http.NewRequest(http.MethodPut, "/dashboard/password", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "do not match") {
			t.Errorf("expected password mismatch error in body")
		}
	})

	t.Run("new password too short", func(t *testing.T) {
		form := url.Values{}
		form.Set("current_password", "password123")
		form.Set("new_password", "short")
		form.Set("confirm_password", "short")

		req, err := http.NewRequest(http.MethodPut, "/dashboard/password", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "at least 8 characters") {
			t.Errorf("expected password length error in body")
		}
	})

	t.Run("wrong current password", func(t *testing.T) {
		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("failed to generate password hash: %v", err)
		}
		s := &Server{
			db:  &mockDBWithPassword{passwordHash: string(hash)},
			cfg: &Config{AppEnv: EnvTest, GinMode: gin.TestMode},
		}
		handler := s.RegisterRoutes(s.cfg)

		form := url.Values{}
		form.Set("current_password", "wrongpassword")
		form.Set("new_password", "newpassword123")
		form.Set("confirm_password", "newpassword123")

		req, err := http.NewRequest(http.MethodPut, "/dashboard/password", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "incorrect") {
			t.Errorf("expected incorrect password error in body")
		}
	})

	t.Run("success", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		s := &Server{
			db:  &mockDBWithPassword{passwordHash: string(hash)},
			cfg: &Config{AppEnv: EnvTest, GinMode: gin.TestMode},
		}
		handler := s.RegisterRoutes(s.cfg)

		form := url.Values{}
		form.Set("current_password", "password123")
		form.Set("new_password", "newpassword123")
		form.Set("confirm_password", "newpassword123")

		req, err := http.NewRequest(http.MethodPut, "/dashboard/password", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "successfully") {
			t.Errorf("expected success message in body")
		}
	})
}
