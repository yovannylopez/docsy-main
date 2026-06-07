-- Rollback migration 000006: remove MFA TOTP indexes.

DROP INDEX IF EXISTS idx_verification_tokens_user_type;
DROP INDEX IF EXISTS idx_verification_tokens_expires_at;
