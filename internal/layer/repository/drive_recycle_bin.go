package repository

import (
	"context"
	"time"
)

type DriveRecycleBinRepository interface {
	Upsert(ctx context.Context, structID int, createdAt time.Time) error
}

type driveRecycleBinRepository struct {
	db DBExecutor
}

func NewDriveRecycleBinRepository(db DBExecutor) DriveRecycleBinRepository {
	return &driveRecycleBinRepository{db: db}
}

func (r *driveRecycleBinRepository) Upsert(ctx context.Context, structID int, createdAt time.Time) error {
	query := `
		INSERT INTO drive_recycle_bin (drive_struct_id, created_at)
		VALUES ($1, $2)
		ON CONFLICT (drive_struct_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, structID, createdAt)
	if err != nil {
		return err
	}

	return nil
}
