--up
CREATE TABLE files
(
    id         UUID      NOT NULL DEFAULT GEN_RANDOM_UUID(),
    size       INT       NOT NULL DEFAULT 0,
    metadata   JSONB              DEFAULT NULL,
    chunk_ids  UUID[]             DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id)
);

CREATE TABLE chunks
(
    id         UUID      NOT NULL DEFAULT GEN_RANDOM_UUID(),
    file_id     UUID      NOT NULL,
    bytes      BYTEA     NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (file_id) REFERENCES files ON DELETE CASCADE,
    PRIMARY KEY (id)
);

--down
DROP TABLE chunks;
DROP TABLE files;
