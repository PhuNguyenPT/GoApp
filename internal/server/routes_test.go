package server

import (
	"GoApp/internal/database"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
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

func (m *mockDB) UpdateUserName(ctx context.Context, arg database.UpdateUserNameParams) (database.User, error) {
	return database.User{ID: arg.ID, Name: arg.Name}, nil
}

func (m *mockDB) UpdateUserPassword(ctx context.Context, arg database.UpdateUserPasswordParams) error {
	return nil
}

func (m *mockDB) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return nil
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

func (m *mockDB) DeleteSessionByID(ctx context.Context, arg database.DeleteSessionByIDParams) error {
	return nil
}

func (m *mockDB) CreateContact(ctx context.Context, arg database.CreateContactParams) (database.Contact, error) {
	return database.Contact{
		Name:    arg.Name,
		Email:   arg.Email,
		Subject: arg.Subject,
		Message: arg.Message,
	}, nil
}

func (m *mockDB) CountContactsByIPToday(ctx context.Context, ipAddress string) (int64, error) {
	return 0, nil
}

func (m *mockDB) CountContactsByEmailToday(ctx context.Context, email string) (int64, error) {
	return 0, nil
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

func TestHomePageHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
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

func TestUnknownRoute(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/does-not-exist", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %v", rr.Code)
	}
}
