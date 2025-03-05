package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type BlockIPRepository interface {
	FindBlocking(ip string, time time.Time) (bool, error)
}

type blockIpRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewBlockIpRepository(ctx context.Context, db *pgxpool.Pool) BlockIPRepository {
	return &blockIpRepository{
		ctx: ctx,
		db:  db,
	}
}

func (ur *blockIpRepository) FindBlocking(ip string, time time.Time) (bool, error) {
	query := `SELECT 1 FROM block_ip WHERE ip = $1 and blocked_until > $2 LIMIT 1`

	row := ur.db.QueryRow(ur.ctx, query, ip, time)

	if err := row.Scan(&ip, &time); err != nil {
		return false, err
	}
	return true, nil
}
