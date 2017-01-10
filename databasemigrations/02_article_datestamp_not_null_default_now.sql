alter table articles alter column datestamp set not null;
alter table articles alter column datestamp set default now();
