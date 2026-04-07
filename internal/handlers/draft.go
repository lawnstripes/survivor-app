package handlers

import (
	"net/http"
	"strconv"

	"survivor-app/internal/models"
	"survivor-app/view"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) HandleDraftRoom(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(userKey).(models.User)

	if !ok || currentUser.ID == 0 {
		// Not authorized! Redirect to login.
		w.Header().Set("HX-Redirect", "/login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	/* The rest of this needs to be implemented...
	1. Handle actual user log
	2. Add a draft season concept to the app to pick the correct season
	3. Deal with logic for snake style draft
	4. ... a lot
	*/
	_, err := h.Store.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	contestants, err := h.Store.GetContestantsBySeasonID(h.AppData.CurrentSeason.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	draftComponent := view.DraftRoom(contestants)

	view.Layout("Draft Room", draftComponent, *h.AppData, currentUser, r.URL.Path).Render(r.Context(), w)
}

func (h *Handler) HandleDraftContestant(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userKey).(models.User)
	if !ok {
		http.Error(w, "You must select a user first!", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	// Get the contestant first to ensure they are not already drafted
	c, err := h.Store.GetContestantByID(id)
	if err != nil {
		http.Error(w, "Contestant not found", http.StatusNotFound)
		return
	}

	// Check if already owned
	if c.OwnerID != nil {
		http.Error(w, "Contestant already drafted", http.StatusConflict)
		return
	}

	// Assign owner and update
	c.OwnerID = &user.ID
	err = h.Store.UpdateContestant(&c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	h.Broker.Notify("update")

	// Refetch to get the joined drafted_by name for the card
	c, err = h.Store.GetContestantByID(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = view.ContestantCard(c).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
