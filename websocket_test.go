package main

import (
	"EE2MesGO/server"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetNameTest(conn *websocket.Conn) {
	getNameResp := map[string]interface{}{"code": 900, "data": ""}

	getNameRespBuf, err := json.Marshal(getNameResp)

	if err != nil {
		log.Fatal(err)
	}

	err = conn.WriteMessage(websocket.TextMessage, getNameRespBuf)

	if err != nil {
		log.Fatal(err)
	}

	messageType, message, err := conn.ReadMessage()

	if err != nil {
		log.Fatal(err)
	}

	if messageType == websocket.TextMessage {

		var messageJSON interface{}
		err := json.Unmarshal(message, &messageJSON)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(messageJSON)
	}
}

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

	ws := "ws" + strings.TrimPrefix(ts.URL, "http") + "/validate-token?" + "deviceToken=" + deviceToken

	conn, _, err := websocket.DefaultDialer.Dial(ws, nil)

	defer conn.Close()

	if err != nil {
		log.Fatal(err)
	}

	for {
		messageType, message, err := conn.ReadMessage()

		if err != nil {
			log.Fatal(err)
		}

		if messageType == websocket.TextMessage {

			var messageJSON interface{}
			err := json.Unmarshal(message, &messageJSON)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(messageJSON)

			GetNameTest(conn)
		}
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

	mgr := server.NewManager()

	fmt.Println("Successful connect to database!")

	mux := http.NewServeMux()

	mux.HandleFunc("/login", srv.Login)
	mux.HandleFunc("/validate-token", func(w http.ResponseWriter, r *http.Request) {
		srv.ValidateDeviceToken(mgr, w, r)
	})

	ts := httptest.NewServer(mux)

	defer ts.Close()

	deviceToken := RegisterTest(ts)

	if deviceToken != "" {
		ValidToken(ts, deviceToken)
	}
}
