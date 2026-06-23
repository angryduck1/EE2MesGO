package server

import "gorm.io/gorm"

type UserInfo struct {
	gorm.Model

	Name         string `gorm:"uniqueIndex"`
	PasswordHash string
	DeviceToken  string `gorm:"uniqueIndex"`
}

type DeviceInfo struct {
	gorm.Model

	DeviceToken string `gorm:"uniqueIndex"`

	UserID uint     `gorm:"index"`
	User   UserInfo `gorm:"foreignKey:UserID"`
}
type RegistrationInfo struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type ResponseInfo struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Server struct {
	DB *gorm.DB
}
type SyncInfo struct {
	Payload []byte

	newChat     chan []byte
	newMessages chan []byte
	Activity    chan []byte
}
