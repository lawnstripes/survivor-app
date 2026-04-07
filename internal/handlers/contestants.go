package handlers

import (
	"net/http"
	"strconv"

	"survivor-app/view"
)

func (h *Handler) HandleContestantGrid(w http.ResponseWriter, r *http.Request) {
	seasonIDStr := r.URL.Query().Get("seasonID")

	selectedSeasonID := h.AppData.CurrentSeason.ID
	if seasonIDStr != "" {
		id, err := strconv.Atoi(seasonIDStr)
		if err != nil {
			http.Error(w, "Invalid season ID", http.StatusBadRequest)
			return
		}
		selectedSeasonID = id
	}

	contestants, err := h.Store.GetContestantsBySeasonID(selectedSeasonID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	gridComponent := view.ContestantGrid(contestants)
	if r.Header.Get("HX-Request") == "true" {
		gridComponent.Render(r.Context(), w)
		return
	}
}
