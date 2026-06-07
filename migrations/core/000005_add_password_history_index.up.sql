-- ================================================
-- MIGRATION 5: PASSWORD HISTORY PERFORMANCE INDEX
-- ================================================
-- Optimizes queries on password_history that filter and sort by user_id and changed_at.
-- Used by GetUserPasswordHistory (ORDER BY changed_at DESC LIMIT N) and
-- CleanOldPasswordHistory (subquery ordered by changed_at DESC LIMIT N).

CREATE INDEX IF NOT EXISTS idx_password_history_user_changed
    ON password_history (user_id, changed_at DESC);
