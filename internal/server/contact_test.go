package server

import (
	"GoApp/internal/database"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestContactPageHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/contact", nil)
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

type mockDBWithContactError struct{ mockDB }

func (m *mockDBWithContactError) CreateContact(ctx context.Context, arg database.CreateContactParams) (database.Contact, error) {
	return database.Contact{}, fmt.Errorf("db error")
}

type mockDBRateLimited struct{ mockDB }

func (m *mockDBRateLimited) CountContactsByIPToday(ctx context.Context, ipAddress string) (int64, error) {
	return maxContactsPerIPPerDay, nil
}

type mockDBEmailRateLimited struct{ mockDB }

func (m *mockDBEmailRateLimited) CountContactsByEmailToday(ctx context.Context, email string) (int64, error) {
	return maxContactsPerEmailPerDay, nil
}

func TestContactFormHandler(t *testing.T) {
	t.Run("success saves contact and renders success", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Test Name")
		form.Set("email", "test@example.com")
		form.Set("subject", "Test Subject")
		form.Set("message", "Test message")

		req, err := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
			t.Errorf("expected HTML content type, got %v", ct)
		}
		if !strings.Contains(rr.Body.String(), "Test Name") {
			t.Errorf("expected response body to contain 'Test Name'")
		}
	})

	t.Run("empty form returns fail view", func(t *testing.T) {
		form := url.Values{}

		req, err := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Something went wrong") {
			t.Errorf("expected error message in body")
		}
	})

	t.Run("name too long returns fail view", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", strings.Repeat("a", 101))
		form.Set("email", "test@example.com")
		form.Set("subject", "Test Subject")
		form.Set("message", "Test message")

		req, _ := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "Something went wrong") {
			t.Errorf("expected error message for name too long")
		}
	})

	t.Run("email too long returns fail view", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Test Name")
		form.Set("email", strings.Repeat("a", 250)+"@x.com") // > 254 bytes
		form.Set("subject", "Test Subject")
		form.Set("message", "Test message")

		req, _ := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "Something went wrong") {
			t.Errorf("expected error message for email too long")
		}
	})

	t.Run("subject too long returns fail view", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Test Name")
		form.Set("email", "test@example.com")
		form.Set("subject", strings.Repeat("a", 151))
		form.Set("message", "Test message")

		req, _ := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "Something went wrong") {
			t.Errorf("expected error message for subject too long")
		}
	})

	t.Run("message too long returns fail view", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", "Test Name")
		form.Set("email", "test@example.com")
		form.Set("subject", "Test Subject")
		form.Set("message", strings.Repeat("a", 5001))

		req, _ := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "Something went wrong") {
			t.Errorf("expected error message for message too long")
		}
	})

	t.Run("unicode name at boundary is valid", func(t *testing.T) {
		form := url.Values{}
		form.Set("name", strings.Repeat("中", 100)) // 100 runes, but 300 bytes
		form.Set("email", "test@example.com")
		form.Set("subject", "Test Subject")
		form.Set("message", "Test message")

		req, _ := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		testHandler.ServeHTTP(rr, req)

		if !strings.Contains(rr.Body.String(), "中") {
			t.Errorf("expected success for 100 unicode runes in name")
		}
	})

	t.Run("db error renders fail view", func(t *testing.T) {
		s := &Server{
			db:  &mockDBWithContactError{},
			cfg: &Config{AppEnv: EnvTest, GinMode: gin.TestMode},
		}
		handler := s.RegisterRoutes(s.cfg)

		form := url.Values{}
		form.Set("name", "Test Name")
		form.Set("email", "test@example.com")
		form.Set("subject", "Test Subject")
		form.Set("message", "Test message")

		req, err := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status %v, got %v", http.StatusOK, rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "Something went wrong") {
			t.Errorf("expected error message in body")
		}
	})

	t.Run("ip rate limited returns rate limit view", func(t *testing.T) {
		s := &Server{
			db:  &mockDBRateLimited{},
			cfg: &Config{AppEnv: EnvTest, GinMode: gin.TestMode},
		}
		handler := s.RegisterRoutes(s.cfg)

		form := url.Values{}
		form.Set("name", "Test Name")
		form.Set("email", "test@example.com")
		form.Set("subject", "Test Subject")
		form.Set("message", "Test message")

		req, err := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "maximum number of messages") {
			t.Errorf("expected rate limit message in body")
		}
	})

	t.Run("email rate limited returns rate limit view", func(t *testing.T) {
		s := &Server{
			db:  &mockDBEmailRateLimited{},
			cfg: &Config{AppEnv: EnvTest, GinMode: gin.TestMode},
		}
		handler := s.RegisterRoutes(s.cfg)

		form := url.Values{}
		form.Set("name", "Test Name")
		form.Set("email", "test@example.com")
		form.Set("subject", "Test Subject")
		form.Set("message", "Test message")

		req, err := http.NewRequest(http.MethodPost, "/contact", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %v", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "maximum number of messages") {
			t.Errorf("expected rate limit message in body")
		}
	})
}
