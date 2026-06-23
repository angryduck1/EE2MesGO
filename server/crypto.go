package server

import (
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

func hashPassword(password string) string {
	key, err := argon2id.CreateHash(password, argon2id.DefaultParams)

	if err != nil {
		fmt.Errorf("hashPassword failed: %v", err)
	}

	return key
}

func matchPassword(password string, passwordHash string) (bool, error) {
	r, err := argon2id.ComparePasswordAndHash(password, passwordHash)

	if err != nil {
		fmt.Errorf("comparePassowrd failed: %v", err)
	}

	return r, err
}

func generateDeviceToken() (string, error) {
	token, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate device token: %v", err)
	}

	return token.String(), nil
}

func generateDeviceToken() (string, error) {
	token, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate device token: %v", err)
	}

	return token.String(), nil
}
