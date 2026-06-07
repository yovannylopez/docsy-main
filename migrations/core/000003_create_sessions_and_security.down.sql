-- ================================================
-- ROLLBACK MIGRATION 3: SESSIONS AND SECURITY
-- ================================================

-- Drop tables in correct order (respecting foreign key constraints)
DROP TABLE IF EXISTS system_config;
DROP TABLE IF EXISTS verification_tokens;
DROP TABLE IF EXISTS password_history;
DROP TABLE IF EXISTS sessions;

-- ================================================
-- ROLLBACK: INITIAL BASIC DATA
-- ================================================

-- Delete system configurations
DELETE FROM system_config WHERE key IN (
    'password.min_length',
    'password.require_uppercase',
    'password.require_lowercase',
    'password.require_numbers',
    'password.require_symbols',
    'password.history_count',
    'session.max_duration_hours',
    'login.max_failed_attempts',
    'login.lockout_duration_minutes',
    'mfa.enforce_for_admins'
);
