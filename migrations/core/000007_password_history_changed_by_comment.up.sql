-- ================================================
-- MIGRATION 7: PASSWORD_HISTORY.CHANGED_BY COMMENT
-- ================================================
-- SDD 008: document explicit actor on self-service change-password (changed_by = user_id).

COMMENT ON COLUMN password_history.changed_by IS
    'UUID of the user who performed the password change (self-service: same as user_id; admin reset: admin user id)';
