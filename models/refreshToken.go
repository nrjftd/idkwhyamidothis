package models

import (
	"time"

	"github.com/uptrace/bun"
)

type RefreshToken struct {
	bun.BaseModel `bun:"table:refresh_tokens"`

	ID        int       `bun:"id,pk,autoincrement"`
	UserID    string    `bun:"user_id,notnull"`
	Token     string    `bun:"token,unique,notnull"`
	ExpireAt  time.Time `bun:"expires_at,notnull"`
	CreatedAt time.Time `bun:"created_at,default:current_timestamp"`
}
