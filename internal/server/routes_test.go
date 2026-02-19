package server

import (
	"GoApp/internal/database"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type mockDB struct{}

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

var testHandler http.Handler

func TestMain(m *testing.M) {
	s := &Server{db: &mockDB{}}
	testHandler = s.RegisterRoutes(&Config{GinMode: gin.TestMode})
	m.Run()
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
		form.Set("email", "test@example.com")
		form.Set("password", "password123")

		req, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Test User") {
			t.Errorf("expected response to contain user name")
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
    })

    t.Run("missing fields", func(t *testing.T) {
        form := url.Values{}
        form.Set("email", "test@example.com")

        req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
        req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
        rr := httptest.NewRecorder()
        testHandler.ServeHTTP(rr, req)

        if !strings.Contains(rr.Body.String(), "Invalid email or password") {
            t.Errorf("expected error message")
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