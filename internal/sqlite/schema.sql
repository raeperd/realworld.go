CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username text NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    bio text,
    image text,
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Trigger that avoids recursion by specifying which columns trigger the update
CREATE TRIGGER update_users_updated_at
    AFTER UPDATE OF username, email, password, bio, image ON users
    FOR EACH ROW
BEGIN
    UPDATE users
    SET updated_at = DATETIME('now')
    WHERE rowid = NEW.rowid;
END;

CREATE TABLE follows (
    follower_id INTEGER NOT NULL,
    followed_id INTEGER NOT NULL,
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, followed_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (followed_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_follows_followed_id ON follows(followed_id);

CREATE TABLE tags (
    id INTEGER PRIMARY KEY,
    name text NOT NULL UNIQUE,
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
);