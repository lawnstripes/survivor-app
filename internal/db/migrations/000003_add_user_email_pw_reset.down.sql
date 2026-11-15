PRAGMA foreign_keys=off;

CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    image_url TEXT,
    avatar_url TEXT,
    password_hash TEXT,
    is_admin BOOLEAN DEFAULT 0
);

INSERT INTO users_new (id, name, slug, image_url, avatar_url, password_hash, is_admin)
SELECT id, name, slug, image_url, avatar_url, password_hash, is_admin FROM users;

DROP TABLE users;

ALTER TABLE users_new RENAME TO users;

PRAGMA foreign_keys=on;
