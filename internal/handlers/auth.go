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
	loginComponent := view.Login()

	if r.Header.Get("HX-Request") == "true" {
		loginComponent.Render(r.Context(), w)
		return
	}

	currentUser, _ := r.Context().Value(userKey).(models.User)

	view.Layout("Login", loginComponent, *h.AppData, currentUser, r.URL.Path).Render(r.Context(), w)
}

func (h *Handler) HandleLoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	slug := r.FormValue("slug")
	password := r.FormValue("password")

	user, err := h.Store.GetUserBySlug(slug)
	if err != nil || user.PasswordHash == nil {
		user, err = h.Store.GetUserByEmail(slug)
		if err != nil || user.PasswordHash == nil {
			component := view.LoginWithErrors("Invalid credentials")
			log.Print("user not found")
			component.Render(r.Context(), w)
			return
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	if err != nil {
		// Password does not match
		component := view.LoginWithErrors("Invalid credentials")
		log.Print("password mismatch")
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

func (h *Handler) HandleLoginForgotPasswordGet(w http.ResponseWriter, r *http.Request) {
	forgotComponent := view.ForgotPassword()

	if r.Header.Get("HX-Request") == "true" {
		forgotComponent.Render(r.Context(), w)
		return
	}

	currentUser, _ := r.Context().Value(userKey).(models.User)
	view.Layout("Forgot Password", forgotComponent, *h.AppData, currentUser, r.URL.Path).Render(r.Context(), w)
}

func (h *Handler) HandleLoginForgotPasswordPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	user, _ := h.Store.GetUserByEmail(email)
	if user.ID > 0 {
		token, err := generateToken()
		if err != nil {
			log.Printf("Failed to generate password reset token: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if err := h.Store.UpdateUserResetToken(user.ID, token); err != nil {
			log.Printf("Failed to update user reset token for user %d: %v", user.ID, err)
			// We still show the confirmation to avoid leaking user information.
		}

		// Send the password reset email using the injected email sender.
		if err := h.EmailSender.SendPasswordReset(*user.Email, token); err != nil {
			log.Printf("Failed to send password reset email to %s: %v", *user.Email, err)
			// We don't want to expose this error to the user, but we should log it.
		}
	}

	view.ForgotPasswordForm(true).Render(r.Context(), w)
}

func (h *Handler) HandlePasswordResetTokenGet(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	user, err := h.Store.GetUserByResetToken(token)
	nowUtc := time.Now().UTC()
	// The token is valid if we found a user without an error, and the token has not expired.
	if err == nil && user.ID > 0 && user.ResetTokenExpiresAt != nil && nowUtc.Before(*user.ResetTokenExpiresAt) {
		resetPw := view.ResetPassword(user, "", token)
		view.Layout("Reset Password", resetPw, *h.AppData, models.User{}, r.URL.Path).Render(r.Context(), w)
		return
	}
	log.Printf("token not found or invalid user: %d, %v", user.ID, err)
	http.Redirect(w, r, "/forgot-password", http.StatusSeeOther)
}

func (h *Handler) HandlePasswordResetPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	// First, validate the token is still valid.
	user, err := h.Store.GetUserByResetToken(token)
	nowUtc := time.Now().UTC()
	if err != nil || user.ID == 0 || user.ResetTokenExpiresAt == nil || nowUtc.After(*user.ResetTokenExpiresAt) {
		view.LoginWithErrors("Invalid or expired password reset link.").Render(r.Context(), w)
		return
	}

	if len(password) < 16 {
		resetPw := view.ResetPassword(user, "Password must be at least 16 characters long.", token)
		resetPw.Render(r.Context(), w)
		return
	}
	if password != password2 {
		resetPw := view.ResetPassword(user, "Passwords do not match.", token)
		resetPw.Render(r.Context(), w)
		return
	}

	// Hash the new password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash new password for user %d: %v", user.ID, err)
		resetPw := view.ResetPassword(user, "An internal error occurred. Please try again.", token)
		resetPw.Render(r.Context(), w)
		return
	}

	if err := h.Store.UpdateUserPassword(user.ID, string(hashedPassword)); err != nil {
		log.Printf("Failed to update password for user %d: %v", user.ID, err)
		resetPw := view.ResetPassword(user, "An internal error occurred. Please try again.", token)
		resetPw.Render(r.Context(), w)
		return
	}
	if err := h.Store.ClearUserResetToken(user.ID); err != nil {
		log.Printf("Failed to clear reset token for user %d: %v", user.ID, err)
	}

	// Render a success message which replaces the form.
	view.PasswordResetSuccess().Render(r.Context(), w)
}

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Best effort to delete the session from the database
		_ = h.Store.DeleteSessionByToken(cookie.Value)
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
