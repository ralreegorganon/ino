create view message_stats as
select	
	type,
	count(1) count,
	min(created_at) as first,
	max(created_at) as last,
	now() - max(created_at) as ago
from
	message
group by	
	type
order by
	type
