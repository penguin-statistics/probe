package wspool

import (
	"github.com/gorilla/websocket"
	"github.com/penguin-statistics/probe/internal/pkg/messages"
	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidWsMessage = mustPrepareMessage("invalid websocket message")
	ErrInternalError    = mustPrepareMessage("internal server error")
)

func mustPrepareMessage(m string) (msg *websocket.PreparedMessage) {
	evt := &messages.ServerErrored{Message: m}
	b, err := proto.Marshal(evt)
	if err != nil {
		panic(err)
	}
	msg, err = websocket.NewPreparedMessage(websocket.BinaryMessage, b)
	if err != nil {
		panic(err)
	}
	return msg
}
