package gorillimiter

import (
	"log"
	"net/http"
	"time"
)

// Limiter is the an LRU based middleware limiter for gorillmux
// It's hardcoded to only remember most recent 1000 IP addresses
// You choose how many requests are allowed per interval
func Limiter(next http.Handler, requestsPerInterval int, interval time.Duration) http.Handler {

	// This is only called once per limiter
	// We'll only cache upto 1000 IP addresses
	// And set the window to flush every 30 seconds (Interval)
	cache, err := NewLRU(1000, interval)
	if err != nil {
		log.Println("Couldn't create a cache - falling back on passthrough", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			return
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getRemoteIP(r)

		// Set a maximum of requestsPerInterval requests per interval
		cnt, underRateLimit := cache.Incr(ip, requestsPerInterval)
		if underRateLimit {
			// we good son
			next.ServeHTTP(w, r)
		} else {
			log.Printf("User [%s] is over rate limit, denying for now, current count [%d]\n", ip, cnt)
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
	})
}
