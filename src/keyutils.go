package main

import (
	"crypto/pbkdf2"
	"crypto/sha256"
	"fmt"
)

func DeriveKey(password string, salt []byte, keyLen int) (key []byte, err error) {
	key, err = pbkdf2.Key(sha256.New, password, salt, 100000, keyLen)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при генерации ключа: %w", err)
	}
	return key, nil
}
