-- +migrate Up
CREATE TABLE users
(
    "id"         bigserial,
    "created_at" timestamptz DEFAULT now(),
    "updated_at" timestamptz DEFAULT now(),
    "deleted_at" timestamptz,
    "email"      varchar(500) NOT NULL UNIQUE,
    "password"   bytea,
    "name"       text,
    PRIMARY KEY ("id"),
    UNIQUE ("email")
);
-- +migrate Down
drop table users;
