-- ================================================
-- MIGRATION ROLLBACK 1: ENUM
-- ================================================

-- Drop enums
DROP TYPE IF EXISTS audit_result_enum;
DROP TYPE IF EXISTS identification_type_enum;

-- Drop extensions
DROP EXTENSION IF EXISTS "pgcrypto";
