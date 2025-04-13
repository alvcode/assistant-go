package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type BlockIPRepository interface {
	FindBlocking(ip string, time time.Time) (bool, error)
	RemoveByDateExpired(time time.Time) error
	SetBlock(ip string, time time.Time) error
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
	query := `SELECT EXISTS(SELECT 1 FROM block_ip WHERE ip = $1 and blocked_until > $2)`

	var exists bool
	err := ur.db.QueryRow(ur.ctx, query, ip, time).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (ur *blockIpRepository) RemoveByDateExpired(time time.Time) error {
	query := `DELETE FROM block_ip WHERE blocked_until < $1`

	_, err := ur.db.Exec(ur.ctx, query, time)
	if err != nil {
		return err
	}
	return nil
}

func (ur *blockIpRepository) SetBlock(ip string, unblockTime time.Time) error {
	query := `INSERT INTO block_ip (ip, blocked_until) VALUES ($1, $2) RETURNING id`

	row := ur.db.QueryRow(ur.ctx, query, ip, unblockTime)

	var id int
	if err := row.Scan(&id); err != nil {
		return err
	}
	return nil
}
