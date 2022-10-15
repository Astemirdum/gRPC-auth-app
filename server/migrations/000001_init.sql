
-- +goose Up
create schema if not exists userdb;

create table if not exists userdb.users (
    id serial primary key,
    email varchar(50) NOT NULL UNIQUE,
    password_hash varchar(100) NOT NULL 
);

-- +goose Down
drop schema if exists userdb cascade;
