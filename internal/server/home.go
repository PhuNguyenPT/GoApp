package server

import (
	"GoApp/internal/views"
	"log"

	"github.com/gin-gonic/gin"
)

func (s *Server) homePageHandler(c *gin.Context) {
	if err := views.HomePage(getUserName(c), getLangStr(c)).Render(c.Request.Context(), c.Writer); err != nil {
		log.Printf("error rendering home page: %v", err)
	}
}
