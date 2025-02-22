package entity

import (
	"encoding/json"
	"time"
)

type Note struct {
	ID         int             `db:"id"`
	CategoryID int             `db:"id"`
	NoteBlocks json.RawMessage `db:"note_blocks"`
	CreatedAt  time.Time       `db:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at"`
}
