-- +goose Up
-- +goose StatementBegin

-- VENUES

CREATE TABLE venues (
    id           text PRIMARY KEY,
    name         text NOT NULL,
    address      text NOT NULL,
    description  text,
    bank_account text
);

-- USERS + CREDENTIALS
CREATE TABLE users (
    id           text PRIMARY KEY,
    first_name   text NOT NULL,
    last_name    text NOT NULL,
    role         int,
    venue_id     text,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE credentials (
    user_id  text PRIMARY KEY,
    login    text UNIQUE NOT NULL,
    password text NOT NULL
);

-- MENU

CREATE TABLE menu_categories (
    id       text PRIMARY KEY,
    name     text NOT NULL,
    venue_id text NOT NULL
);

CREATE TABLE dishes (
    id            text PRIMARY KEY,
    name          text NOT NULL,
    description   text,
    category_id   text NOT NULL,
    price         int NOT NULL,
    weight        int NOT NULL,
    weight_unit   text NOT NULL,
    calories      int,
    protein       int,
    fat           int,
    carbs         int,
    venue_id      text NOT NULL
);

-- TABLES

CREATE TABLE tables (
    id            text PRIMARY KEY,
    number        int NOT NULL,
    venue_id      text NOT NULL,
    status        text NOT NULL,
    waiter_id     text,
    order_id      text,
    qr_token      text,
    session_token text,

    UNIQUE (venue_id, number)
);

-- ORDERS

CREATE TABLE orders (
    id           text PRIMARY KEY,
    table_id     text NOT NULL,
    waiter_id    text,
    status       text NOT NULL,
    total_amount int NOT NULL,
    paid_amount  int NOT NULL,
    wishes       text,
    created_at   timestamptz NOT NULL DEFAULT now()
);

-- ORDER ITEMS

CREATE TABLE order_items (
    id          text PRIMARY KEY,
    order_id    text NOT NULL,
    dish_id     text NOT NULL,
    dish_name   text NOT NULL,
    quantity    int NOT NULL,
    price       int NOT NULL,
    status      text NOT NULL,
    added_later boolean NOT NULL
);

-- TRANSACTIONS

CREATE TABLE transactions (
    id           text PRIMARY KEY,
    order_id     text NOT NULL,
    amount       int NOT NULL,
    tips_amount  int NOT NULL,
    status       text NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE transaction_items (
    transaction_id text,
    order_item_id  text,

    PRIMARY KEY (transaction_id, order_item_id)
);

-- WAITER REQUESTS

CREATE TABLE waiter_requests (
    id           text PRIMARY KEY,
    table_id     text NOT NULL,
    waiter_id    text,
    status       text NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS transaction_items;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS waiter_requests;

DROP TABLE IF EXISTS tables;

DROP TABLE IF EXISTS dishes;
DROP TABLE IF EXISTS menu_categories;

DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS venues;
-- +goose StatementEnd
