package service

import (
	"context"

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

func (s *Bonjour) UIDExists(ctx context.Context, uid string) bool {
	return s.repo.CheckUIDExists(ctx, uid)
}

// Record adds a bonjour request in model.Bonjour to db
func (s *Bonjour) Record(ctx context.Context, b *model.Bonjour) error {
	return s.repo.Insert(ctx, b)
}

// Count counts current bonjour requests from db
func (s *Bonjour) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}
