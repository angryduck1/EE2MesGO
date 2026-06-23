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

	srv.Registration(rr, r)
	
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	
	if deviceToken, ok := response["deviceToken"]; ok {
		fmt.Printf("generated device token: %v\n", deviceToken)
	} else {
		t.Error("device token not found in response")
	}
}
