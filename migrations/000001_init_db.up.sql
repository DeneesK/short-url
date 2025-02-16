CREATE TABLE shorten_url (
    id SERIAL PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    long_url TEXT NOT NULL
);
CREATE INDEX alias_idx ON shorten_url (alias);