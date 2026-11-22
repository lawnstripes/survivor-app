package models

import (
	"path"
	"strconv"
)

type Contestant struct {
	ID                 int    `db:"id"`
	Name               string `db:"name"`
	KnownBy            string `db:"known_by"`
	Status             string `db:"status"`
	ImageURL           string `db:"image_url"`
	EliminationEpisode int    `db:"elimination_episode"`
	AddedEpisode       int    `db:"added_episode"`
	IsWinner           bool   `db:"is_winner"`
	OwnerID            *int   `db:"owner_id"`
	SeasonID           int    `db:"season_id"`
	DraftedBy          string `db:"drafted_by"`
}

func (c Contestant) TargetID() string {
	return "contestant" + strconv.Itoa(c.ID)
}

func (c Contestant) SmallImageURL() string {
	if c.ImageURL == "" {
		return ""
	}
	return path.Join(path.Dir(c.ImageURL), "sm", path.Base(c.ImageURL))
}
