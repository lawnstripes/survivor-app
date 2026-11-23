package models

import (
	"path"
	"time"
)

const (
	DraftStatusSetup     = "setup"
	DraftStatusLive      = "live"
	DraftStatusPaused    = "paused"
	DraftStatusCompleted = "completed"
)

type DraftState struct {
	SeasonID    int    `db:"season_id"`
	Status      string `db:"status"`
	Rounds      int    `db:"rounds"`
	CurrentPick int    `db:"current_pick"`
}

type DraftOrderEntry struct {
	SeasonID int `db:"season_id"`
	UserID   int `db:"user_id"`
	Position int `db:"position"`
}

type DraftPick struct {
	ID                 int       `db:"id"`
	SeasonID           int       `db:"season_id"`
	PickNumber         int       `db:"pick_number"`
	Round              int       `db:"round"`
	UserID             int       `db:"user_id"`
	ContestantID       int       `db:"contestant_id"`
	CreatedAt          time.Time `db:"created_at"`
	UserName           string    `db:"user_name"`
	UserSlug           string    `db:"user_slug"`
	ContestantName     string    `db:"contestant_name"`
	ContestantKnownBy  string    `db:"contestant_known_by"`
	ContestantImageURL string    `db:"contestant_image_url"`
}

func (p DraftPick) ContestantSmallImageURL() string {
	if p.ContestantImageURL == "" {
		return ""
	}
	return path.Join(path.Dir(p.ContestantImageURL), "sm", path.Base(p.ContestantImageURL))
}

type DraftSlot struct {
	PickNumber  int
	Round       int
	PickInRound int
	User        User
	Pick        *DraftPick
}

type DraftBoard struct {
	Season               Season
	State                DraftState
	Order                []User
	Picks                []DraftPick
	AvailableContestants []Contestant
	DraftedByUser        map[int][]Contestant
}

// TotalPicks returns the total number of picks across all configured rounds.
func (b DraftBoard) TotalPicks() int {
	return len(b.Order) * b.State.Rounds
}

// CurrentRound returns the 1-indexed round number for the current pick.
func (b DraftBoard) CurrentRound() int {
	n := len(b.Order)
	if n == 0 || b.State.CurrentPick <= 0 {
		return 1
	}
	total := b.TotalPicks()
	pick := b.State.CurrentPick
	if pick > total && total > 0 {
		pick = total
	}
	return (pick-1)/n + 1
}

// UserForPick calculates which user owns a given 1-indexed pickNumber in a snake draft.
// Odd rounds go 0..N-1; even rounds go N-1..0.
func (b DraftBoard) UserForPick(pickNumber int) (user User, round int, pickInRound int, ok bool) {
	n := len(b.Order)
	if n == 0 || pickNumber < 1 || pickNumber > b.TotalPicks() {
		return User{}, 0, 0, false
	}

	round = (pickNumber-1)/n + 1
	pickInRound = (pickNumber-1)%n + 1

	var idx int
	if round%2 == 1 {
		idx = pickInRound - 1
	} else {
		idx = n - pickInRound
	}

	return b.Order[idx], round, pickInRound, true
}

// OnTheClock returns the user currently on the clock if the draft has picks remaining.
func (b DraftBoard) OnTheClock() (User, bool) {
	if b.State.Status == DraftStatusCompleted {
		return User{}, false
	}
	u, _, _, ok := b.UserForPick(b.State.CurrentPick)
	return u, ok
}

// OnDeck returns the user who picks immediately after the current pick.
func (b DraftBoard) OnDeck() (User, bool) {
	if b.State.Status == DraftStatusCompleted {
		return User{}, false
	}
	u, _, _, ok := b.UserForPick(b.State.CurrentPick + 1)
	return u, ok
}

// CanUserPick returns true if the draft is currently live and the given user is either
// the player on the clock or an admin (proxy picking).
func (b DraftBoard) CanUserPick(currentUser User) bool {
	if b.State.Status != DraftStatusLive || currentUser.ID == 0 {
		return false
	}
	onClock, ok := b.OnTheClock()
	if !ok {
		return false
	}
	return currentUser.IsAdmin || currentUser.ID == onClock.ID
}

// Schedule builds the complete list of pick slots across all rounds, attaching any completed picks.
func (b DraftBoard) Schedule() []DraftSlot {
	total := b.TotalPicks()
	if total <= 0 {
		return nil
	}

	picksByNumber := make(map[int]*DraftPick, len(b.Picks))
	for i := range b.Picks {
		p := &b.Picks[i]
		picksByNumber[p.PickNumber] = p
	}

	slots := make([]DraftSlot, 0, total)
	for pickNum := 1; pickNum <= total; pickNum++ {
		u, round, pickInRound, ok := b.UserForPick(pickNum)
		if !ok {
			continue
		}
		slots = append(slots, DraftSlot{
			PickNumber:  pickNum,
			Round:       round,
			PickInRound: pickInRound,
			User:        u,
			Pick:        picksByNumber[pickNum],
		})
	}
	return slots
}
