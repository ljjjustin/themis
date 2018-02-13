package api

import (
	"net/http"
	"time"
)

func Logger(in http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		in.ServeHTTP(w, r)
		plog.Infof("%s %s %s %s", r.Method, r.RequestURI, name, time.Since(start))
	})
}
