CREATE TABLE IF NOT EXISTS shorten_url (
    id SERIAL PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    long_url TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS alias_idx ON shorten_url (alias);