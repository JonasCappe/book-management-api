CREATE TABLE IF NOT EXISTS history_entries (
    id BIGSERIAL PRIMARY KEY,
    book_id BIGINT NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    change_type TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    changes JSONB NOT NULL,
    CONSTRAINT chk_history_change_type
        CHECK (change_type IN ('created', 'updated', 'deleted')),
    CONSTRAINT fk_history_entry_book
        FOREIGN KEY (book_id)
        REFERENCES books (id)
);

CREATE INDEX IF NOT EXISTS idx_history_entries_book_changed_at
    ON history_entries (book_id, changed_at DESC);
