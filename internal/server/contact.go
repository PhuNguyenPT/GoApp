package server

import (
	"GoApp/internal/database"
	"GoApp/internal/views"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
