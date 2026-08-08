-- Curated OCR / manual key-value metadata (array of {key,label,value}).
ALTER TABLE archive_documents
    ADD COLUMN IF NOT EXISTS extra_fields JSONB;

COMMENT ON COLUMN archive_documents.extra_fields IS
    'optional curated metadata as JSONB array of {key,label,value}; not the raw OCR dump';
