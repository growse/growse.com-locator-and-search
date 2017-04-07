CREATE EXTENSION IF NOT EXISTS postgis;
ALTER TABLE locations
  ADD point GEOGRAPHY(POINT, 4326),
  ADD gisdistance FLOAT;
UPDATE locations
SET point = ST_GeographyFromText('SRID=4326;POINT(' || longitude || ' ' || latitude || ')');
UPDATE locations
SET gisdistance = ST_DISTANCE(point, vtable.prevpoint) FROM (SELECT
                                                               id,
                                                               lag(point)
                                                               OVER (
                                                                 ORDER BY devicetimestamp ASC ) AS prevpoint
                                                             FROM locations) AS vtable
WHERE locations.id = vtable.id;
CREATE INDEX locations_point_idx
  ON locations USING GIST (point);