package server

import (
	"GoApp/internal/database"
	"GoApp/internal/views"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

func (s *Server) authHeaderHandler(c *gin.Context) {
	userName := getUserName(c)
	if err := views.AuthHeaderFragment(userName).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering auth header: %v", err)
	}
}

type RegisterInput struct {
	Name     string `form:"name"     validate:"required"`
	Email    string `form:"email"    validate:"required,email"`
	Password string `form:"password" validate:"required,min=8"`
}

func getClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := c.GetHeader("X-Real-Ip"); ip != "" {
		return ip
	}
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}

func (s *Server) registerPageHandler(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.RegisterPage(getUserName(c)).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering register page: %v", err)
	}
}

func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required."
	case "email":
		return "Please enter a valid email address."
	case "min":
		return e.Field() + " must be at least " + e.Param() + " characters."
	default:
		return "Invalid input."
	}
}

func (s *Server) registerHandler(c *gin.Context) {
	// helper to reduce repetition
	renderError := func(msg string) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := views.RegisterError(msg).Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering register error: %v", err)
		}
	}

	var input RegisterInput
	if err := c.ShouldBind(&input); err != nil {
		renderError("All fields are required.")
		return
	}
	if err := validate.Struct(input); err != nil {
		errs := err.(validator.ValidationErrors)
		renderError(validationMessage(errs[0]))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		renderError("Something went wrong, please try again.")
		return
	}

	user, err := s.db.CreateUser(c.Request.Context(), database.CreateUserParams{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		renderError("Email already in use.")
		return
	}

	token, err := uuid.NewV7()
	if err != nil {
		renderError("Something went wrong, please try again.")
		return
	}
	ip := getClientIP(c)
	userAgent := c.Request.UserAgent()

	_, err = s.db.CreateSession(c.Request.Context(), database.CreateSessionParams{
		UserID:    user.ID,
		Token:     token.String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserAgent: userAgent,
		IpAddress: ip,
	})
	if err != nil {
		renderError("Something went wrong, please try again.")
		return
	}
	secure := s.cfg.AppEnv == EnvProduction
	c.SetCookie("session_token", token.String(), 86400, "/", "", secure, true)
	c.Header("HX-Redirect", "/dashboard")
	c.Status(http.StatusOK)
}

type LoginInput struct {
	Email    string `form:"email"    validate:"required,email"`
	Password string `form:"password" validate:"required"`
}

func (s *Server) loginPageHandler(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.LoginPage(getUserName(c)).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering login page: %v", err)
	}
}

func (s *Server) loginHandler(c *gin.Context) {
	renderError := func(msg string) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := views.LoginError(msg).Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering login error: %v", err)
		}
	}

	var input LoginInput
	if err := c.ShouldBind(&input); err != nil {
		renderError("All fields are required.")
		return
	}
	if err := validate.Struct(input); err != nil {
		renderError("Invalid email or password.")
		return
	}

	user, err := s.db.GetUserByEmail(c.Request.Context(), input.Email)
	if err != nil {
		renderError("Invalid email or password.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		renderError("Invalid email or password.")
		return
	}

	token := uuid.New().String()
	ip := getClientIP(c)
	userAgent := c.Request.UserAgent()

	_, err = s.db.CreateSession(c.Request.Context(), database.CreateSessionParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserAgent: userAgent,
		IpAddress: ip,
	})
	if err != nil {
		renderError("Something went wrong, please try again.")
		return
	}
	secure := s.cfg.AppEnv == EnvProduction

	c.SetCookie("session_token", token, 86400, "/", "", secure, true)
	c.Header("HX-Redirect", "/dashboard")
	c.Status(http.StatusOK)
}

func (s *Server) logoutHandler(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err == nil {
		if err := s.db.DeleteSession(c.Request.Context(), token); err != nil {
			log.Printf("error deleting session: %v", err)
		}
	}
	secure := s.cfg.AppEnv == EnvProduction
	c.SetCookie("session_token", "", -1, "/", "", secure, true)
	c.Redirect(http.StatusFound, "/login")
}

func (s *Server) revokeSessionHandler(c *gin.Context) {
	token, err := c.Cookie("session_token")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	session, err := s.db.GetSessionByToken(c.Request.Context(), token)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := s.db.DeleteSessionByID(c.Request.Context(), database.DeleteSessionByIDParams{
		ID:     sessionID,
		UserID: session.UserID,
	}); err != nil {
		log.Printf("error revoking session: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
