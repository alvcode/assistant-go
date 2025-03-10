package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"assistant-go/internal/service/utils"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/tidwall/gjson"
	"time"
)

type NoteUseCase interface {
	Create(in dto.NoteCreate, userEntity *entity.User, lang string) (*entity.Note, error)
	GetAll(catIdStruct dto.RequiredID, userEntity *entity.User, lang string) ([]*entity.Note, error)
	Update(in dto.NoteUpdate, userEntity *entity.User, lang string) (*entity.Note, error)
	GetOne(noteIdStruct dto.RequiredID, userEntity *entity.User, lang string) (*entity.Note, error)
	DeleteOne(noteIdStruct dto.RequiredID, userEntity *entity.User, lang string) error
	Pin(noteIdStruct dto.RequiredID, userEntity *entity.User, lang string) error
	UnPin(noteIdStruct dto.RequiredID, userEntity *entity.User, lang string) error
}

type noteUseCase struct {
	ctx          context.Context
	repositories repository.Repositories
}

func NewNoteUseCase(ctx context.Context, repositories *repository.Repositories) NoteUseCase {
	return &noteUseCase{
		ctx:          ctx,
		repositories: *repositories,
	}
}

func (uc *noteUseCase) Create(in dto.NoteCreate, userEntity *entity.User, lang string) (*entity.Note, error) {
	_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, in.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(locale.T(lang, "category_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}

	timeNow := time.Now().UTC()

	var pinned bool
	if in.Pinned == nil {
		pinned = false
	} else {
		pinned = *in.Pinned
	}

	noteEntity := entity.Note{
		CategoryID: in.CategoryID,
		NoteBlocks: in.NoteBlocks,
		CreatedAt:  timeNow,
		UpdatedAt:  timeNow,
		Title:      uc.getNoteTitle(string(in.NoteBlocks)),
		Pinned:     pinned,
	}

	data, err := uc.repositories.NoteRepository.Create(noteEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}

func (uc *noteUseCase) GetAll(catIdStruct dto.RequiredID, userEntity *entity.User, lang string) ([]*entity.Note, error) {
	categories, err := uc.repositories.NoteCategoryRepository.FindByIDAndUserWithChildren(userEntity.ID, catIdStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(locale.T(lang, "category_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}

	catIds := make([]int, 0)
	for _, cat := range categories {
		catIds = append(catIds, cat.ID)
	}
	if len(catIds) == 0 {
		return nil, errors.New(locale.T(lang, "category_not_found"))
	}

	notes, err := uc.repositories.NoteRepository.GetMinimalByCategoryIds(catIds)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logging.GetLogger(uc.ctx).Error(err)
			return nil, errors.New(locale.T(lang, "unexpected_database_error"))
		}
	}
	return notes, nil
}

func (uc *noteUseCase) Update(in dto.NoteUpdate, userEntity *entity.User, lang string) (*entity.Note, error) {
	currentNote, err := uc.repositories.NoteRepository.GetById(in.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(locale.T(lang, "note_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}

	// проверим, что текущая категория заметки принадлежит пользователю
	_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, currentNote.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(locale.T(lang, "note_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}

	// проверим, что новая категория заметки принадлежит пользователю, если она изменяется
	if currentNote.CategoryID != in.CategoryID {
		_, err := uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, in.CategoryID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, errors.New(locale.T(lang, "category_not_found"))
			}
			logging.GetLogger(uc.ctx).Error(err)
			return nil, errors.New(locale.T(lang, "unexpected_database_error"))
		}
	}

	var pinned bool
	if in.Pinned == nil {
		pinned = currentNote.Pinned
	} else {
		pinned = *in.Pinned
	}

	currentNote.NoteBlocks = in.NoteBlocks
	currentNote.CategoryID = in.CategoryID
	currentNote.Title = uc.getNoteTitle(string(in.NoteBlocks))
	currentNote.UpdatedAt = time.Now().UTC()
	currentNote.Pinned = pinned

	err = uc.repositories.NoteRepository.Update(currentNote)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return currentNote, nil
}

func (uc *noteUseCase) GetOne(noteIdStruct dto.RequiredID, userEntity *entity.User, lang string) (*entity.Note, error) {
	currentNote, err := uc.repositories.NoteRepository.GetById(noteIdStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(locale.T(lang, "note_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}

	// проверим, что текущая категория заметки принадлежит пользователю
	_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, currentNote.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(locale.T(lang, "note_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}

	return currentNote, nil
}

func (uc *noteUseCase) DeleteOne(noteIdStruct dto.RequiredID, userEntity *entity.User, lang string) error {
	currentNote, err := uc.repositories.NoteRepository.GetById(noteIdStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(locale.T(lang, "note_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}

	// проверим, что текущая категория заметки принадлежит пользователю
	_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, currentNote.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(locale.T(lang, "note_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}

	err = uc.repositories.NoteRepository.DeleteOne(currentNote.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return nil
}

func (uc *noteUseCase) getNoteTitle(blocks string) *string {
	firstBlockText := gjson.Get(blocks, `0.data.text`)
	stringUtils := utils.NewStringUtils()
	titleWithoutHtml := stringUtils.RemoveHTMLTags(firstBlockText.Str)
	titleTruncate := stringUtils.Trim(stringUtils.TruncateString(titleWithoutHtml, 50))

	if titleTruncate == "" {
		return nil
	} else {
		return &titleTruncate
	}
}

func (uc *noteUseCase) Pin(noteIdStruct dto.RequiredID, userEntity *entity.User, lang string) error {
	currentNote, err := uc.repositories.NoteRepository.GetById(noteIdStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(locale.T(lang, "note_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}

	// проверим, что текущая категория заметки принадлежит пользователю
	_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, currentNote.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(locale.T(lang, "note_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}

	err = uc.repositories.NoteRepository.Pin(currentNote.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return nil
}

func (uc *noteUseCase) UnPin(noteIdStruct dto.RequiredID, userEntity *entity.User, lang string) error {
	currentNote, err := uc.repositories.NoteRepository.GetById(noteIdStruct.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(locale.T(lang, "note_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}

	_, err = uc.repositories.NoteCategoryRepository.FindByIDAndUser(userEntity.ID, currentNote.CategoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New(locale.T(lang, "note_not_found"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}

	err = uc.repositories.NoteRepository.UnPin(currentNote.ID)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return nil
}
