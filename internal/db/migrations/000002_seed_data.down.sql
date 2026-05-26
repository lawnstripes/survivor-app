-- The order is important to avoid foreign key constraint errors.
DELETE FROM sessions;
DELETE FROM contestants;
DELETE FROM appdata;
DELETE FROM users;
DELETE FROM seasons;
