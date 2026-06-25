package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

const (
	PollTimeout = 25 * time.Second
)

// LongPoll — GET /poll?deviceToken=...
// устанавливает LongPoll, после таймаута вернет 204, клиенту нужно заново отправить запрос
func (server *Server) LongPoll(mgr *Manager, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendMessage(w, http.StatusMethodNotAllowed, "error", "BAD_REQUEST", "Invalid method")
		return
	}

	sessionID := r.URL.Query().Get("deviceToken")
	if sessionID == "" {
		sendMessage(w, http.StatusBadRequest, "error", "BAD_REQUEST", "Missing deviceToken")
		return
	}

	user, err := server.getUserByToken(r.Context(), sessionID)
	if err != nil {
		if errors.Is(r.Context().Err(), context.Canceled) {
			return
		}
		sendMessage(w, http.StatusUnauthorized, "error", "INVALID_TOKEN", "Invalid session")
		return
	}

	mgr.AddSession(sessionID, user.ID, user.Name, nil)

	s, ok := mgr.GetSession(sessionID)
	if !ok {
		sendMessage(w, http.StatusInternalServerError, "error", "SERVER_ERROR", "Session error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	ctx, cancel := context.WithTimeout(r.Context(), PollTimeout)
	defer cancel()

	select {
	case msg := <-s.OutQueue:
		w.WriteHeader(http.StatusOK)
		w.Write(msg)

	case <-ctx.Done():
		w.WriteHeader(http.StatusNoContent)
	}
}

// PollSend — POST /poll/send?deviceToken=...
// Пишет в ту же общую очередь, если ws недоступен
func (server *Server) PollSend(mgr *Manager, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendMessage(w, http.StatusMethodNotAllowed, "error", "BAD_REQUEST", "Invalid method")
		return
	}

	sessionID := r.URL.Query().Get("deviceToken")
	if sessionID == "" {
		sendMessage(w, http.StatusBadRequest, "error", "BAD_REQUEST", "Missing deviceToken")
		return
	}

	_, err := server.getUserByToken(r.Context(), sessionID)
	if err != nil {
		if errors.Is(r.Context().Err(), context.Canceled) {
			return
		}
		sendMessage(w, http.StatusUnauthorized, "error", "INVALID_TOKEN", "Invalid session")
		return
	}

	var incoming struct {
		Code    TaskCode        `json:"code"`
		Payload json.RawMessage `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		sendMessage(w, http.StatusBadRequest, "error", "BAD_REQUEST", "Invalid body")
		return
	}
	defer r.Body.Close()

	mgr.Enqueue(sessionID, incoming.Code, incoming.Payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"code":   "QUEUED",
	})

	log.Printf("poll send from session %s code %d", sessionID, incoming.Code)
}
