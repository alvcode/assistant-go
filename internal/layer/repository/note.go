package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteRepository interface {
	Create(in entity.Note) (*entity.Note, error)
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
	query := `INSERT INTO note_categories (user_id, name, parent_id) VALUES ($1, $2, $3) RETURNING id`

	row := ur.db.QueryRow(ur.ctx, query, in.UserId, in.Name, in.ParentId)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return &in, nil
}
