package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"RootIndex", "GET", "/", RootIndex,
	},
	Route{
		"HostIndex", "GET", "/hosts", HostIndex,
	},
	Route{
		"HostCreate", "POST", "/hosts", HostCreate,
	},
	Route{
		"HostShow", "GET", "/hosts/{hostId}", HostShow,
	},
	Route{
		"HostUpdate", "PUT", "/hosts/{hostId}", HostUpdate,
	},
	Route{
		"HostDelete", "DELETE", "/hosts/{hostId}", HostDelete,
	},
}

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		var handler http.Handler

		handler = route.Handler
		handler = Logger(handler, route.Name)
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(handler)
	}

	return router
}
