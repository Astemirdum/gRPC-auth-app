
create table if not exists users (
    id serial primary key,
    email varchar(50) NOT NULL UNIQUE,
    password_hash varchar(100) NOT NULL 
);

