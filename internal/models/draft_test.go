package models

import "testing"

func TestSnakeDraftEngine(t *testing.T) {
	users := []User{
		{ID: 1, Name: "Alex", Slug: "alex"},
		{ID: 2, Name: "Jen", Slug: "jen"},
		{ID: 3, Name: "John", Slug: "john"},
	}

	board := DraftBoard{
		State: DraftState{
			SeasonID:    1,
			Status:      DraftStatusLive,
			Rounds:      3,
			CurrentPick: 1,
		},
		Order: users,
	}

	if got := board.TotalPicks(); got != 9 {
		t.Fatalf("expected 9 total picks, got %d", got)
	}

	// Expected snake order for 3 users over 3 rounds:
	// Round 1: Alex (1), Jen (2), John (3)
	// Round 2: John (4), Jen (5), Alex (6)
	// Round 3: Alex (7), Jen (8), John (9)
	expectedUserIDs := []int{1, 2, 3, 3, 2, 1, 1, 2, 3}
	expectedRounds := []int{1, 1, 1, 2, 2, 2, 3, 3, 3}

	for i, wantID := range expectedUserIDs {
		pickNum := i + 1
		u, round, _, ok := board.UserForPick(pickNum)
		if !ok {
			t.Fatalf("pick %d should be valid", pickNum)
		}
		if u.ID != wantID {
			t.Errorf("pick %d: expected user ID %d, got %d (%s)", pickNum, wantID, u.ID, u.Name)
		}
		if round != expectedRounds[i] {
			t.Errorf("pick %d: expected round %d, got %d", pickNum, expectedRounds[i], round)
		}
	}

	// Test Pick 4 (Round 2, Pick 1 -> John on clock, Jen on deck)
	board.State.CurrentPick = 4
	onClock, ok := board.OnTheClock()
	if !ok || onClock.ID != 3 {
		t.Errorf("expected John (3) on the clock at pick 4, got %v", onClock)
	}
	onDeck, ok := board.OnDeck()
	if !ok || onDeck.ID != 2 {
		t.Errorf("expected Jen (2) on deck at pick 4, got %v", onDeck)
	}

	// Permissions: John can pick, Alex cannot, Admin can proxy pick
	if !board.CanUserPick(users[2]) {
		t.Errorf("expected John (on the clock) to be allowed to pick")
	}
	if board.CanUserPick(users[0]) {
		t.Errorf("expected Alex (not on the clock) to be disallowed from picking")
	}
	adminUser := User{ID: 1, Name: "Alex", IsAdmin: true}
	if !board.CanUserPick(adminUser) {
		t.Errorf("expected Admin to be allowed to proxy pick")
	}

	// When paused, nobody can pick
	board.State.Status = DraftStatusPaused
	if board.CanUserPick(users[2]) || board.CanUserPick(adminUser) {
		t.Errorf("expected no one to be allowed to pick while paused")
	}
}
