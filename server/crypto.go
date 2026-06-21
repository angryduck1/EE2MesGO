package server

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func hashPassword(password string) string {
	key, err := argon2id.CreateHash(password, argon2id.DefaultParams)

	if err != nil {
		fmt.Errorf("hashPassword failed: %v", err.Error())
	}

	return key
}

func matchPassword(password string, passwordHash string) (bool, error) {
	r, e := argon2id.ComparePasswordAndHash(password, passwordHash)

	if e != nil {
		fmt.Errorf("comparePassowrd failed: %v", e.Error())
	}

	return r, e
}
