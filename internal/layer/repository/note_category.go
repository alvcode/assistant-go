package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteCategoryRepository interface {
	Create(in entity.NoteCategory) (*entity.NoteCategory, error)
	FindAll(userID int) ([]*entity.NoteCategory, error)
	FindByIDAndUser(userID int, id int) (*entity.NoteCategory, error)
	FindByIDAndUserWithChildren(userID int, id int) ([]*entity.NoteCategory, error)
	DeleteByIds(catIDs []int) error
	DeleteByUserId(userID int) error
	Update(in *entity.NoteCategory) error
	GetMaxPosition(userID int, parentID *int) (int, error)
	UpdatePosition(in *entity.NoteCategory) error
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
	query := `INSERT INTO note_categories (user_id, name, parent_id, position) VALUES ($1, $2, $3, $4) RETURNING id`

	row := ur.db.QueryRow(ur.ctx, query, in.UserId, in.Name, in.ParentId, in.Position)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return &in, nil
}

func (ur *noteCategoryRepository) FindAll(userID int) ([]*entity.NoteCategory, error) {
	query := `SELECT * FROM note_categories WHERE user_id = $1`
	rows, err := ur.db.Query(ur.ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]*entity.NoteCategory, 0)
	for rows.Next() {
		category := &entity.NoteCategory{}
		if err := rows.Scan(&category.ID, &category.UserId, &category.Name, &category.ParentId, &category.Position); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (ur *noteCategoryRepository) FindByIDAndUser(userID int, id int) (*entity.NoteCategory, error) {
	query := `SELECT * FROM note_categories WHERE user_id = $1 and id = $2`
	row := ur.db.QueryRow(ur.ctx, query, userID, id)

	var category entity.NoteCategory
	if err := row.Scan(&category.ID, &category.UserId, &category.Name, &category.ParentId, &category.Position); err != nil {
		return nil, err
	}
	return &category, nil
}

func (ur *noteCategoryRepository) FindByIDAndUserWithChildren(userID int, id int) ([]*entity.NoteCategory, error) {
	query := `
		WITH RECURSIVE subcategories AS (
			SELECT id, user_id, name, parent_id, position
			FROM note_categories
			WHERE id = $1 and user_id = $2
		
			UNION ALL
		
			SELECT c.id, c.user_id, c.name, c.parent_id, c.position
			FROM note_categories c
			INNER JOIN subcategories s ON c.parent_id = s.id
		)
		SELECT id, user_id, name, parent_id, position FROM subcategories
	`

	rows, err := ur.db.Query(ur.ctx, query, id, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]*entity.NoteCategory, 0)
	for rows.Next() {
		category := &entity.NoteCategory{}
		if err := rows.Scan(&category.ID, &category.UserId, &category.Name, &category.ParentId, &category.Position); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (ur *noteCategoryRepository) DeleteByIds(catIDs []int) error {
	query := `DELETE FROM note_categories WHERE id = ANY($1)`
	_, err := ur.db.Exec(ur.ctx, query, catIDs)
	if err != nil {
		return err
	}
	return nil
}

func (ur *noteCategoryRepository) DeleteByUserId(userID int) error {
	query := `DELETE FROM note_categories WHERE user_id = $1`
	_, err := ur.db.Exec(ur.ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *noteCategoryRepository) Update(in *entity.NoteCategory) error {
	query := `UPDATE note_categories SET name = $3, parent_id = $4, position = $5 WHERE id = $1 and user_id = $2`

	_, err := ur.db.Exec(ur.ctx, query, in.ID, in.UserId, in.Name, in.ParentId, in.Position)
	if err != nil {
		return err
	}
	return nil
}

func (ur *noteCategoryRepository) GetMaxPosition(userID int, parentID *int) (int, error) {
	var query string
	query = `SELECT coalesce(max(position), 0) FROM note_categories WHERE user_id = $1`
	if parentID != nil {
		query += " AND parent_id = $2"
	}

	var result int
	var err error
	if parentID != nil {
		err = ur.db.QueryRow(ur.ctx, query, userID, *parentID).Scan(&result)
	} else {
		err = ur.db.QueryRow(ur.ctx, query, userID).Scan(&result)
	}

	if err != nil {
		return 0, err
	}
	return result, nil
}

func (ur *noteCategoryRepository) UpdatePosition(in *entity.NoteCategory) error {
	query := `UPDATE note_categories SET position = $2 WHERE id = $1`

	_, err := ur.db.Exec(ur.ctx, query, in.ID, in.Position)
	if err != nil {
		return err
	}
	return nil
}
