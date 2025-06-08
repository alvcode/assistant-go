package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"time"
)

const (
	typeDirectory = 0
	typeFile      = 1
)

var (
	ErrDriveDirectoryExists  = errors.New("directory exists")
	ErrDriveParentIdNotFound = errors.New("drive parent id does not exist")
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
	if dto.ParentID != nil {
		_, err := uc.repositories.DriveStructRepository.FindByID(*dto.ParentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrDriveParentIdNotFound
			}
		}
	}
	_, err := uc.repositories.DriveStructRepository.FindRow(user.ID, dto.Name, typeDirectory, dto.ParentID)
	fmt.Println(err)
	if err == nil {
		return ErrDriveDirectoryExists
	}
	createEntity := &entity.DriveStruct{
		UserID:    user.ID,
		Name:      dto.Name,
		Type:      typeDirectory,
		ParentID:  dto.ParentID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err = uc.repositories.DriveStructRepository.CreateDirectory(createEntity)
	if err != nil {
		return err
	}

	return errors.New("директории создана, все ок. stop")
}
