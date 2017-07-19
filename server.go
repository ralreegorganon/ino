package ino

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func CreateRouter(server *HTTPServer) (*mux.Router, error) {
	r := mux.NewRouter()
	m := map[string]map[string]HttpApiFunc{
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
			"": options,
		},
	}

	for method, routes := range m {
		for route, handler := range routes {
			localRoute := route
			localHandler := handler
			localMethod := method
			f := makeHttpHandler(localMethod, localRoute, localHandler)

			if localRoute == "" {
				r.Methods(localMethod).HandlerFunc(f)
			} else {
				r.Path(localRoute).Methods(localMethod).HandlerFunc(f)
			}
		}
	}

	return r, nil
}

func makeHttpHandler(localMethod string, localRoute string, handlerFunc HttpApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeCorsHeaders(w, r)
		if err := handlerFunc(w, r, mux.Vars(r)); err != nil {
			httpError(w, err)
		}
	}
}

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

type HttpApiFunc func(w http.ResponseWriter, r *http.Request, vars map[string]string) error

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
		log.WithField("err", err).Error("http error")
		http.Error(w, err.Error(), statusCode)
	}
}

func options(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *HTTPServer) GetVessels(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
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

func (s *HTTPServer) GetVesselByMmsi(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	mmsi, err := strconv.Atoi(vars["mmsi"])
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

func (s *HTTPServer) GetPositionsForVessel(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	mmsi, err := strconv.Atoi(vars["mmsi"])
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

func (s *HTTPServer) GetMessageStats(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	json, err := s.DB.GetMessageStatsJson()
	if err != nil {
		return err
	}
	writeJSONDirect(w, http.StatusOK, json)
	return nil
}

func (s *HTTPServer) GetMessageStatsByVessel(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	json, err := s.DB.GetMessageStatsByVesselJson()
	if err != nil {
		return err
	}
	writeJSONDirect(w, http.StatusOK, json)
	return nil
}

func (s *HTTPServer) GetMessageStatsByVesselForType(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	messageType, err := strconv.Atoi(vars["type"])
	if err != nil {
		return err
	}

	json, err := s.DB.GetMessageStatsByVesselForTypeJson(messageType)
	if err != nil {
		return err
	}
	writeJSONDirect(w, http.StatusOK, json)
	return nil
}

func (s *HTTPServer) GetMessageStatsByVesselForVessel(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	mmsi, err := strconv.Atoi(vars["mmsi"])
	if err != nil {
		return err
	}

	json, err := s.DB.GetMessageStatsByVesselForVesselJson(mmsi)
	if err != nil {
		return err
	}
	writeJSONDirect(w, http.StatusOK, json)
	return nil
}
