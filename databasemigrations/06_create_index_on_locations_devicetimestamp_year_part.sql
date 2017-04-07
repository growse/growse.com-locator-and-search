DROP INDEX IF EXISTS location_year_idx;
CREATE INDEX location_year_idx
  ON locations (date_part('year' :: TEXT, date(devicetimestamp AT TIME ZONE 'UTC')));