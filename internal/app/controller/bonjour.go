package controller

import (
	"errors"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/dchest/uniuri"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/penguin-statistics/probe/internal/app/model"
	"github.com/penguin-statistics/probe/internal/app/service"
	"github.com/penguin-statistics/probe/internal/pkg/commons"
	"github.com/penguin-statistics/probe/internal/pkg/logger"
	"github.com/penguin-statistics/probe/internal/pkg/messages"
	"github.com/penguin-statistics/probe/internal/pkg/wspool"
	"google.golang.org/protobuf/proto"
)

var (
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
	upgrader *websocket.Upgrader
}

// NewBonjour creates a Bonjour controller with service
func NewBonjour(sBonjour *service.Bonjour, sProm *service.Prometheus) *Bonjour {
	hub := wspool.NewHub()
	go hub.Run()

	sProm.RegisterLiveUserFunc(func() float64 {
		return float64(len(hub.Clients))
	})
	sProm.RegisterUsersFunc(func() float64 {
		count, err := sBonjour.Count()
		if err != nil {
			log.Errorln("failed to get users count", err)
		}
		return float64(count)
	})

	return &Bonjour{
		sBonjour: sBonjour,
		sProm:    sProm,
		hub:      hub,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     commons.GenOriginChecker(),
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			},
			EnableCompression: false,
		},
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
	path, err := commons.CleanClientRoute(req.Referer)
	if err != nil {
		log.Debugln("invalid referer provided: failed to clean client route:", err)
		path = "(unspecified)"
	}

	// if a legacy client, we only record basic request info and return ok
	if req.Legacy != 0 {
		// generate a uid for them
		req.UID = uniuri.NewLen(32)

		// record bonjour request - see how many sessions are there
		_ = bc.sBonjour.Record(req)

		bc.sProm.IncUV(platform, path)
		bc.sProm.IncPV(platform, path)

		return ctx.NoContent(http.StatusNoContent)
	}

	// record reconnections
	bc.sProm.RecordReconnection(platform, req.Reconnects)

	// record initial visit records only if this is NOT a reconnecting request
	if req.Reconnects == 0 {
		// increment the uv since this is a probe request that would initiate on and only on reconnect==0
		bc.sProm.IncUV(platform, path)

		// record initial page view that comes with initial probe request
		bc.sProm.IncPV(platform, path)

		// record bonjour request - see how many sessions are there
		err = bc.sBonjour.Record(req)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err)
		}
	}

	// upgrade to websocket
	ws, err := bc.upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		log.Debugln("failed to update http conn to ws conn", err)
		ctx.Response().Header().Set(echo.HeaderUpgrade, "websocket")
		return echo.NewHTTPError(http.StatusUpgradeRequired, err)
	}

	client := wspool.NewClient(bc.hub, ws)

	must := func(err error) error {
		if err != nil {
			client.Send <- wspool.ErrInvalidWsMessage
			log.Traceln(err)
			return err
		}
		return nil
	}

	bc.hub.Register <- client
	go client.Read()
	go client.Write()
	for {
		select {
		case _, more := <-client.Closed:
			if !more {
				return nil
			}
		case r, more := <-client.Received:
			if !more {
				return nil
			}
			switch r.Skeleton.Meta.Type {
			case messages.MessageType_NAVIGATED:
				var body messages.Navigated
				err := must(proto.Unmarshal(r.Body, &body))
				if err != nil {
					break
				}
				s, err := commons.CleanClientRoute(body.Path)
				err = must(err)
				if err != nil {
					break
				}
				bc.sProm.IncPV(platform, s)

			case messages.MessageType_ENTERED_SEARCH_RESULT:
				var body messages.EnteredSearchResult
				err := must(proto.Unmarshal(r.Body, &body))
				if err != nil {
					break
				}
				log.Infoln("Client Report: entered search result: query", body.Query, "(stageId", body.GetStageId(), "itemId", body.GetItemId(), ") at position", body.Position)

			case messages.MessageType_EXECUTED_ADVANCED_QUERY:
				var body messages.ExecutedAdvancedQuery
				err := must(proto.Unmarshal(r.Body, &body))
				if err != nil {
					break
				}
				log.Infoln("Client Report: performed advanced queries:", spew.Sdump(body.Queries))

			default:
				log.Debugln("unknown message type", r.Skeleton.Meta.Type)
				client.Send <- wspool.ErrInvalidWsMessage
				break
			}
		}
	}
}
