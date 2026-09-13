CREATE TABLE IF NOT EXISTS authors (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) unique NOT NULL,
    bio VARCHAR(1024)
);

CREATE TABLE IF NOT EXISTS books (
    id BIGSERIAL NOT NULL,
    title VARCHAR(150) unique NOT NULL,
    description VARCHAR(1024),
    author_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT pk_books PRIMARY KEY (id),
    CONSTRAINT fK_book_author FOREIGN KEY (author_id) REFERENCES authors (id) ON DELETE CASCADE,
    CONSTRAINT chk_books_updated_after_created CHECK (updated_at >= created_at)

);

CREATE TABLE IF NOT EXISTS history_entries (
    id BIGSERIAL NOT NULL,
    book_id BIGINT NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    change_type TEXT NOT NULL,
    description TEXT,
    changes JSONB NOT NULL,
    CONSTRAINT pk_history_book PRIMARY KEY (id),
    CONSTRAINT fk_history_entry_book
        FOREIGN KEY (book_id)
        REFERENCES books (id)
);
