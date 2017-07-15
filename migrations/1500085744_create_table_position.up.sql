create table position
(
    position_id serial not null,
    mmsi int not null,
    latitude double precision not null,
    longitude double precision not null,
    the_geog geography(POINT,4326) not null,
    created_at timestamp with time zone not null default now(),
    constraint position_pkey primary key (position_id)
);
