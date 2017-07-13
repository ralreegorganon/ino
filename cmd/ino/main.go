package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/ralreegorganon/ino"
	"github.com/ralreegorganon/rudia"
	log "github.com/sirupsen/logrus"
)

var version = flag.Bool("version", false, "Print version")

func init() {
	f := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(f)
}

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("Version: %s - Commit: %s - Date: %s\n", Version, GitCommit, BuildDate)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	connectionString := os.Getenv("INO_CONNECTION_STRING")
	var db ino.DB
	if err := db.Open(connectionString); err != nil {
		log.Fatal(err)
	}

	ro := &rudia.RepeaterOptions{
		UpstreamProxyIdleTimeout:    time.Duration(600) * time.Second,
		UpstreamListenerIdleTimeout: time.Duration(600) * time.Second,
		RetryInterval:               time.Duration(10) * time.Second,
	}
	r := rudia.NewRepeater(ro)
	m := ino.NewMonstah(&db)
	r.Proxy("ais1.shipraiser.net:6494")

	port := ":33000"
	go r.ListenAndAcceptClients(port)
	go m.Decode(port)

	<-interrupt
	m.Shutdown()
	r.Shutdown()
}
