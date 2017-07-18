create view position_geojson as 
with
mmsi_lines as
(
	select
		mmsi, 
		st_segmentize(st_makeline(the_geog::geometry order by created_at),100)  the_geom
	from
		position
	group by
		mmsi
)
select
	mmsi, 
	json_build_object(
		'type', 'FeatureCollection',
		'features', json_agg(json_build_object(
			'type', 'Feature',
			'geometry', st_asgeojson(the_geom)::json,
			'properties',json_build_object(
				'mmsi', mmsi
			)
		))
	) geojson
from 
	mmsi_lines
group by
	mmsi