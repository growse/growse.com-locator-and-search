ALTER TABLE public.locations
    ADD CONSTRAINT locations_unique_point_devicetimestamp UNIQUE (point, devicetimestamp);