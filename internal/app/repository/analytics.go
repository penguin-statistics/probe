package repository

import (
	"github.com/penguin-statistics/probe/internal/app/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Probe describes a repository which holds probe requests
type Probe struct {
	DB *gorm.DB
}

// NewProbe returns a repository with probe requests
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
