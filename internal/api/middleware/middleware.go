package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"slices"
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
	MaxAge           int
}

func CORS(config CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(RequestIDKey)
			if requestID == nil {
				requestID = "no-request-id"
			}

			origin := r.Header.Get("Origin")

			isCrossOrigin := (origin != "")

			if isCrossOrigin {
				isOriginAllowed := false
				if len(config.AllowedOrigins) == 0 || (len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*") {
					if config.AllowCredentials {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					} else {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					}
					isOriginAllowed = true
				} else {
					isOriginAllowed = slices.Contains(config.AllowedOrigins, origin)

					if isOriginAllowed {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					} else {
						if r.Method == http.MethodOptions {
							w.WriteHeader(http.StatusForbidden)
							return
						}
					}
				}
			}

			if r.Method == http.MethodOptions && isCrossOrigin {
				if len(config.AllowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowedMethods))
				} else {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				}
			}

			if r.Method == http.MethodOptions && isCrossOrigin {
				reqHeaders := r.Header.Get("Access-Control-Request-Headers")
				if reqHeaders != "" {
					if len(config.AllowedHeaders) > 0 {
						w.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders))
					} else {
						w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
					}
				} else if len(config.AllowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowedHeaders))

				} else {
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept")
				}
			}

			if len(config.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", joinStrings(config.ExposedHeaders))
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions && isCrossOrigin && config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

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
