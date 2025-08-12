-- +goose Up
create view message_stats as
select
    type,
    count(*) as count,
    min(created_at) as first,
    max(created_at) as last,
    now() - max(created_at) as ago
from
    message
group by
    type
order by
    type;

-- +goose Down
drop view message_stats;
