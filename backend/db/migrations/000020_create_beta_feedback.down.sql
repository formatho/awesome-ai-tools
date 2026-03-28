-- Drop beta_feedback table
DROP TABLE IF EXISTS beta_feedback;

-- Drop trigger and function
DROP TRIGGER IF EXISTS beta_feedback_updated_at ON beta_feedback;
DROP FUNCTION IF EXISTS update_beta_feedback_updated_at();
