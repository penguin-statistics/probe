package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/penguin-statistics/probe/internal/app/controller"
	"github.com/penguin-statistics/probe/internal/app/repository"
	"github.com/penguin-statistics/probe/internal/app/service"
	"github.com/penguin-statistics/probe/internal/pkg/commons"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

// Bootstrap starts the http server up
func Bootstrap() error {
	e := echo.New()
	e.Debug = true
	e.Validator = &Validator{}
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} | \u001B[97;34m${status} ${latency_human}\u001B[0m | \033[97;36m${method} ${uri}\033[0m\n",
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		//AllowCredentials: true,
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

	// TODO: config management
	r := repository.NewProbe("host=localhost user=root password=root dbname=penguinprobe port=5432 sslmode=disable TimeZone=Asia/Shanghai")
	sBonjour := service.NewBonjour(r)
	sProm := service.NewPrometheus()
	c := controller.NewBonjour(sBonjour, sProm)

	// TODO: DEV
	//{
	//	e.File("/web", "web/index.html")
	//	e.File("/web/events.js", "web/events.js")
	//}

	e.GET("/", c.LiveHandler)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// TODO: config management
	return e.Start("localhost:8100")
}
