package repository

import (
	"github.com/penguin-statistics/probe/internal/app/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Probe struct {
	DB *gorm.DB
}

func NewProbe(dsn string) *Probe {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&model.Bonjour{})
	if err != nil {
		panic(err)
	}
	return &Probe{
		DB: db,
	}
}
