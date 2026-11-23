package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"survivor-app/internal/models"

	"github.com/jmoiron/sqlx"
)

type SqliteStore struct {
	db *sqlx.DB
}

func NewSqliteStore(db *sqlx.DB) *SqliteStore {
	return &SqliteStore{db: db}
}

func (s *SqliteStore) GetUserBySlug(slug string) (models.User, error) {
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE slug = ?", slug)
	return user, err
}

func (s *SqliteStore) GetUserByID(id int) (models.User, error) {
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE id = ?", id)
	return user, err
}

func (s *SqliteStore) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE email = ?", email)
	return user, err
}

func (s *SqliteStore) GetUserByResetToken(token string) (models.User, error) {
	var user models.User
	err := s.db.Get(&user, "SELECT * FROM users WHERE reset_token = ?", token)
	return user, err
}

func (s *SqliteStore) UpdateUserResetToken(id int, token string) error {
	expiresAt := time.Now().UTC().Add(20 * time.Minute)
	query := "UPDATE users SET reset_token = ?, reset_token_expires_at = ? WHERE id = ?"
	_, err := s.db.Exec(query, token, expiresAt, id)
	return err
}

func (s *SqliteStore) UpdateUserPassword(id int, passwordHash string) error {
	query := "UPDATE users SET password_hash = ?, reset_token = NULL, reset_token_expires_at = NULL WHERE id = ?"
	_, err := s.db.Exec(query, passwordHash, id)
	return err
}

func (s *SqliteStore) ClearUserResetToken(id int) error {
	query := "UPDATE users SET reset_token = NULL, reset_token_expires_at = NULL WHERE id = ?"
	_, err := s.db.Exec(query, id)
	return err
}

func (s *SqliteStore) GetAllUsers() ([]models.User, error) {
	var users []models.User
	err := s.db.Select(&users, "SELECT * FROM users ORDER BY name")
	return users, err
}

func (s *SqliteStore) GetAllUsersNonBench() ([]models.User, error) {
	var users []models.User
	err := s.db.Select(&users, "SELECT * FROM users WHERE slug != 'bench' ORDER BY name")
	return users, err
}

func (s *SqliteStore) GetSeasonByID(id int) (models.Season, error) {
	var season models.Season
	err := s.db.Get(&season, "SELECT * FROM seasons WHERE id = ?", id)
	return season, err
}

func (s *SqliteStore) GetAllSeasons() ([]models.Season, error) {
	var seasons []models.Season
	err := s.db.Select(&seasons, "SELECT * FROM seasons ORDER BY season DESC")
	return seasons, err
}

func (s *SqliteStore) GetSeasonBySeason(season int) (models.Season, error) {
	query := "SELECT * FROM seasons where season = ?"
	var model models.Season
	err := s.db.Get(&model, query, season)
	return model, err
}

func (s *SqliteStore) CreateSeason(season *models.Season) error {
	query := `INSERT INTO seasons (season, year, title, title_long, episodes, logo_url) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, season.Season, season.Year, season.Title, season.TitleLong, season.Episodes, season.LogoURL)
	return err
}

func (s *SqliteStore) UpdateSeason(season *models.Season) error {
	query := `UPDATE seasons SET season=?, year=?, title=?, title_long=?, episodes=?, logo_url=? WHERE id=?`
	_, err := s.db.Exec(query, season.Season, season.Year, season.Title, season.TitleLong, season.Episodes, season.LogoURL, season.ID)
	return err
}

func (s *SqliteStore) GetContestantsBySeasonID(seasonID int) ([]models.Contestant, error) {
	var contestants []models.Contestant
	query := `
		SELECT c.*, COALESCE(u.name, '') as drafted_by
		FROM contestants c
		LEFT JOIN users u ON c.owner_id = u.id
		WHERE c.season_id = ? ORDER BY c.name
	`
	return contestants, s.db.Select(&contestants, query, seasonID)
}

func (s *SqliteStore) GetContestantsByOwnerID(ownerID int) ([]models.Contestant, error) {
	var contestants []models.Contestant
	return contestants, s.db.Select(&contestants, "SELECT * FROM contestants WHERE owner_id = ?", ownerID)
}

func (s *SqliteStore) GetContestantByID(id int) (models.Contestant, error) {
	var contestant models.Contestant
	query := `
		SELECT c.*, COALESCE(u.name, '') as drafted_by
		FROM contestants c
		LEFT JOIN users u ON c.owner_id = u.id
		WHERE c.id = ?
	`
	err := s.db.Get(&contestant, query, id)
	return contestant, err
}

func (s *SqliteStore) CreateContestant(c *models.Contestant) error {
	query := `INSERT INTO contestants (name, known_by, image_url, season_id) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, c.Name, c.KnownBy, c.ImageURL, c.SeasonID)
	return err
}

