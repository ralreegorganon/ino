package ino

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ralreegorganon/nmeaais"
)

type DB struct {
	*sqlx.DB
}

func (db *DB) Open(connectionString string) error {
	d, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return err
	}
	db.DB = d
	return nil
}

func (db *DB) AddPacket(raw string, feedID int) error {
	_, err := db.Exec("insert into packet (raw, feed_id) values ($1, $2)", raw, feedID)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) AddMessage(mmsi int64, messageType int64, message []byte, raw []byte, feedID int) error {
	_, err := db.Exec("insert into message (mmsi, type, message, raw, feed_id) values ($1, $2, $3, $4, $5)", mmsi, messageType, message, raw, feedID)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetVessels() ([]*Vessel, error) {
	vessels := []*Vessel{}
	err := db.Select(&vessels, `
		select 
			mmsi, 
			vessel_name, 
			call_sign, 
			ship_type, 
			length, 
			breadth, 
			draught, 
			latitude, 
			longitude, 
			speed_over_ground, 
			true_heading, 
			course_over_ground, 
			navigation_status, 
			destination, 
			updated_at 
		from 
			vessel
	`)
	if err != nil {
		return nil, err
	}
	return vessels, nil
}

func (db *DB) GetVesselsGeojson() ([]byte, error) {
	var geojson []byte
	err := db.QueryRow("select geojson from vessel_geojson").Scan(&geojson)
	if err != nil {
		return nil, err
	}
	return geojson, nil
}

func (db *DB) GetVessel(mmsi int) (*Vessel, error) {
	vessel := &Vessel{}
	err := db.Get(vessel, `
		select 
			mmsi, 
			vessel_name, 
			call_sign, 
			ship_type, 
			length, 
			breadth, 
			draught, 
			latitude, 
			longitude, 
			speed_over_ground, 
			true_heading, 
			course_over_ground, 
			navigation_status, 
			destination, 
			updated_at 
		from 
			vessel
		where
			mmsi = $1
	`, mmsi)
	if err != nil {
		return nil, err
	}
	return vessel, nil
}

func (db *DB) GetPositionsForVessel(mmsi int) ([]*Position, error) {
	positions := []*Position{}
	err := db.Select(&positions, `
		select 
			mmsi, 
			latitude, 
			longitude, 
			created_at 
		from 
			position
		where
			mmsi = $1
		order by created_at desc
	`, mmsi)
	if err != nil {
		return nil, err
	}
	return positions, nil
}

func (db *DB) GetPositionsForVesselGeojson(mmsi int) ([]byte, error) {
	var geojson []byte
	err := db.QueryRow("select geojson from position_geojson where mmsi = $1", mmsi).Scan(&geojson)
	if err != nil {
		return nil, err
	}
	return geojson, nil
}

func (db *DB) UpdateVesselFromPositionReportClassA(m *nmeaais.PositionReportClassA) error {
	sql := fmt.Sprintf(`
	insert into vessel 
	(mmsi, latitude, longitude, speed_over_ground, true_heading, course_over_ground, navigation_status, the_geog, updated_at)
	values 
	($1, $2, $3, $4, $5, $6, $7, ST_GeographyFromText('SRID=4326;POINT(%[1]f %[2]f)'), now())
	on conflict (mmsi)
	do update set
		latitude = EXCLUDED.latitude,
		longitude = EXCLUDED.longitude,
		speed_over_ground = EXCLUDED.speed_over_ground,
		true_heading = EXCLUDED.true_heading,
		course_over_ground = EXCLUDED.course_over_ground,
		navigation_status = EXCLUDED.navigation_status,
		the_geog = EXCLUDED.the_geog,
		updated_at = EXCLUDED.updated_at
	`, m.Longitude, m.Latitude)

	_, err := db.Exec(sql, m.MMSI, m.Latitude, m.Longitude, m.SpeedOverGround, m.TrueHeading, m.CourseOverGround, m.NavigationStatus)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdateVesselFromStaticAndVoyageRelatedData(m *nmeaais.StaticAndVoyageRelatedData) error {
	sql := `
	insert into vessel 
	(mmsi, vessel_name, call_sign, ship_type, length, breadth, draught, destination, updated_at)
	values 
	($1, $2, $3, $4, $5, $6, $7, $8, now())
	on conflict (mmsi)
	do update set
		vessel_name = EXCLUDED.vessel_name,
		call_sign = EXCLUDED.call_sign,
		ship_type = EXCLUDED.ship_type,
		length = EXCLUDED.length,
		breadth = EXCLUDED.breadth,
		draught = EXCLUDED.draught,
		destination = EXCLUDED.destination,
		updated_at = EXCLUDED.updated_at
	`
	_, err := db.Exec(sql, m.MMSI, m.VesselName, m.CallSign, m.ShipType, m.DimensionToBow+m.DimensionToStern, m.DimensionToPort+m.DimensionToStarboard, m.Draught, m.Destination)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdateVesselFromPositionReportClassBStandard(m *nmeaais.PositionReportClassBStandard) error {
	sql := fmt.Sprintf(`
	insert into vessel 
	(mmsi, latitude, longitude, speed_over_ground, true_heading, course_over_ground, the_geog, updated_at)
	values 
	($1, $2, $3, $4, $5, $6, ST_GeographyFromText('SRID=4326;POINT(%[1]f %[2]f)'), now())
	on conflict (mmsi)
	do update set
		latitude = EXCLUDED.latitude,
		longitude = EXCLUDED.longitude,
		speed_over_ground = EXCLUDED.speed_over_ground,
		true_heading = EXCLUDED.true_heading,
		course_over_ground = EXCLUDED.course_over_ground,
		the_geog = EXCLUDED.the_geog,
		updated_at = EXCLUDED.updated_at
	`, m.Longitude, m.Latitude)

	_, err := db.Exec(sql, m.MMSI, m.Latitude, m.Longitude, m.SpeedOverGround, m.TrueHeading, m.CourseOverGround)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdateVesselFromStaticDataReportA(m *nmeaais.StaticDataReportA) error {
	sql := `
	insert into vessel 
	(mmsi, vessel_name, updated_at)
	values 
	($1, $2, now())
	on conflict (mmsi)
	do update set
		vessel_name = EXCLUDED.vessel_name,
		updated_at = EXCLUDED.updated_at
	`
	_, err := db.Exec(sql, m.MMSI, m.VesselName)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdateVesselFromStaticDataReportB(m *nmeaais.StaticDataReportB) error {
	sql := `
	insert into vessel 
	(mmsi, call_sign, ship_type, length, breadth, updated_at)
	values 
	($1, $2, $3, $4, $5, now())
	on conflict (mmsi)
	do update set
		call_sign = EXCLUDED.call_sign,
		ship_type = EXCLUDED.ship_type,
		length = EXCLUDED.length,
		breadth = EXCLUDED.breadth,
		draught = EXCLUDED.draught,
		updated_at = EXCLUDED.updated_at
	`
	_, err := db.Exec(sql, m.MMSI, m.CallSign, m.ShipType, m.DimensionToBow+m.DimensionToStern, m.DimensionToPort+m.DimensionToStarboard)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdatePositionFromPositionReportClassA(m *nmeaais.PositionReportClassA) error {
	sql := fmt.Sprintf(`
	insert into position 
	(mmsi, latitude, longitude, the_geog, created_at)
	values 
	($1, $2, $3, ST_GeographyFromText('SRID=4326;POINT(%[1]f %[2]f)'), now())
	`, m.Longitude, m.Latitude)

	_, err := db.Exec(sql, m.MMSI, m.Latitude, m.Longitude)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) UpdatePositionFromPositionReportClassBStandard(m *nmeaais.PositionReportClassBStandard) error {
	sql := fmt.Sprintf(`
	insert into position 
	(mmsi, latitude, longitude, the_geog, created_at)
	values 
	($1, $2, $3, ST_GeographyFromText('SRID=4326;POINT(%[1]f %[2]f)'), now())
	`, m.Longitude, m.Latitude)

	_, err := db.Exec(sql, m.MMSI, m.Latitude, m.Longitude)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetMessageStatsJson() ([]byte, error) {
	var json []byte
	err := db.QueryRow("select json_agg(message_stats) json from message_stats").Scan(&json)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func (db *DB) GetMessageStatsByVesselForTypeJson(messageType int) ([]byte, error) {
	var json []byte
	err := db.QueryRow("select json_agg(message_stats_by_vessel) json from message_stats_by_vessel where type = $1", messageType).Scan(&json)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func (db *DB) GetMessageStatsByVesselJson() ([]byte, error) {
	var json []byte
	err := db.QueryRow("select json_agg(message_stats_by_vessel) json from message_stats_by_vessel").Scan(&json)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func (db *DB) GetMessageStatsByVesselForVesselJson(mmsi int) ([]byte, error) {
	var json []byte
	err := db.QueryRow("select json_agg(message_stats_by_vessel) json from message_stats_by_vessel where mmsi = $1", mmsi).Scan(&json)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func (db *DB) GetFeedId(address string) (int, error) {
	var feedIDs []int
	err := db.Select(&feedIDs, "select feed_id from feed where remote_address = $1", address)
	if err != nil {
		return 0, err
	}

	c := len(feedIDs)

	switch c {
	case 1:
		return feedIDs[0], nil
	case 0:
		var feedID int
		err = db.QueryRow("insert into feed (remote_address) values ($1) returning feed_id", address).Scan(&feedID)
		if err != nil {
			return 0, err
		}
		return feedID, nil
	default:
		return 0, errors.New("ino: feed couldn't be found or created")
	}
}

func (db *DB) GetFeeds() ([]*Feed, error) {
	feeds := []*Feed{}
	err := db.Select(&feeds, `
		select 
			feed_id, 
			remote_address,
			active,
			created_at 
		from 
			feed
	`)
	if err != nil {
		return nil, err
	}
	return feeds, nil
}
