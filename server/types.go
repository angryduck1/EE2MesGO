package server

import "gorm.io/gorm"

type UserInfo struct {
	gorm.Model

	Name         string
	PasswordHash string
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
