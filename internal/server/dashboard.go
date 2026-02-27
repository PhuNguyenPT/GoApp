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
	if err := views.DashboardPage(user.Name, maskEmail(user.Email), user.Email, user.CreatedAt, sessions, getLangStr(c)).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering dashboard: %v", err)
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
	lang := getLangStr(c)
	renderError := func(msg string) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := views.DashboardError(msg, lang).Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering dashboard error: %v", err)
		}
	}

	userID := c.MustGet("userID").(uuid.UUID)

	var input UpdateUserNameInput
	if err := c.ShouldBind(&input); err != nil {
		renderError(views.T(lang).ErrNameRequired)
		return
	}
	if err := validate.Struct(input); err != nil {
		renderError(views.T(lang).ErrNameRequired)
		return
	}

	_, err := s.db.UpdateUserName(c.Request.Context(), database.UpdateUserNameParams{
		ID:   userID,
		Name: input.Name,
	})
	if err != nil {
		renderError(views.T(getLangStr(c)).ErrSomethingWrong)
		return
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.DashboardNameSuccess(input.Name, lang).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering dashboard success: %v", err)
	}
}

func (s *Server) updateUserPasswordHandler(c *gin.Context) {
	lang := getLangStr(c)
	renderError := func(msg string) {
		c.Status(http.StatusOK)
		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := views.DashboardError(msg, lang).Render(c.Request.Context(), c.Writer); err != nil {
			log.Printf("error rendering dashboard error: %v", err)
		}
	}

	userID := c.MustGet("userID").(uuid.UUID)

	var input UpdateUserPasswordInput
	if err := c.ShouldBind(&input); err != nil {
		renderError(views.T(getLangStr(c)).ErrAllRequired)
		return
	}
	if err := validate.Struct(input); err != nil {
		errs := err.(validator.ValidationErrors)
		renderError(validationMessage(errs[0]))
		return
	}
	if input.NewPassword != input.ConfirmPassword {
		renderError(views.T(lang).ErrPasswordMismatch)
		return
	}

	user, err := s.db.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		renderError(views.T(getLangStr(c)).ErrSomethingWrong)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		renderError(views.T(lang).ErrWrongPassword)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		renderError(views.T(lang).ErrSomethingWrong)
		return
	}

	if err := s.db.UpdateUserPassword(c.Request.Context(), database.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: string(hash),
	}); err != nil {
		renderError(views.T(getLangStr(c)).ErrSomethingWrong)
		return
	}

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.DashboardSuccess(views.T(lang).PasswordUpdated, lang).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering dashboard success: %v", err)
	}
}
