package db

import "survivor-app/internal/models"

// Store defines the interface for database operations.
// This allows us to swap out the database implementation (e.g., SQLite, Postgres)
// without changing the business logic in the handlers.
type Store interface {
	// User methods
	GetUserBySlug(slug string) (models.User, error)
	GetUserByID(id int) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
	GetAllUsers() ([]models.User, error)
	GetAllUsersNonBench() ([]models.User, error)
	GetUserByResetToken(token string) (models.User, error)

	UpdateUserResetToken(id int, token string) error
	UpdateUserPassword(id int, passwordHash string) error
	ClearUserResetToken(id int) error

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

	// Draft methods
	GetOrCreateDraftState(seasonID int) (models.DraftState, error)
	UpdateDraftState(state models.DraftState) error
	GetDraftOrder(seasonID int) ([]models.User, error)
	SaveDraftOrder(seasonID int, userIDs []int) error
	GetDraftPicks(seasonID int) ([]models.DraftPick, error)
	MakeDraftPick(seasonID, pickNumber, round, userID, contestantID, totalPicks int) (models.DraftState, error)
	UndoLastDraftPick(seasonID int) (models.DraftState, error)
	ResetDraft(seasonID int) error
	GetDraftBoard(seasonID int) (models.DraftBoard, error)
}
