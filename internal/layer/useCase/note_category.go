package useCase

import (
	dtoNoteCategory "assistant-go/internal/layer/dto/noteCategory"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"context"
)

type NoteCategoryUseCase interface {
	Create(in dtoNoteCategory.Create, lang string) (*entity.NoteCategory, error)
}

type noteCategoryUseCase struct {
	ctx                    context.Context
	noteCategoryRepository repository.NoteCategoryRepository
}

func NewNoteCategoryUseCase(ctx context.Context, noteCategoryRepository repository.NoteCategoryRepository) NoteCategoryUseCase {
	return &noteCategoryUseCase{
		ctx:                    ctx,
		noteCategoryRepository: noteCategoryRepository,
	}
}

func (uc *noteCategoryUseCase) Create(in dtoNoteCategory.Create, lang string) (*entity.NoteCategory, error) {
	return nil, nil
}
