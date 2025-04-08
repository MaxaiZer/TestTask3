CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    password BYTEA NOT NULL,
    email TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS currencies (
  code TEXT PRIMARY KEY
);

INSERT INTO currencies VALUES ('USD'), ('EUR'), ('RUB') ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL references users(id),
    currency TEXT NOT NULL references currencies(code),
    amount DECIMAL NOT NULL CHECK ( amount >= 0 ),
    CONSTRAINT unique_user_currency UNIQUE (user_id, currency)
)