ALTER TABLE locations
  ADD CONSTRAINT unique_device_timestamps UNIQUE (devicetimestamp);
ALTER TABLE locations
  DROP COLUMN timedelta;
ALTER TABLE locations
  ALTER COLUMN geocoding DROP NOT NULL;