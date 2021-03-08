--up
create table sessions
(
    id         BYTES NOT NULL,
    token      text  not null,
    ip         inet  not null,
    user_agent text  not null,
    user_id    UUID  not null,
    created_at timestamp default now(),
    updated_at timestamp default now(),

    unique (token),
    primary key (id)
);

--down
drop table sessions;
