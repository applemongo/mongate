package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/applemongo/mongate/gate"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
    "os"
)

type Server interface {
	io.Closer
	http.Handler
}

type server struct {
	sync.Mutex

	r        *mux.Router
	backends map[string]gate.Proxy
	logger   *logrus.Logger
}

func New(logger *logrus.Logger) Server {
    var configPath string = ""
    if len(os.Args) > 1 {
        configPath = os.Args[1]
    }

    config, err := NewConfig(logger, configPath)

    if err != nil {
        fmt.Println("config error: ", err)
        os.Exit(1) //TODO: checkout code definition
    }

	r := mux.NewRouter()

	s := &server{
		r:        r,
		logger:   logger,
		backends: make(map[string]gate.Proxy),
	}
	r.HandleFunc("/", s.listBackends).Methods("GET")

    s.configServer(config)

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.r.ServeHTTP(w, r)
}

func (s *server) Close() error {
	s.Lock()
	defer s.Unlock()

	var err error
	for id, p := range s.backends {
		if nerr := p.Close(); nerr != nil {
			s.logger.WithFields(logrus.Fields{
				"error": err,
				"id":    id,
			}).Error("closing backend proxy")

			err = nerr
		}
	}

	return err
}

func (s *server) configServer(config *Config) {
    s.Lock()
    defer s.Unlock()

    for id, proxy := range config.Proxies {
        s.configProxy(id, &proxy)
    }
}

func (s *server) configProxies(config *Config) {
    s.Lock()
    defer s.Unlock()

    for id, proxy := range config.Proxies {
        s.configProxy(id, &proxy)
    }
}

func (s *server) configProxy(id string, backend *gate.Backend) {
    gate, err := gate.New(backend)
    if err != nil {
        s.logger.WithField("error", err).Error("creating gate " + id)
        return
    }

    s.backends[id] = gate
    if err := gate.Start(); err != nil {
        s.logger.WithFields(logrus.Fields{
            "error": err,
            "id":    id,
        }).Error("starting gate " + id)

        delete(s.backends, id)
        return
    }
}


func (s *server) listBackends(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("listing backends")

	out := []*gate.Backend{}

	s.Lock()
	for _, p := range s.backends {
		out = append(out, p.Backend())
	}
	s.Unlock()

	s.marshal(w, out)
}

func (s *server) getBackend(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	s.logger.WithField("id", id).Debug("getting backend")

	s.Lock()
	p, exists := s.backends[id]
	s.Unlock()

	if !exists {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	s.marshal(w, p.Backend())
}

func (s *server) addBackend(w http.ResponseWriter, r *http.Request) {
	var (
		backend *gate.Backend
		id      = mux.Vars(r)["id"]
	)
    fmt.Printf("id %s", string(id))
	s.logger.WithField("id", id).Debug("adding new backend")

	s.Lock()
	_, exists := s.backends[id]
	s.Unlock()

	if exists {
		http.Error(w, fmt.Sprintf("%s already exists", id), http.StatusConflict)

		return
	}

	if err := json.NewDecoder(r.Body).Decode(&backend); err != nil {
		s.logger.WithField("error", err).Error("decoding backend json")
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}
	backend.Name = id

	s.Lock()
	defer s.Unlock()

	gate, err := gate.New(backend)
	if err != nil {
		s.logger.WithField("error", err).Error("creating new proxy")
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	s.backends[id] = gate

	if err := gate.Start(); err != nil {
		s.logger.WithFields(logrus.Fields{
			"error": err,
			"id":    id,
		}).Error("starting new proxy")

		delete(s.backends, id)

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
}


func (s *server) marshal(w http.ResponseWriter, v interface{}) {
	w.Header().Set("content-type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.WithField("error", err).Error("marshal json")
	}
}
