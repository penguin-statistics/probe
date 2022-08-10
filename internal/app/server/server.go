package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"

	"github.com/penguin-statistics/probe/internal/app/controller"
	"github.com/penguin-statistics/probe/internal/app/repository"
	"github.com/penguin-statistics/probe/internal/app/service"
	"github.com/penguin-statistics/probe/internal/pkg/commons"
)

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
	sBonjour := service.NewBonjour(r)
	sProm := service.NewPrometheus()
	c := controller.NewBonjour(sBonjour, sProm)

	if viper.GetBool("app.debug") {
		e.File("/web", "web/index.html")
		e.File("/web/events.js", "web/events.js")
	}

	e.GET("/", c.LiveHandler)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	e.GET("/health", func(ctx echo.Context) error {
		if err := r.DB.Raw("SELECT 1").Scan(&struct{}{}).Error; err != nil {
			return ctx.String(http.StatusInternalServerError, "database is not ready")
		}

		return ctx.String(http.StatusOK, "ok")
	})

	return e.Start(viper.GetString("http.server"))
}
