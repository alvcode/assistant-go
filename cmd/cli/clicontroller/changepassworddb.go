package clicontroller

import (
	"assistant-go/internal/layer/repository"
	"assistant-go/internal/logging"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func ChangePasswordDBInit(ctx context.Context, db *pgxpool.Pool, login string, password string) {
	fmt.Println("start reset-password cli command")
	logging.GetLogger(ctx).Println("start clean-db cli command")

	userRepository := repository.NewUserRepository(ctx, db)

	user, err := userRepository.Find(login)
	if err != nil {
		fmt.Printf("Error find user: %s", err)
		logging.GetLogger(ctx).Errorf("Error find user: %s", err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		fmt.Printf("Error generate hash password: %s", err)
		logging.GetLogger(ctx).Errorf("Error generate hash password: %s", err)
		return
	}

	err = userRepository.ChangePassword(user.ID, string(hashedPassword))
	if err != nil {
		fmt.Printf("Error change password: %s", err)
		logging.GetLogger(ctx).Errorf("Error change password: %s", err)
		return
	}

	err = userRepository.DeleteUserTokensByID(user.ID)
	if err != nil {
		fmt.Printf("Error delete user token: %s", err)
		logging.GetLogger(ctx).Errorf("Error delete user token: %s", err)
		return
	}

	db.Close()
	fmt.Println("successfully")
	logging.GetLogger(ctx).Println("successfully")
}
