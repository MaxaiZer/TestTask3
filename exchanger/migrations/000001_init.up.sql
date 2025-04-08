CREATE TABLE IF NOT EXISTS currencies (
  code TEXT PRIMARY KEY
);

INSERT INTO currencies VALUES ('USD'), ('EUR'), ('RUB') ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS exchange_rates (
    id SERIAL PRIMARY KEY,
    currency TEXT UNIQUE NOT NULL references currencies(code),
    rate DECIMAL NOT NULL CHECK ( rate >= 0 )
);

INSERT INTO exchange_rates (currency, rate) VALUES
('USD', 1),
('EUR', 0.85),
('RUB', 0.1)
ON CONFLICT (currency) DO UPDATE SET rate = EXCLUDED.rate;