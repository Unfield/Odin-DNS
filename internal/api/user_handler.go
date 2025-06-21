package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Unfield/Odin-DNS/internal/types"
	"github.com/Unfield/Odin-DNS/internal/util"
	"github.com/alexedwards/argon2id"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	SessionID string `json:"session_id"`
	Token     string `json:"token"`
	Username  string `json:"username"`
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	checkSuccessfull := util.CheckForDemoKey(r.URL.Query(), w, h.config.DEMO_KEY)
	if !checkSuccessfull {
		h.logger.Info("Get user attempt with invalid demo key")
		return
	}

	var loginReq LoginRequest

	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUser(loginReq.Username)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		h.logger.Info("Login attempt with non-existing user", "username", loginReq.Username)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	passwordValid, err := argon2id.ComparePasswordAndHash(loginReq.Password, user.PasswordHash)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !passwordValid || user.DeletedAt.Valid {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	sessionId, err := gonanoid.New()
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	sessionToken, err := gonanoid.New(42)
	if err != nil {
		http.Error(w, "Failed to create session token", http.StatusInternalServerError)
		return
	}

	session := &types.Session{
		ID:        sessionId,
		UserID:    user.ID,
		Token:     sessionToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = h.store.CreateSession(session)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := LoginResponse{
		SessionID: session.ID,
		Token:     session.Token,
		Username:  user.Username,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User logged in", "username", user.Username, "session_id", session.ID)
}

type RegisterRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	checkSuccessfull := util.CheckForDemoKey(r.URL.Query(), w, h.config.DEMO_KEY)
	if !checkSuccessfull {
		h.logger.Info("Get user attempt with invalid demo key")
		return
	}

	var registerReq RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&registerReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if registerReq.Username == "" || registerReq.Password == "" || registerReq.PasswordConfirm == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	if registerReq.Password != registerReq.PasswordConfirm {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	userId, err := gonanoid.New()
	if err != nil {
		http.Error(w, "Failed to create user ID", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := argon2id.CreateHash(registerReq.Password, argon2id.DefaultParams)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := &types.User{
		ID:           userId,
		Username:     registerReq.Username,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = h.store.CreateUser(user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User registered", "username", user.Username, "user_id", user.ID)
}

type LogoutRequest struct {
	SessionID string `json:"session_id"`
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	checkSuccessfull := util.CheckForDemoKey(r.URL.Query(), w, h.config.DEMO_KEY)
	if !checkSuccessfull {
		h.logger.Info("Get user attempt with invalid demo key")
		return
	}

	var logoutReq LogoutRequest
	err := json.NewDecoder(r.Body).Decode(&logoutReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if logoutReq.SessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	session, err := h.store.GetSession(logoutReq.SessionID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if session == nil {
		h.logger.Info("Logout attempt with non-existing session", "session_id", logoutReq.SessionID)
		http.Error(w, "Invalid session ID", http.StatusUnauthorized)
		return
	}

	session.DeletedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	err = h.store.UpdateSession(session)
	if err != nil {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	h.logger.Info("User logged out", "session_id", logoutReq.SessionID)
}

type GetUserRequest struct {
	SessionID string `json:"session_id"`
}

type GetUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	checkSuccessfull := util.CheckForDemoKey(r.URL.Query(), w, h.config.DEMO_KEY)
	if !checkSuccessfull {
		h.logger.Info("Get user attempt with invalid demo key")
		return
	}

	var getUserReq GetUserRequest
	err := json.NewDecoder(r.Body).Decode(&getUserReq)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if getUserReq.SessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	session, err := h.store.GetSession(getUserReq.SessionID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if session == nil || session.DeletedAt.Valid {
		h.logger.Info("Get user attempt with non-existing or deleted session", "session_id", getUserReq.SessionID)
		http.Error(w, "Invalid session ID", http.StatusUnauthorized)
		return
	}

	user, err := h.store.GetUserById(session.UserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if user == nil || user.DeletedAt.Valid {
		h.logger.Info("Get user attempt with non-existing or deleted user", "user_id", session.UserID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := GetUserResponse{
		ID:       user.ID,
		Username: user.Username,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User details retrieved", "user_id", user.ID, "username", user.Username)
}
