package app

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

func PrintRoutesReport(r *gin.Engine, env, port string) {
	type row struct {
		Method string
		Path   string
	}

	rows := []row{}
	for _, rt := range r.Routes() {
		rows = append(rows, row{
			Method: rt.Method,
			Path:   rt.Path,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Method == rows[j].Method {
			return rows[i].Path < rows[j].Path
		}
		return rows[i].Method < rows[j].Method
	})

	fmt.Println()
	color.New(color.Bold, color.FgCyan).Println("🚀 UKK KANTIN API SERVER")
	fmt.Println("────────────────────────────────────────")
	fmt.Printf("ENV   : %s\n", env)
	fmt.Printf("PORT  : %s\n", port)
	fmt.Printf("ROUTES: %d endpoints\n", len(rows))
	fmt.Println("────────────────────────────────────────")

	for _, r := range rows {
		methodColor := color.New(color.FgWhite)
		switch r.Method {
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

		fmt.Printf("%s  %s\n",
			methodColor.Sprintf("%-6s", r.Method),
			r.Path,
		)
	}

	fmt.Println("────────────────────────────────────────")
}