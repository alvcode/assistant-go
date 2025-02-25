package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteRepository interface {
	Create(in entity.Note) (*entity.Note, error)
	GetMinimalByCategoryIds(catIds []int) ([]*entity.Note, error)
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
	query := `INSERT INTO notes (category_id, note_blocks, created_at, updated_at, title) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	row := ur.db.QueryRow(ur.ctx, query, in.CategoryID, in.NoteBlocks, in.CreatedAt, in.UpdatedAt, in.Title)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return &in, nil
}

func (ur *noteRepository) GetMinimalByCategoryIds(catIds []int) ([]*entity.Note, error) {
	query := `select n.id, n.category_id, n.created_at, n.updated_at, n.title from notes n where n.category_id = ANY($1)`

	rows, err := ur.db.Query(ur.ctx, query, catIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]*entity.Note, 0)
	for rows.Next() {
		note := &entity.Note{}
		if err := rows.Scan(&note.ID, &note.CategoryID, &note.CreatedAt, &note.UpdatedAt, &note.Title); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}
