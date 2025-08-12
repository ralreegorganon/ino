# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is "ino", a Go application that processes NMEA AIS (Automatic Identification System) marine vessel data. The application receives AIS data streams from remote sources, decodes NMEA messages into structured data, and stores vessel positions, metadata, and message statistics in a PostgreSQL database with PostGIS extensions.

## Development Commands

### Build and Run
- `make build` - Build the binary for the current platform
- `make run` - Build and run locally (requires PostgreSQL running on localhost:5432)
- `make install` - Install the binary to GOPATH/bin

### Database Operations  
- `make migrate` - Run database migrations against local PostgreSQL using goose
- `make migrate-docker` - Run database migrations against dockerized PostgreSQL using goose

### Docker Operations
- `make docker` - Build Docker image
- `make run-docker` - Build and run the complete stack (app + PostgreSQL) in Docker
- `make stop-docker` - Stop Docker containers
- `make docker-logs` - View Docker container logs

### Environment Variables Required
- `INO_CONNECTION_STRING` - PostgreSQL connection string (format: `postgres://user:pass@host:port/db?sslmode=disable`)
- `INO_MIGRATIONS_PATH` - Path to migration files (typically `file://migrations`)

## Architecture

### Core Components

**Main Application (`cmd/ino/main.go`)**
- Entry point that initializes database, runs migrations, starts MonstahManager, and HTTP server
- Listens on port 8989 for HTTP API requests
- Handles graceful shutdown on interrupt signals

**MonstahManager (`monstah_manager.go`)**
- Manages multiple Monstah instances, one per configured feed
- Retrieves feed configurations from database and starts corresponding decoders

**Monstah (`monstah.go`)**
- Core AIS data processor that connects to remote AIS feeds
- Uses rudia library for TCP proxying and connection management
- Decodes NMEA AIS messages using nmeaais library
- Stores raw packets and decoded messages to database
- Updates vessel and position records asynchronously

**HTTP Server (`server.go`)**  
- REST API using chi router with CORS enabled
- Serves vessel data, positions, and message statistics
- Supports both JSON and GeoJSON output formats
- Routes: `/api/vessels`, `/api/vessels/{mmsi}`, `/api/vessels/{mmsi}/positions`, `/api/stats/message/*`

**Database Layer (`db.go`)**
- PostgreSQL with PostGIS for geographic data
- Handles vessel metadata, positions (with geographic indexing), messages, packets, and feeds
- Uses upsert patterns for vessel updates from different AIS message types
- Provides both structured and raw JSON query methods
- Database migrations managed by goose

### Data Flow
1. Remote AIS feeds → Monstah (via rudia proxy) → NMEA decoder → Database
2. Database → HTTP API → JSON/GeoJSON responses
3. Multiple concurrent feeds supported through MonstahManager

### Database Schema
Key tables: `vessel`, `position`, `message`, `packet`, `feed`
Views: `vessel_geojson`, `position_geojson`, `message_stats`, `message_stats_by_vessel`

## Testing

No test framework is currently configured. Tests should be implemented using Go's standard testing package.

## Dependencies

Key external libraries:
- `github.com/ralreegorganon/nmeaais` - NMEA AIS message decoding
- `github.com/ralreegorganon/rudia` - TCP connection proxying/management  
- `github.com/jmoiron/sqlx` - Extended SQL operations
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/pressly/goose/v3` - Database migrations