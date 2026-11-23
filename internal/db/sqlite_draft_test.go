package db

import (
	"testing"

	"survivor-app/internal/db/migrations"
	"survivor-app/internal/models"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

func TestSqliteDraftLifecycle(t *testing.T) {
	database, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory db: %v", err)
	}
	defer database.Close()

	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatalf("failed to create migration source driver: %v", err)
	}
	dbDriver, err := sqlite3.WithInstance(database.DB, &sqlite3.Config{})
	if err != nil {
		t.Fatalf("failed to create migration db driver: %v", err)
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", dbDriver)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	store := NewSqliteStore(database)
	appData, err := store.GetAppData()
	if err != nil {
		t.Fatalf("GetAppData failed: %v", err)
	}
	seasonID := appData.CurrentSeason.ID

	// Reset any seed ownership on the current season for a clean 2-round, 2-player draft test
	if err := store.ResetDraft(seasonID); err != nil {
		t.Fatalf("ResetDraft failed: %v", err)
	}

	board, err := store.GetDraftBoard(seasonID)
	if err != nil {
		t.Fatalf("GetDraftBoard failed: %v", err)
	}
	if len(board.Order) < 2 {
		t.Fatalf("expected at least 2 seeded users in draft order, got %d", len(board.Order))
	}

	// Configure a 2-player, 2-round draft (4 total picks)
	twoUsers := []int{board.Order[0].ID, board.Order[1].ID}
	if err := store.SaveDraftOrder(seasonID, twoUsers); err != nil {
		t.Fatalf("SaveDraftOrder failed: %v", err)
	}
	board.State.Rounds = 2
	board.State.Status = models.DraftStatusLive
	if err := store.UpdateDraftState(board.State); err != nil {
		t.Fatalf("UpdateDraftState failed: %v", err)
	}

	board, err = store.GetDraftBoard(seasonID)
	if err != nil {
		t.Fatalf("GetDraftBoard failed: %v", err)
	}
	if board.TotalPicks() != 4 {
		t.Fatalf("expected 4 total picks, got %d", board.TotalPicks())
	}
	initialAvailable := len(board.AvailableContestants)
	if initialAvailable < 4 {
		t.Fatalf("expected at least 4 available contestants, got %d", initialAvailable)
	}

	// Make all 4 snake picks: User0 -> User1 -> User1 -> User0
	for pickNum := 1; pickNum <= 4; pickNum++ {
		board, err = store.GetDraftBoard(seasonID)
		if err != nil {
			t.Fatalf("GetDraftBoard at pick %d failed: %v", pickNum, err)
		}
		onClock, ok := board.OnTheClock()
		if !ok {
			t.Fatalf("expected OnTheClock user at pick %d", pickNum)
		}
		contestant := board.AvailableContestants[0]
		state, err := store.MakeDraftPick(seasonID, pickNum, board.CurrentRound(), onClock.ID, contestant.ID, board.TotalPicks())
		if err != nil {
			t.Fatalf("MakeDraftPick %d failed: %v", pickNum, err)
		}
		if pickNum < 4 && state.Status != models.DraftStatusLive {
			t.Errorf("expected status live after pick %d, got %s", pickNum, state.Status)
		}
		if pickNum == 4 && state.Status != models.DraftStatusCompleted {
			t.Errorf("expected status completed after final pick, got %s", state.Status)
		}
	}

	// Undo the 4th pick and verify draft reopens at pick 4
	undoneState, err := store.UndoLastDraftPick(seasonID)
	if err != nil {
		t.Fatalf("UndoLastDraftPick failed: %v", err)
	}
	if undoneState.CurrentPick != 4 || undoneState.Status != models.DraftStatusLive {
		t.Errorf("expected pick 4 and status live after undo, got pick=%d status=%s", undoneState.CurrentPick, undoneState.Status)
	}
}
