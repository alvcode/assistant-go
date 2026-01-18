package entity

import (
	"encoding/json"
	"time"
)

type Note struct {
	ID         int             `db:"id"`
	CategoryID int             `db:"category_id"`
	NoteBlocks json.RawMessage `db:"note_blocks"`
	CreatedAt  time.Time       `db:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at"`
	Title      *string         `db:"title"`
	Pinned     bool            `db:"pinned"`
}

type NoteMinimal struct {
	ID         int       `db:"id"`
	CategoryID int       `db:"category_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	Title      *string   `db:"title"`
	Pinned     bool      `db:"pinned"`
	Shared     bool      `db:"shared"`
}
