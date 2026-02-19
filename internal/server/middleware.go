package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		session, err := s.db.GetSessionByToken(c.Request.Context(), token)
		if err != nil {
			secure := s.cfg.AppEnv == EnvProduction
			c.SetCookie("session_token", "", -1, "/", "", secure, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("userID", session.UserID)
		c.Next()
	}
}

func (s *Server) resolveUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil {
			c.Set("userName", "")
			c.Next()
			return
		}

		session, err := s.db.GetSessionByToken(c.Request.Context(), token)
		if err != nil {
			c.Set("userName", "")
			c.Next()
			return
		}

		user, err := s.db.GetUserByID(c.Request.Context(), session.UserID)
		if err != nil {
			log.Printf("resolveUserMiddleware: user not found: %v", err)
			c.Set("userName", "")
			c.Next()
			return
		}

		c.Set("userName", user.Name)
		c.Next()
	}
}

func getUserName(c *gin.Context) string {
	name, _ := c.Get("userName")
	if s, ok := name.(string); ok {
		return s
	}
	return ""
}
