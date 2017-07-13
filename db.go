package ino

import (
	"encoding/json"

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
