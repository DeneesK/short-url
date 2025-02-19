DROP TABLE users;

ALTER TABLE shorten_url
DROP CONSTRAINT fk_user_id;

ALTER TABLE shorten_url
DROP COLUMN user_id;