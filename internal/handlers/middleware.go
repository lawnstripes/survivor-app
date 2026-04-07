package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"survivor-app/internal/models"
)

const (
	csrfHeaderName = "X-CSRF-Token"
	csrfCookieName = "csrf_token"
)

func (h *Handler) UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			// No session cookie, continue as a guest
			next.ServeHTTP(w, r)
			return
		}

		session, err := h.Store.GetSessionByToken(cookie.Value)
		if err != nil {
			// Invalid session token, clear the cookie and continue as guest
			clearSessionCookie(w, r)
			next.ServeHTTP(w, r)
			return
		}

		if time.Now().After(session.ExpiresAt) {
			// Session expired, delete it, clear the cookie, and continue as guest
			h.Store.DeleteSessionByToken(session.Token)
			clearSessionCookie(w, r)
			next.ServeHTTP(w, r)
			return
		}

		user, err := h.Store.GetUserByID(session.UserID)
		if err != nil {
			// User associated with session not found, something is wrong.
			// Clear everything and continue as guest.
			h.Store.DeleteSessionByToken(session.Token)
			clearSessionCookie(w, r)
			next.ServeHTTP(w, r)
			return
		}

		// Valid session, add user to context
		ctx := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// On unsafe methods (POST, PUT, DELETE), we verify the token.
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			headerToken := r.Header.Get(csrfHeaderName)
			cookie, err := r.Cookie(csrfCookieName)

			if err != nil || cookie.Value == "" || headerToken != cookie.Value {
				log.Printf("CSRF Error: token mismatch or cookie not found.")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		// For all requests, we ensure a token is available for the next request
		// and pass it to the context for rendering in templates.
		var token string
		cookie, err := r.Cookie(csrfCookieName)
		if err != nil || cookie.Value == "" {
			// No valid cookie, generate a new token.
			newToken, err := generateToken()
			if err != nil {
				log.Printf("Failed to generate CSRF token: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			token = newToken

			http.SetCookie(w, &http.Cookie{
				Name:     csrfCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true, // Important for security!
				Secure:   r.TLS != nil,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   86400, // 24 hours
			})
		} else {
			token = cookie.Value
		}

		// Add token to context for views to use.
		ctx := context.WithValue(r.Context(), csrfTokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := r.Context().Value(userKey).(models.User)

		if !ok || !currentUser.IsAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Expire immediately
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
}
