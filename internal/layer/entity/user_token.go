package entity

type UserToken struct {
	UserId       uint32 `json:"user_id" db:"user_id"`
	Token        string `json:"token" db:"token"`
	RefreshToken string `json:"refresh_token" db:"refresh_token"`
	ExpiredTo    uint32 `json:"expired_to" db:"expired_to"`
}
