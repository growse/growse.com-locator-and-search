update locations set geocoding='{}' where geocoding ='';
alter table locations add column json_geocoding jsonb;
update locations set json_geocoding=geocoding::jsonb;
alter table locations drop column geocoding;
alter table locations rename json_geocoding to geocoding;
alter table locations alter column geocoding set not null;
