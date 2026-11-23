package handlers

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"survivor-app/internal/models"
	"survivor-app/view"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) HandleAdminSeasonsGet(w http.ResponseWriter, r *http.Request) {
	var seasons []models.Season

	seasons, err := h.Store.GetAllSeasons()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adminComponent := view.AdminSeasons(seasons, r.URL.Path)
	if r.Header.Get("HX-Request") == "true" {
		adminComponent.Render(r.Context(), w)
		return
	}

	currentUser, _ := r.Context().Value(userKey).(models.User)
	view.Layout("Admin: Manage Seasons", adminComponent, *h.AppData, currentUser, r.URL.Path).Render(r.Context(), w)
}

func (h *Handler) HandleAdminSeasonCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	seasonNum, _ := strconv.Atoi(r.FormValue("season"))
	year, _ := strconv.Atoi(r.FormValue("year"))
	episodes, _ := strconv.Atoi(r.FormValue("episodes"))
	title := r.FormValue("title")
	titleLong := r.FormValue("title_long")
	logoURL := r.FormValue("logo_url")

	//_, err := h.DB.Exec(`INSERT INTO seasons (season, year, title, title_long, episodes, logo_url) VALUES (?, ?, ?, ?, ?, ?)`, seasonNum, year, title, titleLong, episodes, logoURL)
	err := h.Store.CreateSeason(&models.Season{
		Season:    seasonNum,
		Year:      year,
		Title:     title,
		TitleLong: titleLong,
		LogoURL:   logoURL,
		Episodes:  episodes,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var seasons []models.Season
	seasons, err = h.Store.GetAllSeasons()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// need to refresh app data for season drop down
	newAppData, err := h.Store.GetAppData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.AppData = newAppData

	view.AdminSeasonsTable(seasons).Render(r.Context(), w)
}

func (h *Handler) HandleAdminSeasonGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid season ID", http.StatusBadRequest)
		return
	}

	var season models.Season
	season, err = h.Store.GetSeasonByID(id)
	if err != nil {
		http.Error(w, "Season not found", http.StatusNotFound)
		return
	}

	view.AdminSeasonRow(season).Render(r.Context(), w)
}

func (h *Handler) HandleAdminSeasonEditGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid season ID", http.StatusBadRequest)
		return
	}

	var season models.Season
	season, err = h.Store.GetSeasonByID(id)
	if err != nil {
		http.Error(w, "Season not found", http.StatusNotFound)
		return
	}
	view.AdminSeasonEditForm(season).Render(r.Context(), w)
}

func (h *Handler) HandleAdminSeasonUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid season ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	seasonNum, err := strconv.Atoi(r.FormValue("season"))
	if err != nil {
		http.Error(w, "Invalid season number provided.", http.StatusBadRequest)
		return
	}
	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		http.Error(w, "Invalid year provided.", http.StatusBadRequest)
		return
	}
	// Episodes can often be optional, so defaulting to 0 might be acceptable here.
	// If it's required, it should be checked as well.
	episodes, _ := strconv.Atoi(r.FormValue("episodes"))
	title := r.FormValue("title")
	titleLong := r.FormValue("title_long")
	logoURL := r.FormValue("logo_url")

	err = h.Store.UpdateSeason(&models.Season{
		ID:        id,
		Season:    seasonNum,
		Year:      year,
		Episodes:  episodes,
		Title:     title,
		TitleLong: titleLong,
		LogoURL:   logoURL,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.HandleAdminSeasonGet(w, r)
}

// HandleAdminContestantsGet displays the page for managing contestants of a specific season.
func (h *Handler) HandleAdminContestantsGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid season ID", http.StatusBadRequest)
		return
	}

	var season models.Season
	season, err = h.Store.GetSeasonByID(id)
	if err != nil {
		http.Error(w, "Season not found", http.StatusNotFound)
		return
	}

	var contestants []models.Contestant
	contestants, err = h.Store.GetContestantsBySeasonID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	adminComponent := view.AdminContestants(season, contestants, r.URL.Path)
	if r.Header.Get("HX-Request") == "true" {
		adminComponent.Render(r.Context(), w)
		return
	}

	currentUser, _ := r.Context().Value(userKey).(models.User)
	view.Layout("Admin: Manage Contestants", adminComponent, *h.AppData, currentUser, r.URL.Path).Render(r.Context(), w)
}

