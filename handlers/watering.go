package handlers

import (
	"home-server/services"
	"home-server/views"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type WateringHandler struct {
	ws services.WaterinService
}

func NewWateringHandler() WateringHandler {
	return WateringHandler{ws: services.NewWateringService()}
}

func (wh *WateringHandler) AddRoutes(g *echo.Group) {
	g.GET("/", wh.Index)
	{
		m := g.Group("/manual")
		m.GET("", wh.GetManual)
		m.GET("/form", wh.GetManualForm)
		m.PUT("", wh.UpdateManual)
	}
	{
		i := g.Group("/interval")
		i.POST("", wh.CreateInterval)
		i.GET("/:id", wh.GetInterval)
		i.GET("/form/:id", wh.GetIntervalForm)
		i.PUT("/:id", wh.UpdateInterval)
		i.DELETE("/:id", wh.DeleteInterval)
	}
}

func (wh *WateringHandler) Index(c echo.Context) error {
	watering := views.Watering(wh.ws.GetManual(), wh.ws.GetIntervals())
	return Render(c, http.StatusOK, views.Index(watering))
}

func (wh *WateringHandler) GetManual(c echo.Context) error {
	wm := wh.ws.GetManual()
	return Render(c, http.StatusOK, views.WateringManual(wm))
}

func (wh *WateringHandler) GetManualForm(c echo.Context) error {
	wm := wh.ws.GetManual()
	return Render(c, http.StatusOK, views.WateringManualForm(wm))
}

func (wh *WateringHandler) UpdateManual(c echo.Context) error {
	on := c.FormValue("on") == "on"

	var areas [3]bool
	for i, area := range views.AREA_NAMES {
		areas[i] = c.FormValue(area) == "on"
	}

	autoOff, err := durForm(c.FormValue("auto-off"))
	if err != nil {
		return echo.ErrBadRequest
	}
	wm := wh.ws.UpdateManual(on, areas, autoOff)
	return Render(c, http.StatusOK, views.WateringManual(wm))
}

func (wh *WateringHandler) CreateInterval(c echo.Context) error {
	wi := wh.ws.CreateInterval()
	return Render(c, http.StatusOK, views.WateringInterval(wi))
}

func (wh *WateringHandler) GetInterval(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest
	}
	wi, ok := wh.ws.GetInterval(id)
	if !ok {
		return echo.ErrBadRequest
	}
	return Render(c, http.StatusOK, views.WateringInterval(wi))
}

func (wh *WateringHandler) GetIntervalForm(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest
	}
	wi, ok := wh.ws.GetInterval(id)
	if !ok {
		return echo.ErrBadRequest
	}
	return Render(c, http.StatusOK, views.WateringIntervalForm(wi))
}

func (wh *WateringHandler) UpdateInterval(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest
	}
	wi := services.WateringInterval{Id: id}
	wi.On = c.FormValue("on") == "on"

	for i, area := range views.AREA_NAMES {
		wi.Areas[i] = c.FormValue(area) == "on"
	}

	for i, day := range views.DAY_NAMES {
		wi.Days[i] = c.FormValue(day) == "on"
	}

	wi.Start, err = durForm(c.FormValue("start"))
	if err != nil {
		return echo.ErrBadRequest
	}

	wi.Duration, err = durForm(c.FormValue("duration"))
	if err != nil {
		return echo.ErrBadRequest
	}

	ok := wh.ws.UpdateInterval(wi)
	if !ok {
		return echo.ErrBadRequest
	}
	return Render(c, http.StatusOK, views.WateringInterval(wi))
}

func (wh *WateringHandler) DeleteInterval(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrBadRequest
	}
	ok := wh.ws.DeleteInterval(id)
	if !ok {
		return echo.ErrBadRequest
	}
	return c.String(http.StatusOK, "")
}
