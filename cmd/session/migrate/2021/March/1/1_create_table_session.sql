--up
CREATE TABLE sessions
(
    id         UUID      NOT NULL,
    token      TEXT      NOT NULL,
    ip         INET      NOT NULL,
    user_agent TEXT      NOT NULL,
    user_id    UUID      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE (token),
    PRIMARY KEY (id)
);

--down
DROP TABLE sessions;
