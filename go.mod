module github.com/ralreegorganon/ino

go 1.24.5

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-chi/cors v1.2.1
	github.com/guregu/null/v5 v5.0.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/pressly/goose/v3 v3.17.0
	github.com/ralreegorganon/nmeaais v0.0.0-20220615002720-1ebe8027bc2b
	github.com/ralreegorganon/rudia v0.0.0-20180322183600-34c80165b6cb
)

require (
	github.com/docker/docker v27.1.1+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/sethvargo/go-retry v0.2.4 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	go.opentelemetry.io/otel/trace v1.28.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
)

replace github.com/ralreegorganon/nmeaais => ../nmeaais

replace github.com/ralreegorganon/rudia => ../rudia
