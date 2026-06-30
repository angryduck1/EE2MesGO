package main

import (
	"fmt"
	"log"
	"net/http"

	server "EE2MesGO/server"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=root dbname=messenger port=5432 sslmode=disable TimeZone=Europe/Moscow"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Fail open bd: %v", err.Error())
	}

	err = db.AutoMigrate(&server.UserInfo{}, &server.DeviceInfo{})

	if err != nil {
		log.Fatalf("Fail migrate bd: %v", err.Error())
	}

	srv := &server.Server{DB: db}

	fmt.Println("Successful connect to database!")

	http.HandleFunc("/registration", srv.Registration)
	http.HandleFunc("/login", srv.Login)
	
	mgr := server.NewManager()

	http.HandleFunc("/validate-token", func(w http.ResponseWriter, r *http.Request) {
		srv.ValidateDeviceToken(mgr, w, r)
	})
	http.HandleFunc("/poll", func(w http.ResponseWriter, r *http.Request) {
		srv.LongPoll(mgr, w, r)
	})
	http.HandleFunc("/poll/send", func(w http.ResponseWriter, r *http.Request) {
		srv.PollSend(mgr, w, r)
	})

	e := http.ListenAndServeTLS("localhost:8080", "./keys/localhost+1.pem", "./keys/localhost+1-key.pem", nil)

	// e := http.ListenAndServe("localhost:8080", nil)

	if e != nil {
		log.Fatalf("Fatal error running server: %v", e.Error())
	}
}
