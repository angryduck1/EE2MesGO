package server

import (
	"context"
	"time"
)

func (server *Server) insertNewUser(ctx context.Context, name, password string) (string, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	hash := hashPassword(password)
	
	deviceToken, err := generateDeviceToken()
	if err != nil {
		return "", err
	}

	newUser := UserInfo{
		Name:         name,
		PasswordHash: hash,
		DeviceToken:  deviceToken,
	}

	e := server.DB.WithContext(dbCtx).Create(&newUser).Error

	if e != nil {
		return "", e
	}

	return deviceToken, nil
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

func (server *Server) getUserByToken(ctx context.Context, deviceToken string) (*UserInfo, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	var userInfo UserInfo

	err := server.DB.WithContext(dbCtx).First(&userInfo, "device_token = ?", deviceToken).Error
	if err != nil {
		return nil, err
	}

	return &userInfo, nil
}
