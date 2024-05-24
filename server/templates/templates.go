package templates

import (
	"github.com/slink-go/disco/common/api"
	"slices"
)

var colorIndex = 0
var colors = []string{
	"lightsalmon",
	"lightgreen",
	"lightblue",
	"lightgoldenrodyellow",
	"lightskyblue",
	"lightgrey",
	"lightpink",
	"lightsalmon",
	"lightsalt",
}

func getColor() string {
	if colorIndex >= len(colors) {
		colorIndex = 0
	}
	color := colors[colorIndex]
	colorIndex++
	return color
}

func Cards(tenants []api.Tenant) []Card {

	var names []string
	var tnts = make(map[string]api.Tenant)
	for _, t := range tenants {
		names = append(names, t.Name())
		tnts[t.Name()] = t
	}
	slices.Sort(names)

	var result []Card
	for _, k := range names {
		result = append(result, newCard(tnts[k]))
	}
	return result

}

func newCard(tenant api.Tenant) Card {
	return Card{
		Tenant:   tenant.Name(),
		Services: getServices(tenant),
		Color:    getColor(),
	}
}

func getServices(tenant api.Tenant) []Service {

	clients := make(map[string][]string)
	for _, c := range tenant.Clients() {
		if _, ok := clients[c.ServiceId()]; !ok {
			clients[c.ServiceId()] = make([]string, 0)
		}
		clients[c.ServiceId()] = append(clients[c.ServiceId()], getEndpoints(c.Endpoints())...)
	}

	var names []string
	for k, _ := range clients {
		names = append(names, k)
		slices.Sort(clients[k])
	}
	slices.Sort(names)

	result := make([]Service, 0)
	for _, k := range names {
		result = append(result, Service{
			Title:     k,
			Endpoints: clients[k],
			Color:     getColor(),
		})
	}
	return result

}

func getEndpoints(endpoints []api.Endpoint) []string {
	result := make([]string, 0)
	for _, endpoint := range endpoints {
		result = append(result, endpoint.Url())
	}
	slices.Sort(result)
	return result
}
