package api

import "github.com/gin-gonic/gin"

func SiswaCreateOrder(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "order create OK"})
}

func SiswaOrdersByMonth(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "order list OK"})
}
