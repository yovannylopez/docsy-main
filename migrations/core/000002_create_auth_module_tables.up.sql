-- ================================================
-- MIGRATION 2: AUTH MODULE TABLES
-- ================================================

-- ================================================
-- MAIN USERS TABLE
-- ================================================
-- Stores all user information for the authentication system
-- Includes authentication, personal information, security and metadata
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- Unique identifier of the user

    -- ================================================
    -- BASIC AUTHENTICATION INFORMATION
    -- ================================================
    email VARCHAR(255) UNIQUE NOT NULL,                             -- User email (unique)
    username VARCHAR(50) UNIQUE,                                    -- Username (unique, optional)
    password_hash VARCHAR(255) NOT NULL,                            -- User password hash

    -- ================================================
    -- PERSONAL INFORMATION
    -- ================================================
    first_name VARCHAR(100) NOT NULL,                               -- User first name
    last_name VARCHAR(100) NOT NULL,                                -- User last name
    identification_number VARCHAR(20) UNIQUE,                       -- Identification number (unique)
    identification_type identification_type_enum DEFAULT 'cc',      -- Identification type: cc, ce, pa, nit, rut
    phone VARCHAR(20),                                              -- User phone number

    -- ================================================
    -- ACCOUNT STATES
    -- ================================================
    is_active BOOLEAN DEFAULT false,                                -- Indicates if the account is active
    is_verified BOOLEAN DEFAULT false,                              -- Indicates if the email has been verified

    -- ================================================
    -- SECURITY AND ACCESS CONTROL
    -- ================================================
    last_login_at TIMESTAMP WITH TIME ZONE,                         -- Date and time of the last login
    failed_login_attempts INTEGER DEFAULT 0,                        -- Number of failed login attempts
    last_failed_login_at TIMESTAMP WITH TIME ZONE,                  -- Date and time of the last failed login attempt
    locked_until TIMESTAMP WITH TIME ZONE,                          -- Date until the account is locked
    password_changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),     -- Date of the last password change
    must_change_password BOOLEAN DEFAULT false,                     -- Indicates if the user must change their password

    -- ================================================
    -- MULTIFACTOR AUTHENTICATION (MFA)
    -- ================================================
    mfa_enabled BOOLEAN DEFAULT false,                              -- Indicates if the two-factor authentication is enabled
    mfa_secret VARCHAR(255),                                        -- Secret for the two-factor authentication

    -- ================================================
    -- METADATA AND AUDITING
    -- ================================================
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of creation of the record
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of the last update of the record
    created_by UUID REFERENCES users(id),                           -- User who created this record
    updated_by UUID REFERENCES users(id)                            -- User who updated this record
);

-- ================================================
-- SYSTEM ROLES TABLE
-- ================================================
-- Define the roles that users can have
-- Each role groups a specific set of permissions
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- Unique identifier of the role
    name VARCHAR(50) UNIQUE NOT NULL,                               -- Role name (unique, used internally)
    display_name VARCHAR(100) NOT NULL,                             -- Friendly name for display in the user interface
    description TEXT,                                               -- Detailed description of the role
    is_system_role BOOLEAN DEFAULT false,                           -- Indicates if it is a system role that cannot be deleted
    is_active BOOLEAN DEFAULT true,                                 -- Indicates if the role is active for use
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of creation of the role
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()               -- Date and time of the last update of the role
);

-- ================================================
-- SYSTEM PERMISSIONS TABLE
-- ================================================
-- Define the granular permissions that control access to resources
-- Follows the resource:action pattern for clarity
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- Unique identifier of the permission
    name VARCHAR(100) UNIQUE NOT NULL,                              -- Permission name (unique, format: resource.action)
    resource VARCHAR(50) NOT NULL,                                  -- Resource to which the permission applies (ej: users, documents)
    action VARCHAR(50) NOT NULL,                                    -- Action allowed on the resource (ej: create, read, update)
    description TEXT,                                               -- Detailed description of the permission
    is_system_permission BOOLEAN DEFAULT false,                     -- Indicates if it is a system permission that cannot be deleted
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of creation of the permission
    
    -- ================================================
    -- INTEGRITY RULES
    -- ================================================
    CONSTRAINT unique_resource_action UNIQUE (resource, action)     -- Avoid duplicates by resource:action
);

