package utils

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("This is not secret key so don't care about it")
var refreshSecretKey = []byte("refeshhhhhhhhhh")

func GenerateJWT(userID, email, userType string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"email":     email,
		"user_type": userType,
		"exp":       time.Now().Add(time.Hour * 1).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)

}
func GenerateRefreshToken(userID, email, userType string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userID,
		"email":     email,
		"user_type": userType,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshSecretKey)
}

func ValidateJWT(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil

}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token is required",
			})
			c.Abort()
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")
		claims, err := ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return

		}
		c.Set("claims", claims)
		c.Next()
	}

}

func RefreshTokenHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return refreshSecretKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	email := claims["email"].(string)
	userType := claims["user_type"].(string)
	newAccessToken, err := GenerateJWT(userID, email, userType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to generate new access token",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access token": newAccessToken,
	})
}

func ValidateRefreshToken(refreshToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return refreshSecretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil
}
