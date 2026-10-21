package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

func main() {
	dbURL := envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/scraps?sslmode=disable")
	addr := envOr("ADDR", ":8080")

	db, err := NewDB(dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	h := &Handler{DB: db}

	router := httprouter.New()
	router.GET("/api/health", h.Health)
	router.POST("/api/users", h.CreateUser)
	router.GET("/api/users", h.ListUsers)
	router.GET("/api/users/:id", h.GetUser)
	router.POST("/api/scraps", h.CreateScrap)
	router.GET("/api/scraps", h.ListScraps)
	router.GET("/api/scraps/:id", h.GetScrap)
	router.POST("/api/scraps/:id/replies", h.CreateReply)
	router.GET("/api/scraps/:id/replies", h.ListReplies)

	srv := &http.Server{
		Addr:         addr,
		Handler:      CORS(router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	log.Printf("server listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}