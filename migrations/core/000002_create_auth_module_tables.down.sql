-- ================================================
-- MIGRATION ROLLBACK 2: AUTH MODULE TABLES
-- ================================================

-- Drop tables in correct order (respecting foreign key constraints)
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;

-- ================================================
-- INITIAL BASIC DATA ROLLBACK
-- ================================================

-- Drop basic permissions
DELETE FROM permissions WHERE name IN (
    'users.create', 'users.read', 'users.update', 'users.delete',
    'roles.create', 'roles.read', 'roles.update', 'roles.delete',
    'audit.read', 'system.config'
);

-- Drop default admin user
DELETE FROM user_roles WHERE user_id = '00000000-0000-0000-0000-000000000001';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-000000000001';

-- Drop basic roles
DELETE FROM roles WHERE name IN ('super_admin', 'user', 'viewer'); 

-- ================================================
-- ENUM TYPES CREATED IN THIS MIGRATION ROLLBACK
-- ================================================

-- Drop identification type enum
DROP TYPE IF EXISTS identification_type_enum; 