package api

import "github.com/gin-gonic/gin"

func SiswaListMenus(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "siswa menus OK"})
}
