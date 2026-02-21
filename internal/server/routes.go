package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterRoutes(cfg *Config) http.Handler {
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/api/health", "/.well-known/appspecific/com.chrome.devtools.json"},
	}))
	r.Use(gin.Recovery())
	r.Use(s.resolveUserMiddleware())
	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/public/site.webmanifest" {
			c.Header("Content-Type", "application/manifest+json")
		}
		c.Next()
	})

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
	r.GET("/auth-header", s.authHeaderHandler)
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./frontend-template/public/favicon.ico")
	})
	r.StaticFile("/", "./frontend-template/public/index.html")
	r.GET("/contact", s.contactPageHandler)
	r.POST("/contact", s.contactFormHandler)
	r.GET("/sitemap.xml", func(c *gin.Context) {
		c.File("./frontend-template/public/sitemap.xml")
	})
	r.HEAD("/sitemap.xml", func(c *gin.Context) {
		c.File("./frontend-template/public/sitemap.xml")
	})
	r.GET("/robots.txt", func(c *gin.Context) {
		c.File("./frontend-template/public/robots.txt")
	})
	r.HEAD("/robots.txt", func(c *gin.Context) {
		c.File("./frontend-template/public/robots.txt")
	})
	r.GET("/register", s.registerPageHandler)
	r.POST("/register", s.registerHandler)
	r.GET("/login", s.loginPageHandler)
	r.POST("/login", s.loginHandler)
	r.GET("/logout", s.logoutHandler)

	protected := r.Group("/")
	protected.Use(s.authMiddleware())
	{
		protected.GET("/dashboard", s.dashboardPageHandler)
	}
	return r
}
