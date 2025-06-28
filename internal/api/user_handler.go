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
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type LoginResponse struct {
	SessionID string `json:"session_id" example:"V1StGXR8_Z5jdHi6B-myT"`
	Token     string `json:"token" example:"kWnEeaODiH5Sb1H1REbfLA3VTl7jbvlpAn4vKNDXEcgOcgdmRhRjRb"`
	Username  string `json:"username" example:"john_doe"`
}

// LoginHandler handles user authentication
// @Summary User login
// @Description Authenticate a user with username and password, returns session information
// @Tags authentication
// @Accept json
// @Produce json
// @Param demo_key query string true "Demo API key for authentication"
// @Param login body LoginRequest true "User login credentials"
// @Success 200 {object} LoginResponse "Login successful"
// @Failure 400 {object} GenericErrorResponse "Bad request - invalid input"
// @Failure 401 {object} GenericErrorResponse "Unauthorized - invalid credentials"
// @Failure 500 {object} GenericErrorResponse "Internal server error"
// @Security DemoKey
// @Router /api/v1/login [post]
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
	Username        string `json:"username" binding:"required" example:"john_doe" minLength:"3" maxLength:"50"`
	Email           string `json:"email" binding:"required" example:"john@example.com" format:"email"`
	Password        string `json:"password" binding:"required" example:"password123" minLength:"8"`
	PasswordConfirm string `json:"password_confirm" binding:"required" example:"password123" minLength:"8"`
}

type RegisterResponse struct {
	ID       string `json:"id" example:"V1StGXR8_Z5jdHi6B-myT"`
	Username string `json:"username" example:"john_doe"`
}

// RegisterHandler handles user registration
// @Summary User registration
// @Description Register a new user account with username, email, and password
// @Tags authentication
// @Accept json
// @Produce json
// @Param demo_key query string true "Demo API key for authentication"
// @Param register body RegisterRequest true "User registration details"
// @Success 201 {object} RegisterResponse "Registration successful"
// @Failure 400 {object} GenericErrorResponse "Bad request - invalid input or passwords don't match"
// @Failure 500 {object} GenericErrorResponse "Internal server error - failed to create user"
// @Security DemoKey
// @Router /api/v1/register [post]
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
	SessionID string `json:"session_id" binding:"required" example:"V1StGXR8_Z5jdHi6B-myT"`
}

type LogoutResponse struct {
	Message string `json:"message" example:"Successfully logged out"`
}

// LogoutHandler handles user logout
// @Summary User logout
// @Description Logout a user by invalidating their session
// @Tags authentication
// @Accept json
// @Produce json
// @Param demo_key query string true "Demo API key for authentication"
// @Param logout body LogoutRequest true "Session information for logout"
// @Success 200 {object} LogoutResponse "Logout successful"
// @Failure 400 {object} GenericErrorResponse "Bad request - invalid session ID"
// @Failure 401 {object} GenericErrorResponse "Unauthorized - invalid session"
// @Failure 500 {object} GenericErrorResponse "Internal server error"
// @Security DemoKey
// @Router /api/v1/logout [post]
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
	ID       string `json:"id" example:"V1StGXR8_Z5jdHi6B-myT"`
	Username string `json:"username" example:"john_doe"`
	Email    string `json:"email" example:"john@example.com"`
}

// GetUserHandler retrieves user information by session ID
// @Summary Get user information
// @Description Retrieve user details using a valid session ID
// @Tags users
// @Produce json
// @Param demo_key query string true "Demo API key for authentication"
// @Param session_id path string true "User session ID" minLength:"1"
// @Success 200 {object} GetUserResponse "User information retrieved successfully"
// @Failure 400 {object} GenericErrorResponse "Bad request - session ID required"
// @Failure 401 {object} GenericErrorResponse "Unauthorized - invalid or expired session"
// @Failure 500 {object} GenericErrorResponse "Internal server error"
// @Security DemoKey
// @Router /api/v1/user/{session_id} [get]
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
