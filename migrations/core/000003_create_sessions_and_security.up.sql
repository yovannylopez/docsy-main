-- ================================================
-- MIGRATION 3: SESSIONS AND SECURITY MODULE
-- ================================================

-- ================================================
-- SESSIONS MANAGEMENT TABLE
-- ================================================
-- Manages the active sessions of users
-- Includes security information, location and access control
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- Unique identifier of the session

    -- ================================================
    -- USER IDENTIFICATION
    -- ================================================
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,   -- User owner of the session

    -- ================================================
    -- AUTHENTICATION TOKENS
    -- ================================================
    refresh_token_hash VARCHAR(255) NOT NULL,                       -- Hash of the refresh token for renewal
    access_token_jti VARCHAR(255),                                  -- JTI of the JWT for selective invalidation

    -- ================================================
    -- CONTEXT INFORMATION
    -- ================================================
    user_agent TEXT,                                                -- User agent of the browser
    ip_address VARCHAR(50),                                         -- User IP address
    location TEXT,                                                  -- User geographical location
    device_fingerprint VARCHAR(255),                                -- Device fingerprint

    -- ================================================
    -- TIME CONTROL
    -- ================================================
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of creation of the session
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),            -- Date and time of the last use of the session
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,                   -- Date and time of expiration of the session

    -- ================================================
    -- STATE AND CONTROL
    -- ================================================
    is_active BOOLEAN DEFAULT true,                                 -- Indicates if the session is active
    revoked_at TIMESTAMP WITH TIME ZONE,                            -- Date and time of session revocation
    revoked_reason TEXT                                             -- Reason for the session revocation
);

-- ================================================
-- PASSWORD HISTORY TABLE
-- ================================================
-- Prevents the reuse of previous passwords by enforcing security policies
-- Implements password security policies
CREATE TABLE IF NOT EXISTS password_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- Unique identifier of the record in the password history
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,   -- User of the password history
    password_hash VARCHAR(255) NOT NULL,                            -- Hash of the previous password
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of the password change
    changed_by UUID REFERENCES users(id)                            -- User who performed the change (admin in resets)
);

-- ================================================
-- VERIFICATION TOKENS TABLE
-- ================================================
-- Manages tokens for email verification and password reset
-- Includes expiration control and unique use
CREATE TABLE IF NOT EXISTS verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- Unique identifier of the token
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,   -- User for which the token is generated
    token_hash VARCHAR(255) NOT NULL UNIQUE,                        -- Hash of the token (not the raw token for security)
    token_type VARCHAR(255) NOT NULL,                               -- Type: email_verification, password_reset, mfa_setup
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,                   -- Date and time of expiration of the token
    used_at TIMESTAMP WITH TIME ZONE,                               -- Date and time of use of the token (NULL if not used)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()               -- Date and time of creation of the token
);

-- ================================================
-- SYSTEM CONFIGURATION TABLE
-- ================================================
-- Stores security policies and system configuration
-- Allows dynamic configuration without restarting the application
CREATE TABLE IF NOT EXISTS system_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- Unique identifier of the configuration
    key VARCHAR(100) UNIQUE NOT NULL,                               -- Configuration key (unique)
    value TEXT NOT NULL,                                            -- Configuration value
    description TEXT,                                               -- Detailed description of the configuration
    is_sensitive BOOLEAN DEFAULT false,                             -- Indicates if the value is sensitive and must be encrypted
    is_public BOOLEAN DEFAULT false,                                -- Indicates if the configuration can be read without authentication
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of creation of the configuration
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of the last update of the configuration
    updated_by UUID REFERENCES users(id)                            -- User who performed the last update of the configuration
);

-- ================================================
-- TABLE AND COLUMN COMMENTS
-- ================================================

-- ================================================
-- COMMENTS FOR THE SESSIONS TABLE
-- ================================================
COMMENT ON TABLE sessions IS 'manages active user sessions with security controls and audit tracking';
COMMENT ON COLUMN sessions.id IS 'unique identifier of the session in the system';
COMMENT ON COLUMN sessions.user_id IS 'reference to the user owner of the session';
COMMENT ON COLUMN sessions.refresh_token_hash IS 'secure hash of the refresh token for session renewal';
COMMENT ON COLUMN sessions.access_token_jti IS 'unique identifier of the JWT for selective token invalidation';
COMMENT ON COLUMN sessions.user_agent IS 'user agent of the browser or client application';
COMMENT ON COLUMN sessions.ip_address IS 'IP address from where the session was established';
COMMENT ON COLUMN sessions.location IS 'approximate geographical location of the user (optional)';
COMMENT ON COLUMN sessions.device_fingerprint IS 'unique fingerprint of the device for anomaly detection';
COMMENT ON COLUMN sessions.created_at IS 'date and time exact of creation of the session';
COMMENT ON COLUMN sessions.last_used_at IS 'date and time of the last active use of the session';
COMMENT ON COLUMN sessions.expires_at IS 'date and time of automatic expiration of the session';
COMMENT ON COLUMN sessions.is_active IS 'indicates if the session is active and valid for use';
COMMENT ON COLUMN sessions.revoked_at IS 'date and time of manual session revocation (NULL if not revoked)';
COMMENT ON COLUMN sessions.revoked_reason IS 'reason for the session revocation (security, logout, etc.)';

