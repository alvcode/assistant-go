package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileNoteLinkRepository interface {
	Upsert(noteID int, fileIDs []int) error
	DeleteByNoteID(noteID int) error
	Set(noteID int, fileIDs []int) error
}

type fileNoteLinkRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewFileNoteLinkRepository(ctx context.Context, db *pgxpool.Pool) FileNoteLinkRepository {
	return &fileNoteLinkRepository{
		ctx: ctx,
		db:  db,
	}
}

func (r *fileNoteLinkRepository) Upsert(noteID int, fileIDs []int) error {
	err := r.DeleteByNoteID(noteID)
	if err != nil {
		return err
	}
	if len(fileIDs) > 0 {
		err = r.Set(noteID, fileIDs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *fileNoteLinkRepository) DeleteByNoteID(noteID int) error {
	query := `DELETE FROM file_note_links WHERE note_id = $1`

	_, err := r.db.Exec(r.ctx, query, noteID)
	if err != nil {
		return err
	}
	return nil
}

func (r *fileNoteLinkRepository) Set(noteID int, fileIDs []int) error {
	for _, fileID := range fileIDs {
		query := `INSERT INTO file_note_links (file_id, note_id) VALUES ($1, $2)`

		_, err := r.db.Exec(r.ctx, query, fileID, noteID)
		if err != nil {
			return err
		}
	}
	return nil
}
