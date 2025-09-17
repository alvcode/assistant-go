package entity

import "time"

type DriveFile struct {
	ID            int       `db:"id"`
	DriveStructID int       `db:"drive_struct_id"`
	Path          string    `db:"path"`
	Ext           string    `db:"ext"`
	Size          int       `db:"size"`
	CreatedAt     time.Time `db:"created_at"`
}
