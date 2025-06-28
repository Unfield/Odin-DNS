package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Unfield/Odin-DNS/internal/datastore"
	"github.com/Unfield/Odin-DNS/internal/models"
	"github.com/Unfield/Odin-DNS/internal/types"
	"github.com/Unfield/Odin-DNS/internal/util"
)

func AuthMiddleware(cache datastore.Driver) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(RequestIDKey)
			if requestID == nil {
				requestID = "no-request-id"
			}
			log.Printf("[%s] AuthMiddleware: Processing request %s %s", requestID, r.Method, r.URL.Path)

			// Crucial: Allow OPTIONS requests to pass through
			// The CORS middleware already handles setting the appropriate headers for OPTIONS.
			// If you uncommented the explicit OPTIONS handlers in your router,
			// this might not be strictly necessary here, but it's good defensive programming.
			if r.Method == http.MethodOptions {
				log.Printf("[%s] AuthMiddleware: Skipping authentication for OPTIONS preflight.", requestID)
				next.ServeHTTP(w, r) // Still pass to next in chain, which could be the actual handler if no more middleware
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("[%s] AuthMiddleware: Authorization header missing.", requestID)
				util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Authorization header required"})
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader { // Check if "Bearer " prefix was not found
				log.Printf("[%s] AuthMiddleware: Invalid token format (missing Bearer prefix).", requestID)
				util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid token format"})
				return
			}
			if token == "" {
				log.Printf("[%s] AuthMiddleware: Empty token after stripping Bearer prefix.", requestID)
				util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid token"})
				return
			}

			session, err := cache.GetSessionByToken(token)
			if err != nil {
				log.Printf("[%s] AuthMiddleware: Error getting session for token %s: %v", requestID, token[:min(len(token), 10)]+"...", err) // Log a prefix of token
				util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid session"})
				return
			}

			if session == nil {
				log.Printf("[%s] AuthMiddleware: Session not found for token %s", requestID, token[:min(len(token), 10)]+"...")
				util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Session not found"})
				return
			}

			if session.DeletedAt.Valid && session.DeletedAt.Time.Before(time.Now()) {
				log.Printf("[%s] AuthMiddleware: Session %s is expired/deleted.", requestID, session.ID)
				util.RespondWithJSON(w, http.StatusUnauthorized, &models.GenericErrorResponse{Error: true, ErrorMessage: "Invalid or expired session"})
				return
			}

			log.Printf("[%s] AuthMiddleware: Session %s (User %s) is valid. Passing to next handler.", requestID, session.ID, session.UserID)
			ctx := context.WithValue(r.Context(), "user_session", &types.SessionContextKey{SessionID: session.ID, UserID: session.UserID, Token: session.Token})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper to get minimum for logging token prefix
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
