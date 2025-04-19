CREATE TABLE IF NOT EXISTS users_tokens (
    id SERIAL PRIMARY KEY,
    guid UUID UNIQUE,
    refresh_token TEXT,
    used BOOLEAN DEFAULT false,
    ip_address VARCHAR(50),
    refresh_token_id VARCHAR(64)
);