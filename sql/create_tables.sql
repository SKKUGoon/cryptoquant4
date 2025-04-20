create table if not exists cryptoquant.trading_metadata (
    key TEXT primary key,
    value text not null,
    description TEXT,
    value_type TEXT,
	created_at timestamptz default now(),
    updated_at timestamptz default now()
);