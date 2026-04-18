package ucase

import (
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"context"
)

type DriveRecycleBinUseCase interface {
	GetAll(ctx context.Context, user *entity.User) ([]*entity.DriveRecycleBinStruct, error)
}

type driveRecycleBinUseCase struct {
	repositories *repository.Repositories
}

func NewDriveRecycleBinUseCase(repositories *repository.Repositories) DriveRecycleBinUseCase {
	return &driveRecycleBinUseCase{
		repositories: repositories,
	}
}

func (uc *driveRecycleBinUseCase) GetAll(ctx context.Context, user *entity.User) ([]*entity.DriveRecycleBinStruct, error) {
	result, err := uc.repositories.DriveRecycleBinRepository.GetAll(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return result, nil
}
