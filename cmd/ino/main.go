package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ralreegorganon/ino"
	log "github.com/sirupsen/logrus"

	"github.com/mattes/migrate"
	_ "github.com/mattes/migrate/database/postgres"
	_ "github.com/mattes/migrate/source/file"
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

	migrationsPath := os.Getenv("INO_MIGRATIONS_PATH")
	g, err := migrate.New(migrationsPath, connectionString)
	if err != nil {
		time.Sleep(30 * time.Second)
		log.Fatal("Couldn't create migrator: ", err)
	}

	if err = g.Up(); err != nil {
		if err != migrate.ErrNoChange {
			log.Fatal(err)
		} else {
			log.Info("Migrations up to date")
		}
	}

	mm := ino.NewMonstahManager(&db)

	server := ino.NewHTTPServer(&db)
	router, err := ino.CreateRouter(server)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", router)

	u := "0.0.0.0:8989"
	go http.ListenAndServe(u, nil)
	log.WithField("address", u).Info("ino web server started")

	<-interrupt
	mm.Shutdown()
}
