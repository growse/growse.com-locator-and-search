drop index location_year_idx if exists;
create index location_year_idx on locations (date_part('year'::text, date(devicetimestamp at time zone 'UTC')));