package repository

import (
	"assistant-go/internal/layer/entity"
	"context"
	"errors"
	"time"

	//"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	//"time"
)

type UserRepository interface {
	Create(in entity.User) (*entity.User, error)
}

type userRepository struct {
	ctx context.Context
	db  *pgxpool.Pool
}

func NewUserRepository(ctx context.Context, db *pgxpool.Pool) UserRepository {
	return &userRepository{
		ctx: ctx,
		db:  db,
	}
}

func (ur *userRepository) Create(in entity.User) (*entity.User, error) {
	query := `INSERT INTO users (login, password, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id`

	// Устанавливаем текущие временные метки
	in.CreatedAt = time.Now()
	in.UpdatedAt = in.CreatedAt

	row := ur.db.QueryRow(ur.ctx, query, in.Login, in.Password, in.CreatedAt, in.UpdatedAt)

	if err := row.Scan(&in.ID); err != nil {
		return nil, errors.New("failed to insert user: " + err.Error())
	}

	return &in, nil
}
