package main

import (
	"fmt"
	"github.com/penguin-statistics/probe/internal/app/server"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	if os.Getenv("PENGUIN_PROBE_PPROF") == "1" {
		go func() {
			fmt.Println("pprof enabled")
			http.ListenAndServe("localhost:8120", nil)
		}()
	}

	panic(server.Bootstrap())
}
