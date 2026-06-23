package main

import (
	"fmt"
	"log"
	"net/http"

	server "EE2MesGO/server"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=root dbname=messenger port=5432 sslmode=disable TimeZone=Europe/Moscow"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Fail open bd: %v", err.Error())
	}

	err = db.AutoMigrate(&server.UserInfo{})

	if err != nil {
		log.Fatalf("Fail migrate bd: %v", err.Error())
	}

	srv := &server.Server{DB: db}

	fmt.Println("Successful connect to database!")

	http.HandleFunc("/registration", srv.Registration)
	http.HandleFunc("/login", srv.Login)
	http.HandleFunc("/validate-token", srv.ValidateDeviceToken)

	e := http.ListenAndServeTLS("localhost:8080", "./keys/localhost+1.pem", "./keys/localhost+1-key.pem", nil)

	if e != nil {
		log.Fatalf("Fatal error running server: %v", e.Error())
	}

}
