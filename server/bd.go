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

	newUser := DeviceInfo{DeviceToken: deviceToken, User: UserInfo{
		Name:         name,
		PasswordHash: hash,
	}}

	e := server.DB.WithContext(dbCtx).Create(&newUser).Error

	if e != nil {
		return "", e
	}

	return deviceToken, nil
}

func (server *Server) getUser(ctx context.Context, name, password string) (bool, UserInfo, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	var userInfo UserInfo

	e := server.DB.WithContext(dbCtx).First(&userInfo, "name = ?", name).Error

	if e != nil {
		return false, UserInfo{}, e
	}

	passwordHash := userInfo.PasswordHash

	match, err := matchPassword(password, passwordHash)

	return match, userInfo, err
}

func (server *Server) getUserByToken(ctx context.Context, deviceToken string) (*UserInfo, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	var deviceInfo DeviceInfo

	err := server.DB.WithContext(dbCtx).Preload("User").First(&deviceInfo, "device_token = ?", deviceToken).Error
	if err != nil {
		return nil, err
	}

	return &deviceInfo.User, nil
}

func (server *Server) addNewDevice(ctx context.Context, userInfo UserInfo, deviceToken string) error {
	dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

	defer cancel()

	newDevice := DeviceInfo{DeviceToken: deviceToken, UserID: userInfo.ID}

	err := server.DB.WithContext(dbCtx).Create(&newDevice).Error

	if err != nil {
		return err
	}

	return nil
}
