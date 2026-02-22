package server

import (
	"GoApp/internal/database"
	"GoApp/internal/views"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) contactPageHandler(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.ContactPage(getUserName(c)).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering contact page: %v", err)
	}
}

const maxContactsPerIPPerDay int64 = 5
const maxContactsPerEmailPerDay int64 = 3

func (s *Server) contactFormHandler(c *gin.Context) {
	renderError := func() {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := views.ContactFail().Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering contact fail: %v", err)
		}
	}

	renderRateLimit := func() {
		log.Printf("rate limit exceeded for ip: %s", c.ClientIP())
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := views.ContactRateLimit().Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering contact rate limit: %v", err)
		}
	}

	ip := c.ClientIP()
	name := c.PostForm("name")
	email := c.PostForm("email")
	subject := c.PostForm("subject")
	message := c.PostForm("message")

	ipCount, err := s.db.CountContactsByIPToday(c.Request.Context(), ip)
	if err != nil {
		log.Printf("error counting contacts by ip: %v", err)
		renderError()
		return
	}
	if ipCount >= maxContactsPerIPPerDay {
		renderRateLimit()
		return
	}

	emailCount, err := s.db.CountContactsByEmailToday(c.Request.Context(), email)
	if err != nil {
		log.Printf("error counting contacts by email: %v", err)
		renderError()
		return
	}
	if emailCount >= maxContactsPerEmailPerDay {
		renderRateLimit()
		return
	}

	_, err = s.db.CreateContact(c.Request.Context(), database.CreateContactParams{
		Name:      name,
		Email:     email,
		Subject:   subject,
		Message:   message,
		IpAddress: ip,
	})
	if err != nil {
		log.Printf("error saving contact: %v", err)
		renderError()
		return
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.ContactSuccess(name).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering contact success: %v", err)
	}
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	if len(parts[0]) <= 1 {
		return "***@" + parts[1]
	}
	return string(parts[0][0]) + "***@" + parts[1]
}

func (s *Server) dashboardPageHandler(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	user, err := s.db.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	sessions, err := s.db.GetActiveSessionsByUserID(c.Request.Context(), userID)
	if err != nil {
		sessions = nil
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.DashboardPage(user.Name, maskEmail(user.Email), user.Email, user.CreatedAt, sessions).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering dashboard: %v", err)
	}
}

func (s *Server) authHeaderHandler(c *gin.Context) {
	userName := getUserName(c)
	if err := views.AuthHeaderFragment(userName).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering auth header: %v", err)
	}
}

type UpdateUserNameInput struct {
	Name string `form:"name" validate:"required"`
}

type UpdateUserPasswordInput struct {
	CurrentPassword string `form:"current_password" validate:"required"`
	NewPassword     string `form:"new_password"     validate:"required,min=8"`
	ConfirmPassword string `form:"confirm_password" validate:"required"`
}

func (s *Server) updateUserNameHandler(c *gin.Context) {
	renderError := func(msg string) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := views.DashboardError(msg).Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering dashboard error: %v", err)
		}
	}

	userID := c.MustGet("userID").(uuid.UUID)

	var input UpdateUserNameInput
	if err := c.ShouldBind(&input); err != nil {
		renderError("Name is required.")
		return
	}
	if err := validate.Struct(input); err != nil {
		renderError("Name is required.")
		return
	}

	_, err := s.db.UpdateUserName(c.Request.Context(), database.UpdateUserNameParams{
		ID:   userID,
		Name: input.Name,
	})
	if err != nil {
		renderError("Something went wrong, please try again.")
		return
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.DashboardNameSuccess(input.Name).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering dashboard success: %v", err)
	}
}

func (s *Server) updateUserPasswordHandler(c *gin.Context) {
	renderError := func(msg string) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := views.DashboardError(msg).Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering dashboard error: %v", err)
		}
	}

	userID := c.MustGet("userID").(uuid.UUID)

	var input UpdateUserPasswordInput
	if err := c.ShouldBind(&input); err != nil {
		renderError("All fields are required.")
		return
	}
	if err := validate.Struct(input); err != nil {
		errs := err.(validator.ValidationErrors)
		renderError(validationMessage(errs[0]))
		return
	}
	if input.NewPassword != input.ConfirmPassword {
		renderError("Passwords do not match.")
		return
	}

	user, err := s.db.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		renderError("Something went wrong, please try again.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		renderError("Current password is incorrect.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		renderError("Something went wrong, please try again")
		return
	}

	if err := s.db.UpdateUserPassword(c.Request.Context(), database.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: string(hash),
	}); err != nil {
		renderError("Something went wrong, please try again.")
		return
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.DashboardSuccess("Password updated successfully.").Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering dashboard success: %v", err)
	}
}