func (s *SqliteStore) UpdateContestant(c *models.Contestant) error {
	query := `
		UPDATE contestants 
		SET name=?, known_by=?, image_url=?, elimination_episode=?, added_episode=?, is_winner=?, owner_id=?
		WHERE id=?
	`
	_, err := s.db.Exec(query, c.Name, c.KnownBy, c.ImageURL, c.EliminationEpisode, c.AddedEpisode, c.IsWinner, c.OwnerID, c.ID)
	return err
}

func (s *SqliteStore) SetContestantTeam(ownerID int, seasonID int, knownby string) error {
	query := "UPDATE contestants SET owner_id = ? WHERE season_id = ? AND known_by = ?"
	_, err := s.db.Exec(query, ownerID, seasonID, knownby)
	return err
}

func (s *SqliteStore) BulkCreateContestants(contestants []models.Contestant, seasonID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, c := range contestants {
		_, err := tx.Exec(`INSERT OR IGNORE INTO contestants (name, known_by, image_url, season_id) VALUES (?, ?, ?, ?)`, c.Name, c.KnownBy, c.ImageURL, seasonID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("could not insert contestant %s: %w", c.Name, err)
		}
	}
	return tx.Commit()
}

func (s *SqliteStore) UpdateAppSettings(seasonID int, draftActive bool) error {
	_, err := s.db.Exec("UPDATE appdata SET current_season = ?, draft_active = ? WHERE id = 0", seasonID, draftActive)
	return err
}

func (s *SqliteStore) GetAppData() (*models.AppData, error) {
	var season models.Season
	err := s.db.Get(&season, "SELECT s.* FROM appdata d JOIN seasons s ON d.current_season = s.id WHERE d.id = 0")
	if err != nil {
		return nil, err
	}
	var draftActive bool
	err = s.db.Get(&draftActive, "SELECT draft_active FROM appdata WHERE id = 0")
	if err != nil {
		return nil, err
	}
	var seasons []models.Season
	err = s.db.Select(&seasons, "SELECT * FROM seasons ORDER BY season DESC")
	if err != nil {
		return nil, err
	}
	return &models.AppData{CurrentSeason: season, IsDraftActive: draftActive, Seasons: seasons}, nil
}

