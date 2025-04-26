package entity

type RateLimiter struct {
	IP                string `db:"ip"`
	AllowanceRequests int    `db:"allowance"`
	Timestamp         int    `db:"timestamp"`
}
