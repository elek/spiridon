package db

import (
	"context"
	"time"
)

type Refresher struct {
	db *Persistence
}

func NewRefresher(db *Persistence) *Refresher {
	return &Refresher{
		db: db,
	}
}

func (r *Refresher) Run(ctx context.Context) error {
	time.Sleep(5 * time.Minute)
	for {
		r.db.RefreshViews(ctx)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Minute):
			continue
		}
	}

}
