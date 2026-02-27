package server

import "github.com/gin-gonic/gin"

func getLang(c *gin.Context) string {
	if q := c.Query("lang"); q == "vi" || q == "en" {
		c.SetCookie("lang", q, 86400*30, "/", "", false, false)
		return q
	}
	if cookie, err := c.Cookie("lang"); err == nil && cookie == "vi" {
		return "vi"
	}
	return "en"
}

func (s *Server) langMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := getLang(c)
		c.Set("lang", lang)
		c.Next()
	}
}

func getLangStr(c *gin.Context) string {
	if lang, ok := c.Get("lang"); ok {
		if s, ok := lang.(string); ok {
			return s
		}
	}
	return "en"
}
