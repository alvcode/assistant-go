package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteCategoryRepository interface {
	Create(in entity.NoteCategory) (*entity.NoteCategory, error)
}

type noteCategoryRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewNoteCategoryRepository(ctx context.Context, db *pgxpool.Pool) NoteCategoryRepository {
	return &noteCategoryRepository{
		ctx: ctx,
		db:  db,
	}
}

func (ur *noteCategoryRepository) Create(in entity.NoteCategory) (*entity.NoteCategory, error) {
	query := `INSERT INTO note_categories (user_id, name, parent_id) VALUES ($1, $2, $3) RETURNING id`

	row := ur.db.QueryRow(ur.ctx, query, in.UserId, in.Name, in.ParentId)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}

	return &in, nil
}
