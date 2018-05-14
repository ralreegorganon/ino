create table feed
(
    feed_id serial not null,
    remote_address character varying not null,
    active boolean not null default true,
    created_at timestamp with time zone not null default now(),
    constraint feed_pkey primary key (feed_id)
);

insert into feed (remote_address) values ('ais1.shipraiser.net:6494'); 