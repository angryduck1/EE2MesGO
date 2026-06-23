package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/alexedwards/argon2id"
)

func (server *Server) Registration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegistrationInfo

	e := json.NewDecoder(r.Body).Decode(&req)

	defer r.Body.Close()

	if e != nil {
		log.Printf("Registration error: %v", e)

		sendMessage(w, http.StatusBadRequest, "error", "BAD_REQUEST", "Bad request")

		return
	}

	name, password := req.Name, req.Password

	deviceToken, err := server.insertNewUser(r.Context(), name, password)
	if err != nil {
		if errors.Is(r.Context().Err(), context.Canceled) {
			return
		}

		log.Printf("BD error: %v", err)

		sendMessage(w, http.StatusInternalServerError, "error", "SERVER_ERROR", "Fail connect to server")

		return
	}

	response := ResponseInfo{
		Status:  "ok",
		Code:    "SUCCESSFUL_REGISTER",
		Message: "Successful register to account",
	}
	
	w.WriteHeader(http.StatusOK)
	if e := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      response.Status,
		"code":        response.Code,
		"message":     response.Message,
		"deviceToken": deviceToken,
	}); e != nil {
		log.Printf("Registration response error: %v", e)
		return
	}
}

func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegistrationInfo

	e := json.NewDecoder(r.Body).Decode(&req)

	defer r.Body.Close()

	if e != nil {
		log.Printf("Login error: %v", e)
		http.Error(w, "Login error", http.StatusBadRequest)
		return
	}

	name, password := req.Name, req.Password

	match, err := server.getUser(r.Context(), name, password)

	if err != nil {
		if errors.Is(r.Context().Err(), context.Canceled) {
			return
		}

		if errors.Is(err, argon2id.ErrIncompatibleVariant) || errors.Is(err, argon2id.ErrIncompatibleVersion) || errors.Is(err, argon2id.ErrInvalidHash) {
			sendMessage(w, http.StatusInternalServerError, "error", "SERVER_ERROR", "Authorization fail")

			log.Printf("Cryption error: %v", err.Error())

			return
		}

		log.Printf("BD error: %v", err)

		sendMessage(w, http.StatusInternalServerError, "error", "SERVER_ERROR", "Server error")

		return
	}

	if match {
		sendMessage(w, http.StatusOK, "ok", "SUCCESSFUL_LOGIN", "Successful login to account")
	} else {
		sendMessage(w, http.StatusUnauthorized, "ok", "LOGIN_FAIL", "Fail login to account")
	}
}

func (server *Server) ValidateDeviceToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		DeviceToken string `json:"deviceToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendMessage(w, http.StatusBadRequest, "error", "BAD_REQUEST", "Invalid request")
		return
	}

	user, err := server.getUserByToken(r.Context(), req.DeviceToken)
	if err != nil {
		sendMessage(w, http.StatusUnauthorized, "error", "INVALID_TOKEN", "Invalid device token")

		return
	}

	response := map[string]interface{}{
		"status": "ok",
		"user": map[string]interface{}{
			"id":   user.ID,
			"name": user.Name,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
