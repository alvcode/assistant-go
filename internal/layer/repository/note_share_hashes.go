package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
)

type NoteShareHashesRepository interface {
	Create(ctx context.Context, in entity.NoteShare) (*entity.NoteShare, error)
	ExistsByNoteID(ctx context.Context, noteID int) (bool, error)
	ExistsByHash(ctx context.Context, hash string) (bool, error)
	GetByNoteID(ctx context.Context, noteID int) (*entity.NoteShare, error)
	DeleteByNoteID(ctx context.Context, noteID int) error
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

func (ur *noteShareHashesRepository) GetByNoteID(ctx context.Context, noteID int) (*entity.NoteShare, error) {
	query := `select * from note_share_hashes where note_id = $1`
	row := ur.db.QueryRow(ctx, query, noteID)
	var noteShare entity.NoteShare
	if err := row.Scan(&noteShare.ID, &noteShare.NoteID, &noteShare.Hash); err != nil {
		return nil, err
	}
	return &noteShare, nil
}

func (ur *noteShareHashesRepository) DeleteByNoteID(ctx context.Context, noteID int) error {
	query := `DELETE FROM note_share_hashes WHERE note_id = $1`

	_, err := ur.db.Exec(ctx, query, noteID)
	if err != nil {
		return err
	}
	return nil
}
