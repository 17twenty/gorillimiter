package gorillimiter

import (
	"log"
	"net/http"
	"time"
)

// MiddlewareWrapper can be used for chained middleware
func MiddlewareWrapper(requestsPerInterval int, interval time.Duration) func(http.Handler) http.Handler {

	cache, err := NewLRU(1000, interval)
	if err != nil {
		log.Println("Couldn't create a cache - falling back to passthrough", err)
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {

				next.ServeHTTP(wr, req)
			})
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			ip := getRemoteIP(req)

			// Set a maximum of requestsPerInterval requests per interval
			cnt, underRateLimit := cache.Inc(ip, requestsPerInterval)
			if underRateLimit {
				// we good son
				next.ServeHTTP(wr, req)
				return
			}

			log.Printf("Address [%s] is over ratelimit, denying for now, current hits [%d]\n", ip, cnt)
			http.Error(wr, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		})
	}
}

// Limiter is the an LRU based limiter for a gorilla mux
// It's hardcoded to only remember most recent 1000 IP addresses
// You choose how many requests are allowed per interval
func Limiter(next http.Handler, requestsPerInterval int, interval time.Duration) http.Handler {

	// This is only called once per limiter
	// We'll only cache upto 1000 IP addresses
	// And set the window to flush every $interval seconds
	// with a max of $requestsPerInterval per $interval
	cache, err := NewLRU(1000, interval)
	if err != nil {
		log.Println("Couldn't create a cache - falling back to passthrough", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			return
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getRemoteIP(r)

		// Set a maximum of requestsPerInterval requests per interval
		cnt, underRateLimit := cache.Inc(ip, requestsPerInterval)
		if underRateLimit {
			// we good son
			next.ServeHTTP(w, r)
			return
		}

		log.Printf("Address [%s] is over ratelimit, denying for now, current hits [%d]\n", ip, cnt)
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		return

	})
}
