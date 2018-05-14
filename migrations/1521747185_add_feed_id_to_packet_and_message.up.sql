alter table message add column feed_id integer references feed (feed_id);
alter table packet add column feed_id integer references feed (feed_id);

update message set feed_id = (select feed_id from feed where remote_address = 'ais1.shipraiser.net:6494');
update packet set feed_id = (select feed_id from feed where remote_address = 'ais1.shipraiser.net:6494');

alter table message alter column feed_id set not null;
alter table packet alter column feed_id set not null;
