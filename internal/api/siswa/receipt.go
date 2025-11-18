package api

import "github.com/gin-gonic/gin"

func SiswaGetReceiptPDF(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "receipt OK"})
}
