-- +goose Up
CREATE TABLE publishers (
    id uuid PRIMARY KEY,
    name text NOT NULL CHECK (name = btrim(name) AND name <> ''),
    CONSTRAINT publishers_name_unique UNIQUE (name)
);
CREATE UNIQUE INDEX publishers_normalized_name_unique ON publishers (lower(btrim(name)));

CREATE TABLE books (
    id uuid PRIMARY KEY,
    author text NOT NULL CHECK (author = btrim(author) AND author <> ''),
    name text NOT NULL CHECK (name = btrim(name) AND name <> ''),
    year integer NOT NULL CHECK (year BETWEEN 1 AND 9999),
    price bigint NOT NULL CHECK (price >= 0),
    publisher_id uuid NOT NULL REFERENCES publishers(id) ON DELETE RESTRICT,
    publication_year integer NOT NULL CHECK (publication_year BETWEEN year AND 9999),
    quantity bigint NOT NULL CHECK (quantity >= 0)
);
CREATE UNIQUE INDEX books_content_unique ON books
    (lower(btrim(author)), lower(btrim(name)), year, publisher_id, publication_year);

-- +goose Down
DROP TABLE books;
DROP TABLE publishers;
