package repository

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DriveStructRepository interface {
	GetByID(ctx context.Context, ID int) (*entity.DriveStruct, error)
	FindRow(ctx context.Context, userID int, name string, rowType int8, parentID *int) (*entity.DriveStruct, error)
	Create(ctx context.Context, entity *entity.DriveStruct) (*entity.DriveStruct, error)
	Update(ctx context.Context, in *entity.DriveStruct) error
	TreeByUserID(ctx context.Context, userID int, parentID *int) ([]*dto.DriveTree, error)
	GetAllRecursive(ctx context.Context, userID int, structID int) ([]*entity.DriveStruct, error)
	DeleteRecursive(ctx context.Context, userID int, structID int) error
}

type driveStructRepository struct {
	db *pgxpool.Pool
}

func NewDriveStructRepository(db *pgxpool.Pool) DriveStructRepository {
	return &driveStructRepository{db: db}
}

func (r *driveStructRepository) GetByID(ctx context.Context, ID int) (*entity.DriveStruct, error) {
	query := `SELECT * FROM drive_structs WHERE id = $1`

	row := r.db.QueryRow(ctx, query, ID)

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
	ctx context.Context,
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

	row := r.db.QueryRow(ctx, query, args...)

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

func (r *driveStructRepository) Create(ctx context.Context, in *entity.DriveStruct) (*entity.DriveStruct, error) {
	query := `
		INSERT INTO drive_structs 
		    (user_id, name, type, parent_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`

	row := r.db.QueryRow(ctx, query, in.UserID, in.Name, in.Type, in.ParentID, in.CreatedAt, in.UpdatedAt)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return in, nil
}

func (r *driveStructRepository) Update(ctx context.Context, in *entity.DriveStruct) error {
	query := `
		UPDATE drive_structs SET user_id = $1, name = $2, type = $3, parent_id = $4, created_at = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(ctx, query, in.UserID, in.Name, in.Type, in.ParentID, in.CreatedAt, in.UpdatedAt, in.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *driveStructRepository) TreeByUserID(ctx context.Context, userID int, parentID *int) ([]*dto.DriveTree, error) {
	var (
		query string
		args  []any
	)

	if parentID == nil {
		query = `
			select 
			    ds.id, ds.user_id, ds.name, ds.type, ds.created_at, ds.updated_at,
			    coalesce((select df.size from drive_files df where df.drive_struct_id = ds.id), 0) as size
			from drive_structs ds 
			where user_id = $1 and parent_id is null
		`
		args = []any{userID}
	} else {
		query = `
			select 
			    ds.id, ds.user_id, ds.name, ds.type, ds.created_at, ds.updated_at,
			    coalesce((select df.size from drive_files df where df.drive_struct_id = ds.id), 0) as size
			from drive_structs ds
			where user_id = $1 and parent_id = $2
		`
		args = []any{userID, parentID}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	structs := make([]*dto.DriveTree, 0)
	for rows.Next() {
		ds := &dto.DriveTree{}
		if err := rows.Scan(
			&ds.ID,
			&ds.UserID,
			&ds.Name,
			&ds.Type,
			&ds.CreatedAt,
			&ds.UpdatedAt,
			&ds.Size,
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

func (r *driveStructRepository) GetAllRecursive(ctx context.Context, userID int, structID int) ([]*entity.DriveStruct, error) {
	query := `
		WITH RECURSIVE structs AS (
			SELECT *
			FROM drive_structs 
			WHERE id = $1 and user_id = $2
		
			UNION ALL
		
			SELECT ds.*
			FROM drive_structs ds
			INNER JOIN structs s ON ds.parent_id = s.id
		)
		SELECT * FROM structs
	`

	rows, err := r.db.Query(ctx, query, structID, userID)
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

func (r *driveStructRepository) DeleteRecursive(ctx context.Context, userID int, structID int) error {
	query := `
		DELETE FROM drive_structs
		WHERE id in (
		    WITH RECURSIVE structs AS (
				SELECT *
				FROM drive_structs 
				WHERE id = $1 and user_id = $2
			
				UNION ALL
			
				SELECT ds.*
				FROM drive_structs ds
				INNER JOIN structs s ON ds.parent_id = s.id
			)
			SELECT id FROM structs
		)
	`

	_, err := r.db.Exec(ctx, query, structID, userID)
	if err != nil {
		return err
	}
	return nil
}