func (s *SqliteStore) CreateSession(session *models.Session) error {
	query := `INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, session.Token, session.UserID, session.ExpiresAt)
	return err
}

func (s *SqliteStore) GetSessionByToken(token string) (*models.Session, error) {
	var session models.Session
	err := s.db.Get(&session, "SELECT * FROM sessions WHERE token = ?", token)
	return &session, err
}

func (s *SqliteStore) DeleteSessionByToken(token string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func (s *SqliteStore) GetOrCreateDraftState(seasonID int) (models.DraftState, error) {
	var state models.DraftState
	err := s.db.Get(&state, "SELECT season_id, status, rounds, current_pick FROM drafts WHERE season_id = ?", seasonID)
	if err == nil {
		return state, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return state, err
	}

	state = models.DraftState{
		SeasonID:    seasonID,
		Status:      models.DraftStatusSetup,
		Rounds:      3,
		CurrentPick: 1,
	}
	_, err = s.db.Exec(
		"INSERT INTO drafts (season_id, status, rounds, current_pick) VALUES (?, ?, ?, ?)",
		state.SeasonID, state.Status, state.Rounds, state.CurrentPick,
	)
	if err != nil {
		return state, err
	}

	// Seed default draft_order with all non-bench users if empty
	var count int
	if err := s.db.Get(&count, "SELECT COUNT(*) FROM draft_order WHERE season_id = ?", seasonID); err == nil && count == 0 {
		users, err := s.GetAllUsersNonBench()
		if err == nil && len(users) > 0 {
			ids := make([]int, len(users))
			for i, u := range users {
				ids[i] = u.ID
			}
			_ = s.SaveDraftOrder(seasonID, ids)
		}
	}

	return state, nil
}

func (s *SqliteStore) UpdateDraftState(state models.DraftState) error {
	_, err := s.db.Exec(
		"UPDATE drafts SET status = ?, rounds = ?, current_pick = ? WHERE season_id = ?",
		state.Status, state.Rounds, state.CurrentPick, state.SeasonID,
	)
	return err
}

func (s *SqliteStore) GetDraftOrder(seasonID int) ([]models.User, error) {
	var users []models.User
	query := `
		SELECT u.*
		FROM draft_order do
		JOIN users u ON do.user_id = u.id
		WHERE do.season_id = ?
		ORDER BY do.position ASC
	`
	err := s.db.Select(&users, query, seasonID)
	return users, err
}

func (s *SqliteStore) SaveDraftOrder(seasonID int, userIDs []int) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM draft_order WHERE season_id = ?", seasonID); err != nil {
		return err
	}

	for i, userID := range userIDs {
		if _, err := tx.Exec(
			"INSERT INTO draft_order (season_id, user_id, position) VALUES (?, ?, ?)",
			seasonID, userID, i+1,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SqliteStore) GetDraftPicks(seasonID int) ([]models.DraftPick, error) {
	var picks []models.DraftPick
	query := `
		SELECT
			dp.id,
			dp.season_id,
			dp.pick_number,
			dp.round,
			dp.user_id,
			dp.contestant_id,
			dp.created_at,
			u.name AS user_name,
			u.slug AS user_slug,
			c.name AS contestant_name,
			COALESCE(NULLIF(c.known_by, ''), c.name) AS contestant_known_by,
			COALESCE(c.image_url, '') AS contestant_image_url
		FROM draft_picks dp
		JOIN users u ON dp.user_id = u.id
		JOIN contestants c ON dp.contestant_id = c.id
		WHERE dp.season_id = ?
		ORDER BY dp.pick_number ASC
	`
	err := s.db.Select(&picks, query, seasonID)
	return picks, err
}

func (s *SqliteStore) MakeDraftPick(seasonID, pickNumber, round, userID, contestantID, totalPicks int) (models.DraftState, error) {
	var state models.DraftState
	tx, err := s.db.Beginx()
	if err != nil {
		return state, err
	}
	defer tx.Rollback()

	if err := tx.Get(&state, "SELECT season_id, status, rounds, current_pick FROM drafts WHERE season_id = ?", seasonID); err != nil {
		return state, err
	}
	if state.Status != models.DraftStatusLive {
		return state, fmt.Errorf("draft is not currently live (status: %s)", state.Status)
	}
	if state.CurrentPick != pickNumber {
		return state, fmt.Errorf("pick number mismatch: expected %d, got %d", state.CurrentPick, pickNumber)
	}

	var existingOwner *int
	if err := tx.Get(&existingOwner, "SELECT owner_id FROM contestants WHERE id = ? AND season_id = ?", contestantID, seasonID); err != nil {
		return state, fmt.Errorf("contestant not found for this season: %w", err)
	}
	if existingOwner != nil {
		return state, errors.New("contestant has already been drafted")
	}

	if _, err := tx.Exec(
		"INSERT INTO draft_picks (season_id, pick_number, round, user_id, contestant_id) VALUES (?, ?, ?, ?, ?)",
		seasonID, pickNumber, round, userID, contestantID,
	); err != nil {
		return state, err
	}

	if _, err := tx.Exec(
		"UPDATE contestants SET owner_id = ?, added_episode = 1 WHERE id = ? AND season_id = ?",
		userID, contestantID, seasonID,
	); err != nil {
		return state, err
	}

	state.CurrentPick = pickNumber + 1
	if totalPicks > 0 && pickNumber >= totalPicks {
		state.Status = models.DraftStatusCompleted
	}

	if _, err := tx.Exec(
		"UPDATE drafts SET status = ?, current_pick = ? WHERE season_id = ?",
		state.Status, state.CurrentPick, seasonID,
	); err != nil {
		return state, err
	}

	if err := tx.Commit(); err != nil {
		return state, err
	}
	return state, nil
}

func (s *SqliteStore) UndoLastDraftPick(seasonID int) (models.DraftState, error) {
	var state models.DraftState
	tx, err := s.db.Beginx()
	if err != nil {
		return state, err
	}
	defer tx.Rollback()

	if err := tx.Get(&state, "SELECT season_id, status, rounds, current_pick FROM drafts WHERE season_id = ?", seasonID); err != nil {
		return state, err
	}

	var lastPick models.DraftPick
	err = tx.Get(&lastPick, "SELECT id, season_id, pick_number, round, user_id, contestant_id FROM draft_picks WHERE season_id = ? ORDER BY pick_number DESC LIMIT 1", seasonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return state, errors.New("no picks to undo")
		}
		return state, err
	}

	if _, err := tx.Exec("UPDATE contestants SET owner_id = NULL WHERE id = ? AND season_id = ?", lastPick.ContestantID, seasonID); err != nil {
		return state, err
	}

	if _, err := tx.Exec("DELETE FROM draft_picks WHERE id = ?", lastPick.ID); err != nil {
		return state, err
	}

	state.CurrentPick = lastPick.PickNumber
	if state.Status == models.DraftStatusCompleted {
		state.Status = models.DraftStatusLive
	}

	if _, err := tx.Exec("UPDATE drafts SET status = ?, current_pick = ? WHERE season_id = ?", state.Status, state.CurrentPick, seasonID); err != nil {
		return state, err
	}

	if err := tx.Commit(); err != nil {
		return state, err
	}
	return state, nil
}

func (s *SqliteStore) ResetDraft(seasonID int) error {
	if _, err := s.GetOrCreateDraftState(seasonID); err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM draft_picks WHERE season_id = ?", seasonID); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE contestants SET owner_id = NULL WHERE season_id = ?", seasonID); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE drafts SET status = ?, current_pick = 1 WHERE season_id = ?", models.DraftStatusSetup, seasonID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SqliteStore) GetDraftBoard(seasonID int) (models.DraftBoard, error) {
	var board models.DraftBoard

	season, err := s.GetSeasonByID(seasonID)
	if err != nil {
		return board, err
	}
	board.Season = season

	state, err := s.GetOrCreateDraftState(seasonID)
	if err != nil {
		return board, err
	}
	board.State = state

	order, err := s.GetDraftOrder(seasonID)
	if err != nil {
		return board, err
	}
	board.Order = order

	picks, err := s.GetDraftPicks(seasonID)
	if err != nil {
		return board, err
	}
	board.Picks = picks

	contestants, err := s.GetContestantsBySeasonID(seasonID)
	if err != nil {
		return board, err
	}

	contestantsByID := make(map[int]models.Contestant, len(contestants))
	for _, c := range contestants {
		contestantsByID[c.ID] = c
	}

	board.DraftedByUser = make(map[int][]models.Contestant)
	addedContestantIDs := make(map[int]bool, len(picks))

	// First add drafted contestants in exact pick_number order so round slots (R1, R2, ...) match pick order
	for _, p := range picks {
		if c, ok := contestantsByID[p.ContestantID]; ok && c.OwnerID != nil && *c.OwnerID == p.UserID {
			board.DraftedByUser[p.UserID] = append(board.DraftedByUser[p.UserID], c)
			addedContestantIDs[c.ID] = true
		}
	}

	// Then categorize remaining contestants (either available or manually assigned outside draft_picks)
	for _, c := range contestants {
		if c.OwnerID == nil {
			board.AvailableContestants = append(board.AvailableContestants, c)
		} else if !addedContestantIDs[c.ID] {
			board.DraftedByUser[*c.OwnerID] = append(board.DraftedByUser[*c.OwnerID], c)
		}
	}

	return board, nil
}
