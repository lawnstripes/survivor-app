package handlers

import (
	"log"
	"net/http"
	"time"

	"survivor-app/internal/models"
	"survivor-app/view"

	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) HandleLoginGet(w http.ResponseWriter, r *http.Request) {
	csrfToken := r.Context().Value(csrfTokenKey).(string)
	loginComponent := view.Login(csrfToken)

	if r.Header.Get("HX-Request") == "true" {
		loginComponent.Render(r.Context(), w)
		return
	}

	currentUser, _ := r.Context().Value(userKey).(models.User)

	view.Layout("Login", loginComponent, *h.AppData, currentUser, r.URL.Path).Render(r.Context(), w)
}

func (h *Handler) HandleLoginPost(w http.ResponseWriter, r *http.Request) {
	log.Print("handle login post")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	slug := r.FormValue("slug")
	password := r.FormValue("password")
	csrfToken := r.Context().Value(csrfTokenKey).(string)

	user, err := h.Store.GetUserBySlug(slug)
	if err != nil || user.PasswordHash == nil {
		// User not found or has no password, render login with an error.
		// It's better to be vague in error messages for security.

		component := view.LoginWithErrors(csrfToken, "Invalid credentials")
		component.Render(r.Context(), w)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	if err != nil {
		// Password does not match
		component := view.LoginWithErrors(csrfToken, "Invalid credentials")
		component.Render(r.Context(), w)
		return
	}

	// Create a new session
	token, err := generateToken()
	if err != nil {
		log.Printf("Failed to generate session token: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session := &models.Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := h.Store.CreateSession(session); err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400, // 24 hours
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})

	// Redirect to the home page on successful login
	w.Header().Set("HX-Redirect", "/")
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Best effort to delete the session from the database
		h.Store.DeleteSessionByToken(cookie.Value)
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Expire immediately
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
