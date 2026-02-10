-- Migration script to add pricings table
-- Run this if you already have an existing database

-- Create pricings table
CREATE TABLE IF NOT EXISTS pricings (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    model_name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    input_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    output_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    currency VARCHAR(20) NOT NULL DEFAULT 'credits',
    unit INTEGER NOT NULL DEFAULT 1000,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    description VARCHAR(500),
    CONSTRAINT unique_model_provider UNIQUE (model_name, provider)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_pricings_deleted_at ON pricings(deleted_at);
CREATE INDEX IF NOT EXISTS idx_pricings_provider ON pricings(provider) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pricings_model_provider ON pricings(model_name, provider) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_pricings_is_active ON pricings(is_active) WHERE deleted_at IS NULL;

-- Add comment
COMMENT ON TABLE pricings IS 'Pricing configurations for models by provider';

-- Insert default pricing configurations
INSERT INTO pricings (model_name, provider, input_price, output_price, currency, unit, is_active, description)
VALUES
    -- OpenAI
    ('gpt-4', 'openai', 30, 60, 'credits', 1000, true, 'GPT-4 定价'),
    ('gpt-4-turbo', 'openai', 10, 30, 'credits', 1000, true, 'GPT-4 Turbo 定价'),
    ('gpt-3.5-turbo', 'openai', 0.5, 1.5, 'credits', 1000, true, 'GPT-3.5 Turbo 定价'),
    
    -- Anthropic
    ('claude-3-opus', 'anthropic', 15, 75, 'credits', 1000, true, 'Claude 3 Opus 定价'),
    ('claude-3-sonnet', 'anthropic', 3, 15, 'credits', 1000, true, 'Claude 3 Sonnet 定价'),
    ('claude-3-haiku', 'anthropic', 0.25, 1.25, 'credits', 1000, true, 'Claude 3 Haiku 定价'),
    
    -- Google
    ('gemini-pro', 'gemini', 0.5, 1.5, 'credits', 1000, true, 'Gemini Pro 定价'),
    ('gemini-ultra', 'gemini', 10, 30, 'credits', 1000, true, 'Gemini Ultra 定价')
ON CONFLICT (model_name, provider) DO NOTHING;
