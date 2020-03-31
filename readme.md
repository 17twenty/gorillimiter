# Gorillimiter

## What is this

[Gorilla Mux](https://www.gorillatoolkit.org/pkg/mux) compatible rate limiter middleware with some opinions

Thanks to https://github.com/hashicorp/golang-lru and
https://www.alexedwards.net/blog/how-to-rate-limit-http-requests for
inspiration.

## Usage

```go
mux := http.NewServeMux()
mux.HandleFunc("/", yourHandler)
...
log.Println("Listening on :5000...")
// Attach your listener here
http.ListenAndServe(":5000", gorillimiter.Limiter(mux, 10, time.Second))
```

## Demo

```bash
$ cd src/demo
$ go build && ./demo 
2020/03/31 16:57:53 Listening on :5000...
... # At this point, smash it with curl -i localhost:5000
2020/03/31 16:54:01 User [::1] is over rate limit, denying for now, current count [13]
2020/03/31 16:54:01 User [::1] is over rate limit, denying for now, current count [14]
2020/03/31 16:54:01 User [::1] is over rate limit, denying for now, current count [15]
2020/03/31 16:54:01 User [::1] is over rate limit, denying for now, current count [16]
2020/03/31 16:54:01 User [::1] is over rate limit, denying for now, current count [17]
2020/03/31 16:54:01 User [::1] is over rate limit, denying for now, current count [18]
```