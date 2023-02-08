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

// RecordBonjour adds a bonjour request in model.Bonjour to db
func (s *Bonjour) RecordBonjour(b *model.Bonjour) error {
	return s.repo.DB.Exec(context.Background(), "insert into bonjours (id, version, platform, uid, legacy) values (?, ?, ?, ?, ?)", b.ID, b.Version, b.Platform, b.UID, b.Legacy)
}

// RecordImpression adds a view request in model.Bonjour to db
func (s *Bonjour) RecordImpression(b *model.Impression) error {
	return s.repo.DB.Exec(context.Background(), "insert into impressions (id, bonjour_id, path) values (?, ?, ?)", b.ID, b.BonjourID, b.Path)
}

// RecordEventSearchResultEntered adds a search result entered event in model.EventSearchResultEntered to db
func (s *Bonjour) RecordEventSearchResultEntered(b *model.EventSearchResultEntered) error {
	return s.repo.DB.Exec(context.Background(), "insert into event_search_result_entered (id, bonjour_id, query, result_position, destination) values (?, ?, ?, ?, ?)", b.ID, b.BonjourID, b.Query, b.ResultPosition, b.Destination)
}

// Count counts current bonjour requests from db
func (s *Bonjour) Count() (uint64, error) {
	var count uint64
	if err := s.repo.DB.QueryRow(context.Background(), "select count(*) from bonjours").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
