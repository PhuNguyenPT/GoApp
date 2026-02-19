package server

import (
	"GoApp/internal/views"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) contactPageHandler(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.ContactPage().Render(c.Request.Context(), c.Writer); err != nil {
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

func (s *Server) homePageHandler(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.HomePage().Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering home page: %v", err)
	}
}