-- ================================================
-- RELATIONSHIPS BETWEEN TABLES
-- ================================================

-- ================================================
-- USER-ROLES RELATIONSHIP
-- ================================================
-- Manages the assignment of roles to users
-- Allows temporary roles and audit of assignments
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                  -- Unique identifier of the role assignment
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,   -- User to which the role is assigned
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,   -- Role assigned to the user
    assigned_by UUID REFERENCES users(id),                          -- User who assigned the role
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),             -- Date and time of the role assignment
    expires_at TIMESTAMP WITH TIME ZONE,                            -- Date of expiration of the role (for temporary roles)
    is_active BOOLEAN DEFAULT true,                                 -- To disable without deleting (soft disable)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),              -- Date and time of creation of the record
    
    -- ================================================
    -- INTEGRITY RULES
    -- ================================================
    CONSTRAINT unique_active_user_role UNIQUE (user_id, role_id)    -- A user can only have one active role specific
);

-- ================================================
-- ROLES-PERMISSIONS RELATIONSHIP
-- ================================================
-- Manages the assignment of permissions to roles
-- Implements the RBAC pattern (Role-Based Access Control)
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),                            -- Unique identifier of the permission assignment
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,             -- Role to which the permission is assigned
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE, -- Permission assigned to the role
    granted_by UUID REFERENCES users(id),                                     -- User who granted the permission
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),                        -- Date and time of creation of the record
    
    -- ================================================
    -- INTEGRITY RULES
    -- ================================================
    CONSTRAINT unique_role_permission UNIQUE (role_id, permission_id)         -- A role can only have one specific permission once
);

-- ================================================
-- TABLE AND COLUMN COMMENTS
-- ================================================

-- ================================================
-- COMMENTS FOR THE USERS TABLE
-- ================================================
COMMENT ON TABLE users IS 'main table storing all user information for the authentication system';
COMMENT ON COLUMN users.id IS 'unique identifier of the user in the system';
COMMENT ON COLUMN users.email IS 'email of the user (unique, used for authentication)';
COMMENT ON COLUMN users.username IS 'optional username for alternative access (unique)';
COMMENT ON COLUMN users.password_hash IS 'secure hash of the user password (bcrypt)';
COMMENT ON COLUMN users.first_name IS 'name of the user for personal identification';
COMMENT ON COLUMN users.last_name IS 'last name of the user for personal identification';
COMMENT ON COLUMN users.identification_number IS 'unique identification number (cedula, NIT, pasaporte)';
COMMENT ON COLUMN users.identification_type IS 'identification type: cc (cedula), ce (cedula extranjera), pa (pasaporte), nit (numero de identificación tributaria), rut (registro unico tributario)';
COMMENT ON COLUMN users.phone IS 'phone number of the user';
COMMENT ON COLUMN users.is_active IS 'indicates if the user account is active in the system';
COMMENT ON COLUMN users.is_verified IS 'indicates if the user email has been verified';
COMMENT ON COLUMN users.last_login_at IS 'date and time of the last successful login';
COMMENT ON COLUMN users.failed_login_attempts IS 'counter of failed login attempts';
COMMENT ON COLUMN users.last_failed_login_at IS 'date and time of the last failed login attempt';
COMMENT ON COLUMN users.locked_until IS 'date until the account is locked for security reasons';
COMMENT ON COLUMN users.password_changed_at IS 'date of the last password change for the user';
COMMENT ON COLUMN users.must_change_password IS 'indicates if the user must change their password in the next login';
COMMENT ON COLUMN users.mfa_enabled IS 'indicates if the two-factor authentication is enabled for the user';
COMMENT ON COLUMN users.mfa_secret IS 'cryptographic secret for two-factor authentication';
COMMENT ON COLUMN users.created_at IS 'date and time of creation of the user record';
COMMENT ON COLUMN users.updated_at IS 'date and time of the last update of the user record';
COMMENT ON COLUMN users.created_by IS 'user who created this record (circular reference for audit)';
COMMENT ON COLUMN users.updated_by IS 'user who performed the last update of the record';

