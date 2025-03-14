package service

import (
	"assistant-go/internal/layer/repository"
	"context"
)

type NoteCategory interface {
	PositionService(ctx context.Context, repositories *repository.Repositories) PositionService
}

type noteCategory struct{}

func NewNoteCategory() NoteCategory {
	return &noteCategory{}
}

func (ps *noteCategory) PositionService(ctx context.Context, repositories *repository.Repositories) PositionService {
	return &positionService{
		ctx:          ctx,
		repositories: repositories,
	}
}
