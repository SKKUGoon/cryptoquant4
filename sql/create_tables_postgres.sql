create table if not exists cryptoquant.trading_metadata (
    key TEXT primary key,
    value text not null,
    description TEXT,
    value_type TEXT,
	created_at timestamptz default now(),
    updated_at timestamptz default now()
);

CREATE TABLE IF NOT EXISTS cryptoquant.strategy_kimchi_order_logs (
    id SERIAL PRIMARY KEY,
    pair_id TEXT NOT NULL,
    order_time timestamptz NOT NULL,
    execution_time timestamptz NOT NULL,
    pair_side TEXT NOT NULL,
    exchange TEXT NOT NULL,
    side TEXT NOT NULL,
    order_price double precision,
    executed_price double precision,
    quantity double precision,
    anchor_price double precision
);
