package main

import (
	"home-server/handlers"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Static("/", "assets")
	wh := handlers.NewWateringHandler()
	wh.AddRoutes(e.Group(""))
	e.Logger.Fatal(e.Start(":3000"))
}
