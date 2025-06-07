package entity

import "time"

type DriveStruct struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Name      string    `db:"name"`
	Type      int8      `db:"type"`
	ParentID  string    `db:"parent_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
