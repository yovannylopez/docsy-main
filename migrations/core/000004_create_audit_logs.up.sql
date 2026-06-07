-- ================================================
-- MIGRATION 4: AUDIT LOGS MODULE
-- ================================================

-- ================================================
-- MAIN AUDIT LOGS TABLE
-- ================================================
-- Registers all system actions for audit and traceability
-- Includes complete context, change data and correlation with application logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- Unique identifier of the audit log

    -- ================================================
    -- USER AND SESSION CONTEXT
    -- ================================================
    user_id UUID REFERENCES users(id),                              -- User who performed the action
    session_id UUID REFERENCES sessions(id),                        -- Session where the action was performed

    -- ================================================
    -- ACTION DETAILS
    -- ================================================
    action VARCHAR(100) NOT NULL,                                   -- Action performed (login, logout, create, update, delete, etc.)
    resource VARCHAR(100),                                          -- Resource affected by the action
    resource_id VARCHAR(100),                                       -- Specific identifier of the affected resource

    -- ================================================
    -- RESULT AND CONTEXT
    -- ================================================
    result audit_result_enum NOT NULL DEFAULT 'success',             -- Result of the action (success, failure, error)
    message TEXT,                                                    -- Descriptive message of the action
    ip_address VARCHAR(50),                                          -- IP address from where the action was performed
    user_agent TEXT,                                                 -- User agent of the browser
    request_id VARCHAR(100),                                         -- Request identifier for correlation

    -- ================================================
    -- CHANGE DATA
    -- ================================================
    previous_data JSONB,                                             -- Previous data in JSONB format (for comparison)
    new_data JSONB,                                                  -- New data in JSONB format (for comparison)
    changed_fields TEXT[],                                           -- Array of field names that were modified

    -- ================================================
    -- METADATA AND AUDIT
    -- ================================================
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of creation of the log
    created_by UUID REFERENCES users(id),                           -- User who created the audit record
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of the last update of the log
    updated_by UUID REFERENCES users(id)                            -- User who performed the last update of the audit record
);

-- ================================================
-- TABLE AND COLUMN COMMENTS
-- ================================================

-- ================================================
-- COMMENTS FOR THE AUDIT_LOGS TABLE
-- ================================================
COMMENT ON TABLE audit_logs IS 'records all system actions for complete audit trail and change traceability';
COMMENT ON COLUMN audit_logs.id IS 'unique identifier of the audit log in the system';
COMMENT ON COLUMN audit_logs.user_id IS 'reference to the user who performed the action (NULL for system actions)';
COMMENT ON COLUMN audit_logs.session_id IS 'reference to the session where the action was performed (for correlation)';
COMMENT ON COLUMN audit_logs.action IS 'specific action performed: login, logout, create, update, delete, assign, etc.';
COMMENT ON COLUMN audit_logs.resource IS 'system resource affected by the action (e.g: users, communications, documents)';
COMMENT ON COLUMN audit_logs.resource_id IS 'specific identifier of the affected resource (e.g: UUID of the modified user)';
COMMENT ON COLUMN audit_logs.result IS 'result of the action using enum: success, failure, error';
COMMENT ON COLUMN audit_logs.message IS 'detailed descriptive message of the performed action';
COMMENT ON COLUMN audit_logs.ip_address IS 'IP address from where the action was performed (for security)';
COMMENT ON COLUMN audit_logs.user_agent IS 'user agent of the browser or client application';
COMMENT ON COLUMN audit_logs.request_id IS 'unique identifier of the request for correlation with application logs';
COMMENT ON COLUMN audit_logs.previous_data IS 'previous data of the resource in JSONB format (for audit of changes)';
COMMENT ON COLUMN audit_logs.new_data IS 'new data of the resource in JSONB format (for audit of changes)';
COMMENT ON COLUMN audit_logs.changed_fields IS 'array of field names that were modified in the action';
COMMENT ON COLUMN audit_logs.created_at IS 'date and time exact of creation of the audit log';
COMMENT ON COLUMN audit_logs.created_by IS 'user who created the audit record (system or user)';
COMMENT ON COLUMN audit_logs.updated_at IS 'date and time of the last update of the audit log';
COMMENT ON COLUMN audit_logs.updated_by IS 'user who performed the last update of the audit record';
