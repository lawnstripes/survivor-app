package db

import "survivor-app/internal/models"

// Store defines the interface for database operations.
// This allows us to swap out the database implementation (e.g., SQLite, Postgres)
// without changing the business logic in the handlers.
type Store interface {
	// User methods
	GetUserBySlug(slug string) (models.User, error)
	GetUserByID(id int) (models.User, error)
	GetAllUsers() ([]models.User, error)
	GetAllUsersNonBench() ([]models.User, error)

	// Season methods
	GetSeasonByID(id int) (models.Season, error)

	GetAllSeasons() ([]models.Season, error)
	CreateSeason(season *models.Season) error
	UpdateSeason(season *models.Season) error

	// Contestant methods
	GetContestantsBySeasonID(seasonID int) ([]models.Contestant, error)
	GetContestantsByOwnerID(ownerID int) ([]models.Contestant, error)
	GetContestantByID(id int) (models.Contestant, error)
	CreateContestant(c *models.Contestant) error
	UpdateContestant(c *models.Contestant) error
	BulkCreateContestants(contestants []models.Contestant, seasonID int) error

	// AppData methods
	UpdateAppSettings(seasonID int, draftActive bool) error
	GetAppData() (*models.AppData, error)

	// Session methods
	CreateSession(session *models.Session) error
	GetSessionByToken(token string) (*models.Session, error)
	DeleteSessionByToken(token string) error
}
