package main

import (
	"context"
	"home-server/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	e.Static("/", "assets")
	wh := handlers.NewWateringHandler(context.Background(), "http://localhost:2727")
	e.GET("/", wh.Index, handlers.ClientIDMiddleware)
	wh.AddRoutes(e.Group("/watering", handlers.ClientIDMiddleware))
	e.Logger.Fatal(e.Start(":3000"))
}
