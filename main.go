package main

import (
	"context"
	"home-server/handlers"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Use(handlers.ClientIDMiddleware)
	e.Static("/", "assets")
	wh := handlers.NewWateringHandler(context.Background(), "http://localhost:2727")
	e.GET("/", wh.Index)
	wh.AddRoutes(e.Group("/watering"))
	e.Logger.Fatal(e.Start(":3000"))
}