// HandleAdminContestantCreate handles the creation of a new contestant for a season.
func (h *Handler) HandleAdminContestantCreate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	seasonID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid season ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form if multipart, otherwise standard URL-encoded form
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}
	}

	name := r.FormValue("name")
	knownBy := r.FormValue("known_by")
	imageURL := r.FormValue("image_url")

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		file, header, err := r.FormFile("image_file")
		// http.ErrMissingFile is okay, it just means no file was uploaded.
		if err != nil && err != http.ErrMissingFile {
			http.Error(w, "Could not get uploaded file", http.StatusBadRequest)
			return
		}

		if file != nil {
			defer file.Close()

			season, err := h.Store.GetSeasonByID(seasonID)
			if err != nil {
				http.Error(w, "Could not find season for upload path", http.StatusInternalServerError)
				return
			}

			// e.g., "contestants/season_49/jake_latimer.jpg"
			objectName := fmt.Sprintf("contestants/season_%d/%s%s",
				season.Season,
				slugify(knownBy),
				filepath.Ext(header.Filename),
			)

			imageURL, err = h.FileStore.Upload(file, objectName)
			if err != nil {
				log.Printf("Failed to upload file: %v", err)
				http.Error(w, "Failed to upload file", http.StatusInternalServerError)
				return
			}
		}
	}

	err = h.Store.CreateContestant(&models.Contestant{
		Name:     name,
		KnownBy:  knownBy,
		ImageURL: imageURL,
		SeasonID: seasonID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var contestants []models.Contestant
	contestants, err = h.Store.GetContestantsBySeasonID(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view.AdminContestantsTable(seasonID, contestants).Render(r.Context(), w)
}

func (h *Handler) HandleAdminContestantGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	contestantID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid contestant ID", http.StatusBadRequest)
		return
	}

	var contestant models.Contestant
	contestant, err = h.Store.GetContestantByID(contestantID)
	if err != nil {
		http.Error(w, "Contestant not found", http.StatusNotFound)
		return
	}

	view.AdminContestantRow(contestant).Render(r.Context(), w)
}

func (h *Handler) HandleAdminContestantEditGet(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	contestantID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid contestant ID", http.StatusBadRequest)
		return
	}

	var contestant models.Contestant
	contestant, err = h.Store.GetContestantByID(contestantID)
	if err != nil {
		http.Error(w, "Contestant not found", http.StatusNotFound)
		return
	}

	var users []models.User
	users, err = h.Store.GetAllUsersNonBench()
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	view.AdminContestantEditForm(contestant, users).Render(r.Context(), w)
}

func (h *Handler) HandleAdminContestantUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid contestant ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	knownBy := r.FormValue("known_by")
	imageURL := r.FormValue("image_url")
	elimEp, _ := strconv.Atoi(r.FormValue("elimination_episode"))
	addedEp, _ := strconv.Atoi(r.FormValue("added_episode"))
	isWinner := r.FormValue("is_winner") == "on"

	ownerID := ptrAtoi(r.FormValue("owner_id"))
	err = h.Store.UpdateContestant(&models.Contestant{
		ID:                 id,
		Name:               name,
		KnownBy:            knownBy,
		ImageURL:           imageURL,
		EliminationEpisode: elimEp,
		AddedEpisode:       addedEp,
		IsWinner:           isWinner,
		OwnerID:            ownerID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.HandleAdminContestantGet(w, r)
}

// HandleAdminContestantBulkCreate handles the bulk creation of contestants from a CSV file.
func (h *Handler) HandleAdminContestantBulkCreate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	seasonID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid season ID", http.StatusBadRequest)
		return
	}

	// Max upload size: 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("contestant_csv")
	if err != nil {
		http.Error(w, "Could not get uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Could not parse CSV file", http.StatusBadRequest)
		return
	}

	var contestantsToCreate []models.Contestant
	// Skip header row (i=0)
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 3 {
			log.Printf("Skipping malformed CSV row: %v", record)
			continue
		}
		contestantsToCreate = append(contestantsToCreate, models.Contestant{
			Name:     record[0],
			KnownBy:  record[1],
			ImageURL: record[2],
		})
	}

	err = h.Store.BulkCreateContestants(contestantsToCreate, seasonID)
	if err != nil {
		log.Printf("Error bulk creating contestants: %s", err.Error())
		http.Error(w, "Error processing CSV file", http.StatusInternalServerError)
		return
	}

	contestants, err := h.Store.GetContestantsBySeasonID(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.AdminContestantsTable(seasonID, contestants).Render(r.Context(), w)
}

