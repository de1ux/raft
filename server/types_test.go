package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppendEntriesToRequestAndBack(t *testing.T) {
	ae := &AppendEntries{
		Term:         1,
		LeaderID:     "2",
		PrevLogIndex: 3,
		PrevLogTerm:  1,
		Entries: []Entry{
			Entry{
				Term: 1,
				Data: []byte(`1`),
			},
			Entry{
				Term: 2,
				Data: []byte(`2`),
			},
		},
		LeaderCommit: 4,
	}
	original := fmt.Sprintf("%+v", ae)

	r, err := ae.ToRequest("")
	if err != nil {
		t.Fatal(err)
	}

	ae2, err := AppendEntriesFromRequest(r)
	if err != nil {
		t.Fatal(err)
	}

	converted := fmt.Sprintf("%+v", ae2)
	if original != converted {
		t.Fatalf("Expected:\n\t%sGot:\n\t%s", original, converted)
	}
}

func TestWriteAppendEntriesResponseAndReceive(t *testing.T) {
	var original string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ar := &AppendEntriesResponse{
			Success: true,
			Term:    1,
		}
		original = fmt.Sprintf("%+v", ar)
		ar.Write(w)
	}))
	defer ts.Close()

	ae := &AppendEntries{
		Term:         1,
		LeaderID:     "2",
		PrevLogIndex: 3,
		PrevLogTerm:  1,
		Entries: []Entry{
			Entry{
				Term: 1,
				Data: []byte(`1`),
			},
			Entry{
				Term: 2,
				Data: []byte(`2`),
			},
		},
		LeaderCommit: 4,
	}

	req, err := ae.ToRequest(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ar2, err := AppendEntriesResponseFromRequest(resp)
	if err != nil {
		t.Fatal(err)
	}

	converted := fmt.Sprintf("%+v", ar2)
	if original != converted {
		t.Fatalf("\nExpected:\n\t%s\nGot:\n\t%s", original, converted)
	}
}

func TestAppendHandlerSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	ae := &AppendEntries{
		Term:         1,
		LeaderID:     "2",
		PrevLogIndex: 0,
		PrevLogTerm:  1,
		Entries: []Entry{
			Entry{
				Term: 1,
				Data: []byte(`1`),
			},
			Entry{
				Term: 1,
				Data: []byte(`2`),
			},
		},
		LeaderCommit: 1,
	}

	req, err := ae.ToRequest("example.com")
	if err != nil {
		t.Fatal(err)
	}

	appendHandler(w, req)

	if w.Result() == nil {
		t.Fatal("Expected a result from appendHandler, got nil")
	}
	ar, err := AppendEntriesResponseFromRequest(w.Result())
	if !ar.Success {
		t.Fatal("Expected to pass")
	}
}

func TestAppendHandlerOlderTerm(t *testing.T) {
	w := httptest.NewRecorder()
	ae := &AppendEntries{
		Term:         -1,
		LeaderID:     "2",
		PrevLogIndex: 0,
		PrevLogTerm:  1,
		Entries: []Entry{
			Entry{
				Term: 1,
				Data: []byte(`1`),
			},
			Entry{
				Term: 1,
				Data: []byte(`2`),
			},
		},
		LeaderCommit: 1,
	}

	req, err := ae.ToRequest("example.com")
	if err != nil {
		t.Fatal(err)
	}

	appendHandler(w, req)

	if w.Result() == nil {
		t.Fatal("Expected a result from appendHandler, got nil")
	}
	ar, err := AppendEntriesResponseFromRequest(w.Result())
	if ar.Success {
		t.Fatal("Expected to fail with old term in AppendEntries RPC")
	}
}

