ALTER TABLE users ADD COLUMN email TEXT;
ALTER TABLE users ADD COLUMN reset_token TEXT;
ALTER TABLE users ADD COLUMN reset_token_expires_at DATETIME;
