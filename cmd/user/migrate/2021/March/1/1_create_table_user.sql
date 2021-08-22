--up
CREATE TABLE users
(
    id         UUID       NOT NULL DEFAULT GEN_RANDOM_UUID(),
    email      TEXT       NOT NULL,
    name       TEXT       NOT NULL,
    pass_hash  BYTEA      NOT NULL,
    avatars    UUID ARRAY NOT NULL DEFAULT '{}',
    created_at TIMESTAMP  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP  NOT NULL DEFAULT NOW(),

    UNIQUE (email),
    UNIQUE (name),

    PRIMARY KEY (id)
);

--down
DROP TABLE users;
