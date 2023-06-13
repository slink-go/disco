package api

import "time"

type PongType uint8

const (
	Ok PongType = iota
	Changed
)

type Pong struct {
	Response PongType `json:"response"`
}

type Ping struct {
}

type JoinRequest struct {
	ServiceId string         `json:"service,omitempty"`
	Endpoints []string       `json:"endpoints,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

type JoinResponse struct {
	ClientId     string        `json:"id,omitempty"`
	PingInterval time.Duration `json:"interval,omitempty"`
}
