-- Migration 000006: add indexes to support MFA TOTP token lookups.
-- Covers the mfa_setup and mfa_challenge token_type queries used by the MFA feature.

CREATE INDEX IF NOT EXISTS idx_verification_tokens_user_type
    ON verification_tokens (user_id, token_type);

CREATE INDEX IF NOT EXISTS idx_verification_tokens_expires_at
    ON verification_tokens (expires_at)
    WHERE used_at IS NULL;
