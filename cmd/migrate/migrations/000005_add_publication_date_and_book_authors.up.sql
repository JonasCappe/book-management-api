ALTER TABLE books
    ADD COLUMN publication_date DATE;

UPDATE books
SET publication_date = created_at::date
WHERE publication_date IS NULL;

ALTER TABLE books
    ALTER COLUMN publication_date SET NOT NULL;

CREATE TABLE book_authors (
    book_id BIGINT NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    author_id BIGINT NOT NULL REFERENCES authors (id) ON DELETE RESTRICT,
    PRIMARY KEY (book_id, author_id)
);

INSERT INTO book_authors (book_id, author_id)
SELECT id, author_id
FROM books;

ALTER TABLE books
    DROP CONSTRAINT fk_book_author,
    DROP COLUMN author_id;

CREATE INDEX idx_book_authors_author_id
    ON book_authors (author_id);
