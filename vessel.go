package ino

import (
	"log/slog"
	"time"

	"github.com/guregu/null/v5"
	"github.com/ralreegorganon/nmeaais"
)

type Vessel struct {
	MMSI             int64       `json:"mmsi" db:"mmsi"`
	VesselName       null.String `json:"vesselName" db:"vessel_name"`
	CallSign         null.String `json:"callSign" db:"call_sign"`
	ShipType         null.String `json:"shipType" db:"ship_type"`
	Length           null.Int    `json:"length" db:"length"`
	Breadth          null.Int    `json:"breadth" db:"breadth"`
	Draught          null.Float  `json:"draught" db:"draught"`
	Latitude         null.Float  `json:"latitude" db:"latitude"`
	Longitude        null.Float  `json:"longitude" db:"longitude"`
	SpeedOverGround  null.Float  `json:"speedOverGround" db:"speed_over_ground"`
	TrueHeading      null.Float  `json:"trueHeading" db:"true_heading"`
	CourseOverGround null.Float  `json:"courseOverGround" db:"course_over_ground"`
	NavigationStatus null.String `json:"navigationStatus" db:"navigation_status"`
	Destination      null.String `json:"destination" db:"destination"`
	UpdatedAt        time.Time   `json:"updatedAt" db:"updated_at"`
}

func (db *DB) UpdateVessel(r nmeaais.DecoderOutput) {
	switch dm := r.DecodedMessage.(type) {
	case *nmeaais.PositionReportClassA:
		err := db.UpdateVesselFromPositionReportClassA(dm)
		if err != nil {
			slog.Error("Couldn't update vessel from PositionReportClassA", slog.Any("error", err))
		}
	case *nmeaais.StaticAndVoyageRelatedData:
		err := db.UpdateVesselFromStaticAndVoyageRelatedData(dm)
		if err != nil {
			slog.Error("Couldn't update vessel from StaticAndVoyageRelatedData", slog.Any("error", err))
		}
	case *nmeaais.PositionReportClassBStandard:
		err := db.UpdateVesselFromPositionReportClassBStandard(dm)
		if err != nil {
			slog.Error("Couldn't update vessel from PositionReportClassBStandard", slog.Any("error", err))
		}
	case *nmeaais.StaticDataReportA:
		err := db.UpdateVesselFromStaticDataReportA(dm)
		if err != nil {
			slog.Error("Couldn't update vessel from StaticDataReportA", slog.Any("error", err))
		}
	case *nmeaais.StaticDataReportB:
		err := db.UpdateVesselFromStaticDataReportB(dm)
		if err != nil {
			slog.Error("Couldn't update vessel from StaticDataReportB", slog.Any("error", err))
		}
	default:
	}
}
