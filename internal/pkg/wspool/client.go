package wspool

import (
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

// Client is a websocket client with op methods
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
	Done chan struct{}
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

	orErr := func(err error) error {
		if err != nil {
			log.Error("malformed client message", err)
			return c.Conn.WritePreparedMessage(ErrInvalidWsMessage)
		}
		return nil
	}

	for {
		s, p, err := c.readSkeleton()
		if err != nil {
			break
		}
		typ := s.Meta.Type
		switch typ {
		case messages.MessageType_NAVIGATED:
			var body messages.Navigated
			err := orErr(proto.Unmarshal(p, &body))
			if err != nil {
				log.Warnln()
			}
		default:
			log.Warnln("unknown message type", s.Meta.Type)
			c.Conn.WritePreparedMessage(ErrInvalidWsMessage)
			break
		}

		c.ack(typ)
	}
}

func (c *Client) ack(messageType messages.MessageType) error {
	m := messages.ServerACK{Type: messageType}
	b, err := proto.Marshal(&m)
	if err != nil {
		log.Errorln("error occurred when sending ack", err)
		return err
	}
	c.Send <- b
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
		return &messages.Skeleton{}, nil, err
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

			err := c.Conn.WriteMessage(websocket.BinaryMessage, message)
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
