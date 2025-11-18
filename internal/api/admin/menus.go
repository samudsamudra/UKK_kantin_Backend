package api

import "github.com/gin-gonic/gin"

func AdminCreateMenu(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "create menu OK"})
}

func AdminUpdateMenu(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "update menu OK"})
}

func AdminDeleteMenu(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "delete menu OK"})
}

func AdminListMenus(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "admin list menus OK"})
}
