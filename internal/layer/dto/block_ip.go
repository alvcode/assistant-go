package dto

import "assistant-go/pkg/vld"

type BlockIP struct {
	IP string `json:"ip" validate:"required"`
}

func (dto *BlockIP) Validate(lang string) error {
	err := vld.Validate.Struct(dto)
	if err != nil {
		return vld.TextFromFirstError(err, lang)
	}
	return nil
}
