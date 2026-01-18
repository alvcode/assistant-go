package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
)

type NoteShareHashesRepository interface {
	Create(ctx context.Context, in entity.NoteShare) (*entity.NoteShare, error)
	ExistsByNoteID(ctx context.Context, noteID int) (bool, error)
	ExistsByHash(ctx context.Context, hash string) (bool, error)
}

type noteShareHashesRepository struct {
	db DBExecutor
}

func NewNoteShareHashesRepository(db DBExecutor) NoteShareHashesRepository {
	return &noteShareHashesRepository{db: db}
}

func (ur *noteShareHashesRepository) Create(ctx context.Context, in entity.NoteShare) (*entity.NoteShare, error) {
	query := `INSERT INTO note_share_hashes (note_id, hash) VALUES ($1, $2) RETURNING id`

	row := ur.db.QueryRow(ctx, query, in.NoteID, in.Hash)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}
	return &in, nil
}

func (ur *noteShareHashesRepository) ExistsByNoteID(ctx context.Context, noteID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM note_share_hashes WHERE note_id = $1)`

	var exists bool
	err := ur.db.QueryRow(ctx, query, noteID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (ur *noteShareHashesRepository) ExistsByHash(ctx context.Context, hash string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM note_share_hashes WHERE hash = $1)`

	var exists bool
	err := ur.db.QueryRow(ctx, query, hash).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
