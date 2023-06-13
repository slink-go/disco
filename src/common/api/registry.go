package api

type Registry interface {
	Join(request JoinRequest) (response JoinResponse, err error)
	//Disconnect(connId string) error
	//Pulse(ping Ping) (Pong, error)
	//List() []Client
	//ListOfType(typ EndpointType) []Client
	//Get(string) Client
}
