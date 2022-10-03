package wspool

import (
	"errors"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/ratelimit"
	"google.golang.org/protobuf/proto"

	"github.com/penguin-statistics/probe/internal/pkg/messages"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Minute

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 30 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// Maximum messages per second
	maxRPS = 3
)

var ErrInvalidMessageType = errors.New("invalid message type")

// ClientRequest is the skeleton-unmarshalled client side request
type ClientRequest struct {
	Skeleton *messages.Skeleton
	Body     []byte
}

// Client is a intermediate module to connect user-side websocket client with hub
type Client struct {
	Hub         *Hub
	Conn        *websocket.Conn
	Received    chan ClientRequest
	Send        chan *websocket.PreparedMessage
	Closed      chan struct{}
	rateLimiter ratelimit.Limiter

	InvalidCount int
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		Hub:          hub,
		Conn:         conn,
		Received:     make(chan ClientRequest, 4),
		Send:         make(chan *websocket.PreparedMessage, 4),
		Closed:       make(chan struct{}, 4),
		rateLimiter:  ratelimit.New(maxRPS),
		InvalidCount: 0,
	}
}

// Read block-reads from the underlying websocket.Conn. It also parses skeleton for further unmarshalling
func (c *Client) Read() {
	defer func() {
		close(c.Received)
		c.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(appData string) error {
		c.Hub.logger.Traceln("got pong from client")
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		c.rateLimiter.Take()
		s, p, err := c.readSkeleton()
		if err != nil {
			break
		}
		typ := s.Meta.Type
		c.Received <- ClientRequest{
			Skeleton: s,
			Body:     p,
		}

		c.ack(typ)
	}
}

func (c *Client) ack(messageType messages.MessageType) error {
	m := messages.ServerACK{Type: messageType}
	b, err := proto.Marshal(&m)
	if err != nil {
		c.Hub.logger.Debugln("error occurred when marshalling ack message", err)
		return err
	}
	p, err := websocket.NewPreparedMessage(websocket.BinaryMessage, b)
	if err != nil {
		c.Hub.logger.Debugln("error occurred when preparing ack message", err)
		return err
	}
	c.Send <- p
	return nil
}

func (c *Client) readSkeleton() (s *messages.Skeleton, p []byte, err error) {
	typ, p, err := c.Conn.ReadMessage()
	if err != nil {
		if !websocket.IsCloseError(err, websocket.CloseGoingAway) {
			c.Hub.logger.Debugln("error occurred when reading message", err)
		}
		return &messages.Skeleton{}, nil, err
	}
	if typ != websocket.BinaryMessage && typ != websocket.PongMessage {
		c.Hub.logger.Debugln("unexpected message type that is not a BinaryMessage type", typ)
		return &messages.Skeleton{}, nil, ErrInvalidMessageType
	}

	var skeleton messages.Skeleton
	err = proto.Unmarshal(p, &skeleton)
	if err != nil {
		c.Hub.logger.Error("message either is not having common header or can't be unmarshalled to Skeleton", err)
		return &messages.Skeleton{}, nil, err
	}
	c.Hub.logger.Traceln("unmarshalled skeleton as", skeleton.String())

	return &skeleton, p, nil
}

func (c *Client) Close() {
	c.Hub.Unregister <- c
	c.Closed <- struct{}{}
	c.Conn.Close()
}

func (c *Client) Write() {
	c.Hub.logger.Traceln("starting ping ticker with period of", pingPeriod)
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		c.Hub.logger.Traceln("stopping writer")
		pingTicker.Stop()
		c.Close()
	}()
	for {
		select {
		case message := <-c.Send:
			c.Hub.logger.Traceln("tries to send ws data", message)
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			err := c.Conn.WritePreparedMessage(message)
			if err != nil {
				c.Hub.logger.Debugln("failed to send message", err)
				return
			}
			c.Hub.logger.Traceln("ws data sent")
		case <-pingTicker.C:
			c.Hub.logger.Traceln("time's up: sending ping to client")

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				c.Hub.logger.Debugln("failed to write ping to client. client probably already gone. disconnecting")
				return
			}
		}
	}
}
