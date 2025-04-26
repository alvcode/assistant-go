package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RateLimiterRepository interface {
	CheckExists(ip string) (bool, error)
	UpsertIP(limiter entity.RateLimiter) (*entity.RateLimiter, error)
	FoundIP(ip string) (*entity.RateLimiter, error)
	UpdateIP(limiter entity.RateLimiter) (*entity.RateLimiter, error)
}

type rateLimiterRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewRateLimiterRepository(ctx context.Context, db *pgxpool.Pool) RateLimiterRepository {
	return &rateLimiterRepository{
		ctx: ctx,
		db:  db,
	}
}

func (ur *rateLimiterRepository) CheckExists(ip string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM rate_limiter WHERE ip = $1)`

	var exists bool
	err := ur.db.QueryRow(ur.ctx, query, ip).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (ur *rateLimiterRepository) UpsertIP(limiter entity.RateLimiter) (*entity.RateLimiter, error) {
	query := `INSERT INTO rate_limiter (ip, allowance, timestamp) VALUES ($1, $2, $3) ON CONFLICT (ip) DO UPDATE SET allowance = $2, timestamp = $3`

	_, err := ur.db.Exec(ur.ctx, query, limiter.IP, limiter.AllowanceRequests, limiter.Timestamp)
	if err != nil {
		return nil, err
	}
	return &limiter, nil
}

func (ur *rateLimiterRepository) FoundIP(ip string) (*entity.RateLimiter, error) {
	query := `SELECT allowance,timestamp FROM rate_limiter WHERE ip = $1`

	row := ur.db.QueryRow(ur.ctx, query, ip)

	var limiter entity.RateLimiter
	if err := row.Scan(&limiter.AllowanceRequests, &limiter.Timestamp); err != nil {
		return nil, err
	}
	return &limiter, nil
}

func (ur *rateLimiterRepository) UpdateIP(limiter entity.RateLimiter) (*entity.RateLimiter, error) {
	query := `UPDATE rate_limiter SET allowance = allowance - 1 WHERE ip = $1 returning allowance`

	row := ur.db.QueryRow(ur.ctx, query, limiter.IP)
	if err := row.Scan(&limiter.AllowanceRequests); err != nil {
		return nil, err
	}
	return &limiter, nil
}
