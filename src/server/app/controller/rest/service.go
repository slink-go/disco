package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ws-slink/disco/common/api"
	"github.com/ws-slink/disco/server/app/jwt"
	"github.com/ws-slink/disco/server/common/util/logger"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// region - REST service

type Service interface {
	Run(address string)
}

func Init(jwt jwt.Jwt, registry api.Registry) (Service, error) {
	return &restServiceImpl{
		jwt:      jwt,
		registry: registry,
	}, nil
}

type restServiceImpl struct {
	jwt      jwt.Jwt
	registry api.Registry
}

func (s *restServiceImpl) Run(address string) {
	router := s.configureRouter()
	logger.Info("Maket proxy service started on %s", address)
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
	router.HandleFunc("/api/token", s.handleGetToken).Methods("GET")
	router.HandleFunc("/api/join", s.authMiddleware(s.handleJoin)).Methods("POST")
	router.HandleFunc("/api/leave", s.authMiddleware(s.handleLeave)).Methods("POST")
	router.HandleFunc("/api/ping", s.authMiddleware(s.handlePing)).Methods("POST")
	router.HandleFunc("/api/list", s.authMiddleware(s.handleList)).Methods("GET")
	router.HandleFunc("/api/test", s.handleTest).Methods("POST")
	return router
}

// endregion
// region - middleware

func (s *restServiceImpl) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			writeResponseStr(w, http.StatusUnauthorized, "Missing Authorization Header")
			return
		}
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		payload, err := s.jwt.Validate(tokenString)
		if err != nil {
			writeResponseStr(w, http.StatusUnauthorized, "Error verifying JWT token: "+err.Error())
			return
		}
		if payload.GetTenant() != "" {
			r = r.WithContext(context.WithValue(r.Context(), api.TenantKey, payload.GetTenant()))
		}
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
	ctx := context.WithValue(r.Context(), api.TenantKey, "test") // TODO: tenant name should be extracted from auth token in auth middleware
	resp, err := s.registry.Join(ctx, rq)
	if err != nil {
		writeResponseStr(w, http.StatusBadRequest, fmt.Sprintf("could not join: %s", err.Error()))
		return
	}
	result, err := json.Marshal(resp)
	if err != nil {
		writeResponseStr(w, http.StatusInternalServerError, fmt.Sprintf("could not marshall json: %s", err.Error()))
		return
	}
	w.Header().Set(ContentTypeHeader, ContentTypeApplicationJson)
	writeResponseStr(w, http.StatusOK, string(result))
}
func (s *restServiceImpl) handleLeave(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("id")
	ctx := context.WithValue(r.Context(), api.TenantKey, "test") // TODO: tenant name should be extracted from auth token in auth middleware
	err := s.registry.Leave(ctx, clientId)
	if err != nil {
		writeResponseStr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeResponseStr(w, http.StatusOK, fmt.Sprintf("left %s", clientId))
}
func (s *restServiceImpl) handlePing(w http.ResponseWriter, r *http.Request) {
	clientId := r.URL.Query().Get("id")
	pong, err := s.registry.Ping(clientId)
	if err != nil {
		writeResponseStr(w, http.StatusInternalServerError, err.Error())
		return
	}
	result, err := json.Marshal(pong)
	if err != nil {
		writeResponseStr(w, http.StatusInternalServerError, fmt.Sprintf("could not marshall json: %s", err.Error()))
		return
	}
	writeResponseBytes(w, http.StatusOK, result)
}
func (s *restServiceImpl) handleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(r.Context(), api.TenantKey, "test") // TODO: tenant name should be extracted from auth token in auth middleware
	service := r.URL.Query().Get("service")
	var list []api.Client
	if service == "" {
		list = s.registry.List(ctx)
	} else {
		for _, v := range s.registry.List(ctx) {
			if v.ServiceId() == service {
				list = append(list, v)
			}
		}
	}
	b, err := json.Marshal(list)
	if err != nil {
		writeResponseStr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeResponseBytes(w, http.StatusOK, b)
}

func (s *restServiceImpl) handleGetToken(w http.ResponseWriter, r *http.Request) {
	token, err := s.jwt.Generate(r.RemoteAddr, "", time.Second*30)
	if err != nil {
		writeResponseStr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeResponseStr(w, http.StatusOK, token)
}
func (s *restServiceImpl) handleTest(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		writeResponseStr(w, http.StatusInternalServerError, err.Error())
	}
	writeResponseBytes(w, http.StatusOK, b)
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

// endregion
