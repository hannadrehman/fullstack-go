package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
)

type Handler struct {
	DB *DB
}

// CORS allows the Vite dev server to call the API.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
	case errors.Is(err, ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Avatar string `json:"avatar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	u, err := h.DB.CreateUser(r.Context(), req.Name, req.Email, req.Avatar)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	users, err := h.DB.ListUsers(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := strconv.ParseInt(ps.ByName("id"), 10, 64)
	u, err := h.DB.GetUser(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) CreateScrap(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req struct {
		UserID  int64  `json:"user_id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	s, err := h.DB.CreateScrap(r.Context(), req.UserID, req.Content)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) ListScraps(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	scraps, err := h.DB.ListScraps(r.Context(), limit)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, scraps)
}

func (h *Handler) GetScrap(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, _ := strconv.ParseInt(ps.ByName("id"), 10, 64)
	s, err := h.DB.GetScrap(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) CreateReply(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	scrapID, _ := strconv.ParseInt(ps.ByName("id"), 10, 64)
	var req struct {
		UserID  int64  `json:"user_id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	rep, err := h.DB.CreateReply(r.Context(), scrapID, req.UserID, req.Content)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rep)
}

func (h *Handler) ListReplies(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	scrapID, _ := strconv.ParseInt(ps.ByName("id"), 10, 64)
	replies, err := h.DB.ListReplies(r.Context(), scrapID)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, replies)
}