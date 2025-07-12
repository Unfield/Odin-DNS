package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Unfield/Odin-DNS/internal/models"
	"github.com/Unfield/Odin-DNS/internal/types"
	"github.com/Unfield/Odin-DNS/internal/util"
	"github.com/alexedwards/argon2id"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// LoginHandler handles user authentication
// @Summary User Login
// @Description Authenticates a user and returns a session token
// @Tags authentication
// @Accept json
// @Produce json
// @Param loginRequest body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse "Login successful"
// @Failure 400 {object} models.GenericErrorResponse "Invalid request body or missing credentials"
// @Failure 401 {object} models.GenericErrorResponse "Invalid username or password"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /api/v1/login [post]
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginReq models.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	if loginReq.Username == "" || loginReq.Password == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "Username and password are required"})
		return
	}

	user, err := h.store.GetUser(loginReq.Username)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Internal server error"})
		return
	}

	if user == nil {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid username or password"})
		return
	}

	passwordValid, err := argon2id.ComparePasswordAndHash(loginReq.Password, user.PasswordHash)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to verify password"})
		return
	}

	if !passwordValid || user.DeletedAt.Valid {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid username or password"})
		return
	}

	sessionId, err := gonanoid.New()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to create session ID"})
		return
	}

	sessionToken, err := gonanoid.New(42)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to create session token"})
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
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to create session"})
		return
	}

	util.RespondWithJSON(w, http.StatusOK, &models.LoginResponse{SessionID: session.ID, Token: session.Token, Username: user.Username})
}

// RegisterHandler handles user registration
// @Summary User Registration
// @Description Creates a new user account
// @Tags authentication
// @Accept json
// @Produce json
// @Param registerRequest body models.RegisterRequest true "Registration details"
// @Success 201 {object} models.RegisterResponse "User created successfully"
// @Failure 400 {object} models.GenericErrorResponse "Invalid request body, missing fields, or passwords don't match"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /api/v1/register [post]
func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var registerReq models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&registerReq)
	if err != nil {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid request body"})
		return
	}

	if registerReq.Username == "" || registerReq.Password == "" || registerReq.PasswordConfirm == "" {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "Username, password, and password confirmation are required"})
		return
	}

	if registerReq.Password != registerReq.PasswordConfirm {
		util.RespondWithJSON(w, http.StatusBadRequest, &models.GenericErrorResponse{Error: true, ErrorMessage: "Passwords do not match"})
		return
	}

	userId, err := gonanoid.New()
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to create user ID"})
		return
	}

	hashedPassword, err := argon2id.CreateHash(registerReq.Password, argon2id.DefaultParams)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to hash password"})
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
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to create user"})
		return
	}

	util.RespondWithJSON(w, http.StatusCreated, &models.RegisterResponse{ID: user.ID, Username: user.Username})
}

// LogoutHandler handles user logout
// @Summary User Logout
// @Description Invalidates the user's session
// @Tags authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.LogoutResponse "Logout successful"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid session"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /api/v1/logout [post]
func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	userSession, sessionValid := r.Context().Value("user_session").(*types.SessionContextKey)

	if !sessionValid || userSession.Token == "" {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Unauthorized - invalid session"})
		return
	}

	session, err := h.store.GetSession(userSession.SessionID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Internal server error"})
		return
	}

	if session == nil {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid session ID"})
		return
	}

	session.DeletedAt = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
	err = h.store.UpdateSession(session)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to log out"})
		return
	}

	util.RespondWithJSON(w, http.StatusOK, &models.LogoutResponse{Message: "Successfully logged out"})
}

// GetUserHandler retrieves user information
// @Summary Get User Information
// @Description Returns information about the authenticated user
// @Tags user
// @Security BearerAuth
// @Produce json
// @Param session_id path string true "Session ID"
// @Success 200 {object} models.GetUserResponse "User information retrieved successfully"
// @Failure 401 {object} models.GenericErrorResponse "Unauthorized - invalid or expired session"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /api/v1/user/{session_id} [get]
func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	userSession, sessionValid := r.Context().Value("user_session").(*types.SessionContextKey)

	if !sessionValid || userSession.Token == "" {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Unauthorized - invalid session"})
		return
	}

	session, err := h.store.GetSession(userSession.SessionID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Internal server error"})
		return
	}

	if session == nil || session.DeletedAt.Valid {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid or expired session"})
		return
	}

	user, err := h.store.GetUserById(session.UserID)
	if err != nil {
		util.RespondWithJSON(w, http.StatusInternalServerError, &models.GenericErrorResponse{Error: true, ErrorMessage: "Failed to retrieve user"})
		return
	}

	if user == nil || user.DeletedAt.Valid {
		util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "User not found or deleted"})
		return
	}

	util.RespondWithJSON(w, http.StatusOK, &models.GetUserResponse{ID: user.ID, Username: user.Username, Email: user.Email})
}
