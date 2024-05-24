package templates

type Service struct {
	Title     string   `json:"title"`
	Endpoints []string `json:"endpoints"`
	Color     string   `json:"color"`
}
type Card struct {
	Tenant   string    `json:"clientId"`
	Services []Service `json:"services"`
	Color    string    `json:"color"`
}
