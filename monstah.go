package ino

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log/slog"
	"net"
	"time"

	"github.com/ralreegorganon/nmeaais"
	"github.com/ralreegorganon/rudia"
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
	slog.Info("Decoding from source", "source", address)

	feedID, err := m.DB.GetFeedId(address)

	if err != nil {
		slog.Error("Error fetching feed id for upstream", "upstream", address, slog.Any("error", err))
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
		slog.Error("Error creating decoding loopback", slog.Any("error", err))
	}
	port := listener.Addr().String()
	err = listener.Close()
	if err != nil {
		slog.Error("Error swapping decoding loopback", slog.Any("error", err))
	}

	go m.r.ListenAndAcceptClients(port)
	go m.receive(port)
	go m.postprocess()
}

func (m *Monstah) Shutdown() {
	slog.Info("Shutting down decoder")
	close(m.d.Input)
	m.r.Shutdown()
}

func (m *Monstah) receive(address string) {
	retryInterval := 10 * time.Second
	fault := make(chan bool)
	for {
		slog.Info("Dialing upstream", "upstream", address)

		conn, err := net.Dial("tcp", address)
		if err != nil {
			slog.Error("Error dialing upstream", "upstream", address, slog.Any("error", err))
			slog.Info("Sleeping before retrying upstream", "upstream", address, "sleep", retryInterval)
			time.Sleep(retryInterval)
			continue
		}

		r := bufio.NewReader(conn)

		for {
			line, err := r.ReadString('\n')
			if err != nil {
				slog.Error("Couldn't read packet", slog.Any("error", err))
				fault <- true
				break
			}
			err = m.DB.AddPacket(line, m.feedID)
			if err != nil {
				slog.Error("Couldn't insert packet to database", slog.Any("error", err))
				continue
			}
			m.d.Input <- nmeaais.DecoderInput{
				Input:     line,
				Timestamp: time.Now(),
			}
		}
		<-fault
		conn.Close()
	}
}

func (m *Monstah) postprocess() {
	for o := range m.d.Output {
		if o.Error != nil {
			slog.Error("Couldn't decode message", "message", o.SourceMessage, slog.Any("error", o.Error))
			continue
		}

		message, err := json.Marshal(o.DecodedMessage)
		if err != nil {
			slog.Error("Couldn't marshal message", "message", o.DecodedMessage, slog.Any("error", err))
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
			slog.Error("Couldn't insert message to database", "message", message, slog.Any("error", err))
			continue
		}

		go m.DB.UpdateVessel(o)
		go m.DB.UpdatePosition(o)
	}
}