-- ================================================
-- COMMENTS FOR THE ROLES TABLE
-- ================================================
COMMENT ON TABLE roles IS 'defines the roles that users can have in the authorization system';
COMMENT ON COLUMN roles.id IS 'unique identifier of the role in the system';
COMMENT ON COLUMN roles.name IS 'internal name of the role (unique, used for system logic)';
COMMENT ON COLUMN roles.display_name IS 'friendly name of the role for display in the user interface';
COMMENT ON COLUMN roles.description IS 'detailed description of the purpose and scope of the role';
COMMENT ON COLUMN roles.is_system_role IS 'indicates if it is a system role that cannot be deleted for security reasons';
COMMENT ON COLUMN roles.is_active IS 'indicates if the role is active and available for assignment';
COMMENT ON COLUMN roles.created_at IS 'date and time of creation of the role in the system';
COMMENT ON COLUMN roles.updated_at IS 'date and time of the last update of the role';

-- ================================================
-- COMMENTS FOR THE PERMISSIONS TABLE
-- ================================================
COMMENT ON TABLE permissions IS 'defines granular permissions that control access to system resources';
COMMENT ON COLUMN permissions.id IS 'unique identifier of the permission in the system';
COMMENT ON COLUMN permissions.name IS 'name of the permission (unique, format: resource.action)';
COMMENT ON COLUMN permissions.resource IS 'system resource to which the permission applies (ej: users, roles, audit)';
COMMENT ON COLUMN permissions.action IS 'specific action allowed on the resource (ej: create, read, update, delete)';
COMMENT ON COLUMN permissions.description IS 'detailed description of the permission and its purpose';
COMMENT ON COLUMN permissions.is_system_permission IS 'indicates if it is a system permission that cannot be deleted';
COMMENT ON COLUMN permissions.created_at IS 'date and time of creation of the permission in the system';

-- ================================================
-- COMMENTS FOR THE USER_ROLES TABLE
-- ================================================
COMMENT ON TABLE user_roles IS 'manages role assignments to users with time control and audit tracking';
COMMENT ON COLUMN user_roles.id IS 'unique identifier of the role assignment to user';
COMMENT ON COLUMN user_roles.user_id IS 'reference to the user to which the role is assigned';
COMMENT ON COLUMN user_roles.role_id IS 'reference to the role assigned to the user';
COMMENT ON COLUMN user_roles.assigned_by IS 'user who performed the role assignment (for audit)';
COMMENT ON COLUMN user_roles.assigned_at IS 'date and time exact of the role assignment';
COMMENT ON COLUMN user_roles.expires_at IS 'date of expiration of the role (NULL for permanent roles)';
COMMENT ON COLUMN user_roles.is_active IS 'indicates if the role assignment is active (soft disable)';
COMMENT ON COLUMN user_roles.created_at IS 'date and time of creation of the assignment record';

-- ================================================
-- COMMENTS FOR THE ROLE_PERMISSIONS TABLE
-- ================================================
COMMENT ON TABLE role_permissions IS 'implements the RBAC pattern by assigning granular permissions to roles';
COMMENT ON COLUMN role_permissions.id IS 'unique identifier of the permission assignment to role';
COMMENT ON COLUMN role_permissions.role_id IS 'reference to the role to which the permission is assigned';
COMMENT ON COLUMN role_permissions.permission_id IS 'reference to the permission assigned to the role';
COMMENT ON COLUMN role_permissions.granted_by IS 'user who granted the permission to the role (for audit)';
COMMENT ON COLUMN role_permissions.created_at IS 'date and time of creation of the permission assignment';

