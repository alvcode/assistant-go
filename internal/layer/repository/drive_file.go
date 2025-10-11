package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
)

type DriveFileRepository interface {
	GetStorageSize(ctx context.Context, userID int) (int64, error)
	GetByStructID(ctx context.Context, structID int) (*entity.DriveFile, error)
	GetLastID(ctx context.Context) (int, error)
	Create(ctx context.Context, in *entity.DriveFile) (*entity.DriveFile, error)
	GetAllRecursive(ctx context.Context, structID int, userID int) ([]*entity.DriveFile, error)
}

type driveFileRepository struct {
	db DBExecutor
}

func NewDriveFileRepository(db DBExecutor) DriveFileRepository {
	return &driveFileRepository{db: db}
}

func (r *driveFileRepository) GetStorageSize(ctx context.Context, userID int) (int64, error) {
	query := `SELECT 
    		coalesce(sum(df.size), 0) 
		FROM drive_structs ds 
		JOIN drive_files df on df.drive_struct_id = ds.id
		WHERE ds.user_id = $1
	`

	var result int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *driveFileRepository) GetByStructID(ctx context.Context, structID int) (*entity.DriveFile, error) {
	query := `select * from drive_files where drive_struct_id = $1`

	var result entity.DriveFile
	err := r.db.QueryRow(ctx, query, structID).Scan(
		&result.ID,
		&result.DriveStructID,
		&result.Path,
		&result.Ext,
		&result.Size,
		&result.CreatedAt,
		&result.IsChunk,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *driveFileRepository) GetLastID(ctx context.Context) (int, error) {
	query := `SELECT coalesce(max(id), 0) FROM drive_files`

	var result int
	err := r.db.QueryRow(ctx, query).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *driveFileRepository) Create(ctx context.Context, in *entity.DriveFile) (*entity.DriveFile, error) {
	query := `
		INSERT INTO drive_files (drive_struct_id, path, ext, size, created_at, is_chunk) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`

	row := r.db.QueryRow(
		ctx,
		query,
		in.DriveStructID,
		in.Path,
		in.Ext,
		in.Size,
		in.CreatedAt,
		in.IsChunk,
	)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return in, nil
}

func (r *driveFileRepository) GetAllRecursive(ctx context.Context, structID int, userID int) ([]*entity.DriveFile, error) {
	query := `
		select * from drive_files df 
		where 
		df.drive_struct_id in (
			WITH RECURSIVE structs AS (
				SELECT id
				FROM drive_structs 
				WHERE id = $1 and user_id = $2
			
				UNION ALL
			
				SELECT ds.id
				FROM drive_structs ds
				INNER JOIN structs s ON ds.parent_id = s.id
			)
			SELECT id FROM structs
		)
	`

	rows, err := r.db.Query(ctx, query, structID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*entity.DriveFile, 0)

	for rows.Next() {
		df := &entity.DriveFile{}
		if err := rows.Scan(&df.ID, &df.DriveStructID, &df.Path, &df.Ext, &df.Size, &df.CreatedAt, &df.IsChunk); err != nil {
			return nil, err
		}
		result = append(result, df)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