func (h *Handler) HandleAdminSettingsGet(w http.ResponseWriter, r *http.Request) {
	var seasons []models.Season
	seasons, err := h.Store.GetAllSeasons()
	if err != nil {
		http.Error(w, "Failed to fetch seasons", http.StatusInternalServerError)
		return
	}
	settingsComponent := view.AdminSettings(*h.AppData, seasons, r.URL.Path)
	if r.Header.Get("HX-Request") == "true" {
		settingsComponent.Render(r.Context(), w)
		return
	}

	currentUser, _ := r.Context().Value(userKey).(models.User)
	view.Layout("Admin: Settings", settingsComponent, *h.AppData, currentUser, r.URL.Path).Render(r.Context(), w)
}

func (h *Handler) HandleAdminSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	seasonID, err := strconv.Atoi(r.FormValue("current_season"))
	if err != nil {
		http.Error(w, "Invalid season ID", http.StatusBadRequest)
		return
	}
	draftActive := r.FormValue("draft_active") == "on"

	err = h.Store.UpdateAppSettings(seasonID, draftActive)
	if err != nil {
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	// Update the in-memory AppData
	newAppData, err := h.Store.GetAppData()
	if err != nil {
		http.Error(w, "Failed to reload app data", http.StatusInternalServerError)
		return
	}
	h.AppData = newAppData

	var seasons []models.Season
	seasons, err = h.Store.GetAllSeasons()
	if err != nil {
		http.Error(w, "Failed to fetch seasons", http.StatusInternalServerError)
		return
	}
	// Re-render the form to show the updated state
	view.AdminSettingsForm(*h.AppData, seasons).Render(r.Context(), w)
}

