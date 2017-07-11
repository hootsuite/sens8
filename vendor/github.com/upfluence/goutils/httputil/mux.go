package httputil

import (
	"net/http"
	"net/http/pprof"
	"os"
)

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthcheck", HealthcheckHandler)

	if os.Getenv("DEBUG") == "true" {
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		mux.HandleFunc("/debug/pprof/heap", pprof.Index)
	}

	return mux
}
