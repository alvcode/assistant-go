package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepository interface {
	Create(in *entity.File) (*entity.File, error)
	GetLastId() (int, error)
	GetByHash(hash string) (*entity.File, error)
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

func (r *fileRepository) Create(in *entity.File) (*entity.File, error) {
	query := `
		INSERT INTO files (user_id, original_filename, file_path, ext, size, hash, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`

	row := r.db.QueryRow(
		r.ctx,
		query,
		in.UserID,
		in.OriginalFilename,
		in.FilePath,
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

func (r *fileRepository) GetLastId() (int, error) {
	query := `SELECT coalesce(max(id), 0) FROM files`

	var result int
	err := r.db.QueryRow(r.ctx, query).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *fileRepository) GetByHash(hash string) (*entity.File, error) {
	query := `select * from files where hash = $1`
	row := r.db.QueryRow(r.ctx, query, hash)
	var file entity.File
	if err := row.Scan(
		&file.ID,
		&file.UserID,
		&file.OriginalFilename,
		&file.FilePath,
		&file.Ext,
		&file.Size,
		&file.Hash,
		&file.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &file, nil
}
