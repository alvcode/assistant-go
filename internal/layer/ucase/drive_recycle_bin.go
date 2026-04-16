package ucase

import (
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"context"
)

type DriveRecycleBinUseCase interface {
	GetAll(ctx context.Context, user *entity.User) error
}

type driveRecycleBin struct {
	repositories *repository.Repositories
}

func NewDriveRecycleBinUseCase(repositories *repository.Repositories) DriveRecycleBinUseCase {
	return &driveRecycleBin{}
}

func (uc *driveRecycleBin) GetAll(ctx context.Context, user *entity.User) error {
	return nil
}
