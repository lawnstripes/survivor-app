package handlers

import (
	"net/http"
	"sort"
	"strconv"

	"survivor-app/internal/db"
	"survivor-app/internal/filestore"
	"survivor-app/internal/models"
	"survivor-app/internal/sse"
	"survivor-app/view"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	Store     db.Store
	Broker    *sse.Broker
	AppData   *models.AppData
	FileStore filestore.FileStore
}

type contextKey string

const (
	userKey      contextKey = "user"
	csrfTokenKey contextKey = "csrfToken"
)

func NewHandler(store db.Store, broker *sse.Broker, appdata *models.AppData, fstore filestore.FileStore) *Handler {
	return &Handler{Store: store, Broker: broker, AppData: appdata, FileStore: fstore}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// CSRF middleware must come before the router and any handlers that need protection
	r.Use(h.CSRFMiddleware)
	r.Use(h.UserMiddleware)

	// static assets
	r.Handle("/static/*", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))

	// routes
	r.Get("/", h.HandleHome)
	r.Get("/user/{slug}", h.HandleUserProfile)

	r.Get("/login", h.HandleLoginGet)
	r.Post("/login", h.HandleLoginPost)
	r.Get("/logout", h.HandleLogout)

	r.Get("/draft", h.HandleDraftRoom)
	r.Post("/draft/{id}", h.HandleDraftContestant)

	r.Get("/events", h.Broker.Stream)
	r.Get("/contestant-grid", h.HandleContestantGrid)

	// Admin routes
	r.Group(func(r chi.Router) {
		r.Use(h.AdminOnlyMiddleware)
		r.Get("/admin/seasons", h.HandleAdminSeasonsGet)
		r.Post("/admin/seasons", h.HandleAdminSeasonCreate)
		r.Get("/admin/seasons/{id}", h.HandleAdminSeasonGet)
		r.Get("/admin/seasons/{id}/edit", h.HandleAdminSeasonEditGet)
		r.Put("/admin/seasons/{id}", h.HandleAdminSeasonUpdate)
		r.Get("/admin/seasons/{id}/contestants", h.HandleAdminContestantsGet)
		r.Post("/admin/seasons/{id}/contestants", h.HandleAdminContestantCreate)
		r.Post("/admin/seasons/{id}/contestants/bulk", h.HandleAdminContestantBulkCreate)
		r.Get("/admin/contestants/{id}", h.HandleAdminContestantGet)
		r.Get("/admin/contestants/{id}/edit", h.HandleAdminContestantEditGet)
		r.Put("/admin/contestants/{id}", h.HandleAdminContestantUpdate)
		r.Get("/admin/settings", h.HandleAdminSettingsGet)
		r.Put("/admin/settings", h.HandleAdminSettingsUpdate)
	})

	return r
}

func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	pageData := h.pageDataForRequest(r)
	selectedSeason := pageData.CurrentSeason

	contestants, err := h.Store.GetContestantsBySeasonID(selectedSeason.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	users, err := h.Store.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	displayMode := r.URL.Query().Get("display")
	if displayMode != "text" {
		displayMode = "avatar"
	}

	timelineRows := BuildTimelineData(users, contestants)
	indexComponent := view.SeasonView(selectedSeason, timelineRows, displayMode)

	if r.Header.Get("HX-Request") == "true" {
		indexComponent.Render(r.Context(), w)
		return
	}

	currentUser, _ := r.Context().Value(userKey).(models.User)

	view.Layout("Survivor Standings", indexComponent, pageData, currentUser, r.URL.Path).Render(r.Context(), w)
}

