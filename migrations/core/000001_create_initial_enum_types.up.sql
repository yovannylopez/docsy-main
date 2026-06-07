-- ================================================
-- MIGRATION 1: FUNDAMENTAL ENUM TYPES FOR THE SYSTEM
-- ================================================

-- ================================================
-- SYSTEM EXTENSIONS
-- ================================================
-- Create extension for UUID support (using pgcrypto for better performance)
-- pgcrypto provides cryptographic functions and more efficient UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ================================================
-- AUDIT RESULT ENUM TYPES
-- ================================================
-- Define the possible results of the actions audited in the system
-- Allows classification and filtering of audit logs
CREATE TYPE audit_result_enum AS ENUM (
    'success',  -- Action completed successfully
    'failure',  -- Action failed due to validation or business rules
    'error'     -- System error or uncontrolled exception
);

-- ================================================
-- IDENTIFICATION TYPE ENUM TYPES
-- ================================================
-- Define the valid identification document types in the system
-- Used for the users table
CREATE TYPE identification_type_enum AS ENUM (
    'cc',  -- Cédula de ciudadanía (Colombia)
    'ce',  -- Cédula de extranjería (Colombia)
    'pa',  -- Pasaporte (Colombia)
    'nit', -- Número de identificación tributaria (Colombia)
    'rut'  -- Registro Unico Tributario (Colombia)
);

-- ================================================
-- ENUM TYPE COMMENTS
-- ================================================

-- ================================================
-- AUDIT RESULT ENUM TYPE COMMENTS
-- ================================================
COMMENT ON TYPE audit_result_enum IS 'possible results of the actions audited in the system';

-- ================================================
-- IDENTIFICATION TYPE ENUM TYPE COMMENTS
-- ================================================
COMMENT ON TYPE identification_type_enum IS 'valid identification document types in the system';
