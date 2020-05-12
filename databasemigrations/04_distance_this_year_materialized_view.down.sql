DROP MATERIALIZED VIEW public.locations_distance_this_year;

alter table locations add column distance distance numeric(12,3);