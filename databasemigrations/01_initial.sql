create table articles(
id serial not null,
datestamp timestamp with time zone not null default now(),
title character varying(255) not null,
shorttitle character varying(255) not null,
description text,
markdown text not null,
idxfti tsvector not null,
published bool not null,
searchtext text not null
);

create table locations (
id integer not null,
timestamp timestamp with time zone not null,
devicetimestamp timestamp with time zone not null,
latitude numeric(9,6) not null,
longitude numeric(9,6) not null,
accuracy numeric(12,6) not null,
distance numeric(12,3),
timedelta interval,
gsmtype character varying(32),
wifissd character varying(32),
geocoding text
)