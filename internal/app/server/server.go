package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"

	"github.com/penguin-statistics/probe/internal/app/controller"
	"github.com/penguin-statistics/probe/internal/app/repository"
	"github.com/penguin-statistics/probe/internal/app/service"
	"github.com/penguin-statistics/probe/internal/pkg/commons"
	"github.com/penguin-statistics/probe/internal/pkg/logger"
	"github.com/penguin-statistics/probe/internal/pkg/wspool"
)

var log = logger.New("cmd")

// Bootstrap starts the http server up
func Bootstrap() error {
	if viper.GetBool("app.debug") {
		fmt.Println("debug enabled")
	}

	e := echo.New()
	e.Debug = viper.GetBool("app.debug")
	e.Validator = &Validator{}
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} | \u001B[97;34m${status} ${latency_human}\u001B[0m | \033[97;36m${method} ${uri}\033[0m\n",
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// AllowCredentials: true,
		AllowMethods: []string{http.MethodGet},
		AllowOrigins: commons.PenguinDomainsOrigin(),
		MaxAge:       int((time.Hour * 24).Seconds()),
	}))
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		e.DefaultHTTPErrorHandler(err, c)
	}

	r := repository.NewProbe(viper.GetString("app.dsn"))
	hub := wspool.NewHub()
	sBonjour := service.NewBonjour(r)
	sProm := service.NewPrometheus()
	c := controller.NewBonjour(sBonjour, sProm, hub)
	e.Server.RegisterOnShutdown(func() {
		go hub.Evict()
	})

	if viper.GetBool("app.debug") {
		e.File("/web", "web/index.html")
		e.File("/web/events.js", "web/events.js")
	}

	e.GET("/", c.LiveHandler)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Start server
	go func() {
		if err := e.Start(viper.GetString("http.server")); err != nil && err != http.ErrServerClosed {
			log.Infoln("server shutdown", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Infoln("received non-nil err from Shutdown()", err)
	}

	if len(hub.Clients) > 0 {
		log.Infoln("waiting for clients to disconnect")
		for {
			l := len(hub.Clients)
			if l == 0 {
				break
			}
			log.Infoln("waiting for", l, "clients to disconnect")
			time.Sleep(time.Minute)
		}
	}

	return nil
}
