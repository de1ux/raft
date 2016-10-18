package server

import (
    "bytes"
    "net/http"
    "io/ioutil"
    "encoding/json"
)

type Role string
const Leader Role = "LEADER"
const Follower Role = "FOLLOWER"

type State struct {
    /***************************************/
    /*** Persistent state on all servers ***/
    /***************************************/
    // currentTerm is the lastest term the server has seen. Initialized to 0 on
    // boot, increases monotonically.
    currentTerm int
    // votedFor is the candidateID that received the vote in the current term,
    // or null if none.
    votedFor    int
    // log is the list of commands to be applied to the state machine, each including
    // the term when the entry was received by the leader. First index is 1
    log         []Log

    /*************************************/
    /*** Volatile state on all servers ***/
    /*************************************/
    // commitIndex is the index of the highest log entry known to be committed
    // Initialized at 0, increases monotonically
    commitIndex     int
    // lastApplied is the index of the highest log entry applied to state machine.
    // Initialized at 0, increases monotonically
    lastApplied     int

    /***************************************************************/
    /*** Volatile state on leaders. Reinitialized after election ***/
    /***************************************************************/
    // nextIndex, for each server, is the index of the next log entry to send
    // to that server. Initialized to leader's last log index + 1
    nextIndex   map[string]int
    // matchIndex, for each server, is the index of the highest log entry
    // known to be replicated on that server. Initialized to 0, increases monotonically
    matchIndex  map[string]int

    // Additions outside the spec:
    // 1. I'm using nextIndex and matchIndex as map[string]int to lookup servers
    //    by a string identifier
    // 2. I'm using role Role to idenify what role the server is actively playing
    role Role
}

type Log struct {
    Term int
    Data []byte
}

// AppendEntry is sent to followers by the leader
type AppendEntry struct {
    // Term is the leaders Term
    Term int
    // LeaderID is given to remind followers where clients can redirect their writes to
    LeaderID string
    // PrevLogIndex term of the previousLogIndex entry
    PrevLogIndex int
    // Entries is the log of entries to store. Empty for heartbeat
    Entries []Log
    // LeaderCommit is the leaders commitIndex
    LeaderCommit int
}

func (ae *AppendEntry) ToRequest(URL string) (*http.Request, error) {
    b, err := json.Marshal(ae)
    if err != nil {
        return nil, err
    }
    bReader := bytes.NewReader(b)
    return http.NewRequest("POST", URL, bReader)
}

func AppendEntryFromRequest(req *http.Request) (*AppendEntry, error) {
    ae := &AppendEntry{}
    b, err := ioutil.ReadAll(req.Body)
    if err != nil {
        return nil, err
    }
    if err := json.Unmarshal(b, ae); err != nil {
        return nil, err
    }
    return ae, nil
}

// AppendEntryResponse is the reply to the leader following an AppendEntry request
type AppendEntryResponse struct {
    // Success is true if the entry was appended on the follower
    Success bool
    // Term that the follower recognizes
    Term int
}

// Write returns the response from the AppendEntry RPC
func (ar *AppendEntryResponse) Write(w http.ResponseWriter) error {
    b, err := json.Marshal(ar)
    if err != nil {
        return err
    }
    w.Write(b)
    return nil
}

func AppendEntryResponseFromRequest(resp *http.Response) (*AppendEntryResponse, error) {
    ar := &AppendEntryResponse{}
    b, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if err := json.Unmarshal(b, ar); err != nil {
        return nil, err
    }
    return ar, nil
}