-- ================================================
-- COMMENTS FOR THE PASSWORD_HISTORY TABLE
-- ================================================
COMMENT ON TABLE password_history IS 'prevents reuse of previous passwords by enforcing security policies';
COMMENT ON COLUMN password_history.id IS 'unique identifier of the record in the password history';
COMMENT ON COLUMN password_history.user_id IS 'reference to the user of the password history';
COMMENT ON COLUMN password_history.password_hash IS 'secure hash of the previous password (bcrypt)';
COMMENT ON COLUMN password_history.changed_at IS 'date and time exact of the password change';
COMMENT ON COLUMN password_history.changed_by IS 'user who performed the change (NULL if it was the own user, admin in resets)';

-- ================================================
-- COMMENTS FOR THE VERIFICATION_TOKENS TABLE
-- ================================================
COMMENT ON TABLE verification_tokens IS 'manages secure tokens for email verification and password reset';
COMMENT ON COLUMN verification_tokens.id IS 'unique identifier of the verification token';
COMMENT ON COLUMN verification_tokens.user_id IS 'reference to the user for which the token is generated';
COMMENT ON COLUMN verification_tokens.token_hash IS 'secure hash of the token (not the raw token for security)';
COMMENT ON COLUMN verification_tokens.token_type IS 'type of token: email_verification, password_reset, mfa_setup';
COMMENT ON COLUMN verification_tokens.expires_at IS 'date and time of automatic expiration of the token';
COMMENT ON COLUMN verification_tokens.used_at IS 'date and time of use of the token (NULL if not used)';
COMMENT ON COLUMN verification_tokens.created_at IS 'date and time of creation of the verification token';

-- ================================================
-- COMMENTS FOR THE SYSTEM_CONFIG TABLE
-- ================================================
COMMENT ON TABLE system_config IS 'stores security policies and dynamic system configuration';
COMMENT ON COLUMN system_config.id IS 'unique identifier of the system configuration';
COMMENT ON COLUMN system_config.key IS 'unique configuration key (ej: password.min_length)';
COMMENT ON COLUMN system_config.value IS 'current value of the configuration (can be text, number, boolean)';
COMMENT ON COLUMN system_config.description IS 'detailed description of the purpose and use of the configuration';
COMMENT ON COLUMN system_config.is_sensitive IS 'indicates if the value is sensitive and must be encrypted in the DB';
COMMENT ON COLUMN system_config.is_public IS 'indicates if the configuration can be read without authentication';
COMMENT ON COLUMN system_config.created_at IS 'date and time of creation of the configuration in the system';
COMMENT ON COLUMN system_config.updated_at IS 'date and time of the last modification of the configuration';
COMMENT ON COLUMN system_config.updated_by IS 'user who performed the last update of the configuration';

-- ================================================
-- INITIAL SYSTEM DATA
-- ================================================

-- ================================================
-- INITIAL SECURITY CONFIGURATIONS
-- ================================================
-- Password and system security policies
-- Each configuration has a specific security purpose
DO $$
BEGIN
    -- Password security policies configurations
    IF NOT EXISTS (SELECT 1 FROM system_config WHERE key = 'password.min_length') THEN
        INSERT INTO system_config (key, value, description, is_public) VALUES
        ('password.min_length', '8', 'minimum password length for system users', false);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM system_config WHERE key = 'password.require_uppercase') THEN
        INSERT INTO system_config (key, value, description, is_public) VALUES
        ('password.require_uppercase', 'true', 'require at least one uppercase letter in the password', false);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM system_config WHERE key = 'password.require_lowercase') THEN
        INSERT INTO system_config (key, value, description, is_public) VALUES
        ('password.require_lowercase', 'true', 'require at least one lowercase letter in the password', false);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM system_config WHERE key = 'password.require_numbers') THEN
        INSERT INTO system_config (key, value, description, is_public) VALUES
        ('password.require_numbers', 'true', 'require at least one number in the password', false);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM system_config WHERE key = 'password.require_symbols') THEN
        INSERT INTO system_config (key, value, description, is_public) VALUES
        ('password.require_symbols', 'true', 'require at least one special symbol in the password', false);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM system_config WHERE key = 'password.history_count') THEN
        INSERT INTO system_config (key, value, description, is_public) VALUES
        ('password.history_count', '5', 'number of previous passwords to remember to prevent reuse', false);
    END IF;
    
    -- Session management configurations
    IF NOT EXISTS (SELECT 1 FROM system_config WHERE key = 'session.max_duration_hours') THEN
        INSERT INTO system_config (key, value, description, is_public) VALUES
        ('session.max_duration_hours', '24', 'maximum session duration in hours before requiring re-authentication', false);
    END IF;
    
    -- Access control and blocking configurations
    IF NOT EXISTS (SELECT 1 FROM system_config WHERE key = 'login.max_failed_attempts') THEN
        INSERT INTO system_config (key, value, description, is_public) VALUES
        ('login.max_failed_attempts', '5', 'maximum number of failed attempts before locking the account', false);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM system_config WHERE key = 'login.lockout_duration_minutes') THEN
        INSERT INTO system_config (key, value, description, is_public) VALUES
        ('login.lockout_duration_minutes', '30', 'account lockout duration in minutes after failed login attempts', false);
    END IF;
    
    -- Multi-factor authentication (MFA) configurations
    IF NOT EXISTS (SELECT 1 FROM system_config WHERE key = 'mfa.enforce_for_admins') THEN
        INSERT INTO system_config (key, value, description, is_public) VALUES
        ('mfa.enforce_for_admins', 'true', 'require two-factor authentication for all administrators', false);
    END IF;
END $$;
