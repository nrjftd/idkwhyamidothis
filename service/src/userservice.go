package service

import (
	"context"
	"errors"
	"fmt"
	"jwt2/models"
	repo "jwt2/repo/src"
	"jwt2/utils"
	"log"
	"time"

	"github.com/google/uuid"
)

type RefreshTokenInterface interface {
	InsertToken(ctx context.Context, token *models.RefreshToken) error
	GetToken(ctx context.Context, token string) (*models.RefreshToken, error)
	DeleteToken(ctx context.Context, token *models.RefreshToken) error
	RefreshToken(ctx context.Context, token string) (string, error)
	GetTokenByUserID(ctx context.Context, userId string) (*models.RefreshToken, error)
}
type UserService struct {
	repo                *repo.UserRepo
	refreshTokenService RefreshTokenInterface
}

func NewService(repo *repo.UserRepo, refreshRepo RefreshTokenInterface) *UserService {
	return &UserService{repo: repo,
		refreshTokenService: refreshRepo,
	}
}

func (service *UserService) RegisterUser(ctx context.Context, user *models.User) error {
	validUserTypes := map[string]bool{"ADMIN": true, "USER": true}
	log.Printf("user_type received: %s", user.User_type)

	if !validUserTypes[user.User_type] {
		return fmt.Errorf("invalid user_type: %s", user.User_type)
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}

	//==========================//

	//===//
	userID := uuid.New().String()
	user.User_id = userID
	user.Password = hashedPassword
	return service.repo.CreateUser(ctx, user)

}
func (service *UserService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := service.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid email ")
	}
	if !utils.CheckPasswordHash(password, user.Password) {
		return "", "", errors.New("invalid password")
	}
	oldToken, err := service.refreshTokenService.GetTokenByUserID(ctx, user.User_id)
	if err == nil && oldToken != nil {
		err := service.refreshTokenService.DeleteToken(ctx, oldToken)

		if err != nil {
			return "", "", errors.New("Failed to delete old token: " + err.Error())
		}
	}
	refreshToken, err := utils.GenerateRefreshToken(user.User_id, user.Email, user.User_type)
	if err != nil {
		return "", "", errors.New(err.Error())
	}
	token, err := utils.GenerateJWT(user.User_id, user.Email, user.User_type)
	if err != nil {
		return "", "", errors.New(err.Error())
	}

	tokenModel := &models.RefreshToken{
		UserID:   user.User_id,
		Token:    refreshToken,
		ExpireAt: time.Now().Add(time.Hour * 24),
	}
	log.Println("Refresh Token Model:", tokenModel)

	err = service.refreshTokenService.InsertToken(ctx, tokenModel)
	if err != nil {
		return "", "", errors.New("Failed to store refresh token: " + err.Error())
	}
	return refreshToken, token, nil
}

func (service *UserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return service.repo.GetUserByID(ctx, id)
}
func (service *UserService) UpdateUser(ctx context.Context, ID int64, user *models.User) error {
	userExist, err := service.repo.GetUserByID(ctx, ID)
	if err != nil {
		return errors.New("user not found")
	}
	if user.Password != "" {
		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	} else {
		user.Password = userExist.Password

	}
	return service.repo.UpdateUser(ctx, user)

}
func (service *UserService) DeleteUser(ctx context.Context, id int64) error {
	_, err := service.repo.GetUserByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}
	return service.repo.DeleteUser(ctx, id)
}

func (service *UserService) GetAllUser(ctx context.Context) ([]models.User, error) {
	return service.repo.GetAllUser(ctx)
}
