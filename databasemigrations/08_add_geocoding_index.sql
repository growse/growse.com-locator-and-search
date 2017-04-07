CREATE INDEX idx_locations_geocoding
  ON locations (geocoding);
CREATE INDEX idx_locations_year
  ON locations (date_part('year' :: TEXT,
                          (date(timezone('UTC' :: TEXT, devicetimestamp))) :: TIMESTAMP WITHOUT TIME ZONE));