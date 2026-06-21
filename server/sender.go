package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func sendMessage(w http.ResponseWriter, statusCode int, status string, code string, message string) {
	w.WriteHeader(statusCode)

	response := ResponseInfo{Status: status, Code: code, Message: message}

	if e := json.NewEncoder(w).Encode(response); e != nil {
		log.Printf("sendMessage error: %v", e)
		return
	}
}
