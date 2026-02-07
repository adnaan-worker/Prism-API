-- API Aggregator Platform Database Schema
-- PostgreSQL 15+

-- Enable UUID extension if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    quota BIGINT NOT NULL DEFAULT 10000,
    used_quota BIGINT NOT NULL DEFAULT 0,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    last_sign_in TIMESTAMP
);

-- Indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin) WHERE deleted_at IS NULL;

-- API Keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL REFERENCES users(id),
    key VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    rate_limit INTEGER NOT NULL DEFAULT 60,
    last_used_at TIMESTAMP
);

-- Indexes for api_keys table
CREATE INDEX IF NOT EXISTS idx_api_keys_deleted_at ON api_keys(deleted_at);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_active ON api_keys(user_id, is_active) WHERE deleted_at IS NULL;

-- API Configs table
CREATE TABLE IF NOT EXISTS api_configs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    base_url TEXT NOT NULL,
    api_key TEXT,
    models JSONB NOT NULL DEFAULT '[]',
    headers JSONB,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    priority INTEGER NOT NULL DEFAULT 100,
    weight INTEGER NOT NULL DEFAULT 1,
    max_rps INTEGER NOT NULL DEFAULT 0,
    timeout INTEGER NOT NULL DEFAULT 30
);

-- Indexes for api_configs table
CREATE INDEX IF NOT EXISTS idx_api_configs_deleted_at ON api_configs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_api_configs_type_active ON api_configs(type, is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_configs_priority ON api_configs(priority DESC) WHERE deleted_at IS NULL AND is_active = true;

-- Load Balancer Configs table
CREATE TABLE IF NOT EXISTS load_balancer_configs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    model_name VARCHAR(255) NOT NULL,
    strategy VARCHAR(50) NOT NULL DEFAULT 'round_robin',
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Indexes for load_balancer_configs table
CREATE INDEX IF NOT EXISTS idx_load_balancer_configs_deleted_at ON load_balancer_configs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_load_balancer_configs_model_name ON load_balancer_configs(model_name);

-- Request Logs table
CREATE TABLE IF NOT EXISTS request_logs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL REFERENCES users(id),
    api_key_id INTEGER NOT NULL REFERENCES api_keys(id),
    api_config_id INTEGER NOT NULL,
    model VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    response_time INTEGER NOT NULL,
    tokens_used INTEGER NOT NULL DEFAULT 0,
    error_msg TEXT
);

-- Indexes for request_logs table
CREATE INDEX IF NOT EXISTS idx_request_logs_deleted_at ON request_logs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_user_id ON request_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_api_key_id ON request_logs(api_key_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_api_config_id ON request_logs(api_config_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_model ON request_logs(model);
CREATE INDEX IF NOT EXISTS idx_request_logs_status_code ON request_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_user_created ON request_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_model_created ON request_logs(model, created_at DESC);

-- Sign In Records table
CREATE TABLE IF NOT EXISTS sign_in_records (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    user_id INTEGER NOT NULL REFERENCES users(id),
    quota_awarded INTEGER NOT NULL DEFAULT 0
);

-- Indexes for sign_in_records table
CREATE INDEX IF NOT EXISTS idx_sign_in_records_deleted_at ON sign_in_records(deleted_at);
CREATE INDEX IF NOT EXISTS idx_sign_in_records_user_id ON sign_in_records(user_id);
CREATE INDEX IF NOT EXISTS idx_sign_in_records_user_created ON sign_in_records(user_id, created_at DESC);

-- Comments for documentation
COMMENT ON TABLE users IS 'User accounts with authentication and quota management';
COMMENT ON TABLE api_keys IS 'API keys for user authentication and rate limiting';
COMMENT ON TABLE api_configs IS 'Third-party API provider configurations';
COMMENT ON TABLE load_balancer_configs IS 'Load balancing strategy configurations per model';
COMMENT ON TABLE request_logs IS 'API request logs for analytics and monitoring';
COMMENT ON TABLE sign_in_records IS 'Daily sign-in records for quota rewards';
