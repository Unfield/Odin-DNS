package models

// GenericErrorResponse represents a standard error response
type GenericErrorResponse struct {
	Error        bool   `json:"error" example:"true" description:"Indicates if an error occurred"`
	ErrorMessage string `json:"error_message" example:"Invalid request body" description:"Human-readable error message"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe" description:"User's username"`
	Password string `json:"password" binding:"required" example:"password123" description:"User's password"`
}

// LoginResponse represents the successful login response
type LoginResponse struct {
	SessionID string `json:"session_id" example:"V1StGXR8_Z5jdHi6B-myT" description:"Unique session identifier"`
	Token     string `json:"token" example:"kWnEeaODiH5Sb1H1REbfLA3VTl7jbvlpAn4vKNDXEcgOcgdmRhRjRb" description:"Bearer token for authentication"`
	Username  string `json:"username" example:"john_doe" description:"Authenticated user's username"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username        string `json:"username" binding:"required" example:"john_doe" minLength:"3" maxLength:"50" description:"Desired username (3-50 characters)"`
	Email           string `json:"email" binding:"required" example:"john@example.com" format:"email" description:"User's email address"`
	Password        string `json:"password" binding:"required" example:"password123" minLength:"8" description:"Password (minimum 8 characters)"`
	PasswordConfirm string `json:"password_confirm" binding:"required" example:"password123" minLength:"8" description:"Password confirmation (must match password)"`
}

// RegisterResponse represents the successful registration response
type RegisterResponse struct {
	ID       string `json:"id" example:"V1StGXR8_Z5jdHi6B-myT" description:"Unique user identifier"`
	Username string `json:"username" example:"john_doe" description:"Registered username"`
}

// LogoutResponse represents the logout response
type LogoutResponse struct {
	Message string `json:"message" example:"Successfully logged out" description:"Confirmation message"`
}

// GetUserResponse represents the user information response
type GetUserResponse struct {
	ID       string `json:"id" example:"V1StGXR8_Z5jdHi6B-myT" description:"Unique user identifier"`
	Username string `json:"username" example:"john_doe" description:"User's username"`
	Email    string `json:"email" example:"john@example.com" description:"User's email address"`
}
