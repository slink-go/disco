package rest

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slink-go/disco/common/api"
	"github.com/slink-go/disco/server/config"
	"github.com/slink-go/disco/server/jwt"
	"github.com/slink-go/disco/server/templates"
	"github.com/slink-go/logging"
	"github.com/xhit/go-str2duration/v2"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"os"
	"strings"
)

type Service interface {
	Run()
}

var (
	ErrUnauthorized          = errors.New("unauthorized")
	ErrBasicAuthNotSupported = errors.New("basic authorization is not enabled")
	ErrNonTokenAuth          = errors.New("non-token auth attempted")
)

func NewDiscoService(jwt jwt.Jwt, registry api.Registry, cfg *config.AppConfig) (Service, error) {
	var httpDuration *prometheus.HistogramVec
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "disco_http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path"})
	return &restServiceImpl{
		jwt:              jwt,
		registry:         registry,
		httpDurationHist: httpDuration,
		cfg:              cfg,
		limiter:          rate.NewLimiter(rate.Limit(cfg.RequestRate), cfg.RequestBurst),
		logger:           logging.GetLogger("service"),
	}, nil
}
func (s *restServiceImpl) Run() {

	if s.cfg.MonitoringPort > 0 {
		go s.startMonitoring()
	}
	if s.cfg.ServicePort > 0 {
		s.startService()
	} else {
		panic("service port not set")
	}

}

type restServiceImpl struct {
	jwt              jwt.Jwt
	registry         api.Registry
	httpDurationHist *prometheus.HistogramVec
	cfg              *config.AppConfig
	limiter          *rate.Limiter
	logger           logging.Logger
}

// region - service

func (s *restServiceImpl) startService() {
	router := s.configureServiceRouter()
	address := fmt.Sprintf(":%d", s.cfg.ServicePort)
	s.logger.Info("Disco service started on %s", address)
	if s.cfg.Secured {
		if s.cfg.SslCertFile != "" && s.cfg.SslCertKey != "" {
			s.serveSslWithCert(address, router)
		} else {
			s.serveSslWithLetsEncrypt(address, router)
		}
	} else {
		s.serveInsecureHttp(address, router)
	}
}

func (s *restServiceImpl) configureServiceRouter() *mux.Router {
	router := mux.NewRouter()

	router.Use(s.rateLimiterMiddleware)

	// https://stackoverflow.com/questions/64768950/how-to-use-specific-middleware-for-specific-routes-in-a-get-subrouter-in-gorilla
	router.Use(s.prometheusMiddleware)
	router.Path("/metrics").Handler(promhttp.Handler())

	//router.HandleFunc("/api/token/{tenant}", s.handleGetToken).Methods("GET")

	router.HandleFunc("/api/join", s.authMiddleware(s.handleJoin)).Methods("POST")
	router.HandleFunc("/api/leave", s.authMiddleware(s.handleLeave)).Methods("POST")
	router.HandleFunc("/api/ping", s.authMiddleware(s.handlePing)).Methods("POST")
	router.HandleFunc("/api/list", s.authMiddleware(s.handleList)).Methods("GET")

	return router
}

func (s *restServiceImpl) handleJoin(w http.ResponseWriter, r *http.Request) {
	//logger.Info(fmt.Sprintf("%s %s", r.Host, r.RemoteAddr))
	var rq api.JoinRequest
	err := decodeJSONBody(w, r, &rq)
	if err != nil {
		writeResponseStr(w, http.StatusBadRequest, fmt.Sprintf("error reading request: %s", err.Error()))
		return
	}
	rq.ServiceId = strings.ToUpper(rq.ServiceId)
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
	durationStr := r.URL.Query().Get("ttl")
	if durationStr == "" {
		durationStr = "30m"
	}
	dur, err := str2duration.ParseDuration(durationStr)
	if err != nil {
		writeResponseError(w, http.StatusInternalServerError, err)
		return
	}
	token, err := s.jwt.Generate(r.RemoteAddr, tenant, dur)
	if err != nil {
		writeResponseError(w, http.StatusInternalServerError, err)
		return
	}
	writeResponseStr(w, http.StatusOK, token)
}

// endregion
// region - monitoring

func (s *restServiceImpl) startMonitoring() {
	router := s.configureMonitoringRouter()
	address := fmt.Sprintf(":%d", s.cfg.MonitoringPort)
	s.logger.Info("Disco monitoring started on %s", address)
	if s.cfg.Secured {
		if s.cfg.SslCertFile != "" && s.cfg.SslCertKey != "" {
			s.serveSslWithCert(address, router)
		} else {
			s.serveSslWithLetsEncrypt(address, router)
		}
	} else {
		s.serveInsecureHttp(address, router)
	}
}

