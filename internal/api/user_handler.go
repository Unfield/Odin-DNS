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

type GenericErrorResponse struct {
	Error        bool   `json:"error"`
	ErrorMessage string `json:"error_message"`
}

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
	var loginReq LoginRequest

	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Username and password are required"})
		return
	}

	user, err := h.store.GetUser(loginReq.Username)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Internal server error"})
		return
	}

	if user == nil {
		util.RespondWithJSON(w, http.StatusUnauthorized, &GenericErrorResponse{Error: true, ErrorMessage: "Invalid username or password"})
		return
	}

	passwordValid, err := argon2id.ComparePasswordAndHash(loginReq.Password, user.PasswordHash)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to verify password"})
		return
	}

	if !passwordValid || user.DeletedAt.Valid {
		util.RespondWithJSON(w, http.StatusUnauthorized, &GenericErrorResponse{Error: true, ErrorMessage: "Invalid username or password"})
		return
	}

	sessionId, err := gonanoid.New()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to create session ID"})
		return
	}

	sessionToken, err := gonanoid.New(42)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to create session token"})
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
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to create session"})
		return
	}

	util.RespondWithJSON(w, http.StatusOK, &LoginResponse{SessionID: session.ID, Token: session.Token, Username: user.Username})
}

type RegisterRequest struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type RegisterResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var registerReq RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&registerReq)
	if err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	if registerReq.Username == "" || registerReq.Password == "" || registerReq.PasswordConfirm == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Username, password, and password confirmation are required"})
		return
	}

	if registerReq.Password != registerReq.PasswordConfirm {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Passwords do not match"})
		return
	}

	userId, err := gonanoid.New()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to create user ID"})
		return
	}

	hashedPassword, err := argon2id.CreateHash(registerReq.Password, argon2id.DefaultParams)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to hash password"})
		return
	}

	user := &types.User{
		ID:           userId,
		Username:     registerReq.Username,
		Email:        registerReq.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = h.store.CreateUser(user)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to create user"})
		return
	}

	util.RespondWithJSON(w, http.StatusCreated, &RegisterResponse{ID: user.ID, Username: user.Username})
}

type LogoutRequest struct {
	SessionID string `json:"session_id"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	var logoutReq LogoutRequest
	err := json.NewDecoder(r.Body).Decode(&logoutReq)
	if err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	if logoutReq.SessionID == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Session ID is required"})
		return
	}

	session, err := h.store.GetSession(logoutReq.SessionID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Internal server error"})
		return
	}

	if session == nil {
		util.RespondWithJSON(w, http.StatusUnauthorized, &GenericErrorResponse{Error: true, ErrorMessage: "Invalid session ID"})
		return
	}

	session.DeletedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	err = h.store.UpdateSession(session)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to log out"})
		return
	}

	util.RespondWithJSON(w, http.StatusOK, &LogoutResponse{Message: "Successfully logged out"})
}

type GetUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := r.PathValue("session_id")

	if sessionId == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &GenericErrorResponse{Error: true, ErrorMessage: "Session ID is required"})
		return
	}

	session, err := h.store.GetSession(sessionId)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Internal server error"})
		return
	}

	if session == nil || session.DeletedAt.Valid {
		util.RespondWithJSON(w, http.StatusUnauthorized, &GenericErrorResponse{Error: true, ErrorMessage: "Invalid or expired session"})
		return
	}

	user, err := h.store.GetUserById(session.UserID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &GenericErrorResponse{Error: true, ErrorMessage: "Failed to retrieve user"})
		return
	}

	if user == nil || user.DeletedAt.Valid {
		util.RespondWithJSON(w, http.StatusUnauthorized, &GenericErrorResponse{Error: true, ErrorMessage: "User not found or deleted"})
		return
	}

	util.RespondWithJSON(w, http.StatusOK, &GetUserResponse{ID: user.ID, Username: user.Username, Email: user.Email})
}
