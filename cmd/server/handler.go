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

// Handler bundles the HTTP handlers together with the database they talk to.
type Handler struct {
	DB *DB
}

// userInput is the body accepted by POST /users.
type userInput struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
}

// postInput is the body shared by scraps and replies.
type postInput struct {
	UserID  int64  `json:"user_id"`
	Content string `json:"content"`
}

// CORS allows the Vite dev server to call the API.
// Preflight requests are answered right here and never reach the next handler.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			// 204: nothing to say, but we are very polite about it.
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// JSON goes in, JSON comes out, magic happens here.
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
		// Something went sideways somewhere deep down. The client gets a polite shrug.
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}

// decodeBody reads a JSON request body into v. On failure it writes the 400
// response itself and reports false, so callers can just return.
func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return false
	}
	return true
}

// pathID pulls the numeric :id segment out of the route. Unparseable ids become 0,
// which simply matches nothing in the database.
func pathID(ps httprouter.Params) int64 {
	id, _ := strconv.ParseInt(ps.ByName("id"), 10, 64)
	return id
}

// queryLimit reads the optional ?limit= query parameter. Missing or junk values become 0.
func queryLimit(r *http.Request) int {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	return limit
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var in userInput
	if !decodeBody(w, r, &in) {
		return
	}
	u, err := h.DB.CreateUser(r.Context(), in.Name, in.Email, in.Avatar)
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
	u, err := h.DB.GetUser(r.Context(), pathID(ps))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) ListUserScraps(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := pathID(ps)
	// Make sure the user exists first, so a bad id gets a proper 404 instead of an empty list.
	if _, err := h.DB.GetUser(r.Context(), userID); err != nil {
		writeErr(w, err)
		return
	}
	scraps, err := h.DB.ListUserScraps(r.Context(), userID, queryLimit(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, scraps)
}

func (h *Handler) CreateScrap(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var in postInput
	if !decodeBody(w, r, &in) {
		return
	}
	s, err := h.DB.CreateScrap(r.Context(), in.UserID, in.Content)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) ListScraps(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	scraps, err := h.DB.ListScraps(r.Context(), queryLimit(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, scraps)
}

func (h *Handler) GetScrap(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s, err := h.DB.GetScrap(r.Context(), pathID(ps))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) CreateReply(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var in postInput
	if !decodeBody(w, r, &in) {
		return
	}
	// A reply is just a scrap-shaped body aimed at the scrap in the URL.
	rep, err := h.DB.CreateReply(r.Context(), pathID(ps), in.UserID, in.Content)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rep)
}

func (h *Handler) ListReplies(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	replies, err := h.DB.ListReplies(r.Context(), pathID(ps))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, replies)
}
