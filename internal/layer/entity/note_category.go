package entity

type NoteCategory struct {
	ID       uint32 `json:"id" db:"id"`
	UserId   uint32 `json:"user_id" db:"user_id"`
	Name     string `json:"name" db:"name"`
	ParentId string `json:"parent_id" db:"parent_id"`
}
