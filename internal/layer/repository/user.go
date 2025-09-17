package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, in entity.User) (*entity.User, error)
	Find(ctx context.Context, login string) (*entity.User, error)
	FindById(ctx context.Context, id int) (*entity.User, error)
	FindUserToken(ctx context.Context, token string) (*entity.UserToken, error)
	SetUserToken(ctx context.Context, in entity.UserToken) (*entity.UserToken, error)
	Delete(ctx context.Context, userID int) error
	DeleteUserTokensByID(ctx context.Context, userID int) error
	ChangePassword(ctx context.Context, userID int, newPassword string) error
	RemoveTokensByDateExpired(ctx context.Context, time int) error
}

type userRepository struct {
	db DBExecutor
}

func NewUserRepository(db DBExecutor) UserRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) Create(ctx context.Context, in entity.User) (*entity.User, error) {
	query := `INSERT INTO users (login, password, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id`

	row := ur.db.QueryRow(ctx, query, in.Login, in.Password, in.CreatedAt, in.UpdatedAt)

	if err := row.Scan(&in.ID); err != nil {
		return nil, err
	}

	return &in, nil
}

func (ur *userRepository) Find(ctx context.Context, login string) (*entity.User, error) {
	query := `SELECT id, login, password, created_at, updated_at FROM users WHERE login = $1`

	row := ur.db.QueryRow(ctx, query, login)

	var user entity.User
	if err := row.Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) FindById(ctx context.Context, id int) (*entity.User, error) {
	query := `SELECT id, login, password, created_at, updated_at FROM users WHERE id = $1`

	row := ur.db.QueryRow(ctx, query, id)

	var user entity.User
	if err := row.Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) FindUserToken(ctx context.Context, token string) (*entity.UserToken, error) {
	query := `SELECT user_id, token, refresh_token, expired_to FROM user_tokens WHERE token = $1`

	row := ur.db.QueryRow(ctx, query, token)

	var userToken entity.UserToken
	if err := row.Scan(&userToken.UserId, &userToken.Token, &userToken.RefreshToken, &userToken.ExpiredTo); err != nil {
		return nil, err
	}
	return &userToken, nil
}

func (ur *userRepository) SetUserToken(ctx context.Context, in entity.UserToken) (*entity.UserToken, error) {
	query := `INSERT INTO user_tokens (user_id, token, refresh_token, expired_to) VALUES ($1, $2, $3, $4)`

	_, err := ur.db.Exec(ctx, query, in.UserId, in.Token, in.RefreshToken, in.ExpiredTo)
	return &in, err
}

func (ur *userRepository) Delete(ctx context.Context, userID int) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := ur.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *userRepository) DeleteUserTokensByID(ctx context.Context, userID int) error {
	query := `DELETE FROM user_tokens WHERE user_id = $1`

	_, err := ur.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *userRepository) ChangePassword(ctx context.Context, userID int, newPassword string) error {
	query := `UPDATE users SET password = $2 WHERE id = $1`

	_, err := ur.db.Exec(ctx, query, userID, newPassword)
	if err != nil {
		return err
	}
	return nil
}

func (ur *userRepository) RemoveTokensByDateExpired(ctx context.Context, time int) error {
	query := `DELETE FROM user_tokens WHERE expired_to < $1`

	_, err := ur.db.Exec(ctx, query, time)
	if err != nil {
		return err
	}
	return nil
}
