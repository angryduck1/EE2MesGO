package main

import (
	"EE2MesGO/server"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func RegisterTest(ts *httptest.Server) string {
	var bodyReq = map[string]interface{}{"name": "meow", "password": "12345"}
	var bodyBuf bytes.Buffer

	err := json.NewEncoder(&bodyBuf).Encode(bodyReq)

	if err != nil {
		log.Fatal(err)
	}

	r, err := http.Post(ts.URL+"/login", "application/json", &bodyBuf)

	if err != nil {
		log.Fatal(err)
	}

	var bodyResp interface{}

	err = json.NewDecoder(r.Body).Decode(&bodyResp)

	defer r.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	if bodyRespMap, ok := bodyResp.(map[string]interface{}); ok {
		fmt.Println(bodyRespMap["message"])
		fmt.Println(bodyRespMap["deviceToken"])

		if r.StatusCode != http.StatusOK {
			return ""
		}

		return bodyRespMap["deviceToken"].(string)
	}

	return ""
}

func ValidToken(ts *httptest.Server, deviceToken string) {
	var bodyReq = map[string]interface{}{"deviceToken": deviceToken}
	var bodyBuf bytes.Buffer

	err := json.NewEncoder(&bodyBuf).Encode(bodyReq)

	if err != nil {
		log.Fatal(err)
	}
}

func TestWebSocket(t *testing.T) {
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

	mux := http.NewServeMux()

	mux.HandleFunc("/login", srv.Login)
	mux.HandleFunc("/validate-token", srv.ValidateDeviceToken)

	ts := httptest.NewServer(mux)

	defer ts.Close()

	deviceToken := RegisterTest(ts)

	if deviceToken != "" {
		ValidToken(ts, deviceToken)
	}
}
