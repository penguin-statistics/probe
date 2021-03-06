package wspool

import (
	"github.com/gorilla/websocket"
	"github.com/penguin-statistics/probe/internal/pkg/messages"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrInvalidWsMessage is the error of an invalid websocket message
	ErrInvalidWsMessage = mustPrepareMessage("invalid websocket message")
	// ErrInternalError describes a server-side error
	ErrInternalError          = mustPrepareMessage("internal server error")
	ErrTooManyInvalidMessages = mustPrepareMessage("too many invalid messages")
)

func mustPrepareMessage(m string) (msg *websocket.PreparedMessage) {
	evt := &messages.ServerACK{Message: m}
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
