package db

import (
	"fmt"

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
