package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// simple login placeholder if you don't have real one yet
func Login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "login endpoint (placeholder)"})
}
