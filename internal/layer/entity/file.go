package entity

import "time"

type File struct {
	ID               int       `db:"id"`
	UserID           int       `db:"user_id"`
	OriginalFilename string    `db:"original_filename"`
	FilePath         string    `db:"file_path"`
	Ext              string    `db:"ext"`
	Size             int       `db:"size"`
	Hash             string    `db:"hash"`
	CreatedAt        time.Time `db:"created_at"`
}
