package hub

import (
    "testing"
)

// Benchmark broadcasting to N subscribers with small payloads.
func BenchmarkHub_Broadcast_Fanout(b *testing.B) {
    for _, subs := range []int{1, 2, 4, 8, 32, 128} {
        b.Run("subs="+itoa(subs), func(b *testing.B) {
            h := New(64)
            // join N subscribers for same user
            var leaves []func()
            for i := 0; i < subs; i++ {
                _, leave := h.Join("u", itoa(i))
                leaves = append(leaves, leave)
            }
            payload := make([]byte, 256)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                h.Broadcast("u", "from", payload)
            }
            b.StopTimer()
            for _, f := range leaves { f() }
        })
    }
}

// local itoa to avoid importing strconv in benchmarks file
func itoa(n int) string {
    if n == 0 { return "0" }
    var buf [32]byte
    i := len(buf)
    for n > 0 {
        i--
        buf[i] = byte('0' + (n % 10))
        n /= 10
    }
    return string(buf[i:])
}

