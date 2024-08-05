package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"home-server/services"
	"home-server/views"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type WateringHandler struct {
	bs BroadCastServer[services.WateringUpdate]
	ws *services.Watering
}

func MapState(state [services.AREA_COUNT]bool) map[string]bool {
	values := make(map[string]bool)
	for i, name := range views.AREA_NAMES {
		values[name] = state[i]
	}
	return values
}

func NewWateringHandler(ctx context.Context, ardURL string) WateringHandler {
	ws, web, ard := services.NewWatering()
	go func() {
		for {
			state := <-ard
			data, err := json.Marshal(MapState(state))
			if err != nil {
				fmt.Print(err.Error())
				continue
			}
			client := http.Client{}
			req, err := http.NewRequest(http.MethodPut, ardURL, bytes.NewBuffer(data))
			if err != nil {
				fmt.Print(err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				fmt.Print(err)
				continue
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Print(err)
				continue
			}
			fmt.Println("ArdUpdate: ", string(b))
			resp.Body.Close()
		}
	}()
	return WateringHandler{
		bs: NewBroadcastServer(ctx, web),
		ws: ws,
	}
}

func (wh *WateringHandler) AddRoutes(g *echo.Group) {
	g.GET("/updates", wh.WebUpdates)
	g.GET("/state", wh.GetState)
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
	watering := views.Watering(wh.ws.GetManual(), wh.ws.GetIntervals(), wh.ws.State())
	return Render(c, http.StatusOK, views.Index(watering))
}

func (wh *WateringHandler) GetState(c echo.Context) error {
	state := MapState(wh.ws.State().Areas)
	return c.JSON(http.StatusOK, state)
}

func (wh *WateringHandler) WebUpdates(c echo.Context) error {
	c.Logger().Debugf("WebUpdate: client connected, ip: %v\n", c.RealIP())

	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	updates := wh.bs.Subscribe()
	for {
		select {
		case <-c.Request().Context().Done():
			wh.bs.CancelSubscription(updates)
			c.Logger().Debugf("WebUpdate: client disconnected, ip: %v\n", c.RealIP())
			return nil
		case u := <-updates:
			data := new(bytes.Buffer)
			views.State(u.State).Render(c.Request().Context(), data)
			event := Event{
				Event: []byte("w-state"),
				Data:  data.Bytes(),
			}
			if err := event.MarshalTo(w); err != nil {
				return err
			}
			data.Reset()
			clientID, err := c.Cookie("client_id")
			if err != nil {
				return err
			}
			if clientID.Value == u.ClientID {
				c.Logger().Debug("WebUpdate: Same id: ", clientID.Value)
			} else {
				var id string
				switch u.Kind {
				case services.UpdateManual:
					views.WateringManual(u.Manual).Render(c.Request().Context(), data)
					id = "w-manual"
				case services.CreateInterval:
					views.WateringInterval(u.Interval).Render(c.Request().Context(), data)
					id = "w-intervals"
				case services.UpdateInterval:
					views.WateringInterval(u.Interval).Render(c.Request().Context(), data)
					id = u.Interval.GetId()
				case services.DeleteInterval:
					id = u.Interval.GetId()
				}
				c.Logger().Debug("WebUpdate: ", u.Kind, id, clientID.Value)
				event = Event{
					Event: []byte(id),
					Data:  data.Bytes(),
				}
				if err := event.MarshalTo(w); err != nil {
					return err
				}
			}
			w.Flush()
		}
	}
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
	clientID, _ := c.Cookie("client_id")
	wm := wh.ws.UpdateManual(on, areas, autoOff, clientID.Value)
	return Render(c, http.StatusOK, views.WateringManual(wm))
}

func (wh *WateringHandler) CreateInterval(c echo.Context) error {
	clientID, _ := c.Cookie("client_id")
	wi := wh.ws.CreateInterval(clientID.Value)
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
	clientID, _ := c.Cookie("client_id")
	ok := wh.ws.UpdateInterval(wi, clientID.Value)
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
	clientID, _ := c.Cookie("client_id")
	ok := wh.ws.DeleteInterval(id, clientID.Value)
	if !ok {
		return echo.ErrBadRequest
	}
	return c.String(http.StatusOK, "")
}
