ALTER TABLE locations
  ADD kalmanlatitude NUMERIC(9, 6),
  ADD kalmanlongitude NUMERIC(9, 6),
  ADD kalmandistance NUMERIC(12, 3),
  ADD kalmanaccuracy NUMERIC(12, 6)
