create view vessel_geojson as 
select
	row_to_json(fc) geojson
from
	(
		select
			'FeatureCollection' as type,
			array_to_json(array_agg(f)) as features
		from
		(
			select
				'Feature' as type,
				st_asgeojson(the_geog)::json as geometry,
				json_build_object(
					'mmsi', mmsi, 
					'vesselName', vessel_name,
					'callSign', call_sign,
					'shipType', ship_type,
					'length', length,
					'breadth', breadth,
					'draught', draught,
					'speedOverGround', speed_over_ground,
					'trueHeading', true_heading,
					'courseOverGround', course_over_ground,
					'navigationStatus', navigation_status,
					'updatedAt', updated_at
				) as properties
			from
				vessel as lg
		) as f
	) as fc;