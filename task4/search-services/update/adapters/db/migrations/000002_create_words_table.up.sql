CREATE TABLE words
(
    id        integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    comics_id INTEGER REFERENCES comics (id) NOT NULL,
    word text NOT NULL
);