create table message
(
      message_id serial not null,
      mmsi int not null,
      type int not null,
      message jsonb not null,
      created_at timestamp with time zone not null default now(),
      constraint message_pkey primary key (message_id)
);
