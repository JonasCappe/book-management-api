ALTER TABLE books
    ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_books_active
    ON books (id)
    WHERE deleted_at IS NULL;
