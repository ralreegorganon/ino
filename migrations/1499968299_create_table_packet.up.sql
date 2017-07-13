create table packet
(
      packet_id serial not null,
      raw character varying not null,
      created_at timestamp with time zone not null default now(),
      constraint packet_pkey primary key (packet_id)
);
