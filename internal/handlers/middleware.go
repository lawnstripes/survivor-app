package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"survivor-app/internal/models"
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

func (h *Handler) StaticFilesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.RequestURI
		if contentType != "" {
			if strings.HasPrefix(contentType, "/static/css") || strings.HasPrefix(contentType, "/static/js") {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Cache-Control", "public, max-age=86400")
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
