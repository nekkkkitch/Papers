SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


create schema if not exists public;

alter schema public owner to pg_database_owner;

SET search_path TO public;

DROP EXTENSION IF EXISTS "uuid-ossp";

CREATE EXTENSION "uuid-ossp" SCHEMA public;

create table if not exists public.users(
    id uuid not null primary key default uuid_generate_v4(),
    login varchar(24) not null,
    password text not null,
    refresh_token text,
    balance real not null default 0
);

create table if not exists public.papers(
    name text not null primary key,
    price real not null,
    past_prices real[]
);

create table if not exists public.storage(
    id uuid not null references public.users(id),
    paper_name text not null references public.papers(name),
    amount int default 0,
    primary key (id, paper_name)
);

insert into public.papers(name, price) values('Dogecoin', '100');
insert into public.papers(name, price) values('Amogus', '12.5');
insert into public.papers(name, price) values('Ichor', '666666');