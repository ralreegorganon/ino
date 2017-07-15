package ino

import (
	"encoding/json"
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

func (db *DB) AddPacket(raw string) error {
	_, err := db.Exec("insert into packet (raw) values ($1)", raw)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) AddMessage(r *nmeaais.DecoderResult) error {
	m, _ := json.Marshal(r.DecodedMessage)
	_, err := db.Exec("insert into message (mmsi, type, message) values ($1, $2, $3)", r.SourceMessage.MMSI, r.SourceMessage.MessageType, m)
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
