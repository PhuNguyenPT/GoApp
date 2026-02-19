package server

import (
	"GoApp/internal/database"
	"GoApp/internal/views"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) registerPageHandler(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.RegisterPage().Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering register page: %v", err)
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

    name := c.PostForm("name")
    email := c.PostForm("email")
    password := c.PostForm("password")

    if name == "" || email == "" || password == "" {
        renderError("All fields are required.")
        return
    }
    if len(password) < 8 {
        renderError("Password must be at least 8 characters.")
        return
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        renderError("Something went wrong, please try again.")
        return
    }

    _, err = s.db.CreateUser(c.Request.Context(), database.CreateUserParams{
        Name:         name,
        Email:        email,
        PasswordHash: string(hash),
    })
    if err != nil {
        renderError("Email already in use.")
        return
    }

    c.Status(http.StatusOK)
    c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.RegisterSuccess(name).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering register success: %v", err)
	}
}