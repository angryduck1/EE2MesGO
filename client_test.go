package main

import (
	"EE2MesGO/server"
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
		log.Fatalf("Fail open bd: %v", err)
	}

	err = db.AutoMigrate(&server.UserInfo{}, &server.DeviceInfo{})

	if err != nil {
		log.Fatalf("Fail migrate bd: %v", err)
	}

	srv := &server.Server{DB: db}

	fmt.Println("Successful connect to database!")

	reqBody := map[string]string{"name": "test", "password": "root"}

	var reqBodyBuf bytes.Buffer

	err = json.NewEncoder(&reqBodyBuf).Encode(reqBody)

	if err != nil {
		log.Fatalf("Fail create NewEncoder: %v", err)
	}

	r := httptest.NewRequest("POST", "/login", &reqBodyBuf)

	rr := httptest.NewRecorder()

	srv.Registration(rr, r)

	var respBody interface{}
	if err := json.NewDecoder(rr.Body).Decode(&respBody); err != nil {
		log.Fatalf("Fail decode response server: %v", err)
	}

	fmt.Println(respBody)
}
