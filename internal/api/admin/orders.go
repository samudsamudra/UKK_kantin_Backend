package admin

import "github.com/gin-gonic/gin"

func AdminUpdateOrderStatus(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "update order status OK"})
}

func AdminOrdersByMonth(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "admin orders by month OK"})
}
