CREATE TABLE articles (
  id          SERIAL                   NOT NULL,
  datestamp   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  title       CHARACTER VARYING(255)   NOT NULL,
  shorttitle  CHARACTER VARYING(255)   NOT NULL,
  description TEXT,
  markdown    TEXT                     NOT NULL,
  idxfti      TSVECTOR                 NOT NULL,
  published   BOOL                     NOT NULL,
  searchtext  TEXT                     NOT NULL
);

CREATE TABLE locations (
  id              INTEGER                  NOT NULL,
  timestamp       TIMESTAMP WITH TIME ZONE NOT NULL,
  devicetimestamp TIMESTAMP WITH TIME ZONE NOT NULL,
  latitude        NUMERIC(9, 6)            NOT NULL,
  longitude       NUMERIC(9, 6)            NOT NULL,
  accuracy        NUMERIC(12, 6)           NOT NULL,
  distance        NUMERIC(12, 3),
  timedelta       INTERVAL,
  gsmtype         CHARACTER VARYING(32),
  wifissd         CHARACTER VARYING(32),
  geocoding       TEXT
)