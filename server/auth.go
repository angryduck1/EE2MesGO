package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
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

	sendAuthMessage(w, http.StatusOK, "ok", "SUCCESSFUL_REGISTER", "Successful register to account", deviceToken)
}

func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegistrationInfo

	e := json.NewDecoder(r.Body).Decode(&req)

	defer r.Body.Close()

	if e != nil {
		log.Printf("Login error: %v", e)

		sendMessage(w, http.StatusBadRequest, "error", "SERVER_ERROR", "Fail connect to server")

		return
	}

	name, password := req.Name, req.Password

	match, userInfo, err := server.getUser(r.Context(), name, password)

	if err != nil {
		if errors.Is(r.Context().Err(), context.Canceled) {
			return
		}

		if errors.Is(err, argon2id.ErrIncompatibleVariant) || errors.Is(err, argon2id.ErrIncompatibleVersion) || errors.Is(err, argon2id.ErrInvalidHash) {
			sendMessage(w, http.StatusInternalServerError, "error", "SERVER_ERROR", "Authorization fail")

			log.Printf("Cryption error: %v", err)

			return
		}

		log.Printf("BD error: %v", err)

		sendMessage(w, http.StatusInternalServerError, "error", "SERVER_ERROR", "Server error")

		return
	}

	if match {
		deviceToken, _ := generateDeviceToken()

		err = server.addNewDevice(r.Context(), userInfo, deviceToken)

		if err != nil {
			if errors.Is(r.Context().Err(), context.Canceled) {
				return
			}

			log.Printf("BD error: %v", err)

			sendMessage(w, http.StatusInternalServerError, "error", "SERVER_ERROR", "Server error")

			return
		}

		sendAuthMessage(w, http.StatusOK, "ok", "SUCCESSFUL_LOGIN", "Successful login to account", deviceToken)
	} else {
		sendMessage(w, http.StatusUnauthorized, "error", "LOGIN_FAIL", "Fail login to account")
	}
}

func (server *Server) ValidateDeviceToken(mgr *Manager, w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("deviceToken")
	if sessionID == "" {
		sendMessage(w, http.StatusBadRequest, "error", "BAD_REQUEST", "Missing sessionId")
		return
	}

	user, err := server.getUserByToken(r.Context(), sessionID)
	if err != nil {
		if errors.Is(r.Context().Err(), context.Canceled) {
			return
		}

		sendMessage(w, http.StatusUnauthorized, "error", "INVALID_TOKEN", "Invalid device token")

		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	mgr.AddSession(sessionID, user.ID, user.Name, conn)

	log.Printf("user %s connected via websocket", user.Name)

	welcome, _ := json.Marshal(map[string]interface{}{
		"status": "ok",
		"code":   "SUCCESSFUL_TOKEN",
		"data": map[string]interface{}{
			"id":   user.ID,
			"name": user.Name,
		},
	})

	conn.WriteMessage(websocket.TextMessage, welcome)

	clientWebSocketActivity(mgr, sessionID, conn)
}
