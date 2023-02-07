package controller

import (
	"errors"
	"net/http"

	"github.com/davecgh/go-spew/spew"
	"github.com/dchest/uniuri"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/proto"

	"github.com/penguin-statistics/probe/internal/app/model"
	"github.com/penguin-statistics/probe/internal/app/service"
	"github.com/penguin-statistics/probe/internal/pkg/commons"
	"github.com/penguin-statistics/probe/internal/pkg/logger"
	"github.com/penguin-statistics/probe/internal/pkg/messages"
	"github.com/penguin-statistics/probe/internal/pkg/wspool"
)

var log = logger.New("controller")

// Bonjour is a bonjour service controller
type Bonjour struct {
	sBonjour *service.Bonjour
	sProm    *service.Prometheus
	hub      *wspool.Hub
	upgrader *websocket.Upgrader
}

// NewBonjour creates a Bonjour controller with service
func NewBonjour(sBonjour *service.Bonjour, sProm *service.Prometheus, hub *wspool.Hub) *Bonjour {
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
			ReadBufferSize:  128,
			WriteBufferSize: 128,
			CheckOrigin:     commons.GenOriginChecker(),
			Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
			},
			EnableCompression: false,
		},
	}
}

// LiveHandler handles probe reports
func (bc *Bonjour) LiveHandler(c echo.Context) error {
	req := new(model.Bonjour)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if req.Platform == nil {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("platform: field is required"))
	}

	req.ID = ulid.Make().String()

	platform := req.Platform.Marshal()

	// get referer path from bonjour request
	path, err := commons.CleanClientRoute(req.Referer)
	if err != nil {
		log.Debugln("invalid referer provided: failed to clean client route:", err)
		path = "(unspecified)"
	}

	impression := &model.Impression{
		ID:        ulid.Make().String(),
		BonjourID: req.ID,
		Path:      path,
	}

	// if a legacy client, we only record basic request info and return ok
	if req.Legacy != 0 {
		// generate a uid for them
		req.UID = uniuri.NewLen(32)

		// record bonjour request - see how many sessions are there
		_ = bc.sBonjour.RecordBonjour(req)

		bc.sProm.IncUV(platform)
		bc.sProm.IncPV(platform)

		return c.NoContent(http.StatusNoContent)
	}

	// record reconnections
	bc.sProm.RecordReconnection(platform, req.Reconnects)

	// record initial visit records only if this is NOT a reconnecting request
	if req.Reconnects == 0 {
		// increment the uv since this is a probe request that would initiate on and only on reconnect==0
		bc.sProm.IncUV(platform)

		// record initial page view that comes with initial probe request
		bc.sProm.IncPV(platform)

		// record bonjour request - see how many sessions are there
		err = bc.sBonjour.RecordBonjour(req)
		if err != nil {
			return err
		}

		err = bc.sBonjour.RecordImpression(impression)
		if err != nil {
			return err
		}
	}

	// upgrade to websocket
	ws, err := bc.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Debugln("failed to update http conn to ws conn", err)
		c.Response().Header().Set(echo.HeaderUpgrade, "websocket")
		return echo.NewHTTPError(http.StatusUpgradeRequired, "failed to upgrade to websocket")
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
				path, err := commons.CleanClientRoute(body.Path)
				if must(err) != nil {
					break
				}
				bc.sProm.IncPV(platform)
				impression := &model.Impression{
					ID:        ulid.Make().String(),
					BonjourID: req.ID,
					Path:      path,
				}
				err = bc.sBonjour.RecordImpression(impression)
				if err != nil {
					log.Warnln("failed to record impression:", err)
				}

			case messages.MessageType_ENTERED_SEARCH_RESULT:
				var body messages.EnteredSearchResult
				err := must(proto.Unmarshal(r.Body, &body))
				if err != nil {
					break
				}

				destination := ""
				if body.GetStageId() != "" {
					destination = "stage:" + body.GetStageId()
				} else if body.GetItemId() != "" {
					destination = "item:" + body.GetItemId()
				} else {
					destination = "unknown"
				}

				err = bc.sBonjour.RecordEventSearchResultEntered(&model.EventSearchResultEntered{
					ID:             ulid.Make().String(),
					BonjourID:      req.ID,
					Query:          body.Query,
					Destination:    destination,
					ResultPosition: body.GetPosition(),
				})
				if err != nil {
					log.Warnln("failed to record impression:", err)
				}

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
			}
		}
	}
}
