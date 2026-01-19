package ucase

import (
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/logging"
	"assistant-go/internal/storage/postgres"
	"assistant-go/pkg/utils"
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNoteShareExists   = errors.New("note share already exists")
	ErrNoteShareNotFound = errors.New("note share not found")
)

type NoteShareUseCase interface {
	Create(ctx context.Context, noteID int, userEntity *entity.User) (*entity.NoteShare, error)
	GetOne(ctx context.Context, noteID int, userEntity *entity.User) (*entity.NoteShare, error)
	Delete(ctx context.Context, noteID int, userEntity *entity.User) error
}

type noteShareUseCase struct {
	repositories repository.Repositories
}

func NewNoteShareUseCase(repositories *repository.Repositories) NoteShareUseCase {
	return &noteShareUseCase{
		repositories: *repositories,
	}
}

func (uc *noteShareUseCase) Create(ctx context.Context, noteID int, userEntity *entity.User) (*entity.NoteShare, error) {
	noteBelongsUser, err := uc.repositories.NoteRepository.BelongsToUser(ctx, noteID, userEntity.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	if !noteBelongsUser {
		return nil, ErrNoteNotFound
	}

	existsByNote, err := uc.repositories.NoteShareHashesRepository.ExistsByNoteID(ctx, noteID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	if existsByNote {
		return nil, ErrNoteShareExists
	}

	stringUtils := utils.NewStringUtils()
	var hash string
	for i := 1; i < 10; i++ {
		h, err := stringUtils.GenerateRandomString(80)
		if err != nil {
			logging.GetLogger(ctx).Error(err)
			return nil, err
		}
		existsByHash, err := uc.repositories.NoteShareHashesRepository.ExistsByHash(ctx, h)
		if err != nil {
			logging.GetLogger(ctx).Error(err)
			return nil, err
		}
		if !existsByHash {
			hash = h
			break
		}
	}

	noteShare := entity.NoteShare{
		NoteID: noteID,
		Hash:   hash,
	}

	data, err := uc.repositories.NoteShareHashesRepository.Create(ctx, noteShare)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	return data, nil
}

func (uc *noteShareUseCase) GetOne(ctx context.Context, noteID int, userEntity *entity.User) (*entity.NoteShare, error) {
	noteBelongsUser, err := uc.repositories.NoteRepository.BelongsToUser(ctx, noteID, userEntity.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}

	if !noteBelongsUser {
		return nil, ErrNoteNotFound
	}

	noteShare, err := uc.repositories.NoteShareHashesRepository.GetByNoteID(ctx, noteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoteShareNotFound
		}
		logging.GetLogger(ctx).Error(err)
		return nil, postgres.ErrUnexpectedDBError
	}
	return noteShare, nil
}

func (uc *noteShareUseCase) Delete(ctx context.Context, noteID int, userEntity *entity.User) error {
	noteBelongsUser, err := uc.repositories.NoteRepository.BelongsToUser(ctx, noteID, userEntity.ID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}

	if !noteBelongsUser {
		return ErrNoteNotFound
	}

	err = uc.repositories.NoteShareHashesRepository.DeleteByNoteID(ctx, noteID)
	if err != nil {
		logging.GetLogger(ctx).Error(err)
		return postgres.ErrUnexpectedDBError
	}
	return nil
}
