-- CREATE TYPE role as ENUM('organizer', 'customer');

CREATE TABLE IF NOT EXISTS users(
    cid SERIAL PRIMARY KEY,
    username VARCHAR(16) UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    profile role DEFAULT 'customer',
    CHECK(email LIKE '%@%.com')
);

CREATE TABLE IF NOT EXISTS temp_users(
    temp_id SERIAL PRIMARY KEY,
    username VARCHAR(16) NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    profile role DEFAULT 'customer',
    CHECK(email LIKE '%@%.com')
);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(cid),
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL,
    UNIQUE(user_id)
);
