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
			"/api/vessels":                         server.GetVessels,
			"/api/vessels/{mmsi:[0-9]+}":           server.GetVesselByMmsi,
			"/api/vessels/{mmsi:[0-9]+}/positions": server.GetPositionsForVessel,
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
	vessels, err := s.DB.GetVessels()

	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, vessels)

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

	positions, err := s.DB.GetPositionsForVessel(mmsi)

	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, positions)

	return nil
}
