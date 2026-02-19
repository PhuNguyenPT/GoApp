package server

import (
	"GoApp/internal/database"
	"GoApp/internal/views"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

type RegisterInput struct {
    Name     string `form:"name"     validate:"required"`
    Email    string `form:"email"    validate:"required,email"`
    Password string `form:"password" validate:"required,min=8"`
}

func (s *Server) registerPageHandler(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.RegisterPage().Render(c.Request.Context(), c.Writer); err != nil {
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

    _, err = s.db.CreateUser(c.Request.Context(), database.CreateUserParams{
        Name:         input.Name,
        Email:        input.Email,
        PasswordHash: string(hash),
    })
    if err != nil {
        renderError("Email already in use.")
        return
    }

    c.Status(http.StatusOK)
    c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.RegisterSuccess(input.Name).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering register success: %v", err)
	}
}