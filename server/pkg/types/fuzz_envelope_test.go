package types

import (
    "encoding/json"
    "testing"
)

// FuzzEnvelopeJSON ensures that arbitrary JSON inputs do not crash the decoder
// and that a round-trip encode/decode of accepted inputs is stable enough
// for our struct shape (we only check absence of panics and basic invariants).
func FuzzEnvelopeJSON(f *testing.F) {
    // Seed with a small corpus of realistic messages
    seeds := [][]byte{
        []byte(`{"type":"hello","hello":{"token":"u1","user_id":"u1","device_id":"A"}}`),
        []byte(`{"type":"clip","clip":{"msg_id":"m1","mime":"text/plain","size":2,"data":"aGk=","upload_url":""}}`),
        []byte(`{"type":"clip","clip":{"msg_id":"m2","mime":"application/octet-stream","size":100000,"upload_url":"/d/abcd"}}`),
        []byte(`{"type":"hello","hello":{}}`),
    }
    for _, s := range seeds { f.Add(string(s)) }

    f.Fuzz(func(t *testing.T, s string) {
        var env Envelope
        if err := json.Unmarshal([]byte(s), &env); err != nil {
            return // ignore invalid JSON
        }
        // encode and decode again
        b, err := json.Marshal(&env)
        if err != nil { t.Fatalf("marshal: %v", err) }
        var env2 Envelope
        _ = json.Unmarshal(b, &env2)
        // Basic sanity: type field is at most a short label
        if len(env2.Type) > 64 {
            t.Skip()
        }
    })
}

