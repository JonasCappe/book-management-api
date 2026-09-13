DROP INDEX IF EXISTS idx_books_active;

ALTER TABLE books
    DROP COLUMN IF EXISTS deleted_at;
