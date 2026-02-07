package app

import (
	"fmt"
	// "strings"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

func PrettyLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()

		methodColor := color.New(color.FgWhite)
		switch method {
		case "GET":
			methodColor = color.New(color.FgGreen)
		case "POST":
			methodColor = color.New(color.FgYellow)
		case "PUT":
			methodColor = color.New(color.FgBlue)
		case "PATCH":
			methodColor = color.New(color.FgMagenta)
		case "DELETE":
			methodColor = color.New(color.FgRed)
		}

		statusColor := color.New(color.FgGreen)
		if status >= 400 {
			statusColor = color.New(color.FgRed)
		}

		fmt.Printf(
			"%s | %s | %s | %s | %v\n",
			color.New(color.FgCyan).Sprint(time.Now().Format("15:04:05")),
			methodColor.Sprintf("%-6s", method),
			statusColor.Sprintf("%d", status),
			color.New(color.FgWhite).Sprint(path),
			latency,
		)
	}
}