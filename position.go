package ino

import (
	"log/slog"
	"time"

	"github.com/guregu/null/v5"
	"github.com/ralreegorganon/nmeaais"
)

type Position struct {
	MMSI      int64      `json:"mmsi" db:"mmsi"`
	Latitude  null.Float `json:"latitude" db:"latitude"`
	Longitude null.Float `json:"longitude" db:"longitude"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
}

func (db *DB) UpdatePosition(r nmeaais.DecoderOutput) {
	switch dm := r.DecodedMessage.(type) {
	case *nmeaais.PositionReportClassA:
		if dm.Latitude == 91 || dm.Longitude == 181 {
			break
		}
		err := db.UpdatePositionFromPositionReportClassA(dm)
		if err != nil {
			slog.Error("Couldn't update position from PositionReportClassA", slog.Any("error", err))
		}
	case *nmeaais.PositionReportClassBStandard:
		if dm.Latitude == 91 || dm.Longitude == 181 {
			break
		}
		err := db.UpdatePositionFromPositionReportClassBStandard(dm)
		if err != nil {
			slog.Error("Couldn't update vessel from PositionReportClassA", slog.Any("error", err))
		}
	default:
	}
}
