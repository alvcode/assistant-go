package repository

import (
	"context"
)

type NoteShareHashesRepository interface {
	ExistsByNoteID(ctx context.Context, noteID int) (bool, error)
}

type noteShareHashesRepository struct {
	db DBExecutor
}

func NewNoteShareHashesRepository(db DBExecutor) NoteShareHashesRepository {
	return &noteShareHashesRepository{db: db}
}

func (ur *noteShareHashesRepository) ExistsByNoteID(ctx context.Context, noteID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM note_share_hashes WHERE note_id = ANY($1))`

	var exists bool
	err := ur.db.QueryRow(ctx, query, noteID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
