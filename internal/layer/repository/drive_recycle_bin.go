package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"time"
)

type DriveRecycleBinRepository interface {
	Upsert(ctx context.Context, structID int, originalPath string, createdAt time.Time) error
	GetAll(ctx context.Context, userID int) ([]*entity.DriveRecycleBinStruct, error)
	GetByID(ctx context.Context, ID int) (*entity.DriveRecycleBin, error)
	DeleteByID(ctx context.Context, ID int) error
}

type driveRecycleBinRepository struct {
	db DBExecutor
}

func NewDriveRecycleBinRepository(db DBExecutor) DriveRecycleBinRepository {
	return &driveRecycleBinRepository{db: db}
}

func (r *driveRecycleBinRepository) Upsert(ctx context.Context, structID int, originalPath string, createdAt time.Time) error {
	query := `
		INSERT INTO drive_recycle_bin (drive_struct_id, created_at, original_path)
		VALUES ($1, $2, $3)
		ON CONFLICT (drive_struct_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, structID, createdAt, originalPath)
	if err != nil {
		return err
	}

	return nil
}

func (r *driveRecycleBinRepository) GetAll(ctx context.Context, userID int) ([]*entity.DriveRecycleBinStruct, error) {
	query := `
		select 
			drb.id, ds.name, ds.type, drb.drive_struct_id, drb.created_at, drb.original_path 
		from drive_recycle_bin drb
		join drive_structs ds on ds.id = drb.drive_struct_id
		where
		ds.user_id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*entity.DriveRecycleBinStruct, 0)
	for rows.Next() {
		i := &entity.DriveRecycleBinStruct{}
		if err := rows.Scan(&i.ID, &i.Name, &i.Type, &i.DriveStructID, &i.CreatedAt, &i.OriginalPath); err != nil {
			return nil, err
		}
		result = append(result, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *driveRecycleBinRepository) GetByID(ctx context.Context, ID int) (*entity.DriveRecycleBin, error) {
	query := `select * from drive_recycle_bin where id = $1`

	row := r.db.QueryRow(ctx, query, ID)

	var result entity.DriveRecycleBin
	if err := row.Scan(&result.ID, &result.DriveStructID, &result.CreatedAt, &result.OriginalPath); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *driveRecycleBinRepository) DeleteByID(ctx context.Context, ID int) error {
	query := `DELETE FROM drive_recycle_bin WHERE id = $1`
	_, err := r.db.Exec(ctx, query, ID)
	if err != nil {
		return err
	}
	return nil
}
