CREATE EXTENSION "pgcrypto";

CREATE TABLE users (
    id UUID DEFAULT gen_random_uuid();
);

ALTER TABLE shorten_url
ADD COLUMN user_id UUID;

ALTER TABLE shorten_url
ADD CONSTRAINT fk_user_id
FOREIGN KEY (user_id) REFERENCES users(id)
ON DELETE CASCADE;