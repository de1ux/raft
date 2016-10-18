package server

import (
	"net/http"
)

var state *State

func init() {
	state = CreateNewState()
}

func main() {
	http.HandleFunc("/direct", directHandler)
	http.HandleFunc("/append", appendHandler)
	println("Running...")
	println(http.ListenAndServe(":8080", nil))
	println("Exiting...")
}
