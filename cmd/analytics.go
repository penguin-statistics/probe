package cmd

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"

	"github.com/penguin-statistics/probe/internal/app/server"
)

func Bootstrap() {
	viper.SetEnvPrefix("penguinprobe")
	viper.AutomaticEnv()
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if viper.GetBool("app.pprof") {
		go func() {
			fmt.Println("pprof enabled on localhost:8120")
			http.ListenAndServe("localhost:8120", nil)
		}()
	}

	err := server.Bootstrap()
	if err != nil {
		panic(err)
	}
}
