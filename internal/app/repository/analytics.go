package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/penguin-statistics/probe/internal/app/model"
)

// Probe describes a repository which holds probe requests
type Probe struct {
	DB *sql.DB
}

// NewProbe returns a repository with probe requests
func NewProbe(dsn string) *Probe {
	opt, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		panic(err)
	}

	db := clickhouse.OpenDB(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		panic(err)
	}

	return &Probe{
		DB: db,
	}
}

func (p *Probe) CheckUIDExists(ctx context.Context, uid string) bool {
	err := p.DB.QueryRowContext(ctx, "SELECT 1 FROM bonjours WHERE uid = $1", uid).Scan(new(int))
	return err == nil
}

func (p *Probe) Insert(ctx context.Context, b *model.Bonjour) error {
	_, err := p.DB.ExecContext(ctx, "INSERT INTO bonjours (version, platform, uid, legacy) VALUES ($1, $2, $3, $4)", b.Version, b.Platform, b.UID, b.Legacy)
	return err
}

func (p *Probe) Count(ctx context.Context) (int64, error) {
	var count int64
	// select bonjours_id_seq to determine the count of bonjours

	err := p.DB.QueryRowContext(ctx, "SELECT last_value FROM bonjours_id_seq").Scan(&count)
	return count, err
}
