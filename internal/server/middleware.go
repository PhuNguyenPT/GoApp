package server

import (
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
