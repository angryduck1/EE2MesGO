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

func sendAuthMessage(w http.ResponseWriter, statusCode int, status string, code string, message string, deviceToken string) {
	w.WriteHeader(statusCode)

	response := ResponseInfo{Status: status, Code: code, Message: message}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      response.Status,
		"code":        response.Code,
		"message":     response.Message,
		"deviceToken": deviceToken,
	}); err != nil {
		log.Printf("sendAuthMessage error: %v", err)
		return
	}
}
