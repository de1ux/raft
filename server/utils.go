package server

import (
	"net/http"
)

func err500(w http.ResponseWriter, err error) {
	println(err.Error())
	w.WriteHeader(http.StatusInternalServerError)
}
