DROP INDEX IF EXISTS idx_settings_scope;
DROP INDEX IF EXISTS idx_settings_scope_key;

ALTER TABLE settings
ADD CONSTRAINT settings_key_key UNIQUE (key);

ALTER TABLE settings
DROP COLUMN IF EXISTS scope;

