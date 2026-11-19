package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/scraps?sslmode=disable"
	defaultListenAddr  = ":8080"
	ioTimeout          = 15 * time.Second // read and write share the same budget, keep it simple
)

func main() {
	// config first, then everything else. the order matters, probably.
	dbURL := envOr("DATABASE_URL", defaultDatabaseURL)
	listenAddr := envOr("ADDR", defaultListenAddr)

	database, err := NewDB(dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	srv := &http.Server{
		Addr:         listenAddr,
		Handler:      CORS(newRouter(&Handler{DB: database})),
		ReadTimeout:  ioTimeout,
		WriteTimeout: ioTimeout,
	}

	log.Printf("server listening on %s", listenAddr)
	log.Fatal(srv.ListenAndServe())
}

// newRouter wires every API route to its handler. Paths are fixed, so nothing here is dynamic.
func newRouter(h *Handler) *httprouter.Router {
	router := httprouter.New()

	// health check: if this one is down, something is very wrong
	router.GET("/api/health", h.Health)

	// users
	router.POST("/api/users", h.CreateUser)
	router.GET("/api/users", h.ListUsers)
	router.GET("/api/users/:id", h.GetUser)
	router.GET("/api/users/:id/scraps", h.ListUserScraps)

	// scraps and their replies
	router.POST("/api/scraps", h.CreateScrap)
	router.GET("/api/scraps", h.ListScraps)
	router.GET("/api/scraps/:id", h.GetScrap)
	router.POST("/api/scraps/:id/replies", h.CreateReply)
	router.GET("/api/scraps/:id/replies", h.ListReplies)

	return router
}

// envOr returns the value of key, or def when the variable is unset or empty.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	// no environment variable was harmed in the making of this fallback
	return def
}
