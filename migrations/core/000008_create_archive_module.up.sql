-- ================================================
-- MIGRATION 8: ARCHIVE MODULE (personal/family file)
-- ================================================

CREATE TABLE IF NOT EXISTS archive_workspaces (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(32) NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users(id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT archive_workspaces_type_chk CHECK (type IN ('personal', 'household', 'organization'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_archive_workspaces_personal_owner
    ON archive_workspaces (owner_user_id)
    WHERE type = 'personal' AND is_active = true;

CREATE INDEX IF NOT EXISTS idx_archive_workspaces_owner ON archive_workspaces (owner_user_id);

CREATE TABLE IF NOT EXISTS archive_workspace_members (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES archive_workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(32) NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT archive_workspace_members_role_chk CHECK (role IN ('owner', 'member', 'viewer')),
    CONSTRAINT archive_workspace_members_unique UNIQUE (workspace_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_archive_workspace_members_user ON archive_workspace_members (user_id);

CREATE TABLE IF NOT EXISTS archive_document_categories (
    code VARCHAR(64) PRIMARY KEY,
    label_es VARCHAR(120) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS archive_documents (
    id UUID PRIMARY KEY,
    workspace_id UUID NOT NULL REFERENCES archive_workspaces(id) ON DELETE CASCADE,
    category_code VARCHAR(64) NOT NULL REFERENCES archive_document_categories(code),
    title VARCHAR(500) NOT NULL,
    document_date DATE,
    due_date DATE,
    issuer VARCHAR(255),
    reference_number VARCHAR(255),
    amount_cents BIGINT,
    currency VARCHAR(8) NOT NULL DEFAULT 'COP',
    notes TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT archive_documents_status_chk CHECK (status IN ('active', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_archive_documents_workspace ON archive_documents (workspace_id);
CREATE INDEX IF NOT EXISTS idx_archive_documents_category ON archive_documents (workspace_id, category_code);
CREATE INDEX IF NOT EXISTS idx_archive_documents_due_date ON archive_documents (workspace_id, due_date);

CREATE TABLE IF NOT EXISTS archive_document_files (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL REFERENCES archive_documents(id) ON DELETE CASCADE,
    storage_key VARCHAR(512) NOT NULL,
    original_name VARCHAR(512) NOT NULL,
    content_type VARCHAR(128) NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    uploaded_by UUID REFERENCES users(id),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_archive_document_files_document ON archive_document_files (document_id);

COMMENT ON TABLE archive_workspaces IS 'multi-tenant containers for personal, household or organization archives';
COMMENT ON TABLE archive_workspace_members IS 'membership of users in archive workspaces';
COMMENT ON TABLE archive_document_categories IS 'seeded document categories for the personal archive module';
COMMENT ON TABLE archive_documents IS 'document metadata (binaries in archive_document_files, iteration C)';
COMMENT ON TABLE archive_document_files IS 'stored binary attachments for archive documents';

-- Categories seed
INSERT INTO archive_document_categories (code, label_es, sort_order, is_active) VALUES
    ('taxes', 'Impuestos', 10, true),
    ('utilities', 'Servicios públicos', 20, true),
    ('invoices', 'Facturas de compra', 30, true),
    ('payments', 'Pagos y recibos', 40, true),
    ('certificates', 'Certificados', 50, true),
    ('health', 'Salud', 60, true),
    ('other', 'Otros', 90, true)
ON CONFLICT (code) DO NOTHING;

-- Permissions
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'archive.read') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('archive.read', 'archive', 'read', 'view personal/family archive workspaces and documents', true);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'archive.write') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('archive.write', 'archive', 'write', 'create and update archive documents', true);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'archive.manage') THEN
        INSERT INTO permissions (name, resource, action, description, is_system_permission) VALUES
        ('archive.manage', 'archive', 'manage', 'manage archive workspace membership and settings', true);
    END IF;
END $$;

-- Assign archive permissions to user and viewer roles (super_admin bypasses checks)
DO $$
DECLARE
    r RECORD;
    p RECORD;
BEGIN
    FOR r IN SELECT id, name FROM roles WHERE name IN ('user', 'viewer') LOOP
        FOR p IN SELECT id, name FROM permissions WHERE name IN ('archive.read', 'archive.write', 'archive.manage') LOOP
            IF r.name = 'viewer' AND p.name <> 'archive.read' THEN
                CONTINUE;
            END IF;
            IF NOT EXISTS (
                SELECT 1 FROM role_permissions WHERE role_id = r.id AND permission_id = p.id
            ) THEN
                INSERT INTO role_permissions (id, role_id, permission_id, created_at)
                VALUES (gen_random_uuid(), r.id, p.id, NOW());
            END IF;
        END LOOP;
    END LOOP;
END $$;
