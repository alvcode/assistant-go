package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DriveStructRepository interface {
	FindRow(userID int, name string, rowType int8, parentID *int) (*entity.DriveStruct, error)
	CreateDirectory(entity *entity.DriveStruct) error
}

type driveStructRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewDriveStructRepository(ctx context.Context, db *pgxpool.Pool) DriveStructRepository {
	return &driveStructRepository{
		ctx: ctx,
		db:  db,
	}
}

func (r *driveStructRepository) FindRow(
	userID int,
	name string,
	rowType int8,
	parentID *int,
) (*entity.DriveStruct, error) {
	var (
		query string
		args  []any
	)

	if parentID == nil {
		query = `SELECT * FROM drive_structs WHERE user_id = $1 AND name = $2 AND type = $3 AND parent_id IS NULL`
		args = []any{userID, name, rowType}
	} else {
		query = `SELECT * FROM drive_structs WHERE user_id = $1 AND name = $2 AND type = $3 AND parent_id = $4`
		args = []any{userID, name, rowType, parentID}
	}

	row := r.db.QueryRow(r.ctx, query, args...)

	var driveStruct entity.DriveStruct
	if err := row.Scan(
		&driveStruct.ID,
		&driveStruct.UserID,
		&driveStruct.Name,
		&driveStruct.Type,
		&driveStruct.ParentID,
		&driveStruct.CreatedAt,
		&driveStruct.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &driveStruct, nil
}

func (r *driveStructRepository) CreateDirectory(in *entity.DriveStruct) error {
	query := `
		INSERT INTO drive_structs 
		    (user_id, name, type, parent_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`

	row := r.db.QueryRow(r.ctx, query, in.UserID, in.Name, in.Type, in.ParentID, in.CreatedAt, in.UpdatedAt)

	if err := row.Scan(&in.ID); err != nil {
		return err
	}
	return nil
}
