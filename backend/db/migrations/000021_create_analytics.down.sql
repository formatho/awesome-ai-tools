-- Drop analytics tables
DROP MATERIALIZED VIEW IF EXISTS analytics_daily_stats;
DROP TABLE IF EXISTS feature_usage;
DROP TABLE IF EXISTS conversion_funnels;
DROP TABLE IF EXISTS analytics_events;
DROP FUNCTION IF EXISTS refresh_analytics_daily_stats();
