package entity

type NoteShare struct {
	ID     int    `db:"id"`
	NoteID int    `db:"note_id"`
	Hash   string `db:"hash"`
}
