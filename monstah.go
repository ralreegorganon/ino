package ino

import (
	"bufio"
	"encoding/json"
	"net"
	"time"

	"bytes"

	"github.com/ralreegorganon/nmeaais"
	"github.com/ralreegorganon/rudia"
	log "github.com/sirupsen/logrus"
)

type Monstah struct {
	feedID int
	r      *rudia.Repeater
	d      *nmeaais.Decoder
	DB     *DB
}

func NewMonstah(db *DB) *Monstah {
	m := &Monstah{
		r: rudia.NewRepeater(&rudia.RepeaterOptions{
			UpstreamProxyIdleTimeout:    time.Duration(600) * time.Second,
			UpstreamListenerIdleTimeout: time.Duration(600) * time.Second,
			RetryInterval:               time.Duration(10) * time.Second,
		}),
		d:  nmeaais.NewDecoder(),
		DB: db,
	}
	return m
}

func (m *Monstah) Decode(address string) {
	log.WithFields(log.Fields{
		"source": address,
	}).Info("Decoding from source")

	feedID, err := m.DB.GetFeedId(address)

	if err != nil {
		log.WithFields(log.Fields{
			"upstream": address,
			"err":      err,
		}).Error("Error fetching feed id for upstream")
		return
	}

	m.feedID = feedID

	m.r.Proxy(address)

	// This is really dumb but I don't want to rework things
	// so that rudia accept a listener, so in order to get a
	// random port I'm just letting the OS pick one, capturing
	// it, closing it, and reusing it. Totally fine! :/
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error creating decoding loopback")
	}
	port := listener.Addr().String()
	err = listener.Close()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error swapping decoding loopback")
	}

	go m.r.ListenAndAcceptClients(port)
	go m.receive(port)
	go m.postprocess()
}

func (m *Monstah) Shutdown() {
	log.Info("Shutting down decoder")
	close(m.d.Input)
	m.r.Shutdown()
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
			err = m.DB.AddPacket(line, m.feedID)
			if err != nil {
				log.WithField("err", err).Error("Couldn't insert packet to database")
				continue
			}
			m.d.Input <- nmeaais.DecoderInput{
				Input:     line,
				Timestamp: time.Now(),
			}
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

		err = m.DB.AddMessage(o.SourceMessage.MMSI, o.SourceMessage.MessageType, message, raw, m.feedID)
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
