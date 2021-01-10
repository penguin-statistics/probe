package service

import (
	"github.com/penguin-statistics/probe/internal/app/model"
	"github.com/penguin-statistics/probe/internal/app/repository"
)

type Bonjour struct {
	repo *repository.Probe
}

func NewBonjour(repo *repository.Probe) *Bonjour {
	return &Bonjour{repo: repo}
}

func (s *Bonjour) Record(b *model.Bonjour) error {
	if err := s.repo.DB.Create(b).Error; err != nil {
		return err
	}
	return nil
}
