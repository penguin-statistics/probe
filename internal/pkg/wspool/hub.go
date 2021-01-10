// modified from https://github.com/gorilla/websocket/blob/master/examples/chat/hub.go

package wspool

// Hub consists of current active clients
type Hub struct {
	Clients    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run listens for channel update and reflects them on the internal Clients property
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
