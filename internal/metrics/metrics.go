package metrics

import (
    "fmt"
    "net/http"
    "sync/atomic"
    "time"
)

var (
    concurrent int64
    rejected   int64
    // naive histogram: total duration and count
    totalDurationNs int64
    totalRequests   int64
)

func IncConcurrent() {
    atomic.AddInt64(&concurrent, 1)
}

func DecConcurrent() {
    atomic.AddInt64(&concurrent, -1)
}

func IncRejected() {
    atomic.AddInt64(&rejected, 1)
}

func ObserveRequest(d time.Duration) {
    atomic.AddInt64(&totalDurationNs, d.Nanoseconds())
    atomic.AddInt64(&totalRequests, 1)
}

func Handler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain; version=0.0.4")
        c := atomic.LoadInt64(&concurrent)
        rej := atomic.LoadInt64(&rejected)
        dur := atomic.LoadInt64(&totalDurationNs)
        cnt := atomic.LoadInt64(&totalRequests)
        avg := float64(0)
        if cnt > 0 {
            avg = float64(dur) / float64(cnt) / 1e6 // ms
        }

        fmt.Fprintf(w, "# HELP lottosmash_current_concurrent_requests Current concurrent HTTP requests\n")
        fmt.Fprintf(w, "# TYPE lottosmash_current_concurrent_requests gauge\n")
        fmt.Fprintf(w, "lottosmash_current_concurrent_requests %d\n", c)

        fmt.Fprintf(w, "# HELP lottosmash_rejected_requests Rejected HTTP requests due to concurrency limit\n")
        fmt.Fprintf(w, "# TYPE lottosmash_rejected_requests counter\n")
        fmt.Fprintf(w, "lottosmash_rejected_requests %d\n", rej)

        fmt.Fprintf(w, "# HELP lottosmash_request_avg_duration_millis Average HTTP request duration in ms\n")
        fmt.Fprintf(w, "# TYPE lottosmash_request_avg_duration_millis gauge\n")
        fmt.Fprintf(w, "lottosmash_request_avg_duration_millis %.3f\n", avg)

        fmt.Fprintf(w, "# HELP lottosmash_total_requests Total HTTP requests observed\n")
        fmt.Fprintf(w, "# TYPE lottosmash_total_requests counter\n")
        fmt.Fprintf(w, "lottosmash_total_requests %d\n", cnt)
    })
}
