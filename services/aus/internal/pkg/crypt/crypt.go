package crypt

import "golang.org/x/crypto/bcrypt"

// Шифрует пароль пользователя
func CryptPassword(password string) ([]byte, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return []byte{}, err
	}
	return passwordHash, nil
}
