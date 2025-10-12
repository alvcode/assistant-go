package entity

type DriveFileChunk struct {
	ID          int    `db:"id"`
	DriveFileID int    `db:"drive_file_id"`
	Path        string `db:"path"`
	Size        int64  `db:"size"`
	ChunkNumber int    `db:"chunk_number"`
}
