package main

import (
	"embed"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	//go:embed all:internal/web
	web       embed.FS
	webAssetFS = echo.MustSubFS(web, "internal/web/assets")
	webDocsFS = echo.MustSubFS(web, "internal/web/docs")
)

func main() {
	e := echo.New()
	// Middleware Config
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfv3339}::${remote_id}::${status}:${method}:${uri}\n",
	}))
	e.Use(middleware.Recover())
	// API Docs
	e.StaticFS("/api", webDocsFS)
	e.StaticFS("/assets", webAssetFS)
	// Start Server
	e.Logger.Fatal(e.Start(":" + "8888"))
}
