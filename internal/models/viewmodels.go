package models

type Draft struct {
	Season      Season
	Contestants []Contestant
	Teams       map[User][]string // key: user value: contestant.knownby
}

type AppData struct {
	CurrentSeason Season
	IsDraftActive bool `db:"draft_active"`
	Seasons       []Season
}

type TimelineRow struct {
	User          User
	ActivePlayers []Contestant
	Eliminations  map[int][]Contestant // Key = Episode Number. Value = Slice of contestants eliminated that episode
	Contestants   []Contestant         // used for the table view
}

type UserProfile struct {
	User     User
	Seasons  []UserSeasonRecord
	BackPath string
}

type UserSeasonRecord struct {
	Season      Season
	Contestants []Contestant
	HasWinner   bool
}
