package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Login     string    `json:"login"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterUserDTO struct {
	Login    string `json:"login" validate:"required,alphanum,min=6,max=32"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=32"`
}

type LoginUserDTO struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}
