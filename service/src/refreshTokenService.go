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

	"github.com/redis/go-redis/v9"
)

type refreshTokenService struct {
	repo        *repo.RefreshTokenRepo
	RedisClient *redis.Client
}
type RefreshTokenService interface {
	InsertToken(ctx context.Context, token *models.RefreshToken) error
	GetToken(ctx context.Context, token string) (*models.RefreshToken, error)
	DeleteToken(ctx context.Context, token *models.RefreshToken) error
	RefreshToken(ctx context.Context, token string) (string, error)
	GetTokenByUserID(ctx context.Context, userId string) (*models.RefreshToken, error)
}

func NewRefreshService(repo *repo.RefreshTokenRepo, RedisClient *redis.Client) RefreshTokenService {
	return &refreshTokenService{repo: repo, RedisClient: RedisClient}
}

func (service *refreshTokenService) InsertToken(ctx context.Context, token *models.RefreshToken) error {
	err := service.repo.InsertToken(ctx, token)
	if err != nil {
		log.Printf("error inserting refresh token (service): %v", err)
		return err
	}
	//redis
	key := "refresh:" + token.UserID
	err = service.RedisClient.Set(ctx, key, token.Token, time.Until(token.ExpireAt)).Err()
	if err != nil {
		log.Printf("Error saving token to redis: %v", err)
		return err
	}
	log.Printf("Refresh token saved to Redis for user: %s", token.UserID)
	return nil
}
func (service *refreshTokenService) GetToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	userID, err := service.RedisClient.Get(ctx, "refresh:"+token).Result()
	if err != nil {
		return &models.RefreshToken{Token: token, UserID: userID}, nil
	} else if err != redis.Nil {
		return nil, err
	}
	reToken, err := service.repo.GetToken(ctx, token)
	if err != nil {
		return nil, err
	}
	fmt.Println("token from service: " + reToken.Token)

	service.RedisClient.Set(ctx, "refresh:"+reToken.Token, reToken.UserID, time.Until(reToken.ExpireAt))
	return reToken, nil
}
func (service *refreshTokenService) DeleteToken(ctx context.Context, token *models.RefreshToken) error {
	err := service.repo.DeleteToken(ctx, token.Token)
	if err != nil {
		return err
	}
	key := "refresh:" + token.UserID
	err = service.RedisClient.Del(ctx, key).Err()
	if err != nil {
		return errors.New(err.Error())
	}
	log.Printf("Successfully deleted refresh token from Redis: %s", key)

	return nil
}
func (s *refreshTokenService) RefreshToken(ctx context.Context, token string) (string, error) {
	storedToken, err := s.repo.GetToken(ctx, token)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}
	if time.Now().After(storedToken.ExpireAt) {
		_ = s.repo.DeleteToken(ctx, token)
		return "", errors.New("refresh token expired")
	}
	claims, err := utils.ValidateRefreshToken(token)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}
	newAccessToken, err := utils.GenerateJWT(
		claims["user_id"].(string),
		claims["email"].(string),
		claims["user_type"].(string),
	)
	if err != nil {
		return "", errors.New("failed to genrate new token")
	}
	return newAccessToken, nil
}
func (service *refreshTokenService) GetTokenByUserID(ctx context.Context, userId string) (*models.RefreshToken, error) {
	reToken, err := service.repo.GetTokenByUserID(ctx, userId)
	if err != nil {
		return nil, err
	}
	return reToken, nil
}
