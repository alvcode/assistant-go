package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DriveFileRepository interface {
	GetStorageSize(userID int) (int64, error)
	GetByStructID(structID int) (*entity.DriveFile, error)
	GetLastID() (int, error)
	Create(in *entity.DriveFile) (*entity.DriveFile, error)
	GetAllRecursive(structID int, userID int) ([]*entity.DriveFile, error)
}

type driveFileRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewDriveFileRepository(ctx context.Context, db *pgxpool.Pool) DriveFileRepository {
	return &driveFileRepository{
		ctx: ctx,
		db:  db,
	}
}

func (r *driveFileRepository) GetStorageSize(userID int) (int64, error) {
	query := `SELECT 
    		coalesce(sum(df.size), 0) 
		FROM drive_structs ds 
		JOIN drive_files df on df.drive_struct_id = ds.id
		WHERE ds.user_id = $1
	`

	var result int64
	err := r.db.QueryRow(r.ctx, query, userID).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *driveFileRepository) GetByStructID(structID int) (*entity.DriveFile, error) {
	query := `select * from drive_files where drive_struct_id = $1`

	var result entity.DriveFile
	err := r.db.QueryRow(r.ctx, query, structID).Scan(
		&result.ID,
		&result.DriveStructID,
		&result.Path,
		&result.Ext,
		&result.Size,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *driveFileRepository) GetLastID() (int, error) {
	query := `SELECT coalesce(max(id), 0) FROM drive_files`

	var result int
	err := r.db.QueryRow(r.ctx, query).Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *driveFileRepository) Create(in *entity.DriveFile) (*entity.DriveFile, error) {
	query := `
		INSERT INTO drive_files (drive_struct_id, path, ext, size, created_at) 
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`

	row := r.db.QueryRow(
		r.ctx,
		query,
		in.DriveStructID,
		in.Path,
		in.Ext,
		in.Size,
		in.CreatedAt,
	)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return in, nil
}

func (r *driveFileRepository) GetAllRecursive(structID int, userID int) ([]*entity.DriveFile, error) {
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

	rows, err := r.db.Query(r.ctx, query, structID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*entity.DriveFile, 0)

	for rows.Next() {
		df := &entity.DriveFile{}
		if err := rows.Scan(&df.ID, &df.DriveStructID, &df.Path, &df.Ext, &df.Size, &df.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, df)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
