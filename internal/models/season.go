package models

type Season struct {
	ID        int    `db:"id"`
	Season    int    `db:"season"`
	Year      int    `db:"year"`
	Title     string `db:"title"`
	TitleLong string `db:"title_long"`
	Episodes  int    `db:"episodes"`
	LogoURL   string `db:"logo_url"`
}