-- ================================================
-- INITIAL SYSTEM DATA
-- ================================================

-- ================================================
-- BASIC ROLES FOR THE SYSTEM
-- ================================================
-- Fundamental roles that cannot be deleted
-- Each role has a specific purpose in the system
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM roles WHERE name = 'super_admin') THEN
        INSERT INTO roles (name, display_name, description, is_system_role) VALUES
        ('super_admin', 'super_admin', 'system super administrator with full access', true);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM roles WHERE name = 'user') THEN
        INSERT INTO roles (name, display_name, description, is_system_role) VALUES
        ('user', 'user', 'standard system user with basic access', true);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM roles WHERE name = 'viewer') THEN
        INSERT INTO roles (name, display_name, description, is_system_role) VALUES
        ('viewer', 'viewer', 'read-only user with no modification permissions', true);
    END IF;
END $$;

-- ================================================
-- BASIC PERMISSIONS FOR THE SYSTEM
-- ================================================
-- Granular permissions that control access to resources
-- Follows the resource:action pattern for clarity
DO $$
BEGIN
    -- User management permissions
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'users.create') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('users.create', 'users', 'create', 'create new users in the system', true);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'users.read') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('users.read', 'users', 'read', 'view user information in the system', true);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'users.update') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('users.update', 'users', 'update', 'update existing user information', true);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'users.delete') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('users.delete', 'users', 'delete', 'delete users from the system', true);
    END IF;
    
    -- Role management permissions
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'roles.create') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('roles.create', 'roles', 'create', 'create new roles in the system', true);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'roles.read') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('roles.read', 'roles', 'read', 'view roles in the system', true);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'roles.update') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('roles.update', 'roles', 'update', 'update role information', true);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'roles.delete') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('roles.delete', 'roles', 'delete', 'delete roles from the system', true);
    END IF;
    
    -- Audit and system permissions
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'audit.read') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('audit.read', 'audit', 'read', 'view system audit logs', true);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'system.config') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('system.config', 'system', 'config', 'configure system parameters', true);
    END IF;
END $$;

-- ================================================
-- DEFAULT ADMIN USER
-- ================================================
-- Initial user to access the system
-- Password: Admin123! (must be changed in the first login)
-- Credentials: yovannyinstructor@gmail.com / Admin123!

-- Create default admin user
DO $$
DECLARE
    admin_role_id UUID;
    admin_user_id UUID := '00000000-0000-0000-0000-000000000001';
BEGIN
    -- Check if the user already exists
    IF NOT EXISTS (SELECT 1 FROM users WHERE id = admin_user_id) THEN
        -- Get the ID of the super_admin role
        SELECT id INTO admin_role_id FROM roles WHERE name = 'super_admin' LIMIT 1;
        
        -- Create default admin user
        INSERT INTO users (
            id, email, username, password_hash, first_name, last_name, identification_number, identification_type, phone,
            is_active, is_verified, password_changed_at, must_change_password,
            created_at, updated_at
        ) VALUES (
            admin_user_id,
            'yovannyinstructor@gmail.com',
            'yovanny',
            '$2a$10$GNqjA2hOqbcHcg61drRJkeFjCMTcBw1l6l18jamcpObw.hn9G5FAa', -- hash of "Admin123!"
            'Yovanny',
            'Lopez',
            '10031632',
            'cc',
            '+573022542943',
            true,
            true,
            NOW(),
            true,  -- Must change the password in the first login
            NOW(),
            NOW()
        );
        
        -- Assign the super_admin role to the user
        INSERT INTO user_roles (id, user_id, role_id, created_at) VALUES (
            gen_random_uuid(),
            admin_user_id,
            admin_role_id,
            NOW()
        );
    END IF;
END $$;

-- ================================================
-- END OF MIGRATION 2
-- ================================================
