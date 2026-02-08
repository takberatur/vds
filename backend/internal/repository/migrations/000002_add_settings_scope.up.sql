ALTER TABLE settings
ADD COLUMN IF NOT EXISTS scope TEXT NOT NULL DEFAULT 'default';

ALTER TABLE settings
DROP CONSTRAINT IF EXISTS settings_key_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_settings_scope_key
ON settings (scope, key);

CREATE INDEX IF NOT EXISTS idx_settings_scope
ON settings (scope);

