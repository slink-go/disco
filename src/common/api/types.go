package api

type EndpointType uint8

const (
	UnknownEndpoint EndpointType = iota
	HttpEndpoint
	HttpsEndpoint
	GrpcEndpoint
)
