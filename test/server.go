package main

import (
	"github.com/dchest/uniuri"
	"github.com/gorilla/websocket"
	"github.com/penguin-statistics/probe/internal/pkg/messages"
	"google.golang.org/protobuf/proto"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

var endpoint = "ws://localhost:8100/"

func handleWsConn(ws *websocket.Conn, wg *sync.WaitGroup, ctr *uint64) {
	//t := time.Now()
	for {
		m, _ := proto.Marshal(&messages.Navigated{
			Meta: &messages.Meta{
				Type:     messages.MessageType_NAVIGATED,
				Language: messages.Language_ZH_CN,
			},
			Path: "/" + uniuri.NewLen(64),
		})

		err := ws.WriteMessage(websocket.BinaryMessage, m)
		if err != nil {
			panic(err)
		}
		_, _, err = ws.ReadMessage()
		if err != nil {
			panic(err)
		}
		atomic.AddUint64(ctr, 1)
		c := atomic.LoadUint64(ctr)
		if c%5000 == 0 {
			log.Printf("requests total: %v times", c)
			break
		}
		time.Sleep(time.Millisecond * 10)
	}
	//ws.WriteMessage(websocket.CloseMessage, []byte{})
	//log.Println(time.Now().Sub(t).Milliseconds())
	//for {
	//	_, _, err := ws.ReadMessage()
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//wg.Done()
}

func main() {
	log.Println("launching")
	var limit uint64 = 1e4
	var current uint64 = 0
	var ctr uint64 = 0
	wg := &sync.WaitGroup{}
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			for {
				v := url.Values{}
				v.Set("v", "v3.4.1")
				v.Set("p", "web")
				v.Set("u", uniuri.NewLen(32))
				u := endpoint + "?" + v.Encode()
				if atomic.LoadUint64(&current) <= limit {
					d := websocket.Dialer{
						Subprotocols: []string{"pb"},
					}
					ws, _, err := d.Dial(u, http.Header{})
					if err != nil {
						panic(err)
					}
					atomic.AddUint64(&current, 1)
					go handleWsConn(ws, wg, &ctr)
				} else {
					break
				}
			}
		}()
	}
	wg.Wait()
}
