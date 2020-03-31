package gorillimiter

import (
	"net"
	"net/http"
	"strings"
)

func getRemoteIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")

		// If we got an array... grab the first IP
		ips := strings.Split(IPAddress, ", ")
		if len(ips) > 1 {
			IPAddress = ips[0]
		}
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}

	possibleAddress, _, err := net.SplitHostPort(IPAddress)
	// Can't use errors.Is() as the net pkg is hard work
	if err == nil {
		IPAddress = possibleAddress
	}
	return IPAddress
}
