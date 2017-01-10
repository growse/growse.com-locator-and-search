alter table locations add constraint unique_device_timestamps unique(devicetimestamp);
alter table locations drop column timedelta;
alter table locations alter column geocoding drop not null;