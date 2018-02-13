package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func RootIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Alive!")
}

func HostIndex(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Content-Type", "application/json; charset=UTF-8")
	// Query storage engine
	w.WriteHeader(http.StatusOK)
}

func HostCreate(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Content-Type", "application/json; charset=UTF-8")
	// Query storage engine
	w.WriteHeader(http.StatusOK)
}

func HostShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostId := vars["hostId"]
	//header := w.Header()
	//header.Set("Content-Type", "application/json; charset=UTF-8")
	// Query storage engine
	fmt.Fprintf(w, "host id: ", hostId)
}

func HostUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostId := vars["hostId"]
	//header := w.Header()
	//header.Set("Content-Type", "application/json; charset=UTF-8")
	// Query storage engine
	fmt.Fprintf(w, "update host: ", hostId)
}

func HostDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostId := vars["hostId"]
	//header := w.Header()
	//header.Set("Content-Type", "application/json; charset=UTF-8")
	// Query storage engine
	fmt.Fprintf(w, "delete host: ", hostId)
}
