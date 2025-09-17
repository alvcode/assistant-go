package repository

import (
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/logging"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepository interface {
	Create(ctx context.Context, in *entity.File) (*entity.File, error)
	GetLastID(ctx context.Context) (int, error)
	GetAllFilesSize(ctx context.Context) (int64, error)
	GetByHash(ctx context.Context, hash string) (*entity.File, error)
	GetByID(ctx context.Context, fileID int) (*entity.File, error)
	GetUnusedFileIDs(ctx context.Context) (<-chan int, error)
	DeleteByID(ctx context.Context, fileID int) error
}

type fileRepository struct {
	db *pgxpool.Pool
}

func NewFileRepository(db *pgxpool.Pool) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(ctx context.Context, in *entity.File) (*entity.File, error) {
	query := `
		INSERT INTO files (user_id, original_filename, file_path, ext, size, hash, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`

	row := r.db.QueryRow(
		ctx,
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

func (r *fileRepository) GetLastID(ctx context.Context) (int, error) {
	query := `SELECT coalesce(max(id), 0) FROM files`

	var result int
	err := r.db.QueryRow(ctx, query).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *fileRepository) GetAllFilesSize(ctx context.Context) (int64, error) {
	query := `SELECT coalesce(sum(size), 0) FROM files`

	var result int64
	err := r.db.QueryRow(ctx, query).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *fileRepository) GetByHash(ctx context.Context, hash string) (*entity.File, error) {
	query := `select * from files where hash = $1`
	row := r.db.QueryRow(ctx, query, hash)
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

func (r *fileRepository) GetByID(ctx context.Context, fileID int) (*entity.File, error) {
	query := `select * from files where id = $1`
	row := r.db.QueryRow(ctx, query, fileID)
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

func (r *fileRepository) DeleteByID(ctx context.Context, fileID int) error {
	query := `DELETE FROM files WHERE id = $1`

	_, err := r.db.Exec(ctx, query, fileID)
	if err != nil {
		return err
	}
	return nil
}
