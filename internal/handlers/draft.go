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
		w.Header().Set("HX-Redirect", "/login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	board, err := h.Store.GetDraftBoard(h.AppData.CurrentSeason.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	draftComponent := view.DraftRoom(board, currentUser)
	if r.Header.Get("HX-Request") == "true" {
		draftComponent.Render(r.Context(), w)
		return
	}

	view.Layout("Live War Room", draftComponent, *h.AppData, currentUser, r.URL.Path).Render(r.Context(), w)
}

func (h *Handler) HandleDraftBoard(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(userKey).(models.User)
	if !ok || currentUser.ID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	board, err := h.Store.GetDraftBoard(h.AppData.CurrentSeason.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.DraftBoardView(board, currentUser).Render(r.Context(), w)
}

func (h *Handler) HandleDraftContestant(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := r.Context().Value(userKey).(models.User)
	if !ok || currentUser.ID == 0 {
		http.Error(w, "You must be logged in to draft", http.StatusUnauthorized)
		return
	}

	contestantID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid contestant ID", http.StatusBadRequest)
		return
	}

	seasonID := h.AppData.CurrentSeason.ID
	board, err := h.Store.GetDraftBoard(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !board.CanUserPick(currentUser) {
		view.DraftBoardView(board, currentUser).Render(r.Context(), w)
		return
	}

	onClockUser, ok := board.OnTheClock()
	if !ok {
		view.DraftBoardView(board, currentUser).Render(r.Context(), w)
		return
	}

	updatedState, err := h.Store.MakeDraftPick(
		seasonID,
		board.State.CurrentPick,
		board.CurrentRound(),
		onClockUser.ID,
		contestantID,
		board.TotalPicks(),
	)
	if err != nil {
		// Re-render current board state so HTMX 4 doesn't swap a plain-text error into #draft-board-container
		if freshBoard, loadErr := h.Store.GetDraftBoard(seasonID); loadErr == nil {
			view.DraftBoardView(freshBoard, currentUser).Render(r.Context(), w)
			return
		}
		w.Header().Set("HX-Reswap", "none")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// If the final pick completed the draft, turn off IsDraftActive automatically
	if updatedState.Status == models.DraftStatusCompleted && h.AppData.IsDraftActive {
		if err := h.Store.UpdateAppSettings(seasonID, false); err == nil {
			if newAppData, err := h.Store.GetAppData(); err == nil {
				h.AppData = newAppData
			}
		}
	}

	// Broadcast to all connected SSE clients so every browser updates immediately
	h.Broker.Notify("update")

	board, err = h.Store.GetDraftBoard(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.DraftBoardView(board, currentUser).Render(r.Context(), w)
}
