package main

import (
	"log"
	"net/http"
)

func directHandler(state *State, w http.ResponseWriter, r *http.Request) {
    // TODO - if role is follower, respond with leader ID
}

// voteHandler is responsible for accepting vote solicitations
func voteHandler(state *State, w http.ResponseWriter, r *http.Request) {
    rr := &RequestVoteResponse{
        VoteGranted: false,
        Term: state.currentTerm,
    }

    rv, err := RequestVoteFromRequest(r)
    if err != nil {
        log.Print("WARN: ", err.Error())
        rr.Write(w)
        return
    }

    if rv.Term < state.currentTerm {
        log.Print("WARN: Received vote solicitation from term ", rv.Term, ", current is ", state.currentTerm)
        rr.Write(w)
        return
    }

    if (state.votedFor == 0 || state.votedFor == rv.CandidateID) && (state.commitIndex == rv.LastLogIndex) {
        rr.VoteGranted = true
        electionTimer.Reset()
    }

    rr.Write(w)
}

// appendHandler receives AppendEntries RPCs from the leader and applies them
func appendHandler(state *State, w http.ResponseWriter, r *http.Request) {
	ar := &AppendEntriesResponse{
		Success: false,
		Term:    state.currentTerm,
	}

	if state.role == LEADER {
		log.Print("WARN: Leaders cannot take AppendEntries RPCs")
		ar.Write(w)
		return
	}

	ae, err := AppendEntriesFromRequest(r)
	if err != nil {
		log.Print("WARN: ", err.Error())
		ar.Write(w)
		return
	}

	// Cannot apply entries from an older term
	if ae.Term < state.currentTerm {
		ar.Write(w)
		return
	}

	// TODO - I allow 0 as an initilization of entries, is this safe?
	if ae.PrevLogIndex == 0 {
		state.Commit(ae)
		ar.Success = true
		ar.Write(w)
		return
	}

	// Cannot apply the continuation of new logs if the old logs don't exist.
	entry := state.log.At(ae.PrevLogIndex)
	if entry == nil {
		ar.Write(w)
		return
	}

	// If there's a conflict in the term for the previous log, reject
	if ae.PrevLogTerm != entry.Term {
		ar.Write(w)
		return
	}

	state.Commit(ae)
    electionTimer.Reset()
    ar.Success = true
	ar.Write(w)
}
