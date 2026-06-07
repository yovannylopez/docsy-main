-- ================================================
-- MIGRATION 5: PASSWORD HISTORY PERFORMANCE INDEX (ROLLBACK)
-- ================================================

DROP INDEX IF EXISTS idx_password_history_user_changed;
