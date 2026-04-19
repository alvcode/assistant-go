package entity

import "time"

type DriveRecycleBin struct {
	ID            int       `db:"id"`
	DriveStructID int       `db:"drive_struct_id"`
	CreatedAt     time.Time `db:"created_at"`
	OriginalPath  string    `db:"original_path"`
}

type DriveRecycleBinStruct struct {
	ID            int       `db:"id"`
	Name          string    `db:"name"`
	Type          int8      `db:"type"`
	DriveStructID int       `db:"drive_struct_id"`
	CreatedAt     time.Time `db:"created_at"`
	OriginalPath  string    `db:"original_path"`
}
