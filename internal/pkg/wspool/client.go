package wspool

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/penguin-statistics/probe/internal/pkg/messages"
	"google.golang.org/protobuf/proto"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	// TODO: fine tuning
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// ClientRequest is the skeleton-unmarshalled client side request
type ClientRequest struct {
	Skeleton *messages.Skeleton
	Body     []byte
}

// Client is a intermediate module to connect user-side websocket client with hub
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Received chan ClientRequest
	Send     chan *websocket.PreparedMessage
	Done     chan struct{}

	InvalidCount int
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		Hub:          hub,
		Conn:         conn,
		Received:     make(chan ClientRequest, 64),
		Send:         make(chan *websocket.PreparedMessage, 256),
		Done:         make(chan struct{}),
		InvalidCount: 0,
	}
}

// Read block-reads from the underlying websocket.Conn. It also parses skeleton for further unmarshalling
func (c *Client) Read() {
	defer func() {
		c.Hub.Unregister <- c
		c.Done <- struct{}{}
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(appData string) error {
		log.Traceln("got pong from client")
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
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
		log.Debugln("error occurred when marshalling ack message", err)
		return err
	}
	p, err := websocket.NewPreparedMessage(websocket.BinaryMessage, b)
	if err != nil {
		log.Debugln("error occurred when preparing ack message", err)
		return err
	}
	c.Send <- p
	return nil
}

func (c *Client) readSkeleton() (s *messages.Skeleton, p []byte, err error) {
	typ, p, err := c.Conn.ReadMessage()
	if err != nil {
		log.Debugln("error occurred when reading message", err)
		return &messages.Skeleton{}, nil, err
	}
	if typ != websocket.BinaryMessage && typ != websocket.PongMessage {
		log.Debugln("unexpected message type that is not a BinaryMessage type", typ)
		return &messages.Skeleton{}, nil, errors.New("unexpected message type")
	}

	var skeleton messages.Skeleton
	err = proto.Unmarshal(p, &skeleton)
	if err != nil {
		log.Error("message either is not having common header or can't be unmarshalled to Skeleton", err)
		return &messages.Skeleton{}, nil, err
	}
	log.Traceln("unmarshalled skeleton as", skeleton.String())

	return &skeleton, p, nil
}

func (c *Client) Write() {
	log.Traceln("starting ping ticker with period of", pingPeriod)
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			log.Traceln("tries to send ws data", message)
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The Hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WritePreparedMessage(message)
			if err != nil {
				log.Warnln("failed to Send message", err)
				break
			}
			log.Traceln("ws data sent")
		case <-pingTicker.C:
			log.Traceln("times up: sending ping to client")

			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				//log.Debugln("failed to write ping to client. client probably already gone. disconnecting")
				return
			}
		}
	}
}
