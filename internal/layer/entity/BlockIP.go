package entity

import "time"

type BlockIP struct {
	IP           string    `db:"ip"`
	BlockedUntil time.Time `db:"blocked_until"`
}
