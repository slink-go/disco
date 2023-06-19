package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slink-go/disco/common/api"
	"github.com/slink-go/disco/common/util/logger"
	"github.com/slink-go/disco/server/jwt"
	"log"
	"net/http"
	"strings"
	"time"
)

// region - REST service

type Service interface {
	Run(address string)
}

func Init(jwt jwt.Jwt, registry api.Registry, monitoringEnabled bool) (Service, error) {

	var httpDuration *prometheus.HistogramVec

	if monitoringEnabled {
		httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "disco_http_duration_seconds",
			Help: "Duration of HTTP requests.",
		}, []string{"path"})
	}

	return &restServiceImpl{
		jwt:               jwt,
		registry:          registry,
		httpDurationHist:  httpDuration,
		monitoringEnabled: monitoringEnabled,
	}, nil

}

type restServiceImpl struct {
	jwt               jwt.Jwt
	registry          api.Registry
	httpDurationHist  *prometheus.HistogramVec
	monitoringEnabled bool
}

func (s *restServiceImpl) Run(address string) {
	router := s.configureRouter()
	logger.Info("Disco service started on %s", address)
	log.Fatal(
		http.ListenAndServe(
			address,
			//handlers.LoggingHandler( // enable basic (mux built-in) request logging
			//	os.Stdout,
			router,
			//),
		),
	)
}
func (s *restServiceImpl) configureRouter() *mux.Router {
	router := mux.NewRouter()

	// https://stackoverflow.com/questions/64768950/how-to-use-specific-middleware-for-specific-routes-in-a-get-subrouter-in-gorilla
	if s.monitoringEnabled {
		router.Use(s.prometheusMiddleware)
		router.Path("/metrics").Handler(promhttp.Handler())
	}

	router.HandleFunc("/api/token/{tenant}", s.handleGetToken).Methods("GET")

	router.HandleFunc("/api/join", s.authMiddleware(s.handleJoin)).Methods("POST")
	router.HandleFunc("/api/leave", s.authMiddleware(s.handleLeave)).Methods("POST")
	router.HandleFunc("/api/ping", s.authMiddleware(s.handlePing)).Methods("POST")
	router.HandleFunc("/api/list", s.authMiddleware(s.handleList)).Methods("GET")

	return router
}

// endregion
// region - middleware

// https://www.robustperception.io/prometheus-middleware-for-gorilla-mux/
func (s *restServiceImpl) prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(s.httpDurationHist.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}

func (s *restServiceImpl) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			writeResponseMessage(w, http.StatusUnauthorized, "error", "missing authorization header")
			return
		}
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		payload, err := s.jwt.Validate(tokenString)
		if err != nil {
			writeResponseError(w, http.StatusUnauthorized, err)
			return
		}
		if payload.GetTenant() != "" {
			r = r.WithContext(context.WithValue(r.Context(), api.TenantKey, payload.GetTenant()))
		} else {
			r = r.WithContext(context.WithValue(r.Context(), api.TenantKey, api.TenantDefault))
		}
		//logger.Info("[auth] tenant: %v, %v %v", r.Context().Value(api.TenantKey), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	}
}

// endregion
// region - handlers

func (s *restServiceImpl) handleJoin(w http.ResponseWriter, r *http.Request) {
	//logger.Info(fmt.Sprintf("%s %s", r.Host, r.RemoteAddr))
	var rq api.JoinRequest
	err := decodeJSONBody(w, r, &rq)
	if err != nil {
		writeResponseStr(w, http.StatusBadRequest, fmt.Sprintf("error reading request: %s", err.Error()))
		return
	}
	resp, err := s.registry.Join(r.Context(), rq)
	if err != nil {
		writeResponseMessage(w, http.StatusBadRequest, "error", fmt.Sprintf("could not join: %s", err.Error()))
		return
	}
	result, err := json.Marshal(resp)
	if err != nil {
		writeResponseMessage(w, http.StatusInternalServerError, "error", fmt.Sprintf("could not marshall json: %s", err.Error()))
		return
	}
	w.Header().Set(api.ContentTypeHeader, api.ContentTypeApplicationJson)
	writeResponseStr(w, http.StatusOK, string(result))
}
func (s *restServiceImpl) handleLeave(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("id")
	err := s.registry.Leave(r.Context(), clientId)
	if err != nil {
		writeResponseError(w, http.StatusInternalServerError, err)
		return
	}
	writeResponseMessage(w, http.StatusOK, "left", clientId)
}
func (s *restServiceImpl) handlePing(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("id")
	pong, err := s.registry.Ping(clientId)
	if err != nil {
		if errors.Is(err, api.NewClientNotFoundError(clientId)) {
			writeResponseError(w, http.StatusNotFound, err)
		} else {
			writeResponseError(w, http.StatusInternalServerError, err)
		}
		return
	}
	result, err := json.Marshal(pong)
	if err != nil {
		writeResponseMessage(w, http.StatusInternalServerError, "error", fmt.Sprintf("could not marshall json: %s", err.Error()))
		return
	}
	writeResponseBytes(w, http.StatusOK, result)
}
func (s *restServiceImpl) handleList(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	var list []api.Client
	if service == "" {
		list = s.registry.List(r.Context())
	} else {
		for _, v := range s.registry.List(r.Context()) {
			if v.ServiceId() == service {
				list = append(list, v)
			}
		}
	}
	b, err := json.Marshal(list)
	if err != nil {
		writeResponseError(w, http.StatusInternalServerError, err)
		return
	}
	writeResponseBytes(w, http.StatusOK, b)
}

func (s *restServiceImpl) handleGetToken(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(time.Duration(rand.Intn(5)) * time.Second) // random delay
	tenant := mux.Vars(r)["tenant"]
	token, err := s.jwt.Generate(r.RemoteAddr, tenant, time.Minute*30)
	if err != nil {
		writeResponseError(w, http.StatusInternalServerError, err)
		return
	}
	writeResponseStr(w, http.StatusOK, token)
}

// endregion
// region - helpers

func writeResponseStr(w http.ResponseWriter, code int, str string) {
	writeResponseBytes(w, code, []byte(fmt.Sprintf("%s\n", str)))
}
func writeResponseBytes(w http.ResponseWriter, code int, data []byte) {
	w.WriteHeader(code)
	_, err := w.Write(data)
	if err != nil {
		return
	}
}
func writeResponseMessage(w http.ResponseWriter, code int, key, value string) {
	writeResponseStr(w, code, fmt.Sprintf("{\"%s\": \"%s\"}", key, value))
}
func writeResponseError(w http.ResponseWriter, code int, err error) {
	writeResponseMessage(w, code, "error", err.Error())
}

// endregion
