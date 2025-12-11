package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zuquanzhi/Chirp/backend/internal/domain"
	"github.com/zuquanzhi/Chirp/backend/internal/service"
)

type contextKey string

const ctxKeyUser contextKey = "user"

func AuthMiddleware(authSvc *service.AuthService, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				http.Error(w, "missing auth", http.StatusUnauthorized)
				return
			}
			var tokenStr string
			_, err := fmt.Sscanf(h, "Bearer %s", &tokenStr)
			if err != nil {
				http.Error(w, "invalid auth header", http.StatusUnauthorized)
				return
			}

			tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, jwt.ErrTokenUnverifiable
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !tok.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := tok.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			sub := claims["sub"]
			var uid int64
			switch v := sub.(type) {
			case float64:
				uid = int64(v)
			case string:
				uid, _ = strconv.ParseInt(v, 10, 64)
			}

			u, err := authSvc.GetUserByID(r.Context(), uid)
			if err != nil || u == nil {
				http.Error(w, "user not found", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUser, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuthMiddleware(authSvc *service.AuthService, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				next.ServeHTTP(w, r)
				return
			}
			var tokenStr string
			_, err := fmt.Sscanf(h, "Bearer %s", &tokenStr)
			if err != nil {
				// Invalid header format, ignore and proceed as anonymous
				next.ServeHTTP(w, r)
				return
			}

			tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, jwt.ErrTokenUnverifiable
				}
				return []byte(jwtSecret), nil
			})

			// If token is invalid, just proceed as anonymous
			if err != nil || !tok.Valid {
				next.ServeHTTP(w, r)
				return
			}

			claims, ok := tok.Claims.(jwt.MapClaims)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			sub := claims["sub"]
			var userID int64
			switch v := sub.(type) {
			case float64:
				userID = int64(v)
			case string:
				userID, _ = strconv.ParseInt(v, 10, 64)
			}

			u, err := authSvc.GetUserByID(r.Context(), userID)
			if err != nil || u == nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUser, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) *domain.User {
	u, _ := ctx.Value(ctxKeyUser).(*domain.User)
	return u
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := GetUserFromContext(r.Context())
		if u == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if u.Role != domain.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware records basic request info and response status.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf("http request method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, rw.status, duration)
	})
}

// RecoverMiddleware prevents panics from crashing the server and logs the stack.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered path=%s err=%v stack=%s", r.URL.Path, rec, debug.Stack())
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
