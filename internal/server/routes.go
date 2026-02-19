package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"GoApp/internal/views"
)


func (s *Server) sitemapHandler(c *gin.Context) {
	c.Header("Content-Type", "application/xml")
	c.File("./frontend-template/public/sitemap.xml")
}

func (s *Server) apiInfoHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Go Server API",
		"version": "0.0",
		"endpoints": gin.H{
			"health": "/api/health",
			"websocket": "/api/websocket",
		},
	})
}

func (s *Server) homePageHandler(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.HomePage().Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering home page: %v", err)
	}
}

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

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) registerPageHandler(c *gin.Context) {
	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := views.RegisterPage().Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering register page: %v", err)
	}
}

func (s *Server) registerHandler(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")

	if name == "" || email == "" || password == "" {
		c.Status(http.StatusBadRequest)
		c.Header("Content-Type", "text/html; charset=utf-8")
		views.RegisterError("All fields are required.").Render(c.Request.Context(), c.Writer)
		return
	}

	if len(password) < 8 {
		c.Status(http.StatusBadRequest)
		c.Header("Content-Type", "text/html; charset=utf-8")
		views.RegisterError("Password must be at least 8 characters.").Render(c.Request.Context(), c.Writer)
		return
	}

	// TODO: hash password and save user via s.db
	log.Printf("Register: %s (%s)", name, email)

	c.Status(http.StatusOK)
	c.Header("Content-Type", "text/html; charset=utf-8")
	views.RegisterSuccess(name).Render(c.Request.Context(), c.Writer)
}

func (s *Server) RegisterRoutes(cfg *Config) http.Handler {
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	apiGroup := r.Group("/api")
	apiGroup.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	{
		apiGroup.GET("/", s.apiInfoHandler)
		apiGroup.GET("/health", s.healthHandler)
		apiGroup.GET("/websocket", s.websocketHandler)
	}

	r.Static("/public", "./frontend-template/public")

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./frontend-template/public/favicon.ico")
	})

	r.GET("/", s.homePageHandler)
	r.GET("/contact", s.contactPageHandler)
	r.POST("/contact", s.contactFormHandler)
	r.GET("/sitemap.xml", s.sitemapHandler)
	r.GET("/robots.txt", func (c *gin.Context)  {
		c.File("./frontend-template/public/robots.txt")
	})
	r.GET("/register", s.registerPageHandler)
	r.POST("/register", s.registerHandler)	
	return r
}

func (s *Server) websocketHandler(c *gin.Context) {
	w := c.Writer
	r := c.Request
	socket, err := websocket.Accept(w, r, nil)

	if err != nil {
		log.Printf("could not open websocket: %v", err)
		_, _ = w.Write([]byte("could not open websocket"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer socket.Close(websocket.StatusGoingAway, "server closing websocket")

	ctx := r.Context()
	socketCtx := socket.CloseRead(ctx)

	for {
		payload := fmt.Sprintf("server timestamp: %d", time.Now().UnixNano())
		err := socket.Write(socketCtx, websocket.MessageText, []byte(payload))
		if err != nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
}