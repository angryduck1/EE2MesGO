package server

import (
	"context"
	"time"
)

func (server *Server) insertNewUser(ctx context.Context, name, password string) error {
	dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	hash := hashPassword(password)

	newUser := UserInfo{Name: name, PasswordHash: hash}

	e := server.DB.WithContext(dbCtx).Create(&newUser).Error

	return e
}

func (server *Server) getUser(ctx context.Context, name, password string) (bool, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	var userInfo UserInfo

	e := server.DB.WithContext(dbCtx).First(&userInfo, "name = ?", name).Error

	if e != nil {
		return false, e
	}

	passwordHash := userInfo.PasswordHash

	match, err := matchPassword(password, passwordHash)

	return match, err
}
