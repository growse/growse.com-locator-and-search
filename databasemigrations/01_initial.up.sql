CREATE TABLE public.locations (
    id integer NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    devicetimestamp timestamp with time zone NOT NULL,
    latitude numeric(9,6) NOT NULL,
    longitude numeric(9,6) NOT NULL,
    accuracy numeric(12,6) NOT NULL,
    distance numeric(12,3),
    gsmtype character varying(32),
    wifissid character varying(32),
    geocoding jsonb,
    batterylevel integer,
    connectiontype character(1),
    doze boolean,
    point public.geography(Point,4326),
    gisdistance double precision
);


CREATE SEQUENCE public.locations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE ONLY public.locations ALTER COLUMN id SET DEFAULT nextval('public.locations_id_seq'::regclass);


--
-- Name: locations locations_pkey; Type: CONSTRAINT; Schema: public; Owner: www_growse_com
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_pkey PRIMARY KEY (id);


--
-- Name: locations unique_device_timestamps; Type: CONSTRAINT; Schema: public; Owner: www_growse_com
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT unique_device_timestamps UNIQUE (devicetimestamp);


--
-- Name: idx_locations_geocoding; Type: INDEX; Schema: public; Owner: www_growse_com
--

CREATE INDEX idx_locations_geocoding ON public.locations USING btree (geocoding);


--
-- Name: idx_locations_year; Type: INDEX; Schema: public; Owner: www_growse_com
--

CREATE INDEX idx_locations_year ON public.locations USING btree (date_part('year'::text, (date(timezone('UTC'::text, devicetimestamp)))::timestamp without time zone));


--
-- Name: locations_point_idx; Type: INDEX; Schema: public; Owner: www_growse_com
--

CREATE INDEX idx_locations_point ON public.locations USING gist (point);


--
-- Name: locations_timestamp_idx; Type: INDEX; Schema: public; Owner: www_growse_com
--

CREATE INDEX idx_locations_timestamp ON public.locations USING btree ("timestamp");

