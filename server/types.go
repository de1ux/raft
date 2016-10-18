package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Role string

const Leader Role = "LEADER"
const Follower Role = "FOLLOWER"

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

	// Update our commit index to
	if ae.LeaderCommit > s.commitIndex {
		if ae.LeaderCommit > s.log.Length() {
			s.commitIndex = s.log.Length()
		} else {
			s.commitIndex = ae.LeaderCommit
		}
	}
}

func (s *State) EntriesToString() string {
	return s.log.String()
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

type Entry struct {
	Term int
	Data []byte
}

// AppendEntries is sent to followers by the leader
type AppendEntries struct {
	// Term is the leaders Term
	Term int
	// LeaderID is given to remind followers where clients can redirect their writes to
	LeaderID string
	// PrevLogIndex term of the previousLogIndex entry
	PrevLogIndex int
	// PrevLogTerm is the previous term the follower should have entries for
	PrevLogTerm int
	// Entries is the log of entries to store. Empty for heartbeat
	Entries []Entry
	// LeaderCommit is the leaders commitIndex
	LeaderCommit int
}

func (ae *AppendEntries) ToRequest(URL string) (*http.Request, error) {
	b, err := json.Marshal(ae)
	if err != nil {
		return nil, err
	}
	bReader := bytes.NewReader(b)
	return http.NewRequest("POST", URL, bReader)
}

func AppendEntriesFromRequest(req *http.Request) (*AppendEntries, error) {
	ae := &AppendEntries{}
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, ae); err != nil {
		return nil, err
	}
	return ae, nil
}

// AppendEntriesResponse is the reply to the leader following an AppendEntries request
type AppendEntriesResponse struct {
	// Success is true if the entry was appended on the follower
	Success bool
	// Term that the follower recognizes
	Term int
}

// Write returns the response from the AppendEntries RPC
func (ar *AppendEntriesResponse) Write(w http.ResponseWriter) error {
	b, err := json.Marshal(ar)
	if err != nil {
		return err
	}
	w.Write(b)
	return nil
}

func AppendEntriesResponseFromRequest(resp *http.Response) (*AppendEntriesResponse, error) {
	ar := &AppendEntriesResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, ar); err != nil {
		return nil, err
	}
	return ar, nil
}
