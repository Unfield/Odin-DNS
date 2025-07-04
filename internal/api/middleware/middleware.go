package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

type Middleware func(http.Handler) http.Handler

type Chain struct {
	middlewares []Middleware
}

func New(middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: append([]Middleware(nil), middlewares...),
	}
}

func (c *Chain) Use(middlewares ...Middleware) *Chain {
	newChain := &Chain{
		middlewares: make([]Middleware, len(c.middlewares)+len(middlewares)),
	}
	copy(newChain.middlewares, c.middlewares)
	copy(newChain.middlewares[len(c.middlewares):], middlewares)
	return newChain
}

func (c *Chain) Then(h http.Handler) http.Handler {
	if h == nil {
		h = http.DefaultServeMux
	}

	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

func (c *Chain) ThenFunc(fn http.HandlerFunc) http.Handler {
	return c.Then(fn)
}

type contextKey string

const (
	RequestIDKey contextKey = "requestID"
	StartTimeKey contextKey = "startTime"
)

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := fmt.Sprintf("%d", time.Now().UnixNano())
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := context.WithValue(r.Context(), StartTimeKey, start)

			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			next.ServeHTTP(wrapped, r.WithContext(ctx))

			duration := time.Since(start)
			requestID := r.Context().Value(RequestIDKey)

			log.Printf(
				"[%s] %s %s %d %v %s",
				requestID,
				r.Method,
				r.URL.Path,
				wrapped.statusCode,
				duration,
				r.RemoteAddr,
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := r.Context().Value(RequestIDKey)
					log.Printf(
						"[%s] PANIC: %v\n%s",
						requestID,
						err,
						debug.Stack(),
					)

					http.Error(w, "Internal Server Error",
						http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int // in seconds
}

// CORS middleware function
func CORS(config CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(RequestIDKey)
			if requestID == nil {
				requestID = "no-request-id"
			}

			origin := r.Header.Get("Origin")
			log.Printf("[%s] CORS: Request from Origin: '%s', Method: '%s', Path: '%s'", requestID, origin, r.Method, r.URL.Path)

			// Determine if it's a cross-origin request
			isCrossOrigin := (origin != "")

			// 1. Set Access-Control-Allow-Origin
			if isCrossOrigin { // Only apply ACAO if an Origin header is present
				isOriginAllowed := false
				if len(config.AllowedOrigins) == 0 || (len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*") {
					// If no specific origins or wildcard.
					// For credentialed requests, *must* echo origin, not *.
					if config.AllowCredentials {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						log.Printf("[%s] CORS: Setting ACAO to explicit origin: '%s' (wildcard & credentials)", requestID, origin)
					} else {
						w.Header().Set("Access-Control-Allow-Origin", "*")
						log.Printf("[%s] CORS: Setting ACAO to '*' (wildcard, no credentials)", requestID)
					}
					isOriginAllowed = true // Treat as allowed if wildcard is used
				} else {
					// Check against specific allowed origins
					for _, allowedOrigin := range config.AllowedOrigins {
						if allowedOrigin == origin {
							isOriginAllowed = true
							break
						}
					}
					if isOriginAllowed {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						log.Printf("[%s] CORS: Setting ACAO to allowed origin: '%s'", requestID, origin)
					} else {
						// Origin is present but not in AllowedOrigins list
						log.Printf("[%s] CORS: Origin '%s' NOT in AllowedOrigins. ACAO header NOT set.", requestID, origin)
						// For preflight, deny access immediately
						if r.Method == http.MethodOptions {
							log.Printf("[%s] CORS: Preflight OPTIONS from disallowed origin '%s'. Returning 403 Forbidden.", requestID, origin)
							w.WriteHeader(http.StatusForbidden)
							return
						}
					}
				}
			} else {
				// No Origin header means same-origin request, CORS headers typically not needed.
				// However, if AllowCredentials is true and Swagger is requesting from 127.0.0.1:8080 and API is on localhost:8080,
				// they are technically different origins, and the browser *should* send an Origin header.
				// The lack of Origin header might be the actual bug here, if the browser thinks it's same-origin.
				log.Printf("[%s] CORS: No Origin header. Assuming same-origin or non-browser request. ACAO not strictly needed for this type.", requestID)
				// For robustness, if you want ACAO on same-origin for some reason:
				// w.Header().Set("Access-Control-Allow-Origin", r.URL.Scheme + "://" + r.URL.Host)
			}

			// The rest of your CORS middleware logic for Methods, Headers, ExposedHeaders,
			// AllowCredentials, MaxAge, and OPTIONS preflight handling can remain largely the same.
			// Ensure these are always set for preflight responses if isCrossOrigin is true

			// Set AllowedMethods for preflight (if OPTIONS and isCrossOrigin)
			if r.Method == http.MethodOptions && isCrossOrigin {
				if len(config.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods))
					log.Printf("[%s] CORS: Setting ACAM: %s", requestID, joinStrings(config.AllowedMethods))
				} else {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
					log.Printf("[%s] CORS: Setting ACAM to default: GET, POST, PUT, DELETE, PATCH, OPTIONS", requestID)
				}
			}

			// Set AllowedHeaders for preflight (if OPTIONS and isCrossOrigin)
			if r.Method == http.MethodOptions && isCrossOrigin {
				reqHeaders := r.Header.Get("Access-Control-Request-Headers")
				if reqHeaders != "" {
					log.Printf("[%s] CORS: Browser requested headers: %s", requestID, reqHeaders)
					if len(config.AllowedHeaders) > 0 {
						w.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders))
						log.Printf("[%s] CORS: Setting ACAH to configured: %s", requestID, joinStrings(config.AllowedHeaders))
					} else {
						w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
						log.Printf("[%s] CORS: Setting ACAH to requested: %s (no configured headers)", requestID, reqHeaders)
					}
				} else if len(config.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders))
					log.Printf("[%s] CORS: Setting ACAH to configured: %s (no requested headers)", requestID, joinStrings(config.AllowedHeaders))
				} else {
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
					log.Printf("[%s] CORS: Setting ACAH to default: Content-Type, Authorization, X-Requested-With, Accept", requestID)
				}
			}

			// Handle ExposedHeaders (if any)
			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", joinStrings(config.ExposedHeaders))
				log.Printf("[%s] CORS: Setting ACEH: %s", requestID, joinStrings(config.ExposedHeaders))
			} else {
				log.Printf("[%s] CORS: No ExposedHeaders configured.", requestID)
			}

			// Handle AllowCredentials (should always be true if your app needs it, and then ACAO cannot be '*')
			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				log.Printf("[%s] CORS: Setting ACAC to true", requestID)
			} else {
				log.Printf("[%s] CORS: AllowCredentials is false or not set.", requestID)
			}

			// Handle MaxAge for preflight
			if r.Method == http.MethodOptions && isCrossOrigin && config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
				log.Printf("[%s] CORS: Setting ACMA: %d seconds", requestID, config.MaxAge)
			} else if r.Method == http.MethodOptions && isCrossOrigin {
				log.Printf("[%s] CORS: No MaxAge configured for preflight.", requestID)
			}

			// Final check for OPTIONS preflight
			if r.Method == http.MethodOptions {
				log.Printf("[%s] CORS: Preflight OPTIONS request handled. Returning 204 No Content.", requestID)
				w.WriteHeader(http.StatusNoContent)
				return // Crucial: Do not call next.ServeHTTP for OPTIONS preflight
			}

			log.Printf("[%s] CORS: Non-OPTIONS request, passing to next handler.", requestID)
			next.ServeHTTP(w, r)
		})
	}
}

func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			now := time.Now()

			if requests, exists := rl.requests[ip]; exists {
				var validRequests []time.Time
				for _, reqTime := range requests {
					if now.Sub(reqTime) < rl.window {
						validRequests = append(validRequests, reqTime)
					}
				}
				rl.requests[ip] = validRequests
			}

			if len(rl.requests[ip]) >= rl.limit {
				http.Error(w, "Rate limit exceeded",
					http.StatusTooManyRequests)
				return
			}

			rl.requests[ip] = append(rl.requests[ip], now)

			next.ServeHTTP(w, r)
		})
	}
}

func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}
