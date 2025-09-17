package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileNoteLinkRepository interface {
	Upsert(ctx context.Context, noteID int, fileIDs []int) error
	DeleteByNoteID(ctx context.Context, noteID int) error
	Set(ctx context.Context, noteID int, fileIDs []int) error
}

type fileNoteLinkRepository struct {
	db *pgxpool.Pool
}

func NewFileNoteLinkRepository(db *pgxpool.Pool) FileNoteLinkRepository {
	return &fileNoteLinkRepository{db: db}
}

func (r *fileNoteLinkRepository) Upsert(ctx context.Context, noteID int, fileIDs []int) error {
	err := r.DeleteByNoteID(ctx, noteID)
	if err != nil {
		return err
	}
	if len(fileIDs) > 0 {
		err = r.Set(ctx, noteID, fileIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *fileNoteLinkRepository) DeleteByNoteID(ctx context.Context, noteID int) error {
	query := `DELETE FROM file_note_links WHERE note_id = $1`

	_, err := r.db.Exec(ctx, query, noteID)
	if err != nil {
		return err
	}
	return nil
}

func (r *fileNoteLinkRepository) Set(ctx context.Context, noteID int, fileIDs []int) error {
	for _, fileID := range fileIDs {
		query := `INSERT INTO file_note_links (file_id, note_id) VALUES ($1, $2)`

		_, err := r.db.Exec(ctx, query, fileID, noteID)
		if err != nil {
			return err
		}
	}
	return nil
}
