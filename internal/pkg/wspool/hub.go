// modified from https://github.com/gorilla/websocket/blob/master/examples/chat/hub.go

package wspool

import (
	"github.com/penguin-statistics/probe/internal/pkg/logger"
	"github.com/sirupsen/logrus"
)

// Hub consists of current active clients
type Hub struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	logger     *logrus.Entry
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		logger:     logger.New("wspool"),
	}
}

// Run listens for channel update on client connect and disconnects
// and reflects them on the internal Clients property
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
			}
		}
	}
}
