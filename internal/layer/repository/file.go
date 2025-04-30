package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepository interface {
	Create(in *entity.File) (*entity.File, error)
	GetLastId() (int, error)
}

type fileRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewFileRepository(ctx context.Context, db *pgxpool.Pool) FileRepository {
	return &fileRepository{
		ctx: ctx,
		db:  db,
	}
}

func (ur *fileRepository) Create(in *entity.File) (*entity.File, error) {
	query := `
		INSERT INTO files (user_id, original_filename, filename, ext, size, hash, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`

	row := ur.db.QueryRow(
		ur.ctx,
		query,
		in.UserID,
		in.OriginalFilename,
		in.Filename,
		in.Ext,
		in.Size,
		in.Hash,
		in.CreatedAt,
	)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return in, nil
}

func (ur *fileRepository) GetLastId() (int, error) {
	query := `SELECT coalesce(max(id), 0) FROM files`

	var result int
	err := ur.db.QueryRow(ur.ctx, query).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}
