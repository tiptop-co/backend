-- +goose Up
-- +goose StatementBegin

-- VENUES

CREATE TABLE venues (
    id           varchar(36) PRIMARY KEY,
    name         varchar(255) NOT NULL,
    address      varchar(255) NOT NULL,
    description  text,
    bank_account varchar(100)
);

-- USERS + CREDENTIALS
CREATE TABLE users (
    id           varchar(36) PRIMARY KEY,
    first_name   varchar(100) NOT NULL,
    last_name    varchar(100) NOT NULL,
    role         int,
    venue_id     varchar(36) REFERENCES venues(id),
    created_at   timestamptz DEFAULT now()
);

CREATE TABLE credentials (
    user_id  varchar(36) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    login    varchar(20) UNIQUE NOT NULL,
    password varchar(255) NOT NULL
);

-- MENU

CREATE TABLE menu_categories (
    id       varchar(36) PRIMARY KEY,
    name     varchar(100) NOT NULL,
    venue_id varchar(36) NOT NULL REFERENCES venues(id)
);

CREATE TABLE dishes (
    id            varchar(36) PRIMARY KEY,
    name          varchar(255) NOT NULL,
    description   text,
    category_id   varchar(36) NOT NULL REFERENCES menu_categories(id),
    price         int NOT NULL,
    weight        int NOT NULL,
    weight_unit   varchar(10) NOT NULL,
    calories      int,
    protein       int,
    fat           int,
    carbs         int,
    venue_id      varchar(36) NOT NULL REFERENCES venues(id)
);

-- TABLES

CREATE TABLE tables (
    id            varchar(36) PRIMARY KEY,
    number        int NOT NULL,
    venue_id      varchar(36) NOT NULL REFERENCES venues(id),
    status        varchar(32) NOT NULL,
    waiter_id     varchar(36) REFERENCES users(id),
    order_id      varchar(36),
    qr_token      varchar(64),
    session_token varchar(64),

    UNIQUE (venue_id, number)
);

-- ORDERS

CREATE TABLE orders (
    id           varchar(36) PRIMARY KEY,
    table_id     varchar(36) NOT NULL REFERENCES tables(id),
    waiter_id    varchar(36) REFERENCES users(id),
    status       varchar(32) NOT NULL,
    total_amount int NOT NULL,
    paid_amount  int NOT NULL,
    wishes       text,
    created_at   timestamptz DEFAULT now()
);

-- ORDER ITEMS

CREATE TABLE order_items (
    id          varchar(36) PRIMARY KEY,
    order_id    varchar(36) NOT NULL REFERENCES orders(id),
    dish_id     varchar(36) NOT NULL REFERENCES dishes(id),
    dish_name   varchar(255) NOT NULL,
    quantity    int NOT NULL,
    price       int NOT NULL,
    status      varchar(32) NOT NULL,
    added_later boolean NOT NULL
);

-- TRANSACTIONS

CREATE TABLE transactions (
    id           varchar(36) PRIMARY KEY,
    order_id     varchar(36) NOT NULL REFERENCES orders(id),
    amount       int NOT NULL,
    tips_amount  int NOT NULL,
    status       varchar(32) NOT NULL,
    created_at   timestamptz DEFAULT now()
);

CREATE TABLE transaction_items (
    transaction_id varchar(36) REFERENCES transactions(id),
    order_item_id  varchar(36) REFERENCES order_items(id),

    PRIMARY KEY (transaction_id, order_item_id)
);

-- WAITER REQUESTS

CREATE TABLE waiter_requests (
    id           varchar(36) PRIMARY KEY,
    table_id     varchar(36) NOT NULL REFERENCES tables(id),
    waiter_id    varchar(36) REFERENCES users(id),
    status       varchar(32) NOT NULL,
    created_at   timestamptz DEFAULT now()
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
