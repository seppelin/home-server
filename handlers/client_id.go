package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func ClientIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := c.Cookie("client_id")
		if err != nil {
			clientID := uuid.New().String()
			println("new id", clientID, c.RealIP(), err.Error())
			c.SetCookie(&http.Cookie{
				Name:     "client_id",
				Value:    clientID,
				Path:     "/",
				SameSite: http.SameSiteStrictMode,
			})
		}
		return next(c)
	}
}
