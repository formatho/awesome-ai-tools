-- Create analytics_events table
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    action VARCHAR(100),
    label TEXT,
    value NUMERIC,
    user_id VARCHAR(255),
    session_id VARCHAR(255) NOT NULL,
    path TEXT,
    url TEXT,
    referrer TEXT,
    user_agent TEXT,
    ip_address VARCHAR(45),
    country VARCHAR(100),
    region VARCHAR(100),
    city VARCHAR(100),
    device_type VARCHAR(50),
    browser VARCHAR(100),
    os VARCHAR(100),
    screen_size VARCHAR(50),
    properties JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX idx_analytics_events_type ON analytics_events(type);
CREATE INDEX idx_analytics_events_name ON analytics_events(name);
CREATE INDEX idx_analytics_events_user_id ON analytics_events(user_id);
CREATE INDEX idx_analytics_events_session_id ON analytics_events(session_id);
CREATE INDEX idx_analytics_events_timestamp ON analytics_events(timestamp DESC);
CREATE INDEX idx_analytics_events_path ON analytics_events(path);
CREATE INDEX idx_analytics_events_category ON analytics_events(category);

-- Create materialized view for daily stats
CREATE MATERIALIZED VIEW IF NOT EXISTS analytics_daily_stats AS
SELECT
    DATE_TRUNC('day', timestamp) AS date,
    COUNT(DISTINCT user_id) AS unique_users,
    COUNT(DISTINCT session_id) AS sessions,
    COUNT(*) FILTER (WHERE type = 'pageview') AS page_views,
    COUNT(*) FILTER (WHERE type = 'event') AS events,
    AVG(
        EXTRACT(EPOCH FROM (
            SELECT MAX(timestamp) - MIN(timestamp)
            FROM analytics_events e2
            WHERE e2.session_id = analytics_events.session_id
        ))
    ) FILTER (WHERE type = 'pageview') AS avg_session_duration
FROM analytics_events
GROUP BY DATE_TRUNC('day', timestamp);

-- Refresh daily stats view every hour
CREATE OR REPLACE FUNCTION refresh_analytics_daily_stats()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY analytics_daily_stats;
END;
$$ LANGUAGE plpgsql;

-- Create table for conversion funnels
CREATE TABLE IF NOT EXISTS conversion_funnels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    steps JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default funnels
INSERT INTO conversion_funnels (name, steps) VALUES
('Pricing Page Conversion', '["landing", "pricing_page", "cta_click", "checkout_start", "checkout_complete"]'),
('Tool Usage', '["landing", "tool_page", "tool_used", "result_copied"]'),
('Beta Signup', '["landing", "beta_page", "form_start", "form_submit", "email_confirmed"]');

-- Create table for feature usage tracking
CREATE TABLE IF NOT EXISTS feature_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feature_name VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    session_id VARCHAR(255),
    usage_count INTEGER DEFAULT 1,
    first_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_feature_usage_name ON feature_usage(feature_name);
CREATE INDEX idx_feature_usage_user_id ON feature_usage(user_id);

-- Comments
COMMENT ON TABLE analytics_events IS 'Stores all analytics events and page views';
COMMENT ON TABLE conversion_funnels IS 'Defines conversion funnels to track';
COMMENT ON TABLE feature_usage IS 'Tracks feature usage by users';
