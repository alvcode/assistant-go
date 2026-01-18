package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteRepository interface {
	Create(ctx context.Context, in entity.Note) (*entity.Note, error)
	Update(ctx context.Context, in *entity.Note) error
	GetById(ctx context.Context, ID int) (*entity.Note, error)
	GetMinimalByCategoryIds(ctx context.Context, catIds []int) ([]*entity.NoteMinimal, error)
	DeleteOne(ctx context.Context, noteID int) error
	CheckExistsByCategoryIDs(ctx context.Context, catIDs []int) (bool, error)
	Pin(ctx context.Context, noteID int) error
	UnPin(ctx context.Context, noteID int) error
	BelongsToUser(ctx context.Context, noteID int, userID int) (bool, error)
}

type noteRepository struct {
	db *pgxpool.Pool
}

func NewNoteRepository(db *pgxpool.Pool) NoteRepository {
	return &noteRepository{db: db}
}

func (ur *noteRepository) Create(ctx context.Context, in entity.Note) (*entity.Note, error) {
	query := `INSERT INTO notes (category_id, note_blocks, created_at, updated_at, title, pinned) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	row := ur.db.QueryRow(ctx, query, in.CategoryID, in.NoteBlocks, in.CreatedAt, in.UpdatedAt, in.Title, in.Pinned)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return &in, nil
}

func (ur *noteRepository) Update(ctx context.Context, in *entity.Note) error {
	query := `UPDATE notes SET category_id = $2, note_blocks = $3, updated_at = $4, title = $5, pinned = $6 WHERE id = $1`

	_, err := ur.db.Exec(ctx, query, in.ID, in.CategoryID, in.NoteBlocks, in.UpdatedAt, in.Title, in.Pinned)
	if err != nil {
		return err
	}
	return nil
}

func (ur *noteRepository) GetById(ctx context.Context, ID int) (*entity.Note, error) {
	query := `select * from notes where id = $1`
	row := ur.db.QueryRow(ctx, query, ID)
	var note entity.Note
	if err := row.Scan(&note.ID, &note.CategoryID, &note.NoteBlocks, &note.CreatedAt, &note.UpdatedAt, &note.Title, &note.Pinned); err != nil {
		return nil, err
	}
	return &note, nil
}

func (ur *noteRepository) GetMinimalByCategoryIds(ctx context.Context, catIDs []int) ([]*entity.NoteMinimal, error) {
	query := `
		select 
		    n.id, 
		    n.category_id, 
		    n.created_at, 
		    n.updated_at, 
		    n.title, 
		    n.pinned,
		    (SELECT EXISTS(SELECT 1 FROM note_share_hashes WHERE note_id = n.id)) as shared
		from notes n 
		where n.category_id = ANY($1)
	`

	rows, err := ur.db.Query(ctx, query, catIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]*entity.NoteMinimal, 0)
	for rows.Next() {
		note := &entity.NoteMinimal{}
		if err := rows.Scan(&note.ID, &note.CategoryID, &note.CreatedAt, &note.UpdatedAt, &note.Title, &note.Pinned, &note.Shared); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (ur *noteRepository) DeleteOne(ctx context.Context, noteID int) error {
	query := `DELETE FROM notes WHERE id = $1`

	_, err := ur.db.Exec(ctx, query, noteID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *noteRepository) CheckExistsByCategoryIDs(ctx context.Context, catIDs []int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM notes WHERE category_id = ANY($1))`

	var exists bool
	err := ur.db.QueryRow(ctx, query, catIDs).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (ur *noteRepository) Pin(ctx context.Context, noteID int) error {
	query := `UPDATE notes SET pinned = true WHERE id = $1`

	_, err := ur.db.Exec(ctx, query, noteID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *noteRepository) UnPin(ctx context.Context, noteID int) error {
	query := `UPDATE notes SET pinned = false WHERE id = $1`

	_, err := ur.db.Exec(ctx, query, noteID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *noteRepository) BelongsToUser(ctx context.Context, noteID int, userID int) (bool, error) {
	query := `
		select EXISTS(
			select 1 from note_categories nc 
			left join notes n on n.category_id = nc.id 
			where
			n.id = $1 and nc.user_id = $2
		)
	`

	var exists bool
	err := ur.db.QueryRow(ctx, query, noteID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
