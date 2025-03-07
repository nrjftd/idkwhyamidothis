package controller

import (
	"context"
	"jwt2/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RefreshServiceInterface interface {
	InsertToken(ctx context.Context, token *models.RefreshToken) error
	GetToken(ctx context.Context, token string) (*models.RefreshToken, error)
	DeleteToken(ctx context.Context, token *models.RefreshToken) error
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
}

type RefreshTokenController struct {
	service RefreshServiceInterface
}

func NewRefreshTokenController(service RefreshServiceInterface) *RefreshTokenController {
	return &RefreshTokenController{service: service}
}

func (s *RefreshTokenController) InsertToken(c *gin.Context) {

}
func (s *RefreshTokenController) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}
	newAccessToken, err := s.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
	})
}
