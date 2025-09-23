CREATE TABLE IF NOT EXISTS counters (
    id integer PRIMARY KEY,
    value bigint NOT NULL
);

INSERT INTO counters (id, value)
VALUES (1, 0)
ON CONFLICT (id) DO NOTHING;


