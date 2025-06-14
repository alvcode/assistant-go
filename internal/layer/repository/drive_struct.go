package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DriveStructRepository interface {
	GetByID(ID int) (*entity.DriveStruct, error)
	FindRow(userID int, name string, rowType int8, parentID *int) (*entity.DriveStruct, error)
	Create(entity *entity.DriveStruct) (*entity.DriveStruct, error)
	ListByUserID(userID int, parentID *int) ([]*entity.DriveStruct, error)
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

func (r *driveStructRepository) GetByID(ID int) (*entity.DriveStruct, error) {
	query := `SELECT * FROM drive_structs WHERE id = $1`

	row := r.db.QueryRow(r.ctx, query, ID)

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

func (r *driveStructRepository) Create(in *entity.DriveStruct) (*entity.DriveStruct, error) {
	query := `
		INSERT INTO drive_structs 
		    (user_id, name, type, parent_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`

	row := r.db.QueryRow(r.ctx, query, in.UserID, in.Name, in.Type, in.ParentID, in.CreatedAt, in.UpdatedAt)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return in, nil
}

func (r *driveStructRepository) ListByUserID(userID int, parentID *int) ([]*entity.DriveStruct, error) {
	var (
		query string
		args  []any
	)

	if parentID == nil {
		query = `select * from drive_structs where user_id = $1 and parent_id is null`
		args = []any{userID}
	} else {
		query = `select * from drive_structs where user_id = $1 and parent_id = $2`
		args = []any{userID, parentID}
	}

	rows, err := r.db.Query(r.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	structs := make([]*entity.DriveStruct, 0)
	for rows.Next() {
		ds := &entity.DriveStruct{}
		if err := rows.Scan(
			&ds.ID,
			&ds.UserID,
			&ds.Name,
			&ds.Type,
			&ds.ParentID,
			&ds.CreatedAt,
			&ds.UpdatedAt,
		); err != nil {
			return nil, err
		}
		structs = append(structs, ds)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return structs, nil
}
