package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteRepository interface {
	Create(in entity.Note) (*entity.Note, error)
	Update(in *entity.Note) error
	GetById(ID int) (*entity.Note, error)
	GetMinimalByCategoryIds(catIds []int) ([]*entity.Note, error)
	DeleteOne(noteID int) error
	CheckExistsByCategoryIDs(catIDs []int) (bool, error)
	Pin(noteID int) error
	UnPin(noteID int) error
}

type noteRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewNoteRepository(ctx context.Context, db *pgxpool.Pool) NoteRepository {
	return &noteRepository{
		ctx: ctx,
		db:  db,
	}
}

func (ur *noteRepository) Create(in entity.Note) (*entity.Note, error) {
	query := `INSERT INTO notes (category_id, note_blocks, created_at, updated_at, title, pinned) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	row := ur.db.QueryRow(ur.ctx, query, in.CategoryID, in.NoteBlocks, in.CreatedAt, in.UpdatedAt, in.Title, in.Pinned)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return &in, nil
}

func (ur *noteRepository) Update(in *entity.Note) error {
	query := `UPDATE notes SET category_id = $2, note_blocks = $3, updated_at = $4, title = $5, pinned = $6 WHERE id = $1`

	_, err := ur.db.Exec(ur.ctx, query, in.ID, in.CategoryID, in.NoteBlocks, in.UpdatedAt, in.Title, in.Pinned)
	if err != nil {
		return err
	}
	return nil
}

func (ur *noteRepository) GetById(ID int) (*entity.Note, error) {
	query := `select * from notes where id = $1`
	row := ur.db.QueryRow(ur.ctx, query, ID)
	var note entity.Note
	if err := row.Scan(&note.ID, &note.CategoryID, &note.NoteBlocks, &note.CreatedAt, &note.UpdatedAt, &note.Title, &note.Pinned); err != nil {
		return nil, err
	}
	return &note, nil
}

func (ur *noteRepository) GetMinimalByCategoryIds(catIds []int) ([]*entity.Note, error) {
	query := `select n.id, n.category_id, n.created_at, n.updated_at, n.title, n.pinned from notes n where n.category_id = ANY($1)`

	rows, err := ur.db.Query(ur.ctx, query, catIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]*entity.Note, 0)
	for rows.Next() {
		note := &entity.Note{}
		if err := rows.Scan(&note.ID, &note.CategoryID, &note.CreatedAt, &note.UpdatedAt, &note.Title, &note.Pinned); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (ur *noteRepository) DeleteOne(noteID int) error {
	query := `DELETE FROM notes WHERE id = $1`

	_, err := ur.db.Exec(ur.ctx, query, noteID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *noteRepository) CheckExistsByCategoryIDs(catIDs []int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM notes WHERE category_id = ANY($1))`

	var exists bool
	err := ur.db.QueryRow(ur.ctx, query, catIDs).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (ur *noteRepository) Pin(noteID int) error {
	query := `UPDATE notes SET pinned = true WHERE id = $1`

	_, err := ur.db.Exec(ur.ctx, query, noteID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *noteRepository) UnPin(noteID int) error {
	query := `UPDATE notes SET pinned = false WHERE id = $1`

	_, err := ur.db.Exec(ur.ctx, query, noteID)
	if err != nil {
		return err
	}
	return nil
}
