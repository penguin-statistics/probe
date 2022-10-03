// modified from https://github.com/gorilla/websocket/blob/master/examples/chat/hub.go

package wspool

import (
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/penguin-statistics/probe/internal/pkg/logger"
)

// Hub consists of current active clients
type Hub struct {
	Register   chan *Client
	Unregister chan *Client
	logger     *logrus.Entry

	clientsmu sync.RWMutex
	Clients   map[*Client]struct{}
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		logger:     logger.New("wspool"),
		Clients:    make(map[*Client]struct{}),
	}
}

// Run listens for channel update on client connect and disconnects
// and reflects them on the internal Clients property
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clientsmu.Lock()
			h.Clients[client] = struct{}{}
			h.clientsmu.Unlock()
		case client := <-h.Unregister:
			h.clientsmu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
			}
			h.clientsmu.Unlock()
		}
	}
}

func (h *Hub) Evict() {
	h.clientsmu.RLock()
	defer h.clientsmu.RUnlock()

	var wg sync.WaitGroup
	limiter := make(chan struct{}, 8)
	for client := range h.Clients {
		limiter <- struct{}{}
		wg.Add(1)
		go func(client *Client) {
			client.Close()
			close(client.GoingAwayClose)
			<-limiter
		}(client)
	}
	wg.Wait()
}
