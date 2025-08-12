-- +goose Up
create view message_stats_by_vessel as
select
    mmsi,
    type,
    count(*) as count,
    min(created_at) as first,
    max(created_at) as last,
    now() - max(created_at) as ago
from
    message
group by
    mmsi,
    type
order by
    mmsi,
    type;

-- +goose Down
drop view message_stats_by_vessel;
