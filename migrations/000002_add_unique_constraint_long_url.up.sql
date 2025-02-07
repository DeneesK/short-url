ALTER TABLE shorten_url
ADD CONSTRAINT long_url_unique_constraint UNIQUE (long_url);