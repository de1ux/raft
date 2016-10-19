package main

import (
    "time"
	"net/http"
)

var state *State
var shouldConsume = true
var electionTimer *ElectionTimer

// Thresholds for random election timeout, in milliseconds
const ELECTION_TIMER_MIN = 150
const ELECTION_TIMER_MAX = 300

func init() {
	state = CreateNewState()
}

func consumeLogs() {
    println("Consumer ready...")
    for shouldConsume {
        if state.lastApplied < state.commitIndex {
            state.lastApplied++
            entry := state.log.At(state.lastApplied)
            println("Consuming ", entry.Data, " during term ", entry.Term)
        }
        time.Sleep(time.Second * 2)
    }
}

func handleElectionTimeout() {
    //state.Transition(CANDIDATE)
}

func StartNewElectionTimer() *ElectionTimer {
    duration := time.Duration(rand(ELECTION_TIMER_MIN, ELECTION_TIMER_MAX)) * time.Millisecond
    e := &ElectionTimer{
        timer: time.AfterFunc(duration, handleElectionTimeout),
        duration: duration,
    }
    return e
}

type ElectionTimer struct {
    timer *time.Timer
    duration time.Duration
}

func (e *ElectionTimer) Reset() {
    e.timer.Reset(e.duration)
}

func main() {
    go consumeLogs()
    electionTimer = StartNewElectionTimer()

    http.HandleFunc("/direct", directHandler)
	http.HandleFunc("/append", appendHandler)
	println("Running...")
	println(http.ListenAndServe(":8080", nil))
	println("Exiting...")

}
