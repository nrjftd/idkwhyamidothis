package repo

import (
	"context"
	"errors"
	"fmt"
	"jwt2/models"

	"github.com/uptrace/bun"
)

type UserRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
	return &UserRepo{db: db}
}
func (repo *UserRepo) CreateUser(ctx context.Context, user *models.User) error {
	_, err := repo.db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to insert user into database: %w", err)
	}
	return err
}

// ================================================================================================================//
func (repo *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := repo.db.NewSelect().Model(&user).Where("email =?", email).Scan(ctx)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

// ================================================================================================================//
func (repo *UserRepo) UpdateUser(ctx context.Context, user *models.User) error {
	_, err := repo.db.NewUpdate().Model(user).Where("id =?", user.ID).Exec(ctx)
	return err
}
func (repo *UserRepo) DeleteUser(ctx context.Context, UserId int64) error {
	_, err := repo.db.NewDelete().Model((*models.User)(nil)).Where("id =?", UserId).Exec(ctx)
	return err
}
func (repo *UserRepo) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := repo.db.NewSelect().Model(&user).Where("id =?", id).Scan(ctx)
	return &user, err
}

func (repo *UserRepo) GetAllUser(ctx context.Context) ([]models.User, error) {
	var user []models.User
	err := repo.db.NewSelect().Model(&user).Scan(ctx)
	return user, err
}
