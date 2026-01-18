package vmodel

import (
	"assistant-go/internal/layer/entity"
)

type NoteShare struct {
	ID     int    `json:"id"`
	NoteID int    `json:"note_id"`
	Hash   string `json:"hash"`
}

func NoteShareFromEntity(entity *entity.NoteShare) *NoteShare {
	return &NoteShare{
		ID:     entity.ID,
		NoteID: entity.NoteID,
		Hash:   entity.Hash,
	}
}
