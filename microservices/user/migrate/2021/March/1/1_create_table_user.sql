--up
create table users
(
    id         UUID NOT NULL DEFAULT gen_random_uuid(),
    email      text not null,
    name       text not null,
    pass_hash  bytea,
    created_at timestamp     default now(),
    updated_at timestamp     default now(),

    unique (email),
    unique (name),
    primary key (id)
);

--down
drop table users;
