package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	typeDirectory = 0
	typeFile      = 1
)

var (
	ErrDriveDirectoryExists = errors.New("directory exists")
)

type DriveUseCase interface {
	CreateDirectory(dto *dto.DriveCreateDirectory, user *entity.User) error
}

type driveUseCase struct {
	ctx          context.Context
	repositories *repository.Repositories
}

func NewDriveUseCase(ctx context.Context, repositories *repository.Repositories) DriveUseCase {
	return &driveUseCase{
		ctx:          ctx,
		repositories: repositories,
	}
}

func (uc *driveUseCase) CreateDirectory(dto *dto.DriveCreateDirectory, user *entity.User) error {
	// нужно проверить существует ли parent_id в принципе

	_, err := uc.repositories.DriveStructRepository.FindRow(user.ID, dto.Name, typeDirectory, dto.ParentId)
	fmt.Println(err)
	if err == nil {
		return ErrDriveDirectoryExists
	}
	createEntity := &entity.DriveStruct{
		UserID:    user.ID,
		Name:      dto.Name,
		Type:      typeDirectory,
		ParentID:  dto.ParentId,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err = uc.repositories.DriveStructRepository.CreateDirectory(createEntity)
	if err != nil {
		return err
	}

	return errors.New("директории создана, все ок. stop")
}
