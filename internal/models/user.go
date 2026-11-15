package models

import "time"

type User struct {
	ID                  int        `db:"id"`
	Name                string     `db:"name"`
	Slug                string     `db:"slug"`
	ImageURL            string     `db:"image_url"`
	AvatarURL           string     `db:"avatar_url"`
	PasswordHash        *string    `db:"password_hash"`
	IsAdmin             bool       `db:"is_admin"`
	Email               *string    `db:"email"`
	ResetToken          *string    `db:"reset_token"`
	ResetTokenExpiresAt *time.Time `db:"reset_token_expires_at"`
}
