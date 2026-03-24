-- Migration: 001_create_subscriptions.sql
-- Description: Creates subscriptions and payments tables for Stripe integration
-- Created: 2026-03-22

-- Create subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    organization_id TEXT,
    tier TEXT NOT NULL DEFAULT 'free',
    status TEXT NOT NULL DEFAULT 'active',
    stripe_customer_id TEXT UNIQUE,
    stripe_subscription_id TEXT UNIQUE,
    stripe_price_id TEXT,
    current_period_start DATETIME,
    current_period_end DATETIME,
    cancel_at_period_end INTEGER DEFAULT 0,
    canceled_at DATETIME,
    trial_start DATETIME,
    trial_end DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for subscriptions
CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_org ON subscriptions(organization_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tier ON subscriptions(tier);

-- Create payments table
CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    subscription_id TEXT,
    stripe_invoice_id TEXT,
    stripe_payment_intent_id TEXT,
    amount REAL NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    status TEXT NOT NULL,
    description TEXT,
    paid_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL
);

-- Create indexes for payments
CREATE INDEX IF NOT EXISTS idx_payments_user ON payments(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_subscription ON payments(subscription_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_created ON payments(created_at);

-- Create usage_records table for tracking usage metrics
CREATE TABLE IF NOT EXISTS usage_records (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    date DATE NOT NULL,
    hour INTEGER NOT NULL DEFAULT 0,
    agents_created INTEGER DEFAULT 0,
    agents_running INTEGER DEFAULT 0,
    tasks_completed INTEGER DEFAULT 0,
    tasks_failed INTEGER DEFAULT 0,
    api_calls INTEGER DEFAULT 0,
    storage_used_mb INTEGER DEFAULT 0,
    agent_pools_used INTEGER DEFAULT 0,
    advanced_skills INTEGER DEFAULT 0,
    custom_integrations INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, date, hour)
);

-- Create index for usage records
CREATE INDEX IF NOT EXISTS idx_usage_org_date ON usage_records(organization_id, date);

-- Create feature_flags table
CREATE TABLE IF NOT EXISTS feature_flags (
    id TEXT PRIMARY KEY,
    key TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    enabled INTEGER DEFAULT 1,
    allowed_tiers TEXT,
    rollout_percentage INTEGER DEFAULT 100,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default feature flags
INSERT OR IGNORE INTO feature_flags (id, key, name, description, allowed_tiers) VALUES
    ('ff-001', 'agent_pools', 'Agent Pools', 'Run multiple agents in parallel', '["pro","team","enterprise"]'),
    ('ff-002', 'advanced_skills', 'Advanced Skills Library', 'Access to premium skills', '["pro","team","enterprise"]'),
    ('ff-003', 'custom_integrations', 'Custom Integrations', 'Create custom integrations', '["pro","team","enterprise"]'),
    ('ff-004', 'team_collaboration', 'Team Collaboration', 'Share agents with team members', '["team","enterprise"]'),
    ('ff-005', 'sso_saml', 'SSO/SAML', 'Single sign-on integration', '["enterprise"]'),
    ('ff-006', 'unlimited_agents', 'Unlimited Agents', 'Create unlimited agents', '["pro","team","enterprise"]'),
    ('ff-007', 'unlimited_tasks', 'Unlimited Tasks', 'Run unlimited tasks per day', '["pro","team","enterprise"]'),
    ('ff-008', 'priority_support', 'Priority Support', 'Get priority support response', '["pro","team","enterprise"]');
