package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	service "assistant-go/internal/layer/service/note_category"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

type FileUseCase interface {
	Upload(in dto.NoteCategoryCreate, userEntity *entity.User) (*entity.NoteCategory, error)
}

type fileUseCase struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func NewFileUseCase(ctx context.Context, repositories *repository.Repositories) FileUseCase {
	return &fileUseCase{
		ctx:          ctx,
		repositories: repositories,
	}
}

func (uc *fileUseCase) Upload(in dto.NoteCategoryCreate, userEntity *entity.User) (*entity.NoteCategory, error) {
	noteCategoryEntity := entity.NoteCategory{
		UserId:   userEntity.ID,
		Name:     in.Name,
		ParentId: in.ParentId,
	}

	if in.ParentId != nil {
		_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, *in.ParentId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrCategoryParentIdNotFound
			}
			logging.GetLogger(uc.ctx).Error(err)
			return nil, postgres.ErrUnexpectedDBError
		}
	}

	positionService := service.NewNoteCategory().PositionService(uc.ctx, uc.repositories)
	newPosition, err := positionService.CalculateForNew(userEntity.ID, in.ParentId)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	noteCategoryEntity.Position = newPosition

	data, err := uc.repositories.NoteCategoryRepository.Create(noteCategoryEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return data, nil
}
