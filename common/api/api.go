package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// region - types

const TenantKey = "tenant"
const TenantDefault = "default"

type PongType uint8

const (
	PongTypeUnknown PongType = iota
	PongTypeOk
	PongTypeChanged
)

var (
	pongTypeNames = map[PongType]string{
		PongTypeUnknown: "UNDEFINED",
		PongTypeOk:      "OK",
		PongTypeChanged: "CHANGED",
	}
	pongTypeValues = map[string]PongType{
		"UNDEFINED": PongTypeUnknown,
		"OK":        PongTypeOk,
		"CHANGED":   PongTypeChanged,
	}
)

func (pt PongType) String() string {
	return pongTypeNames[pt]
}
func (pt *PongType) UnmarshalJSON(data []byte) (err error) {
	var source string
	if err := json.Unmarshal(data, &source); err != nil {
		return err
	}
	if *pt, err = parsePongType(source); err != nil {
		return err
	}
	return err
}
func (pt PongType) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.String())
}

func parsePongType(s string) (PongType, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	value, ok := pongTypeValues[s]
	if !ok {
		var values string
		for k, v := range clientStateValues {
			if ClientStateUnknown != v {
				values += k + ", "
			}
		}
		values = strings.TrimSpace(strings.ToLower(values))
		values = values[:len(values)-1]
		return PongTypeUnknown, fmt.Errorf("%q is not a valid PongType; available values are: %s", s, values)
	}
	return value, nil
}

type EndpointType uint8

const (
	UnknownEndpoint EndpointType = iota
	HttpEndpoint
	HttpsEndpoint
	GrpcEndpoint
)

type ClientState uint8

const (
	ClientStateUnknown ClientState = iota
	ClientStateStarting
	ClientStateUp
	ClientStateFailing
	ClientStateDown
	ClientStateRemoved
)

var (
	clientStateNames = map[ClientState]string{
		ClientStateUnknown:  "UNDEFINED",
		ClientStateStarting: "STARTING",
		ClientStateUp:       "UP",
		ClientStateFailing:  "FAILING",
		ClientStateDown:     "DOWN",
		ClientStateRemoved:  "REMOVED",
	}
	clientStateValues = map[string]ClientState{
		"UNDEFINED": ClientStateUnknown,
		"STARTING":  ClientStateStarting,
		"UP":        ClientStateUp,
		"FAILING":   ClientStateFailing,
		"DOWN":      ClientStateDown,
		"REMOVED":   ClientStateRemoved,
	}
)

func (ds ClientState) String() string {
	return clientStateNames[ds]
}
func (ds *ClientState) UnmarshalJSON(data []byte) (err error) {
	var source string
	if err := json.Unmarshal(data, &source); err != nil {
		return err
	}
	if *ds, err = parseClientState(source); err != nil {
		return err
	}
	return err
}
func (ds ClientState) MarshalJSON() ([]byte, error) {
	return json.Marshal(ds.String())
}

func parseClientState(s string) (ClientState, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	value, ok := clientStateValues[s]
	if !ok {
		var values string
		for k, v := range clientStateValues {
			if ClientStateUnknown != v {
				values += k + ", "
			}
		}
		values = strings.TrimSpace(strings.ToLower(values))
		values = values[:len(values)-1]
		return ClientStateUnknown, fmt.Errorf("%q is not a valid ClientState; available values are: %s", s, values)
	}
	return value, nil
}

// endregion
// region - requests

type Ping struct {
}

type JoinRequest struct {
	ServiceId string         `json:"service,omitempty"`
	Endpoints []string       `json:"endpoints,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// endregion
// region - responses

type Pong struct {
	Response PongType `json:"response"`
	Error    string   `json:"error"`
}

type JoinResponse struct {
	ClientId     string   `json:"id,omitempty"`
	PingInterval Duration `json:"interval,omitempty"`
}

// endregion
// region - endpoints

type Endpoint interface {
	Url() string
	Type() EndpointType
}

type endpointImpl struct {
	UrlStr string `json:"url"`
}

func (e *endpointImpl) Type() EndpointType {
	s := strings.ToLower(e.UrlStr)
	switch {
	case strings.HasPrefix(s, "https://"):
		return HttpsEndpoint
	case strings.HasPrefix(s, "http://"):
		return HttpEndpoint
	case strings.HasPrefix(s, "grpc://"):
		return GrpcEndpoint
	}
	return UnknownEndpoint
}
func (e *endpointImpl) Url() string {
	return e.UrlStr
}

func NewEndpoint(url string) (Endpoint, error) {
	e := endpointImpl{UrlStr: url}
	if e.Type() == UnknownEndpoint {
		return nil, fmt.Errorf("unsupported url protocol: %s", url)
	}
	return &e, nil
}

// endregion
// region - clients

type Client interface {
	ClientId() string
	ServiceId() string
	Tenant() string
	Endpoints() []Endpoint
	Meta() map[string]any
	Ping()
	LastSeen() time.Time
	State() ClientState
	SetState(state ClientState)
	SetDirty(value bool)
	IsDirty() bool
}

// endregion
// region - registry

type Registry interface {
	Join(ctx context.Context, request JoinRequest) (*JoinResponse, error)
	Leave(ctx context.Context, clientId string) error
	List(ctx context.Context) []Client
	Ping(clientId string) (Pong, error)
}

// endregion
