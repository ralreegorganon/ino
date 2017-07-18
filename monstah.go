package ino

import (
	"bufio"
	"encoding/json"
	"net"
	"time"

	"bytes"

	"github.com/ralreegorganon/nmeaais"
	log "github.com/sirupsen/logrus"
)

type Monstah struct {
	d  *nmeaais.Decoder
	DB *DB
}

func NewMonstah(db *DB) *Monstah {
	m := &Monstah{
		d:  nmeaais.NewDecoder(),
		DB: db,
	}
	return m
}

func (m *Monstah) Decode(address string) {
	go m.receive(address)
	go m.postprocess()
}

func (m *Monstah) Shutdown() {
	log.Info("Shutting down decoder")
	close(m.d.Input)
}

func (m *Monstah) receive(address string) {
	retryInterval := 10 * time.Second
	fault := make(chan bool)
	for {
		log.WithFields(log.Fields{
			"upstream": address,
		}).Info("Dialing upstream")

		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.WithFields(log.Fields{
				"upstream": address,
				"err":      err,
			}).Error("Error dialing upstream")

			log.WithFields(log.Fields{
				"upstream": address,
				"sleep":    retryInterval,
			}).Info("Sleeping before retrying upstream")

			time.Sleep(retryInterval)
			continue
		}

		r := bufio.NewReader(conn)

		for {
			line, err := r.ReadString('\n')
			if err != nil {
				log.WithField("err", err).Error("Couldn't read packet")
				fault <- true
				break
			}
			err = m.DB.AddPacket(line)
			if err != nil {
				log.WithField("err", err).Error("Couldn't insert packet to database")
				continue
			}
			m.d.Input <- line
		}
		_ = <-fault
		conn.Close()
	}
}

func (m *Monstah) postprocess() {
	for o := range m.d.Output {
		if o.Error != nil {
			log.WithFields(log.Fields{
				"message": o.SourceMessage,
				"err":     o.Error,
			}).Error("Couldn't decode message")
			continue
		}

		message, err := json.Marshal(o.DecodedMessage)
		if err != nil {
			log.WithFields(log.Fields{
				"message": o.DecodedMessage,
				"err":     err,
			}).Error("Couldn't marshal message")
			message = []byte("{}")
		}

		var rawBuf bytes.Buffer
		length := len(o.SourcePackets)
		for i, p := range o.SourcePackets {
			rawBuf.WriteString(p.Raw)
			if i+1 != length {
				rawBuf.WriteString("\n")
			}
		}

		raw := rawBuf.Bytes()

		err = m.DB.AddMessage(o.SourceMessage.MMSI, o.SourceMessage.MessageType, message, raw)
		if err != nil {
			log.WithFields(log.Fields{
				"message": message,
				"err":     err,
			}).Error("Couldn't insert message to database")
			continue
		}

		go m.DB.UpdateVessel(o)
		go m.DB.UpdatePosition(o)
	}
}
