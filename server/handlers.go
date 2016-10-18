package server

import (
    "net/http"
)

func directHandler(w http.ResponseWriter, r *http.Request) {

}

// appendHandler receives AppendEntry RPCs from the leader and applies them
func appendHandler(w http.ResponseWriter, r *http.Request) {
    _, err := AppendEntryFromRequest(r)
    if err != nil {
        err500(w, err)
    }

}

