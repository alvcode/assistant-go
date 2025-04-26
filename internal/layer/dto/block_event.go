package dto

type BlockEventsStat struct {
	All               int `db:"all"`
	ValidateInputData int `db:"validate_input_data"`
	DecodeBody        int `db:"decode_body"`
	SignIn            int `db:"sign_in"`
	Unauthorized      int `db:"unauthorized"`
	RefreshToken      int `db:"refresh_token"`
	PageNotFound      int `db:"page_not_found"`
	TooManyRequests   int `db:"too_many_requests"`
}
