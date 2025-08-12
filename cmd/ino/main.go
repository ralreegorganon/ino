package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ralreegorganon/ino"

	"github.com/pressly/goose/v3"
	_ "github.com/lib/pq"
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
	if migrationsPath == "" {
		migrationsPath = "migrations"
	} else if strings.HasPrefix(migrationsPath, "file://") {
		migrationsPath = strings.TrimPrefix(migrationsPath, "file://")
	}

	goose.SetDialect("postgres")
	if err := goose.Up(db.DB.DB, migrationsPath); err != nil {
		time.Sleep(30 * time.Second)
		slog.Error("Couldn't run migrations", slog.Any("error", err))
		os.Exit(1)
	} else {
		slog.Info("Migrations completed successfully")
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
