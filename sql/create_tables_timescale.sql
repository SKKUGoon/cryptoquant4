CREATE TABLE IF NOT EXISTS cryptoquant.premium_logs (
    time        TIMESTAMPTZ NOT NULL,
    symbol           TEXT NOT NULL,
    anchor_price     DOUBLE PRECISION,
    kimchi_best_bid  DOUBLE PRECISION,
    kimchi_best_ask  DOUBLE PRECISION,
    cefi_best_bid    DOUBLE PRECISION,
    cefi_best_ask    DOUBLE PRECISION
);
SELECT create_hypertable('cryptoquant.premium_logs', 'time', if_not_exists => TRUE);
SELECT add_retention_policy('cryptoquant.premium_logs', INTERVAL '5 days');

CREATE TABLE IF NOT EXISTS cryptoquant.account_snapshots (
    time        TIMESTAMPTZ NOT NULL,
    exchange TEXT not null,
    available DOUBLE PRECISION NOT NULL,
    reserved DOUBLE PRECISION NOT NULL,
    total DOUBLE PRECISION NOT NULL,
    wallet_balance_usdt DOUBLE PRECISION NOT NULL,
    wallet_balance_krw DOUBLE PRECISION NOT NULL
);
SELECT create_hypertable('cryptoquant.account_snapshots', 'time', if_not_exists => TRUE);
