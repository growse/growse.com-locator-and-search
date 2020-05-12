
CREATE MATERIALIZED VIEW public.locations_distance_this_year
    TABLESPACE pg_default
AS
SELECT sum(a.distance) AS distance
FROM ( SELECT st_distance(locations.point, lag(locations.point, 1, locations.point) OVER (ORDER BY locations.devicetimestamp)) AS distance
       FROM locations
       WHERE date_part('year'::text, date(timezone('UTC'::text, locations.devicetimestamp))::timestamp without time zone) = date_part('year'::text, now())) a
WITH DATA;

ALTER TABLE public.locations_distance_this_year
    OWNER TO www_growse_com;

alter table public.locations drop column distance;