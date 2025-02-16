package ucase

import (
	"assistant-go/internal/layer/dto"
	"assistant-go/internal/layer/entity"
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/locale"
	"assistant-go/internal/logging"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserUseCase interface {
	Create(in dto.UserLoginAndPassword, lang string) (*entity.User, error)
	Login(in dto.UserLoginAndPassword, lang string) (*entity.UserToken, error)
	RefreshToken(in dto.UserRefreshToken, lang string) (*entity.UserToken, error)
}

type userUseCase struct {
	ctx            context.Context
	userRepository repository.UserRepository
}

func NewUserUseCase(ctx context.Context, userRepository repository.UserRepository) UserUseCase {
	return &userUseCase{
		ctx:            ctx,
		userRepository: userRepository,
	}
}

func (uc *userUseCase) Create(in dto.UserLoginAndPassword, lang string) (*entity.User, error) {
	existingUser, err := uc.userRepository.Find(in.Login)
	if err == nil && existingUser != nil {
		return nil, errors.New(locale.T(lang, "user_already_exists"))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), 11)
	if err != nil {
		return nil, err
	}

	userEntity := entity.User{
		Login:     in.Login,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := uc.userRepository.Create(userEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}

func (uc *userUseCase) Login(in dto.UserLoginAndPassword, lang string) (*entity.UserToken, error) {
	existingUser, err := uc.userRepository.Find(in.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(locale.T(lang, "incorrect_username_or_password"))
		}
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(in.Password))
	if err != nil {
		return nil, errors.New(locale.T(lang, "incorrect_username_or_password"))
	}

	userTokenEntity, err := uc.generateTokenPair(int(existingUser.ID))
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_error"))
	}

	data, err := uc.userRepository.SetUserToken(*userTokenEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}

func (uc *userUseCase) RefreshToken(in dto.UserRefreshToken, lang string) (*entity.UserToken, error) {
	existingToken, err := uc.userRepository.FindUserToken(in.Token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New(locale.T(lang, "refresh_token_not_found"))
		}
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	if existingToken.RefreshToken != in.RefreshToken {
		return nil, errors.New(locale.T(lang, "refresh_token_not_found"))
	}

	userTokenEntity, err := uc.generateTokenPair(int(existingToken.UserId))
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_error"))
	}
	data, err := uc.userRepository.SetUserToken(*userTokenEntity)
	if err != nil {
		logging.GetLogger(uc.ctx).Error(err)
		return nil, errors.New(locale.T(lang, "unexpected_database_error"))
	}
	return data, nil
}

func (uc *userUseCase) generateTokenPair(userId int) (*entity.UserToken, error) {
	var userTokenEntity *entity.UserToken
	for {
		token, err := generateAPIToken()
		if err != nil {
			return nil, err
		}
		refreshToken, err := generateAPIToken()
		if err != nil {
			return nil, err
		}

		_, err = uc.userRepository.FindUserToken(token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				userTokenEntity = &entity.UserToken{
					UserId:       userId,
					Token:        token,
					RefreshToken: refreshToken,
					ExpiredTo:    int(time.Now().Add(4 * time.Hour).Unix()),
				}
				break
			}
		}
	}
	return userTokenEntity, nil
}

func generateAPIToken() (string, error) {
	bytes := make([]byte, 48)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
