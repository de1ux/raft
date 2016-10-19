package main

import (
	"fmt"
)

type Role int
const (
    LEADER Role = iota
    CANDIDATE
    FOLLOWER
)

// TODO - clean this up
func CreateNewState() *State {
	return &State{
		currentTerm: 0,
		commitIndex: 0,
		lastApplied: 0,
		log:         &Log{},
	}
}

type State struct {
	/***************************************/
	/*** Persistent state on all servers ***/
	/***************************************/
	// currentTerm is the lastest term the server has seen. Initialized to 0 on
	// boot, increases monotonically.
	currentTerm int
	// votedFor is the candidateID that received the vote in the current term,
	// or null if none.
	votedFor int
	// log is the list of commands to be applied to the state machine, each including
	// the term when the entry was received by the leader. First index is 1
	log *Log

	/*************************************/
	/*** Volatile state on all servers ***/
	/*************************************/
	// commitIndex is the index of the highest log entry known to be committed
	// Initialized at 0, increases monotonically
	commitIndex int
	// lastApplied is the index of the highest log entry applied to state machine.
	// Initialized at 0, increases monotonically
	lastApplied int

	/***************************************************************/
	/*** Volatile state on leaders. Reinitialized after election ***/
	/***************************************************************/
	// nextIndex, for each server, is the index of the next log entry to send
	// to that server. Initialized to leader's last log index + 1
	nextIndex map[string]int
	// matchIndex, for each server, is the index of the highest log entry
	// known to be replicated on that server. Initialized to 0, increases monotonically
	matchIndex map[string]int

	// Additions outside the spec:
	// 1. I'm using nextIndex and matchIndex as map[string]int to lookup servers
	//    by a string identifier
	// 2. I'm using role Role to idenify what role the server is actively playing
	role Role
}

// Commit is responsible for applying a set of AppendEntries to the log
func (s *State) Commit(ae *AppendEntries) {
	keep := s.log.Length()
	// If the follower has more records, verify the terms are correct
	if s.log.Length() > ae.PrevLogIndex {
		for i := 0; i < len(ae.Entries); i++ {
			existingEntry := s.log.At(i + ae.PrevLogIndex)
			if existingEntry == nil {
				break
			}
			if existingEntry.Term != ae.Entries[i].Term {
				keep = i + ae.PrevLogIndex
				break
			}
		}
	}

    s.log.Rollback(keep)
    s.log.Append(ae.Entries)
    s.UpdateCommitIndex(ae.LeaderCommit)
}

// UpdateCommitIndex updates state's commit index if the AppendEntries RPC has
// a higher leader commit index.
// TODO - when would log.Length be greater than leaderCommitIndex?
func (s *State) UpdateCommitIndex(leaderCommitIndex int) {
    if leaderCommitIndex > s.commitIndex {
        s.commitIndex = min(leaderCommitIndex, s.log.Length())
    }
}

func (s *State) EntriesToString() string {
	return s.log.String()
}

// Entry is the command to be committed to the state machine's log
type Entry struct {
	Term int
	Data []byte
}

// Log is an encaspulation of the entries and includes helper functions for safely
// getting/setting entry data
type Log struct {
	entries []Entry
}

func (l *Log) String() string {
	return fmt.Sprintf("%+v", l.entries)
}

func (l *Log) Rollback(count int) {
	l.entries = l.entries[0:count]
}

func (l *Log) Append(entries []Entry) {
	l.entries = append(l.entries, entries...)
}

func (l *Log) Length() int {
	return len(l.entries)
}

func (l *Log) At(index int) *Entry {
	if index >= len(l.entries) {
		return nil
	}
	return &l.entries[index]
}

func (l *Log) AtIndexAndTerm(index, term int) *Entry {
	entry := l.At(index)
	if entry == nil {
		return nil
	}
	if entry.Term != term {
		return nil
	}
	return entry
}