func (h *Handler) renderAdminDraftPanel(w http.ResponseWriter, r *http.Request) {
	seasonID := h.AppData.CurrentSeason.ID
	board, err := h.Store.GetDraftBoard(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allUsers, err := h.Store.GetAllUsersNonBench()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view.AdminDraftPanel(board, allUsers).Render(r.Context(), w)
}

func (h *Handler) HandleAdminDraftGet(w http.ResponseWriter, r *http.Request) {
	seasonID := h.AppData.CurrentSeason.ID
	board, err := h.Store.GetDraftBoard(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allUsers, err := h.Store.GetAllUsersNonBench()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	draftComponent := view.AdminDraft(board, allUsers, "/admin/draft")
	if r.Header.Get("HX-Request") == "true" {
		draftComponent.Render(r.Context(), w)
		return
	}

	currentUser, _ := r.Context().Value(userKey).(models.User)
	view.Layout("Admin: Manage Draft", draftComponent, *h.AppData, currentUser, r.URL.Path).Render(r.Context(), w)
}

func (h *Handler) HandleAdminDraftRounds(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	rounds, err := strconv.Atoi(r.FormValue("rounds"))
	if err != nil || rounds < 1 || rounds > 20 {
		http.Error(w, "Rounds must be between 1 and 20", http.StatusBadRequest)
		return
	}

	seasonID := h.AppData.CurrentSeason.ID
	state, err := h.Store.GetOrCreateDraftState(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state.Rounds = rounds
	if err := h.Store.UpdateDraftState(state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Broker.Notify("update")
	h.renderAdminDraftPanel(w, r)
}

func (h *Handler) HandleAdminDraftStatus(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	newStatus := r.FormValue("status")
	switch newStatus {
	case models.DraftStatusSetup, models.DraftStatusLive, models.DraftStatusPaused, models.DraftStatusCompleted:
	default:
		http.Error(w, "Invalid draft status", http.StatusBadRequest)
		return
	}

	seasonID := h.AppData.CurrentSeason.ID
	state, err := h.Store.GetOrCreateDraftState(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state.Status = newStatus
	if err := h.Store.UpdateDraftState(state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Keep AppData.IsDraftActive synchronized with live/paused draft state
	draftActive := newStatus == models.DraftStatusLive || newStatus == models.DraftStatusPaused
	if err := h.Store.UpdateAppSettings(seasonID, draftActive); err == nil {
		if newAppData, err := h.Store.GetAppData(); err == nil {
			h.AppData = newAppData
		}
	}

	h.Broker.Notify("update")
	if r.FormValue("source") == "warroom" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	h.renderAdminDraftPanel(w, r)
}

func (h *Handler) HandleAdminDraftOrderMove(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	direction := r.FormValue("direction")

	seasonID := h.AppData.CurrentSeason.ID
	order, err := h.Store.GetDraftOrder(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ids := make([]int, len(order))
	idx := -1
	for i, u := range order {
		ids[i] = u.ID
		if u.ID == userID {
			idx = i
		}
	}

	if idx != -1 {
		if direction == "up" && idx > 0 {
			ids[idx], ids[idx-1] = ids[idx-1], ids[idx]
			_ = h.Store.SaveDraftOrder(seasonID, ids)
		} else if direction == "down" && idx < len(ids)-1 {
			ids[idx], ids[idx+1] = ids[idx+1], ids[idx]
			_ = h.Store.SaveDraftOrder(seasonID, ids)
		}
	}

	h.Broker.Notify("update")
	h.renderAdminDraftPanel(w, r)
}

func (h *Handler) HandleAdminDraftOrderRemove(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	seasonID := h.AppData.CurrentSeason.ID
	order, err := h.Store.GetDraftOrder(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ids := make([]int, 0, len(order))
	for _, u := range order {
		if u.ID != userID {
			ids = append(ids, u.ID)
		}
	}

	if err := h.Store.SaveDraftOrder(seasonID, ids); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Broker.Notify("update")
	h.renderAdminDraftPanel(w, r)
}

func (h *Handler) HandleAdminDraftOrderAdd(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	seasonID := h.AppData.CurrentSeason.ID
	order, err := h.Store.GetDraftOrder(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ids := make([]int, 0, len(order)+1)
	exists := false
	for _, u := range order {
		ids = append(ids, u.ID)
		if u.ID == userID {
			exists = true
		}
	}
	if !exists {
		ids = append(ids, userID)
		if err := h.Store.SaveDraftOrder(seasonID, ids); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.Broker.Notify("update")
	h.renderAdminDraftPanel(w, r)
}

func (h *Handler) HandleAdminDraftOrderRandomize(w http.ResponseWriter, r *http.Request) {
	seasonID := h.AppData.CurrentSeason.ID
	order, err := h.Store.GetDraftOrder(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ids := make([]int, len(order))
	for i, u := range order {
		ids[i] = u.ID
	}

	rand.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})

	if err := h.Store.SaveDraftOrder(seasonID, ids); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Broker.Notify("update")
	h.renderAdminDraftPanel(w, r)
}

func (h *Handler) HandleAdminDraftUndo(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	seasonID := h.AppData.CurrentSeason.ID
	state, err := h.Store.UndoLastDraftPick(seasonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// If undoing reopened a completed draft, re-enable draft_active
	if state.Status == models.DraftStatusLive && !h.AppData.IsDraftActive {
		if err := h.Store.UpdateAppSettings(seasonID, true); err == nil {
			if newAppData, err := h.Store.GetAppData(); err == nil {
				h.AppData = newAppData
			}
		}
	}

	h.Broker.Notify("update")
	if r.FormValue("source") == "warroom" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	h.renderAdminDraftPanel(w, r)
}

func (h *Handler) HandleAdminDraftReset(w http.ResponseWriter, r *http.Request) {
	seasonID := h.AppData.CurrentSeason.ID
	if err := h.Store.ResetDraft(seasonID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.Store.UpdateAppSettings(seasonID, false); err == nil {
		if newAppData, err := h.Store.GetAppData(); err == nil {
			h.AppData = newAppData
		}
	}

	h.Broker.Notify("update")
	h.renderAdminDraftPanel(w, r)
}

func ptrAtoi(s string) *int {
	if s == "" || s == "0" {
		return nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &i
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9_]+`)

// create a clean, URL-friendly string.
func slugify(s string) string {
	return strings.Trim(nonAlphanumericRegex.ReplaceAllString(strings.ToLower(s), "_"), "_")
}
