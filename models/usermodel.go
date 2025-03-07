package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`
	ID            int64     `bun:",pk,autoincrement" json:"id"`
	First_name    string    `bun:"first_name,notnull" json:"first_name" validate:"required, min=2, max=100"`
	Last_name     string    `bun:"last_name,notnull" json:"last_name" validate:"required, min=2, max=100"`
	Password      string    `bun:"password,notnull" json:"password" validate:"required, min=6"`
	Email         string    `bun:"email,unique,notnull" json:"email" validate:"email, required"`
	Phone         string    `bun:"phone,notnull" json:"phone" validate:"required"`
	Token         *string   `bun:"token" json:"token,omitempty"`
	User_type     string    `bun:"user_type,notnull" json:"user_type" validate:"required, oneof=ADMIN USER"`
	Refresh_token *string   `bun:"refresh_token" json:"refresh_token,omitempty"`
	Created_at    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	Updated_at    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	User_id       string    `bun:"user_id,unique,notnull" json:"user_id"`
}
