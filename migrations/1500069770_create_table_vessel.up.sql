create table vessel
(
      mmsi int not null,
      vessel_name character varying, 
      call_sign character varying, 
      ship_type character varying,
      length int,
      breadth int,
      draught real, 
      latitude double precision,
      longitude double precision,
      speed_over_ground real, 
      true_heading real,
      course_over_ground real,
      navigation_status character varying,
      destination character varying,
      the_geog geography(POINT,4326),
      updated_at timestamp with time zone not null,
      constraint vessel_pkey primary key (mmsi)
);
