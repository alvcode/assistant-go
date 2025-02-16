package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteCategoryRepository interface {
	Create(in entity.NoteCategory) (*entity.NoteCategory, error)
	FindAll(userId int) ([]*entity.NoteCategory, error)
	FindByIDAndUser(userId int, id int) (*entity.NoteCategory, error)
	DeleteById(catId int) error
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

func (ur *noteCategoryRepository) FindAll(userId int) ([]*entity.NoteCategory, error) {
	query := `SELECT * FROM note_categories WHERE user_id = $1`
	rows, err := ur.db.Query(ur.ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]*entity.NoteCategory, 0)
	for rows.Next() {
		category := &entity.NoteCategory{}
		if err := rows.Scan(&category.ID, &category.UserId, &category.Name, &category.ParentId); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (ur *noteCategoryRepository) FindByIDAndUser(userId int, id int) (*entity.NoteCategory, error) {
	query := `SELECT * FROM note_categories WHERE user_id = $1 and id = $2`
	row := ur.db.QueryRow(ur.ctx, query, userId, id)

	var noteCategory entity.NoteCategory
	if err := row.Scan(&noteCategory.ID, &noteCategory.UserId, &noteCategory.Name, &noteCategory.ParentId); err != nil {
		return nil, err
	}
	return &noteCategory, nil
}

func (ur *noteCategoryRepository) DeleteById(catId int) error {
	query := `DELETE FROM note_categories WHERE id = $1`
	_, err := ur.db.Exec(ur.ctx, query, catId)
	if err != nil {
		return err
	}
	return nil
}
