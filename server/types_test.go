package server

import (
    "fmt"
    "testing"
    "net/http"
    "net/http/httptest"
)

func TestAppendEntryToRequestAndBack(t *testing.T) {
    ae := &AppendEntry{
        Term: 1,
        LeaderID: "2",
        PrevLogIndex: 3,
        Entries: []Log{
            Log{
                Term: 1,
                Data: []byte(`1`),
            },
            Log{
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

    ae2, err := AppendEntryFromRequest(r)
    if err != nil {
        t.Fatal(err)
    }

    converted := fmt.Sprintf("%+v", ae2)
    if original != converted {
        t.Fatalf("Expected:\n\t%sGot:\n\t%s", original, converted)
    }
}


func TestWriteAppendEntryResponseAndReceive(t *testing.T) {
    var original string
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ar := &AppendEntryResponse{
            Success: true,
            Term: 1,
        }
        original = fmt.Sprintf("%+v", ar)
        ar.Write(w)
    }))
    defer ts.Close()

    ae := &AppendEntry{
        Term: 1,
        LeaderID: "2",
        PrevLogIndex: 3,
        Entries: []Log{
            Log{
                Term: 1,
                Data: []byte(`1`),
            },
            Log{
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

    ar2, err := AppendEntryResponseFromRequest(resp)
    if err != nil {
        t.Fatal(err)
    }

    converted := fmt.Sprintf("%+v", ar2)
    if original != converted {
        t.Fatalf("Expected:\n\t%sGot:\n\t%s", original, converted)
    }
}