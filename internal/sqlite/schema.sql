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

CREATE TABLE articles (
    id INTEGER PRIMARY KEY,
    slug text NOT NULL UNIQUE,
    title text NOT NULL,
    description text NOT NULL,
    body text NOT NULL,
    author_id INTEGER NOT NULL,
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TRIGGER update_articles_updated_at
    AFTER UPDATE OF title, description, body ON articles
    FOR EACH ROW
BEGIN
    UPDATE articles
    SET updated_at = DATETIME('now')
    WHERE rowid = NEW.rowid;
END;

CREATE INDEX idx_articles_author_id ON articles(author_id);
CREATE INDEX idx_articles_created_at ON articles(created_at DESC);

CREATE TABLE article_tags (
    article_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (article_id, tag_id),
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE INDEX idx_article_tags_tag_id ON article_tags(tag_id);

CREATE TABLE favorites (
    user_id INTEGER NOT NULL,
    article_id INTEGER NOT NULL,
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, article_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
);

CREATE INDEX idx_favorites_article_id ON favorites(article_id);

CREATE TABLE comments (
    id INTEGER PRIMARY KEY,
    body text NOT NULL,
    article_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TRIGGER update_comments_updated_at
    AFTER UPDATE OF body ON comments
    FOR EACH ROW
BEGIN
    UPDATE comments
    SET updated_at = DATETIME('now')
    WHERE rowid = NEW.rowid;
END;

CREATE INDEX idx_comments_article_id ON comments(article_id);
CREATE INDEX idx_comments_author_id ON comments(author_id);