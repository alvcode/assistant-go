package entity

type NoteCategory struct {
	ID       int    `db:"id"`
	UserId   int    `db:"user_id"`
	Name     string `db:"name"`
	ParentId *int   `db:"parent_id"`
	Position int    `db:"position"`
}
