package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type BlockIPRepository interface {
	FindBlocking(ctx context.Context, ip string, time time.Time) (bool, error)
	RemoveByDateExpired(ctx context.Context, time time.Time) error
	SetBlock(ctx context.Context, ip string, time time.Time) error
}

type blockIpRepository struct {
	db *pgxpool.Pool
}

func NewBlockIpRepository(db *pgxpool.Pool) BlockIPRepository {
	return &blockIpRepository{db: db}
}

func (ur *blockIpRepository) FindBlocking(ctx context.Context, ip string, time time.Time) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM block_ip WHERE ip = $1 and blocked_until > $2)`

	var exists bool
	err := ur.db.QueryRow(ctx, query, ip, time).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (ur *blockIpRepository) RemoveByDateExpired(ctx context.Context, time time.Time) error {
	query := `DELETE FROM block_ip WHERE blocked_until < $1`

	_, err := ur.db.Exec(ctx, query, time)
	if err != nil {
		return err
	}
	return nil
}

func (ur *blockIpRepository) SetBlock(ctx context.Context, ip string, unblockTime time.Time) error {
	query := `INSERT INTO block_ip (ip, blocked_until) VALUES ($1, $2) RETURNING id`

	row := ur.db.QueryRow(ctx, query, ip, unblockTime)

	var id int
	if err := row.Scan(&id); err != nil {
		return err
	}
	return nil
}
