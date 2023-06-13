package api

type Client interface {
	Endpoints() []Endpoint
	Meta() map[string]any
}

type client struct {
	id        string
	endpoints []Endpoint
	meta      map[string]any
}
