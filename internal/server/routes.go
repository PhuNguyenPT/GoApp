package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

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
	r.GET("/robots.txt", func(c *gin.Context) {
		c.File("./frontend-template/public/robots.txt")
	})
	r.GET("/register", s.registerPageHandler)
	r.POST("/register", s.registerHandler)
	return r
}
