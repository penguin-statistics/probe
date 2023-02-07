package repository

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/spf13/viper"
)

// Probe describes a repository which holds probe requests
type Probe struct {
	DB driver.Conn
}

// NewProbe returns a repository with probe requests
func NewProbe() *Probe {
	db, err := clickhouse.Open(&clickhouse.Options{
		Addr: viper.GetStringSlice("clickhouse.addr"),
		Auth: clickhouse.Auth{
			Database: viper.GetString("clickhouse.database"),
			Username: viper.GetString("clickhouse.user"),
			Password: viper.GetString("clickhouse.password"),
		},
		Debug:       viper.GetBool("app.debug"),
		DialTimeout: time.Second * 20,
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "penguin-statistics/probe", Version: "0.1.0"},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	err = db.Ping(context.Background())
	if err != nil {
		panic(err)
	}
	return &Probe{
		DB: db,
	}
}
