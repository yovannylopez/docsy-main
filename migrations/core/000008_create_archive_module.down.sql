-- ================================================
-- DOWN: ARCHIVE MODULE
-- ================================================

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN ('archive.read', 'archive.write', 'archive.manage')
);

DELETE FROM permissions WHERE name IN ('archive.read', 'archive.write', 'archive.manage');

DROP TABLE IF EXISTS archive_document_files;
DROP TABLE IF EXISTS archive_documents;
DROP TABLE IF EXISTS archive_document_categories;
DROP TABLE IF EXISTS archive_workspace_members;
DROP TABLE IF EXISTS archive_workspaces;
