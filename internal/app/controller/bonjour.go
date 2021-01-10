package controller

import (
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/penguin-statistics/probe/internal/app/model"
	"github.com/penguin-statistics/probe/internal/app/service"
	"github.com/penguin-statistics/probe/internal/pkg/logger"
	"github.com/penguin-statistics/probe/internal/pkg/utils"
	"github.com/penguin-statistics/probe/internal/pkg/wspool"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// TODO: DEV
			return true || utils.IsValidDomain(r.URL)
		},
		Subprotocols: []string{"pb"},
	}
	log = logger.New("controller")
)

type Bonjour struct {
	service *service.Bonjour
	hub *wspool.Hub
}

func NewBonjour(service *service.Bonjour) *Bonjour {
	hub := wspool.NewHub()
	go hub.Run()
	return &Bonjour{
		service: service,
		hub: hub,
	}
}

func (ct *Bonjour) LiveHandler(c echo.Context) error {
	r := new(model.Bonjour)
	if err := c.Bind(r); err != nil {
		log.Errorln("failed to bind", err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := c.Validate(r); err != nil {
		log.Errorln("failed to validate", err)
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err := ct.service.Record(r)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Errorln("failed to update http conn to ws conn", err)
		return err
	}

	send := make(chan []byte, 256)
	done := make(chan struct{})

	client := wspool.Client{Hub: ct.hub, Conn: ws, Send: send, Done: done}
	ct.hub.Register <- &client
	go client.Read()
	go client.Write()

	<-done
	return nil
}

