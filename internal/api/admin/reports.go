package admin

import "github.com/gin-gonic/gin"

func AdminMonthlyReport(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "admin monthly report OK"})
}
