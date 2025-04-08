package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type BlockEventRepository interface {
	SetEvent(ip string, eventName string, time time.Time) (int, error)
}

type blockEventRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewBlockEventRepository(ctx context.Context, db *pgxpool.Pool) BlockEventRepository {
	return &blockEventRepository{
		ctx: ctx,
		db:  db,
	}
}

func (ur *blockEventRepository) SetEvent(ip string, eventName string, time time.Time) (int, error) {
	query := `INSERT INTO block_events (ip, event, created_at) VALUES ($1, $2, $3) RETURNING id`

	row := ur.db.QueryRow(ur.ctx, query, ip, eventName, time)

	var resID int
	if err := row.Scan(&resID); err != nil {
		return 0, err
	}
	return resID, nil
}
