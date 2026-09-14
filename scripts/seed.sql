INSERT INTO authors (name, bio)
VALUES
    (
        'J.R.R. Tolkien',
        'English writer and philologist.'
    ),
    (
        'George Orwell',
        'English novelist and essayist.'
    ),
    (
        'Ursula K. Le Guin',
        'American author known for speculative fiction.'
    )
ON CONFLICT (name) DO NOTHING;