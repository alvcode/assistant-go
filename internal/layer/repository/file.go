package repository

import (
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/logging"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepository interface {
	Create(in *entity.File) (*entity.File, error)
	GetLastID() (int, error)
	GetAllFilesSize() (int64, error)
	GetByHash(hash string) (*entity.File, error)
	GetByID(fileID int) (*entity.File, error)
	GetUnusedFileIDs(ctx context.Context) (<-chan int, error)
	DeleteByID(fileID int) error
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

func (r *fileRepository) GetLastID() (int, error) {
	query := `SELECT coalesce(max(id), 0) FROM files`

	var result int
	err := r.db.QueryRow(r.ctx, query).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *fileRepository) GetAllFilesSize() (int64, error) {
	query := `SELECT coalesce(sum(size), 0) FROM files`

	var result int64
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

func (r *fileRepository) GetByID(fileID int) (*entity.File, error) {
	query := `select * from files where id = $1`
	row := r.db.QueryRow(r.ctx, query, fileID)
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

func (r *fileRepository) GetUnusedFileIDs(ctx context.Context) (<-chan int, error) {
	ch := make(chan int)
	query := `select id from files f
		left join file_note_links fnl on fnl.file_id = f.id 
		where 
		fnl.note_id is null`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		defer rows.Close()

		for rows.Next() {
			var ID int
			if err := rows.Scan(&ID); err != nil {
				logging.GetLogger(ctx).Errorf("error scanning row GetUnusedFileIDs: %s", err)
				return
			}

			select {
			case <-ctx.Done():
				return
			case ch <- ID:
			}
		}
	}()

	return ch, nil
}

func (r *fileRepository) DeleteByID(fileID int) error {
	query := `DELETE FROM files WHERE id = $1`

	_, err := r.db.Exec(r.ctx, query, fileID)
	if err != nil {
		return err
	}
	return nil
}
