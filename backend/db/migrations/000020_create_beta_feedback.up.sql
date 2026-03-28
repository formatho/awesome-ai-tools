-- Create beta_feedback table
CREATE TABLE IF NOT EXISTS beta_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    rating INTEGER DEFAULT 0,
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    priority VARCHAR(50) DEFAULT 'medium',
    browser TEXT,
    steps_to_reproduce TEXT,
    expected_behavior TEXT,
    actual_behavior TEXT,
    attachments JSONB DEFAULT '[]',
    status VARCHAR(50) DEFAULT 'new',
    resolution TEXT,
    assigned_to VARCHAR(255),
    beta_tester_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX idx_beta_feedback_type ON beta_feedback(type);
CREATE INDEX idx_beta_feedback_status ON beta_feedback(status);
CREATE INDEX idx_beta_feedback_priority ON beta_feedback(priority);
CREATE INDEX idx_beta_feedback_email ON beta_feedback(email);
CREATE INDEX idx_beta_feedback_created_at ON beta_feedback(created_at DESC);
CREATE INDEX idx_beta_feedback_beta_tester_id ON beta_feedback(beta_tester_id);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_beta_feedback_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER beta_feedback_updated_at
    BEFORE UPDATE ON beta_feedback
    FOR EACH ROW
    EXECUTE FUNCTION update_beta_feedback_updated_at();

-- Add comments
COMMENT ON TABLE beta_feedback IS 'Stores feedback from beta testers';
COMMENT ON COLUMN beta_feedback.type IS 'Type of feedback: bug, feature, testimonial, general';
COMMENT ON COLUMN beta_feedback.rating IS 'Rating from 1-5 (for testimonials and general feedback)';
COMMENT ON COLUMN beta_feedback.priority IS 'Priority level: low, medium, high, critical';
COMMENT ON COLUMN beta_feedback.status IS 'Status: new, in_progress, resolved, closed';
