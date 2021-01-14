package controller

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/penguin-statistics/probe/internal/app/model"
	"github.com/penguin-statistics/probe/internal/app/service"
	"github.com/penguin-statistics/probe/internal/pkg/commons"
	"github.com/penguin-statistics/probe/internal/pkg/logger"
	"github.com/penguin-statistics/probe/internal/pkg/messages"
	"github.com/penguin-statistics/probe/internal/pkg/wspool"
	"google.golang.org/protobuf/proto"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// TODO: DEV
			return true || commons.IsValidDomain(r.URL)
		},
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		},
		EnableCompression: false,
	}
	log = logger.New("controller")
	// Maximum invalid messages a client may send
	// if client sent invalid messages more than this count, their connection may be forcely closed
	maxInvalidTolerance = 16
)

// Bonjour is a bonjour service controller
type Bonjour struct {
	sBonjour *service.Bonjour
	sProm    *service.Prometheus
	hub      *wspool.Hub
}

// NewBonjour creates a Bonjour controller with service
func NewBonjour(sbonjour *service.Bonjour, sprom *service.Prometheus) *Bonjour {
	hub := wspool.NewHub()
	go hub.Run()
	return &Bonjour{
		sBonjour: sbonjour,
		sProm:    sprom,
		hub:      hub,
	}
}

// LiveHandler handles probe reports
func (bc *Bonjour) LiveHandler(ctx echo.Context) error {
	req := new(model.Bonjour)
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := ctx.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if req.Platform == nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("platform: field is required"))
	}

	platform := req.Platform.Marshal()

	// get referer path from bonjour request
	path, err := commons.CleanClientRoute(ctx.Request().Referer())
	if err != nil {
		log.Warnln("invalid referer provided: failed to clean client route:", err)
		path = "(unspecified)"
	}

	bc.sProm.RecordReconnection(platform, req.Reconnects)

	// record initial visit records only if this is NOT a reconnecting request
	if req.Reconnects == 0 {
		// if uid doesn't exist before, increment the UV value
		if !bc.sBonjour.UIDExists(req.UID) {
			bc.sProm.IncUV(platform, path)
		}

		// record initial page view
		bc.sProm.IncPV(platform, path)

		// record bonjour request - see how many sessions are there
		err = bc.sBonjour.Record(req)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
	}

	// upgrade to websocket
	ws, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		log.Debugln("failed to update http conn to ws conn", err)
		ctx.Response().Header().Set(echo.HeaderUpgrade, "websocket")
		return echo.NewHTTPError(http.StatusUpgradeRequired, err)
	}

	client := wspool.NewClient(bc.hub, ws)

	must := func(err error) error {
		if err != nil {
			log.Error("malformed client message", err, "tolerance for now but adds InvalidCount")
			errMsg := wspool.ErrInvalidWsMessage

			client.InvalidCount++
			if client.InvalidCount >= maxInvalidTolerance {
				errMsg = wspool.ErrTooManyInvalidMessages
			}

			client.Send <- errMsg
			return nil
		}
		return nil
	}

	bc.hub.Register <- client
	go client.Read()
	go client.Write()
	go func() {
		for {
			r := <-client.Received
			switch r.Skeleton.Meta.Type {
			case messages.MessageType_NAVIGATED:
				var body messages.Navigated
				err := must(proto.Unmarshal(r.Body, &body))
				if err != nil {
					log.Debugln(err)
					break
				}
				s, err := commons.CleanClientRoute(body.Path)
				err = must(err)
				if err != nil {
					log.Debugln(err)
					break
				}
				bc.sProm.IncPV(req.Platform.Marshal(), s)

			default:
				log.Warnln("unknown message type", r.Skeleton.Meta.Type)
				client.Send <- wspool.ErrInvalidWsMessage
				break
			}
		}
	}()

	<-client.Done
	return nil
}
