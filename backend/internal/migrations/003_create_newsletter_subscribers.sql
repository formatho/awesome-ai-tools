-- Migration: Create newsletter_subscribers table
-- Created: 2026-03-22
-- Purpose: Store email newsletter subscriptions from marketing website

CREATE TABLE IF NOT EXISTS newsletter_subscribers (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    source TEXT NOT NULL, -- 'homepage', 'pricing', 'blog'
    subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    unsubscribed_at TIMESTAMP,
    metadata TEXT, -- JSON string for additional data
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_newsletter_email ON newsletter_subscribers(email);
CREATE INDEX IF NOT EXISTS idx_newsletter_source ON newsletter_subscribers(source);
CREATE INDEX IF NOT EXISTS idx_newsletter_subscribed_at ON newsletter_subscribers(subscribed_at);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_newsletter_subscribers_updated_at
AFTER UPDATE ON newsletter_subscribers
FOR EACH ROW
BEGIN
    UPDATE newsletter_subscribers SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;
