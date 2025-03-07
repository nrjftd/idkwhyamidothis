package repo

import (
	"context"
	"jwt2/models"

	"github.com/uptrace/bun"
)

type RefreshTokenRepository interface {
	InsertToken(ctx context.Context, refreshToken *models.RefreshToken) error
	GetToken(ctx context.Context, refreshToken string) (*models.RefreshToken, error)
	//GetTokenByEmail(ctx context.Context, email string) (*models.RefreshToken, error)
	DeleteToken(ctx context.Context, token string) error
	GetTokenByUserID(ctx context.Context, userID string) (*models.RefreshToken, error)
}
type RefreshTokenRepo struct {
	db *bun.DB
}

func NewRefreshToken(db *bun.DB) *RefreshTokenRepo {

	return &RefreshTokenRepo{db: db}
}

func (repo *RefreshTokenRepo) InsertToken(ctx context.Context, token *models.RefreshToken) error {
	_, err := repo.db.NewInsert().Model(token).Exec(ctx)
	return err
}
func (repo *RefreshTokenRepo) GetToken(ctx context.Context, refreshToken string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := repo.db.NewSelect().Model(&token).Where("token =?", refreshToken).Scan(ctx)
	return &token, err
}

//	func (repo *RefreshTokenRepo) GetTokenByEmail(ctx context.Context, email string) (*models.RefreshToken, error) {
//		var token models.RefreshToken
//		err := repo.db.NewSelect().Model(&token).Where("email =?", email).Scan(ctx)
//		return &token, err
//	}
func (repo *RefreshTokenRepo) GetTokenByUserID(ctx context.Context, userID string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := repo.db.NewSelect().Model(&token).Where("user_id=?", userID).Scan(ctx)
	return &token, err
}

func (repo *RefreshTokenRepo) DeleteToken(ctx context.Context, token string) error {
	_, err := repo.db.NewDelete().Model((*models.RefreshToken)(nil)).Where("token=?", token).Exec(ctx)
	return err
}
