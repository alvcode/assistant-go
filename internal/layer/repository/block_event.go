package repository

import (
	"assistant-go/internal/layer/dto"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type BlockEventRepository interface {
	SetEvent(ip string, eventName string, time time.Time) (int, error)
	GetStat(ip string, timeFrom time.Time) (*dto.BlockEventsStat, error)
	RemoveByDateExpired(time time.Time) error
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

func (ur *blockEventRepository) GetStat(ip string, timeFrom time.Time) (*dto.BlockEventsStat, error) {
	query := `
		WITH error_counts AS (
			SELECT 
				event,
				COUNT(*) AS event_count
			FROM 
				block_events
			WHERE 
				ip = $1
				AND created_at >= $2
			GROUP BY 
				event
		),
		total_errors AS (
			SELECT 
				'all' as event,
				COUNT(*) AS event_count
			FROM 
				block_events
			WHERE 
				ip = $1
				AND created_at >= $2
		)
		SELECT 
			ec.event,
			ec.event_count
		FROM 
			error_counts ec
		union 
		SELECT 
			te.event,
			te.event_count
		FROM 
			total_errors te;
	`

	rows, err := ur.db.Query(ur.ctx, query, ip, timeFrom)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stat := &dto.BlockEventsStat{}
	for rows.Next() {
		var event string
		var count int

		if err := rows.Scan(&event, &count); err != nil {
			return nil, err
		}

		switch event {
		case "all":
			stat.All = count
		case "validate_input_data":
			stat.ValidateInputData = count
		case "decode_body":
			stat.DecodeBody = count
		case "sign_in":
			stat.SignIn = count
		case "unauthorized":
			stat.Unauthorized = count
		case "refresh_token":
			stat.RefreshToken = count
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stat, nil
}

func (ur *blockEventRepository) RemoveByDateExpired(time time.Time) error {
	query := `DELETE FROM block_events WHERE created_at < $1`

	_, err := ur.db.Exec(ur.ctx, query, time)
	if err != nil {
		return err
	}
	return nil
}
