package entity

import "time"

type DriveRecycleBinStruct struct {
	ID            int       `db:"id"`
	Name          string    `db:"name"`
	Type          int8      `db:"type"`
	DriveStructID int       `db:"drive_struct_id"`
	CreatedAt     time.Time `db:"created_at"`
	OriginalPath  string    `db:"original_path"`
}
