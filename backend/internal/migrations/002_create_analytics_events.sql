-- Migration: Create analytics_events table for conversion funnel tracking
-- Created: 2026-03-22
-- Description: Tracks user journey events from pricing page to checkout

-- Create analytics_events table
CREATE TABLE IF NOT EXISTS analytics_events (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    session_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    event_name TEXT NOT NULL,
    properties TEXT, -- JSON blob for additional event data
    page_url TEXT,
    referrer TEXT,
    user_agent TEXT,
    ip_address TEXT,
    created_at DATETIME NOT NULL
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_name ON analytics_events(event_name);
CREATE INDEX IF NOT EXISTS idx_analytics_events_session ON analytics_events(session_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_user ON analytics_events(user_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_created ON analytics_events(created_at);
CREATE INDEX IF NOT EXISTS idx_analytics_events_funnel ON analytics_events(event_name, created_at);

-- Create composite index for common funnel queries
CREATE INDEX IF NOT EXISTS idx_analytics_funnel_query ON analytics_events(event_type, event_name, created_at);

-- Create index for date range queries
CREATE INDEX IF NOT EXISTS idx_analytics_date_range ON analytics_events(date(created_at));
