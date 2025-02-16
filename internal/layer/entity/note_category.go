package entity

type NoteCategory struct {
	ID       int    `json:"id" db:"id"`
	UserId   int    `json:"user_id" db:"user_id"`
	Name     string `json:"name" db:"name"`
	ParentId *int   `json:"parent_id" db:"parent_id"`
}
