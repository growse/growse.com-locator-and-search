
CREATE or replace FUNCTION public.location_update_distance_view()
    RETURNS trigger
    language plpgsql


AS $$
begin
    refresh materialized view concurrently public.locations_distance_this_year;
    return null;
end
$$;

ALTER FUNCTION public.location_update_distance_view()
    OWNER TO www_growse_com;


create trigger refresh_mat_view
    after insert or update or delete or truncate
    on locations for each statement
execute procedure public.location_update_distance_view();