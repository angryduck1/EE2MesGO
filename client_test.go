package main

import (
	"FirstBackend/server"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http/httptest"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRegistration(t *testing.T) {
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

	regBody := map[string]string{"name": "test", "password": "root"}

	registerBody, err := json.Marshal(regBody)

	if err != nil {
		t.Fatalf("%v", err.Error())
	}

	r := httptest.NewRequest("POST", "/register", bytes.NewBuffer(registerBody))

	rr := httptest.NewRecorder()

	body := rr.Body.Bytes()

	fmt.Println(body)

	srv.Registration(rr, r)
}
