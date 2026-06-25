package server

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WorkerCount     = 8
	TaskQueueSize   = 512
	SessionTimeOut  = 30 * time.Second
	CleanupInterval = 30 * time.Second
	PollQueueSize   = 64
)

// Коды как в c++
const (
	CodeNewChat = 600
	CodeSync    = 700
	CodeMessage = 800
	CodeGetName = 900
)

type TaskCode int

type Task struct {
	SessionID string
	Code      TaskCode
	Payload   []byte
}

type Session struct {
	UserID     uint
	UserName   string
	Conn       *websocket.Conn
	LastActive time.Time
	OutQueue   chan []byte
	mu         sync.Mutex
}

type Manager struct {
	sessions  map[string]*Session
	mu        sync.RWMutex
	taskQueue chan Task
}

func NewManager() *Manager {
	m := &Manager{
		sessions:  make(map[string]*Session),
		taskQueue: make(chan Task, TaskQueueSize),
	}

	for i := 0; i < WorkerCount; i++ {
		go m.worker(i)
	}

	go m.cleanupLoop()
	return m
}

// Либо создает новую сессию, либо если сессия активна меняет conn
func (m *Manager) AddSession(sessionID string, userID uint, userName string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[sessionID]; ok {
		s.mu.Lock()
		s.Conn = conn
		s.LastActive = time.Now()
		s.mu.Unlock()
		return
	}

	m.sessions[sessionID] = &Session{
		UserID:     userID,
		UserName:   userName,
		Conn:       conn,
		LastActive: time.Now(),
		OutQueue:   make(chan []byte, PollQueueSize),
	}
}

func (m *Manager) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[sessionID]; ok {
		if s.Conn != nil {
			s.Conn.Close()
		}
		delete(m.sessions, sessionID)
	}
}

func (m *Manager) GetSession(sessionID string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.sessions[sessionID]
	return s, ok
}

func (m *Manager) Enqueue(sessionID string, code TaskCode, payload []byte) {
	select {
	case m.taskQueue <- Task{SessionID: sessionID, Code: code, Payload: payload}:
	default:
		log.Printf("task queue full, dropping task for session %s code %d", sessionID, code)
	}
}

// Отправляет по ws, если Conn == nil то добавляет в очередь на отправку poll
func (m *Manager) Send(sessionId string, data []byte) {
	s, ok := m.GetSession(sessionId)
	if !ok {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.LastActive = time.Now()

	if s.Conn != nil {
		if err := s.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("ws send error session %s: %v", sessionId, err)
			s.Conn = nil
		} else {
			return
		}
	}

	select {
	case s.OutQueue <- data:
	default:
		log.Printf("poll queue full for session %s", sessionId)
	}
}

// горутина воркер, читает с общей очереди
func (m *Manager) worker(id int) {
	for task := range m.taskQueue {
		m.handleTask(task)
	}
}

func (m *Manager) handleTask(task Task) {
	s, ok := m.GetSession(task.SessionID)
	if !ok {
		return
	}
	s.mu.Lock()
	s.LastActive = time.Now()
	s.mu.Unlock()

	switch task.Code {
	case CodeNewChat:
		m.handleNewChat(task.SessionID, s, task.Payload)
	case CodeSync:
		m.handleSync(task.SessionID, s, task.Payload)
	case CodeMessage:
		m.handleMessage(task.SessionID, s, task.Payload)
	case CodeGetName:
		m.handleGetName(task.SessionID, s)
	default:
		log.Printf("unknown task code %d for session %s", task.Code, task.SessionID)
	}
}

func (m *Manager) handleNewChat(sessionID string, s *Session, payload []byte) {
	// TODO: логика создания чата

	log.Printf("new_chat from %s", s.UserName)

	resp, _ := json.Marshal(map[string]interface{}{
		"status": "new_chat",
		"code":   CodeNewChat,
		"data":   "",
	})
	m.Send(sessionID, resp)
}

func (m *Manager) handleMessage(sessionID string, s *Session, payload []byte) {
	// TODO: логика сообщения
	log.Printf("new_message from %s", s.UserName)

	resp, _ := json.Marshal(map[string]interface{}{
		"status": "new_message_ok",
		"code":   CodeMessage,
		"data":   "",
	})
	m.Send(sessionID, resp)
}

func (m *Manager) handleSync(sessionID string, s *Session, payload []byte) {
	// TODO: логика синка
	log.Printf("sync from %s", s.UserName)

	resp, _ := json.Marshal(map[string]interface{}{
		"status": "sync_data",
		"code":   CodeSync,
		"data":   []interface{}{},
	})
	m.Send(sessionID, resp)
}

func (m *Manager) handleGetName(sessionID string, s *Session) {
	resp, _ := json.Marshal(map[string]interface{}{
		"status": "get_name",
		"code":   CodeGetName,
		"data":   map[string]string{"name": s.UserName},
	})
	m.Send(sessionID, resp)
}

func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanup()
	}
}

func (m *Manager) cleanup() {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, s := range m.sessions {
		s.mu.Lock()
		inactive := now.Sub(s.LastActive) > SessionTimeOut
		s.mu.Unlock()

		if inactive {
			if s.Conn != nil {
				s.Conn.Close()
			}
			delete(m.sessions, id)
		}
	}
}

func clientWebSocketActivity(mgr *Manager, sessionID string, conn *websocket.Conn) {
	defer mgr.RemoveSession(sessionID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("ws read error session %s: %v", sessionID, err)
			return
		}

		var incoming struct {
			Code    TaskCode `json:"code"`
			Payload []byte   `json:"data"`
		}

		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Printf("ws parse error session %s: %v", sessionID, err)
			continue
		}

		mgr.Enqueue(sessionID, incoming.Code, incoming.Payload)
	}
}
