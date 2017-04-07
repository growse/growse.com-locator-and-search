UPDATE locations
SET geocoding = '{}'
WHERE geocoding = '';
ALTER TABLE locations
  ADD COLUMN json_geocoding JSONB;
UPDATE locations
SET json_geocoding = geocoding :: JSONB;
ALTER TABLE locations
  DROP COLUMN geocoding;
ALTER TABLE locations
  RENAME json_geocoding TO geocoding;
ALTER TABLE locations
  ALTER COLUMN geocoding SET NOT NULL;
