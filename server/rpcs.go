package main

import (
    "bytes"
    "io/ioutil"
    "net/http"
    "encoding/json"
)

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


// RequestVote is an RPC for candidate election
type RequestVote struct {
    Term int
    CandidateID int
    LastLogIndex int
    LastLogTerm int
}

func RequestVoteFromRequest(req *http.Request) (*RequestVote, error) {
    rv := &RequestVote{}
    b, err := ioutil.ReadAll(req.Body)
    if err != nil {
        return nil, err
    }
    if err := json.Unmarshal(b, rv); err != nil {
        return nil, err
    }
    return rv, nil
}

func (rv *RequestVote) Write(w http.ResponseWriter) error {
    b, err := json.Marshal(rv)
    if err != nil {
        return err
    }
    w.Write(b)
    return nil
}

type RequestVoteResponse struct {
    Term int
    VoteGranted bool
}

func RequestVoteResponseFromRequest(req *http.Request) (*RequestVoteResponse, error) {
    rr := &RequestVoteResponse{}
    b, err := ioutil.ReadAll(req.Body)
    if err != nil {
        return nil, err
    }
    if err := json.Unmarshal(b, rr); err != nil {
        return nil, err
    }
    return rr, nil
}

func (rr *RequestVoteResponse) Write (w http.ResponseWriter) error {
    b, err := json.Marshal(rr)
    if err != nil {
        return err
    }
    w.Write(b)
    return nil
}