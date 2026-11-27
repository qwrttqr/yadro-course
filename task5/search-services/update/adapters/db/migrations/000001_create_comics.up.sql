CREATE TABLE comics
(
    id    integer PRIMARY KEY,
    url   text   NOT NULL,
    words text[] NOT NULL DEFAULT '{}'
);