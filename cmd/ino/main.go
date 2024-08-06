package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ralreegorganon/ino"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var version = flag.Bool("version", false, "Print version")

func init() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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
		slog.Error("Couldn't connect to database", slog.Any("error", err))
		os.Exit(1)
	}

	migrationsPath := os.Getenv("INO_MIGRATIONS_PATH")
	g, err := migrate.New(migrationsPath, connectionString)
	if err != nil {
		time.Sleep(30 * time.Second)
		slog.Error("Couldn't create migrator", slog.Any("error", err))
		os.Exit(1)
	}

	if err = g.Up(); err != nil {
		if err != migrate.ErrNoChange {
			slog.Error("Couldn't migrate", slog.Any("error", err))
			os.Exit(1)
		} else {
			slog.Info("Migrations up to date")
		}
	}

	mm, err := ino.NewMonstahManager(&db)
	if err != nil {
		slog.Error("Couldn't create feed manager", slog.Any("error", err))
		os.Exit(1)
	}

	server := ino.NewHTTPServer(&db)
	router, err := ino.CreateRouter(server)
	if err != nil {
		slog.Error("Couldn't create router", slog.Any("error", err))
		os.Exit(1)
	}
	http.Handle("/", router)

	u := "0.0.0.0:8989"
	go http.ListenAndServe(u, nil)
	slog.Info("ino web server started", "address", u)

	<-interrupt
	mm.Shutdown()
}
