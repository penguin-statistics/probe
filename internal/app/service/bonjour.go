package service

import (
	"github.com/penguin-statistics/probe/internal/app/model"
	"github.com/penguin-statistics/probe/internal/app/repository"
)

// Bonjour is the bonjour service
type Bonjour struct {
	repo *repository.Probe
}

// NewBonjour creates a bonjour request-related service with repo
func NewBonjour(repo *repository.Probe) *Bonjour {
	return &Bonjour{repo: repo}
}

func (s *Bonjour) UIDExists(uid string) bool {
	var req model.Bonjour
	err := s.repo.DB.First(&req, &model.Bonjour{UID: uid}).Error
	if err != nil {
		return false
	}
	return true
}

// Record adds a bonjour request in model.Bonjour to db
func (s *Bonjour) Record(b *model.Bonjour) error {
	if err := s.repo.DB.Create(b).Error; err != nil {
		return err
	}
	return nil
}

// Count counts current bonjour requests from db
func (s *Bonjour) Count() (int64, error) {
	var count int64
	if err := s.repo.DB.Model(&model.Bonjour{}).Count(&count).Error; err != nil {
		return -1, err
	}
	return count, nil
}
