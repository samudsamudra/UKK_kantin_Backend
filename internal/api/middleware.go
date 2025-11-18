package api

import "github.com/gin-gonic/gin"

// if you have real JWTAuth in another place, ignore this
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
