package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"mini-poc-02/backend/internal/config"
	"mini-poc-02/backend/internal/db"
	"mini-poc-02/backend/internal/httpapi"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// d, err := db.Open(cfg.MySQLDSN())
	dbConn, err := db.OpenPostgres(cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	// defer d.SQL.Close()
	defer dbConn.Close()

	// handlers := httpapi.NewHandlers(d.SQL)
	handlers := httpapi.NewHandlers(dbConn)
	router := httpapi.NewRouter(handlers)

	addr := ":" + cfg.AppPort

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Bind dulu supaya tidak misleading: kalau bind gagal, tidak akan log "listening".
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to bind %s (is it already in use?): %v", addr, err)
	}

	log.Printf("API listening on http://localhost:%s", cfg.AppPort)

	// Serve menggunakan listener yang sudah dipastikan berhasil.
	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
