package httputil

import "net/http"

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
