package server

import (
	"GoApp/internal/database"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockDB struct {
	deleteExpiredSessionsCalled int
}

func (m *mockDB) Health() map[string]string {
	return map[string]string{"status": "up", "message": "It's healthy"}
}

func (m *mockDB) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.User, error) {
	return database.User{Name: arg.Name, Email: arg.Email}, nil
}

func (m *mockDB) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	return database.User{
		Name:         "Test User",
		Email:        email,
		PasswordHash: string(hash),
	}, nil
}

func (m *mockDB) GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error) {
	return database.User{Name: "Test User", Email: "test@example.com"}, nil
}

func (m *mockDB) CreateSession(ctx context.Context, arg database.CreateSessionParams) (database.Session, error) {
	return database.Session{Token: arg.Token}, nil
}

func (m *mockDB) GetSessionByToken(ctx context.Context, token string) (database.Session, error) {
	return database.Session{}, nil
}

func (m *mockDB) DeleteSession(ctx context.Context, token string) error {
	return nil
}

func (m *mockDB) DeleteExpiredSessions(ctx context.Context) error {
	m.deleteExpiredSessionsCalled++
	return nil
}

func (m *mockDB) GetActiveSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]database.Session, error) {
	return []database.Session{
		{UserAgent: "Mozilla/5.0 Test Browser"},
	}, nil
}

var testHandler http.Handler

func TestMain(m *testing.M) {
	if err := os.Chdir("../../"); err != nil {
		log.Fatalf("failed to change directory: %v", err)
	}
	s := &Server{
		db:  &mockDB{},
		cfg: &Config{AppEnv: EnvTest, GinMode: gin.TestMode},
	}
	testHandler = s.RegisterRoutes(s.cfg)
	os.Exit(m.Run())
}

func TestApiInfoHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON %v", err)
	}

	if body["message"] != "Go Server API" {
		t.Errorf("unexpected message: %v", body["message"])
	}

	if body["version"] != "0.0" {
		t.Errorf("unexpected version: %v", body["version"])
	}
}

func TestHealthHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if body["status"] != "up" {
		t.Errorf("expected status up, got %v", body["status"])
	}

	if body["message"] != "It's healthy" {
		t.Errorf("expected healthy message, got %v", body["message"])
	}
}

func TestHomePageHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %v", ct)
	}
}

func TestContactFormHandler(t *testing.T) {
	form := url.Values{}
	form.Set("name", "Test Name")
	form.Set("email", "test@example.com")
	form.Set("subject", "Test Subject")
	form.Set("message", "Test message")

	req, _ := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Test Nam") {
		t.Errorf("expected response body to contain 'Test Name'")
	}
}

func TestContactPageHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/contact", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %v", ct)
	}
}

func TestUnknownRoute(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %v", rr.Code)
	}
}

func TestRegisterPageHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/register", nil)
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("expected HTML content type, got %v", ct)
	}
}

func TestRegisterHandler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Test User")
		form.Set("email", "newuser@example.com")
		form.Set("password", "password123")

		req, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
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
		req, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
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

		req, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
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
			req, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)
			if !strings.Contains(rr.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got: %s", tt.wantBody, rr.Body.String())
			}
		})
	}
}

func TestLoginPageHandler(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
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

		req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
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

		req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
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

		req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "Invalid email or password") {
			t.Errorf("expected error message")
		}
	})
}

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
		req, _ := http.NewRequest(http.MethodGet, "/dashboard", nil)
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
		req, _ := http.NewRequest(http.MethodGet, "/dashboard", nil)
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
		if !strings.Contains(rr.Body.String(), "1") {
			t.Errorf("expected active session count in dashboard body")
		}
	})
}

func TestLogoutHandler(t *testing.T) {
	t.Run("clears cookie and redirects", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/logout", nil)
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

func TestAuthHeaderHandler(t *testing.T) {
	t.Run("unauthenticated returns sign in fragment", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/auth-header", nil)
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "/login") {
			t.Errorf("expected sign in link in unauthenticated fragment")
		}
		if !strings.Contains(rr.Body.String(), "/register") {
			t.Errorf("expected register link in unauthenticated fragment")
		}
	})

	t.Run("authenticated returns user fragment", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/auth-header", nil)
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
