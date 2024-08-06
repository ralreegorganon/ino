package ino

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func CreateRouter(server *HTTPServer) (*chi.Mux, error) {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	m := map[string]map[string]HTTPApiFunc{
		"GET": {
			"/api/vessels":                             server.GetVessels,
			"/api/vessels/{mmsi:[0-9]+}":               server.GetVesselByMmsi,
			"/api/vessels/{mmsi:[0-9]+}/positions":     server.GetPositionsForVessel,
			"/api/stats/message":                       server.GetMessageStats,
			"/api/stats/message/vessels":               server.GetMessageStatsByVessel,
			"/api/stats/message/{type:[0-9]+}/vessels": server.GetMessageStatsByVesselForType,
			"/api/stats/message/vessels/{mmsi:[0-9]+}": server.GetMessageStatsByVesselForVessel,
		},
		"PUT": {},
		"OPTIONS": {
			"/": options,
		},
	}

	for method, routes := range m {
		for route, handler := range routes {
			localRoute := route
			localHandler := handler
			localMethod := method
			f := makeHTTPHandler(localMethod, localRoute, localHandler)
			r.Method(localMethod, localRoute, f)
		}
	}
	return r, nil
}

func makeHTTPHandler(_ string, _ string, handlerFunc HTTPApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handlerFunc(w, r); err != nil {
			httpError(w, err)
		}
	}
}

type HTTPApiFunc func(w http.ResponseWriter, r *http.Request) error

type HTTPServer struct {
	DB *DB
}

func NewHTTPServer(db *DB) *HTTPServer {
	s := &HTTPServer{
		DB: db,
	}

	return s
}

func writeJSON(w http.ResponseWriter, code int, thing interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	val, err := json.Marshal(thing)
	w.Write(val)
	return err
}

func writeGeoJSON(w http.ResponseWriter, code int, thing []byte) {
	w.Header().Set("Content-Type", "application/vnd.geo+json")
	w.WriteHeader(code)
	w.Write(thing)
}

func writeJSONDirect(w http.ResponseWriter, code int, thing []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(thing)
}

func httpError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError

	if err != nil {
		slog.Error("http error", slog.Any("error", err))
		http.Error(w, err.Error(), statusCode)
	}
}

func options(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *HTTPServer) GetVessels(w http.ResponseWriter, r *http.Request) error {
	query := r.URL.Query()
	format := query.Get("f")

	if format == "geojson" {
		geojson, err := s.DB.GetVesselsGeojson()
		if err != nil {
			return err
		}

		writeGeoJSON(w, http.StatusOK, geojson)
	} else {
		vessels, err := s.DB.GetVessels()
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, vessels)
	}

	return nil
}

func (s *HTTPServer) GetVesselByMmsi(w http.ResponseWriter, r *http.Request) error {
	mmsi, err := strconv.Atoi(chi.URLParam(r, "mmsi"))
	if err != nil {
		return err
	}
	vessel, err := s.DB.GetVessel(mmsi)
	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, vessel)

	return nil
}

func (s *HTTPServer) GetPositionsForVessel(w http.ResponseWriter, r *http.Request) error {
	mmsi, err := strconv.Atoi(chi.URLParam(r, "mmsi"))
	if err != nil {
		return err
	}

	query := r.URL.Query()
	format := query.Get("f")

	if format == "geojson" {
		geojson, err := s.DB.GetPositionsForVesselGeojson(mmsi)
		if err != nil {
			return err
		}

		writeGeoJSON(w, http.StatusOK, geojson)
	} else {
		positions, err := s.DB.GetPositionsForVessel(mmsi)
		if err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, positions)
	}

	return nil
}

func (s *HTTPServer) GetMessageStats(w http.ResponseWriter, r *http.Request) error {
	json, err := s.DB.GetMessageStatsJSON()
	if err != nil {
		return err
	}
	writeJSONDirect(w, http.StatusOK, json)
	return nil
}

func (s *HTTPServer) GetMessageStatsByVessel(w http.ResponseWriter, r *http.Request) error {
	json, err := s.DB.GetMessageStatsByVesselJSON()
	if err != nil {
		return err
	}
	writeJSONDirect(w, http.StatusOK, json)
	return nil
}

func (s *HTTPServer) GetMessageStatsByVesselForType(w http.ResponseWriter, r *http.Request) error {
	messageType, err := strconv.Atoi(chi.URLParam(r, "type"))
	if err != nil {
		return err
	}

	json, err := s.DB.GetMessageStatsByVesselForTypeJSON(messageType)
	if err != nil {
		return err
	}
	writeJSONDirect(w, http.StatusOK, json)
	return nil
}

func (s *HTTPServer) GetMessageStatsByVesselForVessel(w http.ResponseWriter, r *http.Request) error {
	mmsi, err := strconv.Atoi(chi.URLParam(r, "mmsi"))
	if err != nil {
		return err
	}

	json, err := s.DB.GetMessageStatsByVesselForVesselJSON(mmsi)
	if err != nil {
		return err
	}
	writeJSONDirect(w, http.StatusOK, json)
	return nil
}
