ALTER TABLE books
    ADD COLUMN author_id BIGINT;

UPDATE books
SET author_id = (
    SELECT MIN(book_authors.author_id)
    FROM book_authors
    WHERE book_authors.book_id = books.id
);

ALTER TABLE books
    ALTER COLUMN author_id SET NOT NULL,
    ADD CONSTRAINT fk_book_author
        FOREIGN KEY (author_id) REFERENCES authors (id) ON DELETE RESTRICT;

DROP TABLE book_authors;

ALTER TABLE books
    DROP COLUMN publication_date;
