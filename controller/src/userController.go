package controller

import (
	"context"
	"fmt"
	"jwt2/models"
	service "jwt2/service/src"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// type RefreshTokenInterface interface {
// 	InsertToken(ctx context.Context, token models.RefreshToken) (*models.RefreshToken, error)
// 	GetToken(ctx context.Context, token string) (*models.RefreshToken, error)
// 	DeleteToken(ctx context.Context, token string) error
// }

type UserController struct {
	service             *service.UserService
	refreshTokenService service.RefreshTokenService
}

func NewUserController(service *service.UserService, refreshTokenService service.RefreshTokenService) *UserController {
	return &UserController{service: service,
		refreshTokenService: refreshTokenService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user by providing user details
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.User true "User registration data"
// @Success 201 {object} map[string]string "Response when registration is successful"
// @Failure 400 {object} map[string]string "Error response for invalid input"
// @Failure 500 {object} map[string]string "Error response for internal server error"
// @Router /register [post]
func (control *UserController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Printf("Invalid input: %v", err)
		return
	}
	err := control.service.RegisterUser(context.Background(), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to register user",
		})
		log.Printf("error while registering user :%v", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
	})
}

// Login godoc
// @Summary Login a user
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body map[string]string true "Login credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
func (control *UserController) Login(c *gin.Context) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	refreshToken, token, err := control.service.Login(context.Background(), creds.Email, creds.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{

		"refresh token": refreshToken,
		"token":         token,
	})
}

func (control *UserController) Logout(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
	}
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}
	fmt.Println("Received token for logout:", req.RefreshToken)
	reToken, err := control.refreshTokenService.GetToken(c, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "refresh token not found",
		})
		return
	}
	fmt.Println(reToken)
	reToken.UserID = userId.(string)
	err = control.refreshTokenService.DeleteToken(context.Background(), reToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete refresh token",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Logged out successfully ",
	})

}

// GetUserByID godoc
// @Summary GetUserID
// @Description search user by id
// @Tags user management
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User "User details"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 404 {object} map[string]string "User not found"
// @Security BearerAuth
// @Router /admin/users/{id} [get]
func (control *UserController) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}
	user, err := control.service.GetUserByID(context.Background(), userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary UpdateUser
// @Description update user detail by ID
// @Tags user management
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @param user body models.User true "User data to update"
// @Success 200 {object} models.User "User updated successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "User not found"
// @Security BearerAuth
// @Router /users/profile/udpate/{id} [put]
func (control *UserController) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = control.service.UpdateUser(context.Background(), userId, &user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
	})
}

// GetAllUser godoc
// @Summary Get all users
// @Description Fetch a list of all users (Admin access required)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {array} models.User "List of users"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Security BearerAuth
// @Router /admin/users [get]
func (control *UserController) GetAllUser(c *gin.Context) {
	users, err := control.service.GetAllUser(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user",
		})
		return
	}
	c.JSON(http.StatusOK, users)
}

// DeleteUserByID godoc
// @Summary Delete user
// @Description Delete a user (Admin access required)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "User deleted successfully"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 404 {object} map[string]string "User not found"
// @Security BearerAuth
// @Router /admin/user/delete/{id} [delete]

func (control *UserController) DeleteUserById(c *gin.Context) {
	id := c.Param("id")
	userId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user ID",
		})
		return
	}
	err = control.service.DeleteUser(context.Background(), userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
