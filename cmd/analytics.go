package cmd

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/spf13/viper"

	"github.com/penguin-statistics/probe/internal/app/server"
)

func Bootstrap() {
	viper.SetEnvPrefix("penguinprobe")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if viper.GetBool("app.pprof") {
		go func() {
			fmt.Println("pprof enabled")
			http.ListenAndServe("localhost:8120", nil)
		}()
	}

	panic(server.Bootstrap())
}