func (s *restServiceImpl) configureMonitoringRouter() *mux.Router {
	router := mux.NewRouter()

	router.Use(s.prometheusMiddleware)
	router.Path("/metrics").Handler(promhttp.Handler())

	staticFilePath := os.Getenv("STATIC_FILE_PATH")
	if staticFilePath != "" {
		staticFilePath = "/static/"
	}
	if !strings.HasSuffix(staticFilePath, "/") {
		staticFilePath = staticFilePath + "/"
	}

	fs := http.FileServer(http.Dir(staticFilePath))
	router.PathPrefix("/s/").Handler(http.StripPrefix("/s/", fs))
	router.HandleFunc("/", s.monitoringPage).Methods("GET")

	return router
}
func (s *restServiceImpl) monitoringPage(w http.ResponseWriter, r *http.Request) {
	cards := templates.Cards(s.registry.ListAll())
	tmpl := templates.RegistryPage(cards)
	if err := tmpl.Render(context.Background(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// endregion
// region - common

// region -> starters

func (s *restServiceImpl) serveInsecureHttp(address string, router *mux.Router) {
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
func (s *restServiceImpl) serveSslWithCert(address string, router *mux.Router) {
	log.Fatal(
		http.ListenAndServeTLS(
			address,
			s.cfg.SslCertFile,
			s.cfg.SslCertKey,
			router,
		),
	)
}
func (s *restServiceImpl) serveSslWithLetsEncrypt(address string, router *mux.Router) {
	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		//HostPolicy: autocert.HostWhitelist("example.com"),
		Cache: autocert.DirCache("/tmp/certs"), //Folder for storing certificates
	}
	server := &http.Server{
		//Addr:    ":https",
		Addr:    address,
		Handler: router,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}
	go func() {
		_ = http.ListenAndServe(":http", certManager.HTTPHandler(nil))
	}()
	log.Fatal(server.ListenAndServeTLS("", "")) //Key and cert are coming from Let's Encrypt
}

// endregion
// region -> middleware

func (s *restServiceImpl) prometheusMiddleware(next http.Handler) http.Handler {
	// https://www.robustperception.io/prometheus-middleware-for-gorilla-mux/
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(s.httpDurationHist.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}

func (s *restServiceImpl) rateLimiterMiddleware(next http.Handler) http.Handler {
	// https://www.alexedwards.net/blog/how-to-rate-limit-http-requests
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.limiter.Allow() == false {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *restServiceImpl) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			writeResponseMessage(w, http.StatusUnauthorized, "error", "missing authorization header")
			return
		}

		// try token auth
		tenant, err := s.tokenAuth(r)
		if err != nil && !errors.Is(err, ErrNonTokenAuth) {
			writeResponseError(w, http.StatusUnauthorized, err)
			return
		}

		if errors.Is(err, ErrNonTokenAuth) {
			// try basic auth
			tenant, err = s.basicAuth(r)
			if err != nil {
				writeResponseError(w, http.StatusUnauthorized, err)
				return
			}
		}

		r = r.WithContext(context.WithValue(r.Context(), api.TenantKey, tenant))
		next.ServeHTTP(w, r)
	}
}
func (s *restServiceImpl) basicAuth(r *http.Request) (string, error) {
	//https://www.alexedwards.net/blog/basic-authentication-in-go
	username, password, ok := r.BasicAuth()
	if !ok {
		return "", ErrUnauthorized
	}

	if s.cfg.RegisteredUsers == nil || len(s.cfg.RegisteredUsers) == 0 {
		return "", ErrBasicAuthNotSupported
	}

	usernameHash := sha256.Sum256([]byte(username))
	passwordHash := sha256.Sum256([]byte(password))

	var userMatch bool
	var passMatch bool

	for _, cr := range s.cfg.RegisteredUsers {
		userMatch = s.checkHash(usernameHash, cr.Login)
		passMatch = s.checkHash(passwordHash, cr.Password)
		if userMatch && passMatch {
			return username, nil
		}
	}

	return "", ErrUnauthorized
}
func (s *restServiceImpl) tokenAuth(r *http.Request) (string, error) {
	authStr := r.Header.Get("Authorization")
	if !strings.Contains(authStr, "Bearer ") {
		return "", ErrNonTokenAuth
	}
	authStr = strings.Replace(authStr, "Bearer ", "", 1)
	payload, err := s.jwt.Validate(authStr)
	if err != nil {
		return "", err
	}
	if payload.GetTenant() != "" {
		return payload.GetTenant(), nil
	} else {
		return api.TenantDefault, nil
	}
}
func (s *restServiceImpl) checkHash(hash [sha256.Size]byte, str string) bool {
	check := sha256.Sum256([]byte(str))
	return subtle.ConstantTimeCompare(hash[:], check[:]) == 1
}

// endregion
// region -> helpers

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

// endregion
