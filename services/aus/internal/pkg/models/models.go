package models

import "github.com/google/uuid"

// Данные о пользователе
type User struct {
	ID       uuid.UUID `json:"id"`
	Login    string    `json:"login"`
	Password string    `json:"password"`
}

type Paper struct {
	Name   string  `json:"name"`
	Price  float32 `json:"price"`
	Amount int32   `json:"amount"`
}

// Данные для аутентификации
type AuthData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Money struct {
	ID   uuid.UUID `json:"id"`
	Cash float32   `json:"cash"`
}
