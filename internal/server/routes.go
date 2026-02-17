package server

import (
	"net/http"
	"fmt"
	"log"
	"time"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/coder/websocket"
	
	"GoApp/internal/views"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

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

	return r
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
	views.HomePage().Render(c.Request.Context(), c.Writer)
}

func (s *Server) contactPageHandler(c *gin.Context) {
	views.ContactPage().Render(c.Request.Context(), c.Writer)
}

func (s *Server) contactFormHandler(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	subject := c.PostForm("subject")
	message := c.PostForm("message")
	
	// Simulate processing delay (e.g., sending email, database operation)
	time.Sleep(1 * time.Second)
	
	log.Printf("Contact form: %s (%s) - %s: %s", name, email, subject, message)
	
	views.ContactSuccess(name).Render(c.Request.Context(), c.Writer)
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"
	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
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