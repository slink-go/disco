package api

import (
	"fmt"
	"strings"
)

type Endpoint struct {
	Url string `json:"url,omitempty"`
}

func (e *Endpoint) Type() EndpointType {
	s := strings.ToLower(e.Url)
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
func NewEndpoint(url string) (*Endpoint, error) {
	e := Endpoint{Url: url}
	if e.Type() == UnknownEndpoint {
		return nil, fmt.Errorf("unsupported url protocol: %s", url)
	}
	return &e, nil
}
