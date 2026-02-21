package server

import (
	"GoApp/internal/views"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) contactPageHandler(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.ContactPage(getUserName(c)).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering contact page: %v", err)
	}
}

func (s *Server) contactFormHandler(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	subject := c.PostForm("subject")
	message := c.PostForm("message")

	// Simulate processing delay (e.g., sending email, database operation)
	time.Sleep(1 * time.Second)

	log.Printf("Contact form: %s (%s) - %s: %s", name, email, subject, message)

	if err := views.ContactSuccess(name).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering contact success: %v", err)
	}
}

func (s *Server) sitemapHandler(c *gin.Context) {
	c.Header("Content-Type", "application/xml")
	c.File("./frontend-template/public/sitemap.xml")
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
	if err := views.DashboardPage(user.Name, maskEmail(user.Email), user.Email, user.CreatedAt, len(sessions)).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering dashboard: %v", err)
	}
}

func (s *Server) authHeaderHandler(c *gin.Context) {
	userName := getUserName(c)
	if err := views.AuthHeaderFragment(userName).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering auth header: %v", err)
	}
}