func (h *Handler) HandleUserProfile(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	seasonIDStr := r.URL.Query().Get("season_id")

	user, err := h.Store.GetUserBySlug(slug)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	allSeasons, err := h.Store.GetAllSeasons()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	contestants, err := h.Store.GetContestantsByOwnerID(user.ID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	backPath := "/"
	if seasonIDStr != "" {
		backPath = "/?season_id=" + seasonIDStr
	}

	profile := models.UserProfile{User: user, Seasons: make([]models.UserSeasonRecord, 0), BackPath: backPath}

	contestantsBySeason := make(map[int][]models.Contestant)
	for _, c := range contestants {
		contestantsBySeason[c.SeasonID] = append(contestantsBySeason[c.SeasonID], c)
	}

	for _, s := range allSeasons {
		if seasonContestants, ok := contestantsBySeason[s.ID]; ok {
			hasWinner := false
			for _, c := range seasonContestants {
				if c.IsWinner {
					hasWinner = true
					break
				}
			}
			record := models.UserSeasonRecord{
				Season:      s,
				Contestants: seasonContestants,
				HasWinner:   hasWinner,
			}
			profile.Seasons = append(profile.Seasons, record)
		}
	}

	// Sort seasons by most recent first for a better user experience.
	sort.Slice(profile.Seasons, func(i, j int) bool {
		// Sort by year descending
		if profile.Seasons[i].Season.Year != profile.Seasons[j].Season.Year {
			return profile.Seasons[i].Season.Year > profile.Seasons[j].Season.Year
		}
		// If years are the same, sort by season number descending
		return profile.Seasons[i].Season.Season > profile.Seasons[j].Season.Season
	})

	profileComponent := view.UserProfile(profile)
	if r.Header.Get("HX-Request") == "true" {
		profileComponent.Render(r.Context(), w)
		return
	}

	currentUser, _ := r.Context().Value(userKey).(models.User)
	pageData := h.pageDataForRequest(r)
	view.Layout(user.Name+"'s Profile", profileComponent, pageData, currentUser, r.URL.Path).Render(r.Context(), w)
}

func (h *Handler) pageDataForRequest(r *http.Request) models.AppData {
	pageData := *h.AppData
	seasonIDStr := r.URL.Query().Get("season_id")

	if seasonIDStr != "" {
		seasonID, err := strconv.Atoi(seasonIDStr)
		if err == nil { // Silently ignore non-integer season_id
			for _, s := range pageData.Seasons {
				if s.ID == seasonID {
					pageData.CurrentSeason = s
					break
				}
			}
		}
	}
	return pageData
}

func BuildTimelineData(users []models.User, contestants []models.Contestant) []models.TimelineRow {
	timeline := make([]models.TimelineRow, 0, len(users)+1)

	benchRow := models.TimelineRow{
		User:          models.User{Name: "Sit-Out Bench", Slug: "bench"},
		ActivePlayers: []models.Contestant{},
		Eliminations:  make(map[int][]models.Contestant),
		Contestants:   []models.Contestant{},
	}

	for _, user := range users {

		row := models.TimelineRow{
			User:          user,
			ActivePlayers: []models.Contestant{},
			Eliminations:  make(map[int][]models.Contestant),
			Contestants:   []models.Contestant{},
		}

		for _, c := range contestants {
			if c.OwnerID != nil && *c.OwnerID == user.ID {
				if c.EliminationEpisode > 0 {
					row.Eliminations[c.EliminationEpisode] = append(row.Eliminations[c.EliminationEpisode], c)
				} else {
					row.ActivePlayers = append(row.ActivePlayers, c)
				}
				row.Contestants = append(row.Contestants, c)
			}
		}

		// Only append if they actually drafted players this season
		if len(row.ActivePlayers) > 0 || len(row.Eliminations) > 0 || len(row.Contestants) > 0 {
			timeline = append(timeline, row)
		}
	}

	// Populate the Bench row with unowned players, or players who started on the bench
	for _, c := range contestants {
		if c.OwnerID == nil || c.AddedEpisode > 1 {
			benchRow.Contestants = append(benchRow.Contestants, c)
			if c.EliminationEpisode > 0 && (c.OwnerID == nil || c.EliminationEpisode < c.AddedEpisode) {
				benchRow.Eliminations[c.EliminationEpisode] = append(benchRow.Eliminations[c.EliminationEpisode], c)
			} else if c.OwnerID == nil {
				benchRow.ActivePlayers = append(benchRow.ActivePlayers, c)
			}
		}
	}
	if len(benchRow.Contestants) > 0 {
		timeline = append(timeline, benchRow)
	}

	for k := range timeline {
		sort.Slice(timeline[k].Contestants, func(i, j int) bool {
			epI := timeline[k].Contestants[i].EliminationEpisode
			epJ := timeline[k].Contestants[j].EliminationEpisode
			// normalize active players
			if epI <= 0 {
				epI = -1
			}
			if epJ <= 0 {
				epJ = -1
			}
			// is I eliminated after J ?
			if epI != epJ {
				return epI < epJ
			}
			// same elimination episode, sort by name
			return timeline[k].Contestants[i].KnownBy < timeline[k].Contestants[j].KnownBy
		})
	}

	return timeline
}
