package main

import (
	"EE2MesGO/server"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// т.к файл содержит виндусовский стиль окончания строк
func normalizeNewlines(s string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(s, "\n", "\r\n")
	}
	return s
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

func sendGetName(conn *websocket.Conn) {
	getNameResp := map[string]interface{}{"code": 900, "data": ""}

	getNameRespBuf, err := json.Marshal(getNameResp)

	if err != nil {
		log.Fatal(err)
	}

	err = conn.WriteMessage(websocket.TextMessage, getNameRespBuf)

	if err != nil {
		log.Fatal(err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	messageType, message, err := conn.ReadMessage()

	if err != nil {
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			log.Println("Read timeout - skipping response")
			return
		}
		log.Fatal(err)
	}

	if messageType == websocket.TextMessage {

		var messageJSON interface{}
		err := json.Unmarshal(message, &messageJSON)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("GetName response:", messageJSON)
	}
}

func sendNewChat(conn *websocket.Conn) {
	newChatReq := map[string]interface{}{
		"code": 600,
		"data": map[string]interface{}{
			"chat_name": "test_chat",
			"users":     []string{"meow", "test_user"},
		},
	}

	newChatBuf, err := json.Marshal(newChatReq)

	if err != nil {
		log.Fatal(err)
	}

	err = conn.WriteMessage(websocket.TextMessage, newChatBuf)

	if err != nil {
		log.Fatal(err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	messageType, message, err := conn.ReadMessage()

	if err != nil {
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			log.Println("Read timeout - skipping response")
			return
		}
		log.Fatal(err)
	}

	if messageType == websocket.TextMessage {

		var messageJSON interface{}
		err := json.Unmarshal(message, &messageJSON)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("NewChat response:", messageJSON)
	}
}

func sendSync(conn *websocket.Conn) {
	syncReq := map[string]interface{}{
		"code": 700,
		"data": map[string]interface{}{
			"last_sync": time.Now().Unix(),
		},
	}

	syncBuf, err := json.Marshal(syncReq)

	if err != nil {
		log.Fatal(err)
	}

	err = conn.WriteMessage(websocket.TextMessage, syncBuf)

	if err != nil {
		log.Fatal(err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	messageType, message, err := conn.ReadMessage()

	if err != nil {
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			log.Println("Read timeout - skipping response")
			return
		}
		log.Fatal(err)
	}

	if messageType == websocket.TextMessage {

		var messageJSON interface{}
		err := json.Unmarshal(message, &messageJSON)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Sync response:", messageJSON)
	}
}

func sendMessage(conn *websocket.Conn) {
	messageReq := map[string]interface{}{
		"code": 800,
		"data": map[string]interface{}{
			"chat_id":   1,
			"content":   "test message",
			"timestamp": time.Now().Unix(),
		},
	}

	messageBuf, err := json.Marshal(messageReq)

	if err != nil {
		log.Fatal(err)
	}

	err = conn.WriteMessage(websocket.TextMessage, messageBuf)

	if err != nil {
		log.Fatal(err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	messageType, message, err := conn.ReadMessage()

	if err != nil {
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			log.Println("Read timeout - skipping response")
			return
		}
		log.Fatal(err)
	}

	if messageType == websocket.TextMessage {

		var messageJSON interface{}
		err := json.Unmarshal(message, &messageJSON)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Message response:", messageJSON)
	}
}

func ValidToken(ts *httptest.Server, deviceToken string) {
	var wsURL string
	if runtime.GOOS == "windows" {
		wsURL = "ws" + strings.TrimPrefix(ts.URL, "http") + "/validate-token?" + "deviceToken=" + deviceToken
	} else {
		wsURL = "ws" + strings.TrimPrefix(ts.URL, "http") + "/validate-token?" + "deviceToken=" + deviceToken
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

	defer conn.Close()

	if err != nil {
		log.Fatal(err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	messageType, message, err := conn.ReadMessage()

	if err != nil {
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			log.Println("Timeout reading welcome message")
			return
		}
		log.Fatal(err)
	}

	if messageType == websocket.TextMessage {

		var messageJSON interface{}
		err := json.Unmarshal(message, &messageJSON)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Welcome message:", messageJSON)
	}

	sendGetName(conn)
	sendNewChat(conn)
	sendSync(conn)
	sendMessage(conn)

	time.Sleep(100 * time.Millisecond)
}

func TestWebSocket(t *testing.T) {
	dsn := "host=localhost user=postgres password=root dbname=messenger port=5432 sslmode=disable TimeZone=Europe/Moscow"

	//if runtime.GOOS == "windows" {
	//    dsn = "host=localhost user=postgres password=root dbname=messenger port=5432 sslmode=disable TimeZone=Europe/Moscow"
	//}

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