func TestAppendHandlerNotFollower(t *testing.T) {
	w := httptest.NewRecorder()
	ae := &AppendEntries{}

	state.role = Leader

	req, err := ae.ToRequest("example.com")
	if err != nil {
		t.Fatal(err)
	}

	appendHandler(w, req)

	if w.Result() == nil {
		t.Fatal("Expected a result from appendHandler, got nil")
	}
	ar, err := AppendEntriesResponseFromRequest(w.Result())
	if ar.Success {
		t.Fatal("Expected to fail with old term in AppendEntries RPC")
	}
}

func TestCommitWithAdditionalFollowerEntriesFails(t *testing.T) {
	state := CreateNewState()
	state.log.entries = []Entry{
		Entry{
			Term: 1,
			Data: []byte(`111`),
		},
		Entry{
			Term: 1,
			Data: []byte(`222`),
		},
		Entry{
			Term: 9000, // term the leader doesnt recognize
			Data: []byte(`lolwut`),
		},
		Entry{
			Term: 9001, // term the leader doesnt recognize
			Data: []byte(`wutlol`),
		},
	}
	ae := &AppendEntries{
		Term:         1,
		LeaderID:     "2",
		PrevLogIndex: 2,
		PrevLogTerm:  1,
		Entries: []Entry{
			Entry{
				Term: 1,
				Data: []byte(`333`),
			},
			Entry{
				Term: 1,
				Data: []byte(`444`),
			},
			Entry{
				Term: 1,
				Data: []byte(`555`),
			},
		},
		LeaderCommit: 1,
	}

	commitAndVerifyLog(t, state, ae, []Entry{
		Entry{
			Term: 1,
			Data: []byte(`111`),
		},
		Entry{
			Term: 1,
			Data: []byte(`222`),
		},
		Entry{
			Term: 1,
			Data: []byte(`333`),
		},
		Entry{
			Term: 1,
			Data: []byte(`444`),
		},
		Entry{
			Term: 1,
			Data: []byte(`555`),
		},
	})
}

func TestCommitOnInitialization(t *testing.T) {
	state := CreateNewState()
	ae := &AppendEntries{
		Term:         1,
		LeaderID:     "2",
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries: []Entry{
			Entry{
				Term: 1,
				Data: []byte(`111`),
			},
			Entry{
				Term: 1,
				Data: []byte(`222`),
			},
			Entry{
				Term: 1,
				Data: []byte(`333`),
			},
		},
		LeaderCommit: 1,
	}

	commitAndVerifyLog(t, state, ae, []Entry{
		Entry{
			Term: 1,
			Data: []byte(`111`),
		},
		Entry{
			Term: 1,
			Data: []byte(`222`),
		},
		Entry{
			Term: 1,
			Data: []byte(`333`),
		},
	})
	if state.commitIndex != 1 {
		t.Fatal("Expected the leader commit to trump the last log index")
	}
}

func TestCommitIndexUpdatedWithLastEntryIndex(t *testing.T) {
	state := CreateNewState()
	ae := &AppendEntries{
		Term:         1,
		LeaderID:     "2",
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		Entries: []Entry{
			Entry{
				Term: 1,
				Data: []byte(`111`),
			},
			Entry{
				Term: 1,
				Data: []byte(`222`),
			},
			Entry{
				Term: 1,
				Data: []byte(`333`),
			},
		},
		LeaderCommit: 9,
	}

	commitAndVerifyLog(t, state, ae, []Entry{
		Entry{
			Term: 1,
			Data: []byte(`111`),
		},
		Entry{
			Term: 1,
			Data: []byte(`222`),
		},
		Entry{
			Term: 1,
			Data: []byte(`333`),
		},
	})
	if state.commitIndex != 3 {
		t.Fatal("Expected the last log index to trump the leader commit")
	}
}

func commitAndVerifyLog(t *testing.T, state *State, ae *AppendEntries, expected []Entry) {
	state.Commit(ae)
	actualStr := state.EntriesToString()
	expectedStr := fmt.Sprintf("%+v", expected)
	if expectedStr != actualStr {
		t.Fatalf("\nExpected:\n\t%s\nGot:\n\t%s", expectedStr, actualStr)
	}
}
