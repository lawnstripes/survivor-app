CREATE TABLE IF NOT EXISTS seasons (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	season INTEGER NOT NULL UNIQUE,
	year INTEGER,
	title TEXT,
	title_long TEXT,
	episodes INTEGER,
	logo_url TEXT
);
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	slug TEXT NOT NULL UNIQUE,
	image_url TEXT,
	avatar_url TEXT,
	password_hash TEXT,
	is_admin BOOLEAN DEFAULT 0
);
CREATE TABLE IF NOT EXISTS contestants (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	known_by TEXT, 
	status TEXT DEFAULT 'Active',
	image_url TEXT,
	elimination_episode INTEGER DEFAULT -1,
	added_episode INTEGER DEFAULT 1,
	is_winner BOOLEAN DEFAULT 0,
	owner_id INTEGER REFERENCES users(id),
	season_id INTEGER REFERENCES seasons(id),
	UNIQUE (name, season_id)
);
CREATE TABLE IF NOT EXISTS appdata (
	id INTEGER PRIMARY KEY CHECK (id = 0),
	draft_active BOOLEAN DEFAULT 0, 
	current_season INTEGER REFERENCES seasons(id)
);