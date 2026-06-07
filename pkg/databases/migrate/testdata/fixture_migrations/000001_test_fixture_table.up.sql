-- Minimum fixture for golang-migrate integration tests (not production schema).
CREATE TABLE IF NOT EXISTS migrate_integration_fixture (
    id SERIAL PRIMARY KEY
);
